package session

import (
	"encoding/json"
	"github.com/df-mc/dragonfly/dragonfly/block"
	"github.com/df-mc/dragonfly/dragonfly/internal/entity_internal"
	"github.com/df-mc/dragonfly/dragonfly/internal/nbtconv"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/inventory"
	"github.com/df-mc/dragonfly/dragonfly/player/form"
	"github.com/df-mc/dragonfly/dragonfly/player/skin"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/gamemode"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"math"
	"net"
	"strings"
	"sync/atomic"
	_ "unsafe" // Imported for compiler directives.
)

// closeCurrentContainer closes the container the player might currently have open.
func (s *Session) closeCurrentContainer() {
	if atomic.LoadUint32(s.containerOpened) == 0 {
		return
	}
	s.closeWindow()
	pos := s.openedPos.Load().(world.BlockPos)
	if container, ok := s.c.World().Block(pos).(block.Container); ok {
		container.RemoveViewer(s, s.c.World(), pos)
	}
}

// SendRespawn spawns the controllable of the session client-side in the world, provided it is has died.
func (s *Session) SendRespawn() {
	s.writePacket(&packet.Respawn{
		Position:        vec64To32(s.c.Position().Add(mgl64.Vec3{0, entityOffset(s.c)})),
		State:           packet.RespawnStateReadyToSpawn,
		EntityRuntimeID: selfEntityRuntimeID,
	})
}

// sendInv sends the inventory passed to the client with the window ID.
func (s *Session) sendInv(inv *inventory.Inventory, windowID uint32) {
	pk := &packet.InventoryContent{
		WindowID: windowID,
		Content:  make([]protocol.ItemInstance, 0, s.inv.Size()),
	}
	for _, i := range inv.All() {
		pk.Content = append(pk.Content, instanceFromItem(i))
	}
	s.writePacket(pk)
}

const (
	containerArmour               = 6
	containerChest                = 7
	containerInventoryChestOpened = 12
	containerCraftingGrid         = 13
	containerHotbar               = 27
	containerInventory            = 28
	containerOffHand              = 33
	containerCursor               = 58
	containerCreativeOutput       = 59
)

// invByID attempts to return an inventory by the ID passed. If found, the inventory is returned and the bool
// returned is true.
func (s *Session) invByID(id int32) (*inventory.Inventory, bool) {
	switch id {
	case containerCraftingGrid, containerCreativeOutput, containerCursor:
		// UI inventory.
		return s.ui, true
	case containerHotbar, containerInventory, containerInventoryChestOpened:
		// Hotbar 'inventory', rest of inventory, inventory when container is opened.
		return s.inv, true
	case containerOffHand:
		return s.offHand, true
	case containerArmour:
		// Armour inventory.
		return s.armour.Inv(), true
	case containerChest:
		// Chests, potentially other containers too.
		if atomic.LoadUint32(s.containerOpened) == 1 {
			return s.openedWindow.Load().(*inventory.Inventory), true
		}
	}
	return nil, false
}

// Disconnect disconnects the client and ultimately closes the session. If the message passed is non-empty,
// it will be shown to the client.
func (s *Session) Disconnect(message string) {
	if s != Nop {
		s.writePacket(&packet.Disconnect{
			HideDisconnectionScreen: message == "",
			Message:                 message,
		})
		_ = s.conn.Flush()
		_ = s.conn.Close()
	}
}

// SendSpeed sends the speed of the player in an UpdateAttributes packet, so that it is updated client-side.
func (s *Session) SendSpeed(speed float64) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{{
			Name:    "minecraft:movement",
			Value:   float32(speed),
			Max:     math.MaxFloat32,
			Min:     0,
			Default: 0.1,
		}},
	})
}

