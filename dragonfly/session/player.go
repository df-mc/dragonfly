package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/internal/nbtconv"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/inventory"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/player/form"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/player/skin"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/gamemode"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"math"
	"net"
	"strings"
	"sync/atomic"
	_ "unsafe" // Imported for compiler directives.
)

// handleMovePlayer ...
func (s *Session) handleMovePlayer(pk *packet.MovePlayer) error {
	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return fmt.Errorf("incorrect entity runtime ID %v: runtime ID must be 1", pk.EntityRuntimeID)
	}
	pk.Position = pk.Position.Sub(mgl32.Vec3{0, 1.62}) // Subtract the base offset of players from the pos.

	s.c.Move(pk.Position.Sub(s.c.Position()))
	s.c.Rotate(pk.Yaw-s.c.Yaw(), pk.Pitch-s.c.Pitch())

	s.chunkLoader.Load().(*world.Loader).Move(pk.Position)
	s.writePacket(&packet.NetworkChunkPublisherUpdate{
		Position: protocol.BlockPos{int32(pk.Position[0]), int32(pk.Position[1]), int32(pk.Position[2])},
		Radius:   uint32(s.chunkRadius * 16),
	})
	return nil
}

// handleMobEquipment ...
func (s *Session) handleMobEquipment(pk *packet.MobEquipment) error {
	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return fmt.Errorf("incorrect entity runtime ID %v: runtime ID must be 1", pk.EntityRuntimeID)
	}
	// The slot that the player might have selected must be within the hotbar: The held item cannot be in a
	// different place in the inventory.
	if pk.InventorySlot > 8 {
		return fmt.Errorf("slot exceeds hotbar range 0-8: slot is %v", pk.InventorySlot)
	}
	if pk.WindowID != protocol.WindowIDInventory {
		return fmt.Errorf("MobEquipmentPacket should only involve main inventory, got window ID %v", pk.WindowID)
	}

	// We first change the held slot.
	atomic.StoreUint32(s.heldSlot, uint32(pk.InventorySlot))

	for _, viewer := range s.c.World().Viewers(s.c.Position()) {
		viewer.ViewEntityItems(s.c)
	}
	return nil
}

// handlePlayerAction ...
func (s *Session) handlePlayerAction(pk *packet.PlayerAction) error {
	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return fmt.Errorf("PlayerAction packet must only have runtime ID of the own entity")
	}
	blockPos, face := world.BlockPos{int(pk.BlockPosition[0]), int(pk.BlockPosition[1]), int(pk.BlockPosition[2])}, world.Face(pk.BlockFace)

	switch pk.ActionType {
	case packet.PlayerActionStartSprint:
		s.c.StartSprinting()
	case packet.PlayerActionStopSprint:
		s.c.StopSprinting()
	case packet.PlayerActionStartSneak:
		s.c.StartSneaking()
	case packet.PlayerActionStopSneak:
		s.c.StopSneaking()
	case packet.PlayerActionStartBreak:
		s.c.StartBreaking(blockPos)
	case packet.PlayerActionAbortBreak:
		s.c.AbortBreaking()
	case packet.PlayerActionStopBreak:
		s.c.FinishBreaking()
	case packet.PlayerActionContinueBreak:
		s.c.ContinueBreaking(face)
	}
	return nil
}

// handleModalFormResponse ...
func (s *Session) handleModalFormResponse(pk *packet.ModalFormResponse) error {
	s.formMu.Lock()
	f, ok := s.forms[pk.FormID]
	delete(s.forms, pk.FormID)
	s.formMu.Unlock()

	if !ok {
		return fmt.Errorf("form with ID %v not currently open", pk.FormID)
	}
	if bytes.Equal(pk.ResponseData, []byte("null")) {
		// The form was cancelled: The cross in the top right corner was clicked.
		return nil
	}
	if err := f.SubmitJSON(pk.ResponseData, s.c); err != nil {
		return fmt.Errorf("error parsing form data: %w", err)
	}
	return nil
}

// handleContainerClose ...
func (s *Session) handleContainerClose(pk *packet.ContainerClose) error {
	switch pk.WindowID {
	case byte(atomic.LoadUint32(&s.openedWindowID)):
		s.closeCurrentContainer()
	}
	return nil
}

