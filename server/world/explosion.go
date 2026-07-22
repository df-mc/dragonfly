package world

import (
	"github.com/go-gl/mathgl/mgl64"
)

// ExplosionSource represents the source of the explosion.
type ExplosionSource interface {
	Position() mgl64.Vec3
	Size() float64
}

// EntityExplosionSource is used for explosion caused by entity.
type EntityExplosionSource struct {
	Entity        Entity
	ExplosionSize float64
}

// Position ...
func (e EntityExplosionSource) Position() mgl64.Vec3 {
	return e.Entity.Position()
}

// Size ...
func (e EntityExplosionSource) Size() float64 {
	return e.ExplosionSize
}

// BlockExplosionSource is used for explosion caused by block.
type BlockExplosionSource struct {
	Block         Block
	Pos           mgl64.Vec3
	ExplosionSize float64
}

// Position ...
func (b BlockExplosionSource) Position() mgl64.Vec3 {
	return b.Pos
}

// Size ...
func (b BlockExplosionSource) Size() float64 {
	return b.ExplosionSize
}
