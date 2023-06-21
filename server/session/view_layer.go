package session

import "github.com/df-mc/dragonfly/server/world"

type LayerViewer interface {
	ViewLayer() *world.ViewLayer
}

func (s *Session) ViewLayer() *world.ViewLayer {
	return s.viewLayer
}