// closeCurrentContainer closes the container the player might currently have open.
func (s *Session) closeCurrentContainer() {
	if atomic.LoadUint32(&s.containerOpened) == 0 {
		return
	}
	s.closeWindow()
	pos := s.openedPos.Load().(world.BlockPos)
	if container, ok := s.c.World().Block(pos).(block.Container); ok {
		container.RemoveViewer(s, s.c.World(), pos)
	}
}

// handleRespawn ...
func (s *Session) handleRespawn(pk *packet.Respawn) error {
	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return fmt.Errorf("entity runtime ID in Respawn packet must always be the player's (%v), but got %v", selfEntityRuntimeID, pk.EntityRuntimeID)
	}
	if pk.State != packet.RespawnStateClientReadyToSpawn {
		return fmt.Errorf("respawn state in Respawn packet must always be %v, but got %v", packet.RespawnStateClientReadyToSpawn, pk.State)
	}
	s.c.Respawn()
	s.SendRespawn()
	return nil
}

// SendRespawn spawns the controllable of the session client-side in the world, provided it is has died.
func (s *Session) SendRespawn() {
	s.writePacket(&packet.Respawn{
		Position:        s.c.Position().Add(mgl32.Vec3{0, 1.62}),
		State:           packet.RespawnStateReadyToSpawn,
		EntityRuntimeID: selfEntityRuntimeID,
	})
	s.writePacket(&packet.InventoryContent{
		WindowID: protocol.WindowIDCreative,
		Content:  creativeItems(),
	})
}

// handleInventoryTransaction ...
func (s *Session) handleInventoryTransaction(pk *packet.InventoryTransaction) error {
	switch data := pk.TransactionData.(type) {
	case *protocol.NormalTransactionData:
		if len(pk.Actions) == 0 {
			return nil
		}
		if err := s.verifyTransaction(pk.Actions); err != nil {
			s.sendInv(s.inv, protocol.WindowIDInventory)
			s.sendInv(s.ui, protocol.WindowIDUI)
			s.sendInv(s.offHand, protocol.WindowIDOffHand)
			s.log.Debugf("%v: %v", s.c.Name(), err)
			return nil
		}
		atomic.StoreUint32(&s.inTransaction, 1)
		s.executeTransaction(pk.Actions)
		atomic.StoreUint32(&s.inTransaction, 0)
	case *protocol.UseItemOnEntityTransactionData:
		e, ok := s.entityFromRuntimeID(data.TargetEntityRuntimeID)
		if !ok {
			return fmt.Errorf("invalid entity interaction: no entity found with runtime ID %v", data.TargetEntityRuntimeID)
		}
		switch data.ActionType {
		case protocol.UseItemOnEntityActionInteract:
			s.c.UseItemOnEntity(e)
		case protocol.UseItemOnEntityActionAttack:
			s.c.AttackEntity(e)
		}
	case *protocol.UseItemTransactionData:
		pos := world.BlockPos{int(data.BlockPosition[0]), int(data.BlockPosition[1]), int(data.BlockPosition[2])}
		switch data.ActionType {
		case protocol.UseItemActionBreakBlock:
			s.c.BreakBlock(pos)
		case protocol.UseItemActionClickBlock:
			// We reset the inventory so that we can send the held item update without the client already
			// having done that client-side.
			s.sendInv(s.inv, protocol.WindowIDInventory)
			s.c.UseItemOnBlock(pos, world.Face(data.BlockFace), data.ClickedPosition)
		case protocol.UseItemActionClickAir:
			s.c.UseItem()
		}
	}
	return nil
}

// sendInv sends the inventory passed to the client with the window ID.
func (s *Session) sendInv(inv *inventory.Inventory, windowID uint32) {
	pk := &packet.InventoryContent{
		WindowID: windowID,
		Content:  make([]protocol.ItemStack, 0, s.inv.Size()),
	}
	for _, i := range inv.All() {
		pk.Content = append(pk.Content, stackFromItem(i))
	}
	s.writePacket(pk)
}

