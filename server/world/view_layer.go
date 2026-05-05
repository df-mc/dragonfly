package world

import (
	"maps"
	"slices"
)

// layer stores the appearance overrides that a ViewLayer applies to an entity.
type layer struct {
	nameTag    *string
	scoreTag   *string
	visibility VisibilityLevel
}

// ViewLayerUpdater handles immediate updates after a ViewLayer changes how an entity is viewed.
type ViewLayerUpdater interface {
	// ViewLayerEntityChanged handles an entity whose view-layer overrides changed.
	ViewLayerEntityChanged(entity Entity)
}

type viewLayerViewer interface {
	ViewLayer() *ViewLayer
}

// ViewLayer holds overrides for how entities are viewed by a single viewer. It allows entities to be
// viewed differently by different players, such as with a different name tag or visibility state.
type ViewLayer struct {
	entities map[*EntityHandle]layer
	updater  ViewLayerUpdater
}

// NewViewLayer returns a new ViewLayer.
func NewViewLayer(updater ViewLayerUpdater) *ViewLayer {
	return &ViewLayer{
		entities: map[*EntityHandle]layer{},
		updater:  updater,
	}
}

// Entities returns the handles of all entities with overrides in the view layer.
func (v *ViewLayer) Entities() []*EntityHandle {
	return slices.Collect(maps.Keys(v.entities))
}

// ViewNameTag overwrites the public name tag of the entity and allows this ViewLayer to view a different name tag.
// Passing an empty name tag removes the name tag for this ViewLayer.
func (v *ViewLayer) ViewNameTag(entity Entity, nameTag string) {
	handle := entity.H()
	l := v.entities[handle]
	l.nameTag = &nameTag
	v.entities[handle] = l
	v.refresh(entity)
}

// ViewPublicNameTag removes the name tag override from the entity, causing the public name tag to be
// viewed again.
func (v *ViewLayer) ViewPublicNameTag(entity Entity) {
	handle := entity.H()
	l := v.entities[handle]
	l.nameTag = nil
	if l.empty() {
		delete(v.entities, handle)
	} else {
		v.entities[handle] = l
	}
	v.refresh(entity)
}

// NameTag returns the overwritten name tag of the entity and whether an override was set.
func (v *ViewLayer) NameTag(entity Entity) (string, bool) {
	nameTag := v.entities[entity.H()].nameTag
	if nameTag == nil {
		return "", false
	}
	return *nameTag, true
}

// ViewScoreTag overwrites the public score tag of the entity and allows this ViewLayer to view a different score tag.
// Passing an empty score tag removes the score tag for this ViewLayer.
func (v *ViewLayer) ViewScoreTag(entity Entity, scoreTag string) {
	handle := entity.H()
	l := v.entities[handle]
	l.scoreTag = &scoreTag
	v.entities[handle] = l
	v.refresh(entity)
}

// ViewPublicScoreTag removes the score tag override from the entity, causing the public score tag to be
// viewed again.
func (v *ViewLayer) ViewPublicScoreTag(entity Entity) {
	handle := entity.H()
	l := v.entities[handle]
	l.scoreTag = nil
	if l.empty() {
		delete(v.entities, handle)
	} else {
		v.entities[handle] = l
	}
	v.refresh(entity)
}

// ScoreTag returns the overwritten score tag of the entity and whether an override was set.
func (v *ViewLayer) ScoreTag(entity Entity) (string, bool) {
	scoreTag := v.entities[entity.H()].scoreTag
	if scoreTag == nil {
		return "", false
	}
	return *scoreTag, true
}

// ViewVisibility overwrites the public visibility of the entity and allows this ViewLayer to view
// this entity as (in)visible depending on the VisibilityLevel.
func (v *ViewLayer) ViewVisibility(entity Entity, level VisibilityLevel) {
	handle := entity.H()
	l := v.entities[handle]
	l.visibility = level
	if l.empty() {
		delete(v.entities, handle)
	} else {
		v.entities[handle] = l
	}
	v.refresh(entity)
}

// Visibility returns the visibility of the entity.
func (v *ViewLayer) Visibility(entity Entity) VisibilityLevel {
	return v.entities[entity.H()].visibility
}

// Remove removes all overrides for the entity from the ViewLayer.
func (v *ViewLayer) Remove(entity Entity) {
	v.remove(entity)
	v.refresh(entity)
}

// remove removes all overrides for the entity from the ViewLayer without refreshing entity metadata.
func (v *ViewLayer) remove(entity Entity) {
	handle := entity.H()
	delete(v.entities, handle)
}

// Close closes the view layer.
func (v *ViewLayer) Close() error {
	clear(v.entities)
	return nil
}

// empty checks if the layer does not override any public entity metadata.
func (l layer) empty() bool {
	return l.nameTag == nil && l.scoreTag == nil && l.visibility == PublicVisibility()
}

func (v *ViewLayer) refresh(entity Entity) {
	if v.updater != nil {
		v.updater.ViewLayerEntityChanged(entity)
	}
}
