package world

import "sync"

// LayerViewer represents an entity whose appearance may be overridden by a ViewLayer for individual
// viewers.
type LayerViewer interface {
	// ViewLayer returns the ViewLayer attached to the viewer.
	ViewLayer() *ViewLayer
}

// layer stores the appearance overrides that a ViewLayer applies to a LayerViewer.
type layer struct {
	nameTag    *string
	scoreTag   *string
	visibility VisibilityLevel
}

// ViewLayer holds per-viewer overrides for entities. It allows entities to be viewed differently by
// different players, such as with a different name tag or visibility state.
type ViewLayer struct {
	viewerMu sync.RWMutex
	viewers  map[LayerViewer]layer
}

// NewViewLayer returns a new ViewLayer.
func NewViewLayer() *ViewLayer {
	return &ViewLayer{
		viewers: map[LayerViewer]layer{},
	}
}

// Viewers returns all viewers of the view layer.
func (v *ViewLayer) Viewers() []LayerViewer {
	v.viewerMu.RLock()
	defer v.viewerMu.RUnlock()
	viewers := make([]LayerViewer, 0, len(v.viewers))
	for viewer := range v.viewers {
		viewers = append(viewers, viewer)
	}
	return viewers
}

// ViewNameTag overwrites the public name tag of the viewer and allows this ViewLayer to view a different name tag.
// Passing an empty name tag removes the name tag for this ViewLayer.
func (v *ViewLayer) ViewNameTag(viewer LayerViewer, nameTag string) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	l := v.viewers[viewer]
	l.nameTag = &nameTag
	v.viewers[viewer] = l
}

// ViewPublicNameTag removes the name tag override from the viewer, causing the public name tag to be
// viewed again.
func (v *ViewLayer) ViewPublicNameTag(viewer LayerViewer) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	l := v.viewers[viewer]
	l.nameTag = nil
	if l.empty() {
		delete(v.viewers, viewer)
		return
	}
	v.viewers[viewer] = l
}

// NameTag returns the overwritten name tag of the viewer and whether an override was set.
func (v *ViewLayer) NameTag(viewer LayerViewer) (string, bool) {
	v.viewerMu.RLock()
	defer v.viewerMu.RUnlock()
	nameTag := v.viewers[viewer].nameTag
	if nameTag == nil {
		return "", false
	}
	return *nameTag, true
}

// ViewScoreTag overwrites the public score tag of the viewer and allows this ViewLayer to view a different score tag.
// Passing an empty score tag removes the score tag for this ViewLayer.
func (v *ViewLayer) ViewScoreTag(viewer LayerViewer, scoreTag string) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	l := v.viewers[viewer]
	l.scoreTag = &scoreTag
	v.viewers[viewer] = l
}

// ViewPublicScoreTag removes the score tag override from the viewer, causing the public score tag to be
// viewed again.
func (v *ViewLayer) ViewPublicScoreTag(viewer LayerViewer) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	l := v.viewers[viewer]
	l.scoreTag = nil
	if l.empty() {
		delete(v.viewers, viewer)
		return
	}
	v.viewers[viewer] = l
}

// ScoreTag returns the overwritten score tag of the viewer and whether an override was set.
func (v *ViewLayer) ScoreTag(viewer LayerViewer) (string, bool) {
	v.viewerMu.RLock()
	defer v.viewerMu.RUnlock()
	scoreTag := v.viewers[viewer].scoreTag
	if scoreTag == nil {
		return "", false
	}
	return *scoreTag, true
}

// ViewVisibility overwrites the public visibility of the viewer and allows this ViewLayer to view
// this viewer as (in)visible depending on the VisibilityLevel.
func (v *ViewLayer) ViewVisibility(viewer LayerViewer, level VisibilityLevel) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	l := v.viewers[viewer]
	l.visibility = level
	if l.empty() {
		delete(v.viewers, viewer)
		return
	}
	v.viewers[viewer] = l
}

// Visibility returns the visibility of the viewer.
func (v *ViewLayer) Visibility(viewer LayerViewer) VisibilityLevel {
	v.viewerMu.RLock()
	defer v.viewerMu.RUnlock()
	return v.viewers[viewer].visibility
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