// executeTransaction executes all actions of a transaction passed. It assumes the actions are all valid,
// which would otherwise be checked by calling verifyTransaction.
func (s *Session) executeTransaction(actions []protocol.InventoryAction) {
	for _, action := range actions {
		if action.SourceType == protocol.InventoryActionSourceCreative {
			continue
		}
		// The window IDs are already checked when using verifyTransaction, so we don't need to check again.
		inv, _ := s.invByID(action.WindowID)
		_ = inv.SetItem(int(action.InventorySlot), stackToItem(action.NewItem))
	}
}

// verifyTransaction verifies a transaction composed of the actions passed. The method makes sure the old
// items are precisely equal to the new items: No new items must be added or removed.
// verifyTransaction also checks if all window IDs sent match some inventory, and if the old items match the
// items found in that inventory.
func (s *Session) verifyTransaction(a []protocol.InventoryAction) error {
	// Allocate a big inventory and add all new items to it.
	temp := inventory.New(128, nil)
	actions := make([]protocol.InventoryAction, 0, len(a))
	for _, action := range a {
		if action.OldItem.Count == 0 && action.NewItem.Count == 0 {
			continue
		}
		actions = append(actions, action)
	}
	for _, action := range actions {
		inv, ok := s.invByID(action.WindowID)
		if !ok {
			// The inventory with that window ID did not exist.
			return fmt.Errorf("unknown inventory ID %v", action.WindowID)
		}
		actualOld, err := inv.Item(int(action.InventorySlot))
		if err != nil {
			// The slot passed actually exceeds the inventory size, meaning we can't actually get an item at
			// that slot.
			return fmt.Errorf("slot %v out of range for inventory %v", action.InventorySlot, action.WindowID)
		}
		old := stackToItem(action.OldItem)
		if !actualOld.Comparable(old) || actualOld.Count() != old.Count() {
			if _, creative := s.c.GameMode().(gamemode.Creative); !creative || action.SourceType != protocol.InventoryActionSourceCreative {
				// Either the type or the count of the old item as registered by the server and the client are
				// mismatched.
				return fmt.Errorf("slot %v holds a different item than the client expects: %v is actually %v", action.InventorySlot, old, actualOld)
			}
		}
		if _, err := temp.AddItem(old); err != nil {
			return fmt.Errorf("inventory was full: %w", err)
		}
	}
	for _, action := range actions {
		newItem := stackToItem(action.NewItem)
		if err := temp.RemoveItem(newItem); err != nil {
			return fmt.Errorf("item %v removed was not present in old items: %w", newItem, err)
		}
	}
	// Now that we made sure every new item was also present in the old items, we must also check if every old
	// item is present as a new item. We can do that by simply checking if the inventory has any items left in
	// it.
	if !temp.Empty() {
		return fmt.Errorf("new items and old items must be balanced")
	}
	return nil
}

// invByID attempts to return an inventory by the ID passed. If found, the inventory is returned and the bool
// returned is true.
func (s *Session) invByID(id int32) (*inventory.Inventory, bool) {
	switch id {
	case protocol.WindowIDInventory:
		return s.inv, true
	case protocol.WindowIDOffHand:
		return s.offHand, true
	case protocol.WindowIDUI:
		return s.ui, true
	case protocol.WindowIDArmour:
		return s.armour.Inv(), true
	case int32(atomic.LoadUint32(&s.openedWindowID)):
		if atomic.LoadUint32(&s.containerOpened) == 1 {
			return s.openedWindow.Load().(*inventory.Inventory), true
		}
	}
	return nil, false
}

// Disconnect disconnects the client and ultimately closes the session. If the message passed is non-empty,
// it will be shown to the client.
func (s *Session) Disconnect(message string) {
	s.writePacket(&packet.Disconnect{
		HideDisconnectionScreen: message == "",
		Message:                 message,
	})
	if s != Nop {
		_ = s.conn.Flush()
		_ = s.conn.Close()
	}
}

// SendSpeed sends the speed of the player in an UpdateAttributes packet, so that it is updated client-side.
func (s *Session) SendSpeed(speed float32) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{{
			Name:    "minecraft:movement",
			Value:   speed,
			Max:     math.MaxFloat32,
			Min:     0,
			Default: 0.1,
		}},
	})
}

