package world

import (
	"sync"
)

type LayerViewer interface {
	ViewLayer() *ViewLayer
}

type Layer struct {
	nameTag   string
	invisible bool
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

// ViewNameTag adds a name tag to a viewer.
func (v *ViewLayer) ViewNameTag(viewer LayerViewer, nameTag string) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	l := v.viewers[viewer]
	l.nameTag = nameTag
	v.viewers[viewer] = l
}

// NameTag returns the name tag of a viewer.
func (v *ViewLayer) NameTag(viewer LayerViewer) string {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()
	return v.viewers[viewer].nameTag
}

// ViewVisible makes a viewer be visible.
func (v *ViewLayer) ViewVisible(viewer LayerViewer) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	l := v.viewers[viewer]
	l.invisible = false
	v.viewers[viewer] = l
}

// ViewInvisible makes a viewer be invisible.
func (v *ViewLayer) ViewInvisible(viewer LayerViewer) {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()

	l := v.viewers[viewer]
	l.invisible = true
	v.viewers[viewer] = l
}

// Invisible returns the invisibility of a viewer.
func (v *ViewLayer) Invisible(viewer LayerViewer) bool {
	v.viewerMu.Lock()
	defer v.viewerMu.Unlock()
	return v.viewers[viewer].invisible
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