// SendFood ...
func (s *Session) SendFood(food int, saturation, exhaustion float64) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{
			{
				Name:  "minecraft:player.hunger",
				Value: float32(food),
				Max:   20, Min: 0, Default: 20,
			},
			{
				Name:  "minecraft:player.saturation",
				Value: float32(saturation),
				Max:   20, Min: 0, Default: 20,
			},
			{
				Name:  "minecraft:player.exhaustion",
				Value: float32(exhaustion),
				Max:   5, Min: 0, Default: 0,
			},
		},
	})
}

// SendVelocity sends the velocity of the player to the client.
func (s *Session) SendVelocity(velocity mgl64.Vec3) {
	s.writePacket(&packet.SetActorMotion{
		EntityRuntimeID: selfEntityRuntimeID,
		Velocity:        vec64To32(velocity),
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

	h := s.handlers[packet.IDModalFormResponse].(*ModalFormResponseHandler)
	id := atomic.AddUint32(h.currentID, 1)

	h.mu.Lock()
	if len(h.forms) > 10 {
		s.log.Debugf("SendForm %v: more than 10 active forms: dropping an existing one.", s.c.Name())
		for k := range h.forms {
			delete(h.forms, k)
			break
		}
	}
	h.forms[id] = f
	h.mu.Unlock()

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
func (s *Session) SendHealth(health *entity_internal.HealthManager) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{{
			Name:    "minecraft:health",
			Value:   float32(math.Ceil(health.Health())),
			Max:     float32(math.Ceil(health.MaxHealth())),
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
		if slot == int(atomic.LoadUint32(s.heldSlot)) {
			for _, viewer := range s.c.World().Viewers(s.c.Position()) {
				viewer.ViewEntityItems(s.c)
			}
		}
		if atomic.LoadUint32(s.inTransaction) == 0 {
			s.writePacket(&packet.InventorySlot{
				WindowID: protocol.WindowIDInventory,
				Slot:     uint32(slot),
				NewItem:  instanceFromItem(item),
			})
		}
	})
	s.offHand = inventory.New(2, func(slot int, item item.Stack) {
		for _, viewer := range s.c.World().Viewers(s.c.Position()) {
			viewer.ViewEntityItems(s.c)
		}
		if atomic.LoadUint32(s.inTransaction) == 0 {
			s.writePacket(&packet.InventorySlot{
				WindowID: protocol.WindowIDOffHand,
				Slot:     uint32(slot),
				NewItem:  instanceFromItem(item),
			})
		}
	})
	s.armour = inventory.NewArmour(func(slot int, item item.Stack) {
		for _, viewer := range s.c.World().Viewers(s.c.Position()) {
			viewer.ViewEntityArmour(s.c)
		}
		if atomic.LoadUint32(s.inTransaction) == 0 {
			s.writePacket(&packet.InventorySlot{
				WindowID: protocol.WindowIDArmour,
				Slot:     uint32(slot),
				NewItem:  instanceFromItem(item),
			})
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

// instanceFromItem converts an item.Stack to its network ItemInstance representation.
func instanceFromItem(it item.Stack) protocol.ItemInstance {
	return protocol.ItemInstance{
		StackNetworkID: item_id(it),
		Stack:          stackFromItem(it),
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
func creativeItems() []protocol.CreativeItem {
	it := make([]protocol.CreativeItem, 0, len(item.CreativeItems()))
	for index, i := range item.CreativeItems() {
		it = append(it, protocol.CreativeItem{
			CreativeItemNetworkID: uint32(index) + 1,
			Item:                  stackFromItem(i),
		})
	}
	return it
}

// The following functions use the go:linkname directive in order to make sure the item.byID and item.toID
// functions do not need to be exported.

//go:linkname world_itemByID github.com/df-mc/dragonfly/dragonfly/world.itemByID
//noinspection ALL
func world_itemByID(id int32, meta int16) (world.Item, bool)

//go:linkname item_id github.com/df-mc/dragonfly/dragonfly/item.id
//noinspection ALL
func item_id(s item.Stack) int32
