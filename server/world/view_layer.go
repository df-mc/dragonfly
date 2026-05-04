package world

import "sync"

// layer stores the appearance overrides that a ViewLayer applies to a viewer.
type layer struct {
	viewer     Entity
	nameTag    *string
	scoreTag   *string
	visibility VisibilityLevel
}

// ViewLayer holds per-viewer overrides for entities. It allows entities to be viewed differently by
// different players, such as with a different name tag or visibility state.
type ViewLayer struct {
	viewerMu sync.RWMutex
	viewers  map[*EntityHandle]layer
}

// NewViewLayer returns a new ViewLayer.
func NewViewLayer() *ViewLayer {
	return &ViewLayer{
		viewers: map[*EntityHandle]layer{},
	}
}

// Viewers returns all viewers with overrides in the view layer.
func (v *ViewLayer) Viewers() []Entity {
	v.viewerMu.RLock()
	defer v.viewerMu.RUnlock()
	viewers := make([]Entity, 0, len(v.viewers))
	for _, l := range v.viewers {
		viewers = append(viewers, l.viewer)
	}
	return viewers
}

// ViewNameTag overwrites the public name tag of the viewer and allows this ViewLayer to view a different name tag.
// Passing an empty name tag removes the name tag for this ViewLayer.
func (v *ViewLayer) ViewNameTag(viewer Entity, nameTag string) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	handle := viewer.H()
	l := v.viewers[handle]
	l.viewer = viewer
	l.nameTag = &nameTag
	v.viewers[handle] = l
}

// ViewPublicNameTag removes the name tag override from the viewer, causing the public name tag to be
// viewed again.
func (v *ViewLayer) ViewPublicNameTag(viewer Entity) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	handle := viewer.H()
	l := v.viewers[handle]
	l.nameTag = nil
	if l.empty() {
		delete(v.viewers, handle)
		return
	}
	l.viewer = viewer
	v.viewers[handle] = l
}

// NameTag returns the overwritten name tag of the viewer and whether an override was set.
func (v *ViewLayer) NameTag(viewer Entity) (string, bool) {
	v.viewerMu.RLock()
	defer v.viewerMu.RUnlock()
	nameTag := v.viewers[viewer.H()].nameTag
	if nameTag == nil {
		return "", false
	}
	return *nameTag, true
}

// ViewScoreTag overwrites the public score tag of the viewer and allows this ViewLayer to view a different score tag.
// Passing an empty score tag removes the score tag for this ViewLayer.
func (v *ViewLayer) ViewScoreTag(viewer Entity, scoreTag string) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	handle := viewer.H()
	l := v.viewers[handle]
	l.viewer = viewer
	l.scoreTag = &scoreTag
	v.viewers[handle] = l
}

// ViewPublicScoreTag removes the score tag override from the viewer, causing the public score tag to be
// viewed again.
func (v *ViewLayer) ViewPublicScoreTag(viewer Entity) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	handle := viewer.H()
	l := v.viewers[handle]
	l.scoreTag = nil
	if l.empty() {
		delete(v.viewers, handle)
		return
	}
	l.viewer = viewer
	v.viewers[handle] = l
}

// ScoreTag returns the overwritten score tag of the viewer and whether an override was set.
func (v *ViewLayer) ScoreTag(viewer Entity) (string, bool) {
	v.viewerMu.RLock()
	defer v.viewerMu.RUnlock()
	scoreTag := v.viewers[viewer.H()].scoreTag
	if scoreTag == nil {
		return "", false
	}
	return *scoreTag, true
}

// ViewVisibility overwrites the public visibility of the viewer and allows this ViewLayer to view
// this viewer as (in)visible depending on the VisibilityLevel.
func (v *ViewLayer) ViewVisibility(viewer Entity, level VisibilityLevel) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	handle := viewer.H()
	l := v.viewers[handle]
	l.viewer = viewer
	l.visibility = level
	if l.empty() {
		delete(v.viewers, handle)
		return
	}
	v.viewers[handle] = l
}

// Visibility returns the visibility of the viewer.
func (v *ViewLayer) Visibility(viewer Entity) VisibilityLevel {
	v.viewerMu.RLock()
	defer v.viewerMu.RUnlock()
	return v.viewers[viewer.H()].visibility
}

// Remove removes all overrides for the viewer from the ViewLayer.
func (v *ViewLayer) Remove(viewer Entity) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	handle := viewer.H()
	delete(v.viewers, handle)
}

// Close closes the view layer.
func (v *ViewLayer) Close() error {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()
	clear(v.viewers)
	return nil
}

func (l layer) empty() bool {
	return l.nameTag == nil && l.scoreTag == nil && l.visibility == PublicVisibility()
}