// SendVelocity sends the velocity of the player to the client.
func (s *Session) SendVelocity(velocity mgl32.Vec3) {
	s.writePacket(&packet.SetActorMotion{
		EntityRuntimeID: selfEntityRuntimeID,
		Velocity:        velocity,
	})
}

// SendForm sends a form to the client of the connection. The Submit method of the form is called when the
// client submits the form.
func (s *Session) SendForm(f form.Form) {
	var n []map[string]interface{}
	m := map[string]interface{}{}

	switch frm := f.(type) {
	case form.Custom:
		m["type"], m["title"] = "custom_form", frm.Title()
		for _, e := range frm.Elements() {
			n = append(n, elemToMap(e))
		}
		m["content"] = n
	case form.Menu:
		m["type"], m["title"], m["content"] = "form", frm.Title(), frm.Body()
		for _, button := range frm.Buttons() {
			v := map[string]interface{}{"text": button.Text}
			if button.Image != "" {
				buttonType := "path"
				if strings.HasPrefix(button.Image, "http:") || strings.HasPrefix(button.Image, "https:") {
					buttonType = "url"
				}
				v["image"] = map[string]interface{}{"type": buttonType, "data": button.Image}
			}
			n = append(n, v)
		}
		m["buttons"] = n
	case form.Modal:
		m["type"], m["title"], m["content"] = "modal", frm.Title(), frm.Body()
		buttons := frm.Buttons()
		m["button1"], m["button2"] = buttons[0].Text, buttons[1].Text
	}
	b, _ := json.Marshal(m)

	s.formMu.Lock()
	if len(s.forms) > 10 {
		s.log.Debug("more than 10 active forms: dropping an existing one.")
		for k := range s.forms {
			delete(s.forms, k)
			break
		}
	}
	id := s.formID
	s.forms[id] = f
	s.formID++
	s.formMu.Unlock()

	s.writePacket(&packet.ModalFormRequest{
		FormID:   id,
		FormData: b,
	})
}

// elemToMap encodes a form element to its representation as a map to be encoded to JSON for the client.
func elemToMap(e form.Element) map[string]interface{} {
	switch element := e.(type) {
	case form.Toggle:
		return map[string]interface{}{
			"type":    "toggle",
			"text":    element.Text,
			"default": element.Default,
		}
	case form.Input:
		return map[string]interface{}{
			"type":        "input",
			"text":        element.Text,
			"default":     element.Default,
			"placeholder": element.Placeholder,
		}
	case form.Label:
		return map[string]interface{}{
			"type": "label",
			"text": element.Text,
		}
	case form.Slider:
		return map[string]interface{}{
			"type":    "slider",
			"text":    element.Text,
			"min":     element.Min,
			"max":     element.Max,
			"step":    element.StepSize,
			"default": element.Default,
		}
	case form.Dropdown:
		return map[string]interface{}{
			"type":    "dropdown",
			"text":    element.Text,
			"default": element.DefaultIndex,
			"options": element.Options,
		}
	case form.StepSlider:
		return map[string]interface{}{
			"type":    "step_slider",
			"text":    element.Text,
			"default": element.DefaultIndex,
			"steps":   element.Options,
		}
	}
	panic("should never happen")
}

// Transfer transfers the player to a server with the IP and port passed.
func (s *Session) Transfer(ip net.IP, port int) {
	s.writePacket(&packet.Transfer{
		Address: ip.String(),
		Port:    uint16(port),
	})
}

