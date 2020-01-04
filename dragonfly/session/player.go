package session

import (
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/entity"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/skin"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"net"
	"sync/atomic"
)

// handleMovePlayer ...
func (s *Session) handleMovePlayer(pk *packet.MovePlayer) error {
	if pk.EntityRuntimeID != s.conn.GameData().EntityRuntimeID {
		return fmt.Errorf("incorrect entity runtime ID %v: runtime ID must be equal to %v", pk.EntityRuntimeID, s.conn.GameData().EntityRuntimeID)
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

// addToPlayerList adds the player of a session to the player list of this session. It will be shown in the
// in-game pause menu screen.
func (s *Session) addToPlayerList(session *Session) {
	c := session.c

	s.entityMutex.Lock()
	var runtimeID uint64
	if session != s {
		runtimeID = atomic.AddUint64(&s.currentEntityRuntimeID, 1)
	} else {
		runtimeID = 1
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
