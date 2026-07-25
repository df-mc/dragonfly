package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

// defaultExplosionSize is the size used if a source does not specify one.
const defaultExplosionSize = 4

// ExplosionSource represents the source of an explosion.
type ExplosionSource interface {
	// Position returns the position at the centre of the explosion. It must
	// return the same position for the duration of an explosion.
	Position() mgl64.Vec3
	// Size returns the radius which entities/blocks are affected within.
	Size() float64
}

// EntityExplosionSource is used for an explosion caused by an entity.
type EntityExplosionSource struct {
	// Entity is the entity that caused the explosion.
	Entity Entity
	// ExplosionSize is the size of the explosion. Defaults to 4 if 0.
	ExplosionSize float64
}

// Position ...
func (e EntityExplosionSource) Position() mgl64.Vec3 {
	return e.Entity.Position()
}

// Size ...
func (e EntityExplosionSource) Size() float64 {
	if e.ExplosionSize == 0 {
		return defaultExplosionSize
	}
	return e.ExplosionSize
}

// BlockExplosionSource is used for an explosion caused by a block.
type BlockExplosionSource struct {
	// Block is the block that caused the explosion.
	Block Block
	// Pos is the position of the block that caused the explosion.
	Pos cube.Pos
	// ExplosionSize is the size of the explosion. Defaults to 4 if 0.
	ExplosionSize float64
}

// Position ...
func (b BlockExplosionSource) Position() mgl64.Vec3 {
	return b.Pos.Vec3Centre()
}

// Size ...
func (b BlockExplosionSource) Size() float64 {
	if b.ExplosionSize == 0 {
		return defaultExplosionSize
	}
	return b.ExplosionSize
}
