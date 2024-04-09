package world

import (
	"sync"
)

type LayerViewer interface {
	ViewLayer() *ViewLayer
}

type Layer struct {
	nameTag    string
	visibility VisibilityLevel
}

// ViewLayer is a view layer that can be used to add a layer to the view of a player.
type ViewLayer struct {
	viewerMu sync.RWMutex
	viewers  map[LayerViewer]Layer
}

// NewViewLayer returns a new ViewLayer.
func NewViewLayer() *ViewLayer {
	return &ViewLayer{
		viewers: map[LayerViewer]Layer{},
	}
}

// Viewers returns all viewers of the view layer.
func (v *ViewLayer) Viewers() []LayerViewer {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()
	viewers := make([]LayerViewer, 0, len(v.viewers))
	for viewer := range v.viewers {
		viewers = append(viewers, viewer)
	}
	return viewers
}

// ViewNameTag overwrites the public name tag of the viewer and allows this ViewLayer to view a different name tag.
// Leaving the name tag empty reverts this behaviour.
func (v *ViewLayer) ViewNameTag(viewer LayerViewer, nameTag string) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	l := v.viewers[viewer]
	l.nameTag = nameTag
	v.viewers[viewer] = l
}

// NameTag returns the overwritten name tag of the viewer.
func (v *ViewLayer) NameTag(viewer LayerViewer) string {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()
	return v.viewers[viewer].nameTag
}

// ViewVisibility overwrites the public visibility of the viewer and allows this ViewLayer to view
// this viewer as (in)visible depending on the VisibilityLevel.
func (v *ViewLayer) ViewVisibility(viewer LayerViewer, level VisibilityLevel) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	l := v.viewers[viewer]
	l.visibility = level
	v.viewers[viewer] = l
}

// Visibility returns the visibility of the viewer.
func (v *ViewLayer) Visibility(viewer LayerViewer) VisibilityLevel {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()
	return v.viewers[viewer].visibility
}

// Close closes the view layer.
func (v *ViewLayer) Close() error {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()
	for viewer := range v.viewers {
		delete(v.viewers, viewer)
	}
	return nil
}