// SendGameMode sends the game mode of the Controllable of the session to the client. It makes sure the right
// flags are set to create the full game mode.
func (s *Session) SendGameMode(mode gamemode.GameMode) {
	flags, id := uint32(0), int32(packet.GameTypeSurvival)
	switch mode.(type) {
	case gamemode.Creative:
		flags = packet.AdventureFlagAllowFlight
		id = packet.GameTypeCreative
	case gamemode.Adventure:
		flags |= packet.AdventureFlagWorldImmutable
		id = packet.GameTypeAdventure
	case gamemode.Spectator:
		flags, id = packet.AdventureFlagWorldImmutable|packet.AdventureFlagAllowFlight|packet.AdventureFlagMuted|packet.AdventureFlagNoClip|packet.AdventureFlagNoPVP, packet.GameTypeCreativeSpectator
	}
	s.writePacket(&packet.AdventureSettings{
		Flags:             flags,
		PermissionLevel:   packet.PermissionLevelMember,
		PlayerUniqueID:    1,
		ActionPermissions: uint32(packet.ActionPermissionBuildAndMine | packet.ActionPermissionDoorsAndSwitched | packet.ActionPermissionOpenContainers | packet.ActionPermissionAttackPlayers | packet.ActionPermissionAttackMobs),
	})
	s.writePacket(&packet.SetPlayerGameType{GameType: id})
}

// SendHealth sends the health and max health to the player.
func (s *Session) SendHealth(health, max float32) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{{
			Name:    "minecraft:health",
			Value:   float32(math.Ceil(float64(health))),
			Max:     float32(math.Ceil(float64(max))),
			Default: 20,
		}},
	})
}

// SendGameRules sends all the provided game rules to the player. Once sent, they will be immediately updated
// on the client if they are valid.
func (s *Session) sendGameRules(gameRules map[string]interface{}) {
	s.writePacket(&packet.GameRulesChanged{GameRules: gameRules})
}

// EnableCoordinates will either enable or disable coordinates for the player depending on the value given.
func (s *Session) EnableCoordinates(enable bool) {
	//noinspection SpellCheckingInspection
	s.sendGameRules(map[string]interface{}{"showcoordinates": enable})
}

// addToPlayerList adds the player of a session to the player list of this session. It will be shown in the
// in-game pause menu screen.
func (s *Session) addToPlayerList(session *Session) {
	c := session.c

	s.entityMutex.Lock()
	runtimeID := uint64(1)
	if session != s {
		runtimeID = atomic.AddUint64(&s.currentEntityRuntimeID, 1)
	}
	s.entityRuntimeIDs[c] = runtimeID
	s.entities[runtimeID] = c
	s.entityMutex.Unlock()

	s.writePacket(&packet.PlayerList{
		ActionType: packet.PlayerListActionAdd,
		Entries: []protocol.PlayerListEntry{{
			UUID:           c.UUID(),
			EntityUniqueID: int64(runtimeID),
			Username:       c.Name(),
			XUID:           c.XUID(),
			Skin:           skinToProtocol(c.Skin()),
		}},
	})
}

// skinToProtocol converts a skin to its protocol representation.
func skinToProtocol(s skin.Skin) protocol.Skin {
	var animations []protocol.SkinAnimation
	for _, animation := range s.Animations {
		protocolAnim := protocol.SkinAnimation{
			ImageWidth:    uint32(animation.Bounds().Max.X),
			ImageHeight:   uint32(animation.Bounds().Max.Y),
			ImageData:     animation.Pix,
			AnimationType: 0,
			FrameCount:    float32(animation.FrameCount),
		}
		switch animation.Type() {
		case skin.AnimationHead:
			protocolAnim.AnimationType = protocol.SkinAnimationHead
		case skin.AnimationBody32x32:
			protocolAnim.AnimationType = protocol.SkinAnimationBody32x32
		case skin.AnimationBody128x128:
			protocolAnim.AnimationType = protocol.SkinAnimationBody128x128
		}
		animations = append(animations, protocolAnim)
	}

	return protocol.Skin{
		SkinID:            uuid.New().String(),
		SkinResourcePatch: s.ModelConfig.Encode(),
		SkinImageWidth:    uint32(s.Bounds().Max.X),
		SkinImageHeight:   uint32(s.Bounds().Max.Y),
		SkinData:          s.Pix,
		CapeImageWidth:    uint32(s.Cape.Bounds().Max.X),
		CapeImageHeight:   uint32(s.Cape.Bounds().Max.Y),
		CapeData:          s.Cape.Pix,
		SkinGeometry:      s.Model,
		PersonaSkin:       s.Persona,
		CapeID:            uuid.New().String(),
		FullSkinID:        uuid.New().String(),
		Animations:        animations,
	}
}

