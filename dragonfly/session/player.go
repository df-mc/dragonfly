package session

import (
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/entity"
	"github.com/dragonfly-tech/dragonfly/dragonfly/item"
	"github.com/dragonfly-tech/dragonfly/dragonfly/item/inventory"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/skin"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world/gamemode"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"net"
	"sync/atomic"
)

// handleMovePlayer ...
func (s *Session) handleMovePlayer(pk *packet.MovePlayer) error {
	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return fmt.Errorf("incorrect entity runtime ID %v: runtime ID must be 1", pk.EntityRuntimeID)
	}
	entity.Move(s.c, pk.Position.Sub(s.c.Position()))
	entity.Rotate(s.c, pk.Yaw-s.c.Yaw(), pk.Pitch-s.c.Pitch())

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

// handleInventoryTransaction ...
func (s *Session) handleInventoryTransaction(pk *packet.InventoryTransaction) error {
	switch data := pk.TransactionData.(type) {
	case *protocol.UseItemTransactionData:
		switch data.ActionType {
		case protocol.UseItemActionBreakBlock:
			_ = s.c.BreakBlock(world.BlockPos{int(data.BlockPosition[0]), int(data.BlockPosition[1]), int(data.BlockPosition[2])})
		}
	}
	return nil
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
	inv = inventory.New(36, func(slot int, item item.Stack) {
		s.writePacket(&packet.InventorySlot{
			WindowID: protocol.WindowIDInventory,
			Slot:     uint32(slot),
			NewItem:  stackFromItem(item),
		})
	})
	offHand = inventory.New(1, func(slot int, item item.Stack) {
		s.writePacket(&packet.InventorySlot{
			WindowID: protocol.WindowIDOffHand,
			Slot:     uint32(slot),
			NewItem:  stackFromItem(item),
		})
	})
	heldSlot = s.heldSlot

	return
}

// stackFromItem converts an item.Stack to its network ItemStack representation.
func stackFromItem(item item.Stack) protocol.ItemStack {
	id, meta := item.Item().EncodeItem()
	return protocol.ItemStack{
		ItemType: protocol.ItemType{
			NetworkID:     id,
			MetadataValue: meta,
		},
		Count: int16(item.Count()),
	}
}
