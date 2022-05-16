package world

import "github.com/go-gl/mathgl/mgl64"

type Explodion struct {
	w *World
}

var rays int8 = 16

// Explode explodes the blocks in the center of the provided vec3
func (e Explodion) Explode(center mgl64.Vec3) {
	if e.w == nil {
		return
	}

}
