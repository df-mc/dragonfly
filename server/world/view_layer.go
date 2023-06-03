package world

import (
	"sync"
)

type LayerViewer interface {
	ViewLayer() *ViewLayer
}

type Layer struct {
	nameTag string
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
	viewers := make([]LayerViewer, 0, len(v.viewers))
	for viewer := range v.viewers {
		viewers = append(viewers, viewer)
	}
	v.viewerMu.Unlock()
	return viewers
}

// ViewNameTag adds a name tag to a viewer.
func (v *ViewLayer) ViewNameTag(viewer LayerViewer, nameTag string) {
	v.viewerMu.Lock()
	v.viewers[viewer] = Layer{nameTag: nameTag}
	v.viewerMu.Unlock()
}

// NameTag returns the name tag of a viewer.
func (v *ViewLayer) NameTag(viewer LayerViewer) string {
	v.viewerMu.Lock()
	nameTag := v.viewers[viewer].nameTag
	v.viewerMu.Unlock()
	return nameTag
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
