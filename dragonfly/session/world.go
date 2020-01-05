package session

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/world"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync/atomic"
)

// handleRequestChunkRadius ...
func (s *Session) handleRequestChunkRadius(pk *packet.RequestChunkRadius) error {
	if pk.ChunkRadius > s.maxChunkRadius {
		pk.ChunkRadius = s.maxChunkRadius
	}
	s.chunkRadius = pk.ChunkRadius

	_ = s.chunkLoader.Load().(*world.Loader).Close()
	s.chunkLoader.Store(world.NewLoader(int(s.chunkRadius), s.world, s))

	s.writePacket(&packet.ChunkRadiusUpdated{ChunkRadius: s.chunkRadius})
	return nil
}

// SendNetherDimension sends the player to the nether dimension
func (s *Session) SendNetherDimension() {
	s.writePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionNether,
		Position:  mgl32.Vec3{},
		Respawn:   false,
	})
}

// SendEndDimension sends the player to the end dimension
func (s *Session) SendEndDimension() {
	s.writePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionEnd,
		Position:  mgl32.Vec3{},
		Respawn:   false,
	})
}

// SendNetherDimension sends the player to the overworld dimension
func (s *Session) SendOverworldDimension() {
	s.writePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionOverworld,
		Position:  mgl32.Vec3{},
		Respawn:   false,
	})
}

// ViewChunk ...
func (s *Session) ViewChunk(pos world.ChunkPos, c *chunk.Chunk) {
	data := chunk.NetworkEncode(c)

	count := 16
	for y := 15; y >= 0; y-- {
		if data.SubChunks[y] == nil {
			count--
			continue
		}
		break
	}
	for y := 0; y < count; y++ {
		if data.SubChunks[y] == nil {
			_ = s.chunkBuf.WriteByte(chunk.SubChunkVersion)
			// We write zero here, meaning the sub chunk has no block storages: The sub chunk is completely
			// empty.
			_ = s.chunkBuf.WriteByte(0)
			continue
		}
		_, _ = s.chunkBuf.Write(data.SubChunks[y])
	}
	_, _ = s.chunkBuf.Write(data.Data2D)
	_, _ = s.chunkBuf.Write(data.BlockNBT)

	s.writePacket(&packet.LevelChunk{
		ChunkX:        pos[0],
		ChunkZ:        pos[1],
		SubChunkCount: uint32(count),
		RawPayload:    append([]byte(nil), s.chunkBuf.Bytes()...),
	})
	s.chunkBuf.Reset()
}

// ViewEntity ...
func (s *Session) ViewEntity(e world.Entity) {
	if s.entityRuntimeID(e) == selfEntityRuntimeID {
		return
	}
	var runtimeID uint64

	s.entityMutex.Lock()
	_, controllable := e.(Controllable)

	if id, ok := s.entityRuntimeIDs[e]; ok && controllable {
		runtimeID = id
	} else {
		runtimeID = atomic.AddUint64(&s.currentEntityRuntimeID, 1)
		s.entityRuntimeIDs[e] = runtimeID
	}
	s.entityMutex.Unlock()

	switch v := e.(type) {
	case Controllable:
		s.writePacket(&packet.AddPlayer{
			UUID:            v.UUID(),
			Username:        v.Name(),
			EntityUniqueID:  int64(runtimeID),
			EntityRuntimeID: runtimeID,
			Position:        e.Position(),
			Pitch:           e.Pitch(),
			Yaw:             e.Yaw(),
			HeadYaw:         e.Yaw(),
		})
	default:
		s.writePacket(&packet.AddActor{
			EntityUniqueID:  int64(runtimeID),
			EntityRuntimeID: runtimeID,
			// TODO: Add methods for entity types.
			EntityType: "",
			Position:   e.Position(),
			Pitch:      e.Pitch(),
			Yaw:        e.Yaw(),
			HeadYaw:    e.Yaw(),
		})
	}
}

// HideEntity ...
func (s *Session) HideEntity(e world.Entity) {
	s.entityMutex.Lock()
	id, ok := s.entityRuntimeIDs[e]
	if _, controllable := e.(Controllable); !controllable {
		delete(s.entityRuntimeIDs, e)
	}
	s.entityMutex.Unlock()
	if !ok {
		s.log.Debugf("cannot hide entity %T with runtime ID %v: entity is not shown", e, id)
		return
	}
	s.writePacket(&packet.RemoveActor{EntityUniqueID: int64(id)})
}

// ViewEntityMovement ...
func (s *Session) ViewEntityMovement(e world.Entity, deltaPos mgl32.Vec3, deltaYaw, deltaPitch float32) {
	id := s.entityRuntimeID(e)

	if id == selfEntityRuntimeID {
		return
	}

	switch e.(type) {
	case Controllable:
		s.writePacket(&packet.MovePlayer{
			EntityRuntimeID: id,
			Position:        e.Position().Add(deltaPos),
			Pitch:           e.Pitch() + deltaPitch,
			Yaw:             e.Yaw() + deltaYaw,
			HeadYaw:         e.Yaw() + deltaYaw,
		})
	default:
		s.writePacket(&packet.MoveActorAbsolute{
			EntityRuntimeID: id,
			Position:        e.Position().Add(deltaPos),
			Rotation:        mgl32.Vec3{e.Pitch() + deltaPitch, e.Yaw() + deltaYaw},
		})
	}
}

// ViewTime ...
func (s *Session) ViewTime(time int) {
	s.writePacket(&packet.SetTime{Time: int32(time)})
}

// ViewEntityTeleport ...
func (s *Session) ViewEntityTeleport(e world.Entity, position mgl32.Vec3) {
	id := s.entityRuntimeID(e)

	if id == selfEntityRuntimeID {
		s.chunkLoader.Load().(*world.Loader).Move(position)
	}

	switch e.(type) {
	case Controllable:
		s.writePacket(&packet.MovePlayer{
			EntityRuntimeID: id,
			Position:        position,
			Pitch:           e.Pitch(),
			Yaw:             e.Yaw(),
			HeadYaw:         e.Yaw(),
			Mode:            packet.MoveModeTeleport,
		})
	default:
		s.writePacket(&packet.MoveActorAbsolute{
			EntityRuntimeID: id,
			Position:        position,
			Rotation:        mgl32.Vec3{e.Pitch(), e.Yaw()},
			Flags:           packet.MoveFlagTeleport,
		})
	}
}

// entityRuntimeID returns the runtime ID of the entity passed.
func (s *Session) entityRuntimeID(e world.Entity) uint64 {
	s.entityMutex.RLock()
	id, _ := s.entityRuntimeIDs[e]
	s.entityMutex.RUnlock()
	return id
}
