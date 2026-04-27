package session

import "github.com/df-mc/dragonfly/server/world"

// LayerViewer represents an entity whose appearance may be overridden by a session ViewLayer.
type LayerViewer interface {
	// ViewLayer returns the ViewLayer attached to the viewer.
	ViewLayer() *world.ViewLayer
}

// ViewLayer returns the session's ViewLayer. The layer may be used to override how entities are viewed
// by this session, such as with a different name tag or visibility state.
func (s *Session) ViewLayer() *world.ViewLayer {
	return s.viewLayer
}