// removeFromPlayerList removes the player of a session from the player list of this session. It will no
// longer be shown in the in-game pause menu screen.
func (s *Session) removeFromPlayerList(session *Session) {
	c := session.c

	s.entityMutex.Lock()
	delete(s.entityRuntimeIDs, c)
	delete(s.entities, s.entityRuntimeIDs[c])
	s.entityMutex.Unlock()

	s.writePacket(&packet.PlayerList{
		ActionType: packet.PlayerListActionRemove,
		Entries: []protocol.PlayerListEntry{{
			UUID: c.UUID(),
		}},
	})
}

// HandleInventories starts handling the inventories of the Controllable of the session. It sends packets when
// slots in the inventory are changed.
func (s *Session) HandleInventories() (inv, offHand *inventory.Inventory, armour *inventory.Armour, heldSlot *uint32) {
	s.inv = inventory.New(36, func(slot int, item item.Stack) {
		if atomic.LoadUint32(&s.inTransaction) == 1 {
			return
		}
		s.writePacket(&packet.InventorySlot{
			WindowID: protocol.WindowIDInventory,
			Slot:     uint32(slot),
			NewItem:  stackFromItem(item),
		})
		if slot == int(atomic.LoadUint32(s.heldSlot)) {
			for _, viewer := range s.c.World().Viewers(s.c.Position()) {
				viewer.ViewEntityItems(s.c)
			}
		}
	})
	s.offHand = inventory.New(1, func(slot int, item item.Stack) {
		if atomic.LoadUint32(&s.inTransaction) == 1 {
			return
		}
		s.writePacket(&packet.InventorySlot{
			WindowID: protocol.WindowIDOffHand,
			Slot:     uint32(slot),
			NewItem:  stackFromItem(item),
		})
		for _, viewer := range s.c.World().Viewers(s.c.Position()) {
			viewer.ViewEntityItems(s.c)
		}
	})
	s.armour = inventory.NewArmour(func(slot int, item item.Stack) {
		if atomic.LoadUint32(&s.inTransaction) == 1 {
			return
		}
		s.writePacket(&packet.InventorySlot{
			WindowID: protocol.WindowIDArmour,
			Slot:     uint32(slot),
			NewItem:  stackFromItem(item),
		})
		for _, viewer := range s.c.World().Viewers(s.c.Position()) {
			viewer.ViewEntityArmour(s.c)
		}
	})
	return s.inv, s.offHand, s.armour, s.heldSlot
}

// stackFromItem converts an item.Stack to its network ItemStack representation.
func stackFromItem(it item.Stack) protocol.ItemStack {
	if it.Empty() {
		return protocol.ItemStack{}
	}
	id, meta := it.Item().EncodeItem()
	return protocol.ItemStack{
		ItemType: protocol.ItemType{
			NetworkID:     id,
			MetadataValue: meta,
		},
		Count:   int16(it.Count()),
		NBTData: nbtconv.ItemToNBT(it, true),
	}
}

// stackToItem converts a network ItemStack representation back to an item.Stack.
func stackToItem(it protocol.ItemStack) item.Stack {
	t, ok := world_itemByID(it.NetworkID, it.MetadataValue)
	if !ok {
		t = block.Air{}
	}
	//noinspection SpellCheckingInspection
	if nbter, ok := t.(world.NBTer); ok && len(it.NBTData) != 0 {
		t = nbter.DecodeNBT(it.NBTData).(world.Item)
	}
	s := item.NewStack(t, int(it.Count))
	return nbtconv.ItemFromNBT(it.NBTData, &s)
}

// creativeItems returns all creative inventory items as protocol item stacks.
func creativeItems() []protocol.ItemStack {
	it := make([]protocol.ItemStack, 0, len(item.CreativeItems()))
	for _, i := range item.CreativeItems() {
		it = append(it, stackFromItem(i))
	}
	return it
}

// The following functions use the go:linkname directive in order to make sure the item.byID and item.toID
// functions do not need to be exported.

//go:linkname world_itemByID git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world.itemByID
//noinspection ALL
func world_itemByID(id int32, meta int16) (world.Item, bool)
