package session

import (
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/block"
	"github.com/dragonfly-tech/dragonfly/dragonfly/item"
	"github.com/dragonfly-tech/dragonfly/dragonfly/item/inventory"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/skin"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world/gamemode"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"math"
	"net"
	"sync/atomic"
)

// handleMovePlayer ...
func (s *Session) handleMovePlayer(pk *packet.MovePlayer) error {
	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return fmt.Errorf("incorrect entity runtime ID %v: runtime ID must be 1", pk.EntityRuntimeID)
	}
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
	switch pk.ActionType {
	case packet.PlayerActionStartSprint:
		s.c.StartSprinting()
	case packet.PlayerActionStopSprint:
		s.c.StopSprinting()
	case packet.PlayerActionStartSneak:
		s.c.StartSneaking()
	case packet.PlayerActionStopSneak:
		s.c.StopSneaking()
	}
	return nil
}

// handleInventoryTransaction ...
func (s *Session) handleInventoryTransaction(pk *packet.InventoryTransaction) error {
	switch data := pk.TransactionData.(type) {
	case *protocol.NormalTransactionData:
		if len(pk.Actions) == 0 {
			return nil
		}
		if err := s.verifyTransaction(pk.Actions); err != nil {
			return fmt.Errorf("invalid inventory transaction: %v", err)
		}
		atomic.StoreUint32(&s.inTransaction, 1)
		s.executeTransaction(pk.Actions)
		atomic.StoreUint32(&s.inTransaction, 0)
	case *protocol.UseItemTransactionData:
		switch data.ActionType {
		case protocol.UseItemActionBreakBlock:
			_ = s.c.BreakBlock(block.Position{int(data.BlockPosition[0]), int(data.BlockPosition[1]), int(data.BlockPosition[2])})
		case protocol.UseItemActionClickBlock:
			_ = s.c.UseItemOnBlock(block.Position{int(data.BlockPosition[0]), int(data.BlockPosition[1]), int(data.BlockPosition[2])}, block.Face(data.BlockFace), data.ClickedPosition)
		case protocol.UseItemActionClickAir:
			_ = s.c.UseItem()
		}
	}
	return nil
}

// executeTransaction executes all actions of a transaction passed. It assumes the actions are all valid,
// which would otherwise be checked by calling verifyTransaction.
func (s *Session) executeTransaction(actions []protocol.InventoryAction) {
	for _, action := range actions {
		// The window IDs are already checked when using verifyTransaction, so we don't need to check again.
		inv, _ := s.invByID(action.WindowID)
		_ = inv.SetItem(int(action.InventorySlot), stackToItem(action.NewItem))
	}
}

// verifyTransaction verifies a transaction composed of the actions passed. The method makes sure the old
// items are precisely equal to the new items: No new items must be added or removed.
// verifyTransaction also checks if all window IDs sent match some inventory, and if the old items match the
// items found in that inventory.
func (s *Session) verifyTransaction(actions []protocol.InventoryAction) error {
	// Allocate a big inventory and add all new items to it.
	temp := inventory.New(128, nil)
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
			// Either the type or the count of the old item as registered by the server and the client are
			// mismatched.
			return fmt.Errorf("slot %v holds a different item than the client expects: %v is actually %v", action.InventorySlot, old, actualOld)
		}
		if err := temp.AddItem(old); err != nil {
			return fmt.Errorf("inventory was full: %v", err)
		}
	}
	for _, action := range actions {
		if err := temp.RemoveItem(stackToItem(action.NewItem)); err != nil {
			return fmt.Errorf("item removed was not present in old items: %v", err)
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

// SendGameRules sends all the provided game rules to the player. Once sent, they
// will be immediately updated on the client if they are valid.
func (s *Session) sendGameRules(gamerules map[string]interface{}) {
	s.writePacket(&packet.GameRulesChanged{
		GameRules: gamerules,
	})
}

// EnableCoordinates will either enable or disable coordinates for the
// player depending on the value given.
func (s *Session) EnableCoordinates(enable bool) {
	gamerules := make(map[string]interface{})
	gamerules["showCoordinates"] = enable

	s.sendGameRules(gamerules)
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
	s.entityMutex.Unlock()

	var animations []protocol.SkinAnimation
	for _, animation := range c.Skin().Animations {
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

	playerSkin := c.Skin()
	s.writePacket(&packet.PlayerList{
		ActionType: packet.PlayerListActionAdd,
		Entries: []protocol.PlayerListEntry{{
			UUID:           c.UUID(),
			EntityUniqueID: int64(runtimeID),
			Username:       c.Name(),
			XUID:           c.XUID(),
			Skin: protocol.Skin{
				SkinID:            uuid.New().String(),
				SkinResourcePatch: playerSkin.ModelConfig.Encode(),
				SkinImageWidth:    uint32(playerSkin.Bounds().Max.X),
				SkinImageHeight:   uint32(playerSkin.Bounds().Max.Y),
				SkinData:          playerSkin.Pix,
				CapeImageWidth:    uint32(playerSkin.Cape.Bounds().Max.X),
				CapeImageHeight:   uint32(playerSkin.Cape.Bounds().Max.Y),
				CapeData:          playerSkin.Cape.Pix,
				SkinGeometry:      playerSkin.Model,
				PersonaSkin:       session.conn.ClientData().PersonaSkin,
				CapeID:            uuid.New().String(),
				FullSkinID:        uuid.New().String(),
				Animations:        animations,
			},
		}},
	})
}

// removeFromPlayerList removes the player of a session from the player list of this session. It will no
// longer be shown in the in-game pause menu screen.
func (s *Session) removeFromPlayerList(session *Session) {
	c := session.c

	s.entityMutex.Lock()
	delete(s.entityRuntimeIDs, c)
	s.entityMutex.Unlock()

	s.writePacket(&packet.PlayerList{
		ActionType: packet.PlayerListActionRemove,
		Entries: []protocol.PlayerListEntry{{
			UUID: c.UUID(),
		}},
	})
}

// handleInventories starts handling the inventories of the Controllable of the session. It sends packets when
// slots in the inventory are changed.
func (s *Session) HandleInventories() (inv, offHand *inventory.Inventory, heldSlot *uint32) {
	s.inv = inventory.New(36, func(slot int, item item.Stack) {
		if atomic.LoadUint32(&s.inTransaction) == 1 {
			return
		}
		s.writePacket(&packet.InventorySlot{
			WindowID: protocol.WindowIDInventory,
			Slot:     uint32(slot),
			NewItem:  stackFromItem(item),
		})
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
	})
	return s.inv, s.offHand, s.heldSlot
}

// stackFromItem converts an item.Stack to its network ItemStack representation.
func stackFromItem(it item.Stack) protocol.ItemStack {
	if it.Empty() {
		return protocol.ItemStack{}
	}
	id, meta := item.ToID(it.Item())
	return protocol.ItemStack{
		ItemType: protocol.ItemType{
			NetworkID:     id,
			MetadataValue: meta,
		},
		Count: int16(it.Count()),
	}
}

// stackToItem converts a network ItemStack representation back to an item.Stack.
func stackToItem(it protocol.ItemStack) item.Stack {
	// TODO: Handle item NBT.
	t, ok := item.ByID(it.NetworkID, it.MetadataValue)
	if !ok {
		t = block.Air{}
	}
	return item.NewStack(t, int(it.Count))
}
