package session

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// ViewLayer returns the session's ViewLayer. The layer may be used to override how entities are viewed
// by this session, such as with a different name tag or visibility state.
func (s *Session) ViewLayer() *world.ViewLayer {
	return s.viewLayer
}

// ViewNameTag overwrites the public name tag of the entity and immediately refreshes it for this session.
func (s *Session) ViewNameTag(entity world.Entity, nameTag string) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.ViewNameTag(entity, nameTag)
}

// ViewPublicNameTag removes the name tag override from the entity and immediately refreshes it for this session.
func (s *Session) ViewPublicNameTag(entity world.Entity) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.ViewPublicNameTag(entity)
}

// ViewScoreTag overwrites the public score tag of the entity and immediately refreshes it for this session.
func (s *Session) ViewScoreTag(entity world.Entity, scoreTag string) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.ViewScoreTag(entity, scoreTag)
}

// ViewPublicScoreTag removes the score tag override from the entity and immediately refreshes it for this session.
func (s *Session) ViewPublicScoreTag(entity world.Entity) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.ViewPublicScoreTag(entity)
}

// ViewVisibility overwrites the public visibility of the entity and immediately refreshes it for this session.
func (s *Session) ViewVisibility(entity world.Entity, level world.VisibilityLevel) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.ViewVisibility(entity, level)
}

// ViewBlock overwrites the public block at the position passed and immediately refreshes it for this session.
func (s *Session) ViewBlock(pos cube.Pos, b world.Block) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.ViewBlock(pos, b)
}

// ViewPublicBlock removes the block override at the position passed.
func (s *Session) ViewPublicBlock(pos cube.Pos) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.ViewPublicBlock(pos)
}

// RemoveViewLayer removes all overrides for the entity and immediately refreshes it for this session.
func (s *Session) RemoveViewLayer(entity world.Entity) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.Remove(entity)
}

// ViewLayerEntityChanged refreshes the entity metadata for this session if the entity is currently visible.
func (s *Session) ViewLayerEntityChanged(e world.Entity) {
	if s.entityHidden(e) || !s.viewingEntity(e.H()) {
		return
	}
	s.ViewEntityState(e)
}

// ViewLayerBlockChanged refreshes a block override for this session if its chunk is currently visible.
func (s *Session) ViewLayerBlockChanged(pos cube.Pos) {
	if !s.viewingBlock(pos) {
		return
	}
	if b, ok := s.viewLayer.Block(pos); ok {
		s.viewBlockUpdate(pos, b, 0)
	}
}

// viewingEntity checks if this session currently has a runtime ID assigned to the entity handle.
func (s *Session) viewingEntity(handle *world.EntityHandle) bool {
	s.entityMutex.RLock()
	_, ok := s.entityRuntimeIDs[handle]
	s.entityMutex.RUnlock()
	return ok
}

// viewingBlock returns true if the block position is loaded for this session.
func (s *Session) viewingBlock(pos cube.Pos) bool {
	if s.chunkLoader == nil {
		return false
	}
	col, ok := s.chunkLoader.Chunk(world.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)})
	if !ok {
		return false
	}
	return !pos.OutOfBounds(col.Range())
}
