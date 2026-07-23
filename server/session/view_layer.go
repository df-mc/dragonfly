package session

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

type viewLayerArmour struct {
	helmet, chestplate, leggings, boots item.Stack
}

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

// ViewAlwaysShowNameTag overrides whether the entity's name tag is shown at all distances for this session.
func (s *Session) ViewAlwaysShowNameTag(entity world.Entity, alwaysShow bool) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.ViewAlwaysShowNameTag(entity, alwaysShow)
}

// ViewPublicAlwaysShowNameTag removes the always-show name tag override from the entity for this session.
func (s *Session) ViewPublicAlwaysShowNameTag(entity world.Entity) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.ViewPublicAlwaysShowNameTag(entity)
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
	s.viewLayerArmourChanged(entity)
}

// ViewArmour overwrites the public armour of the entity and immediately refreshes it for this session.
func (s *Session) ViewArmour(entity world.Entity, helmet, chestplate, leggings, boots item.Stack) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayerArmourMu.Lock()
	s.viewLayerArmour[entity.H()] = viewLayerArmour{
		helmet:     helmet,
		chestplate: chestplate,
		leggings:   leggings,
		boots:      boots,
	}
	s.viewLayerArmourMu.Unlock()
	s.viewLayerArmourChanged(entity)
}

// ViewNoArmour hides the armour of the entity and immediately refreshes it for this session.
func (s *Session) ViewNoArmour(entity world.Entity) {
	s.ViewArmour(entity, item.Stack{}, item.Stack{}, item.Stack{}, item.Stack{})
}

// ViewPublicArmour removes the armour override from the entity and immediately refreshes it for this session.
func (s *Session) ViewPublicArmour(entity world.Entity) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayerArmourMu.Lock()
	delete(s.viewLayerArmour, entity.H())
	s.viewLayerArmourMu.Unlock()
	s.viewLayerArmourChanged(entity)
}

// RemoveViewLayer removes all overrides for the entity and immediately refreshes it for this session.
func (s *Session) RemoveViewLayer(entity world.Entity) {
	if s.viewLayer == nil {
		return
	}
	s.viewLayer.Remove(entity)
	s.viewLayerArmourMu.Lock()
	delete(s.viewLayerArmour, entity.H())
	s.viewLayerArmourMu.Unlock()
	s.viewLayerArmourChanged(entity)
}

// ViewLayerEntityChanged refreshes the entity metadata for this session if the entity is currently visible.
func (s *Session) ViewLayerEntityChanged(e world.Entity) {
	if s.entityHidden(e) || !s.viewingEntity(e.H()) {
		return
	}
	s.ViewEntityState(e)
}

// viewLayerArmourChanged refreshes the entity armour for this session if the entity is currently visible.
func (s *Session) viewLayerArmourChanged(e world.Entity) {
	if s.entityHidden(e) || !s.viewingEntity(e.H()) {
		return
	}
	s.ViewEntityArmour(e)
}

func (s *Session) viewedArmour(entity world.Entity) (viewLayerArmour, bool) {
	s.viewLayerArmourMu.RLock()
	armour, ok := s.viewLayerArmour[entity.H()]
	s.viewLayerArmourMu.RUnlock()
	return armour, ok
}

// viewingEntity checks if this session currently has a runtime ID assigned to the entity handle.
func (s *Session) viewingEntity(handle *world.EntityHandle) bool {
	s.entityMutex.RLock()
	_, ok := s.entityRuntimeIDs[handle]
	s.entityMutex.RUnlock()
	return ok
}
