package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

type Painting struct {
	transform

	motive    PaintingMotive
	direction cube.Direction
}

// NewPainting creates a new Painting entity.
func NewPainting(motive PaintingMotive, direction cube.Direction, pos mgl64.Vec3) *Painting {
	b := &Painting{
		motive:    motive,
		direction: direction,
	}
	b.transform = newTransform(b, pos)
	return b
}

// Motive returns the motive of the painting.
func (p *Painting) Motive() PaintingMotive {
	return p.motive
}

// Direction returns the direction the painting is facing.
func (p *Painting) Direction() cube.Direction {
	return p.direction
}

// Name ...
func (p *Painting) Name() string {
	return "Painting"
}

// EncodeEntity ...
func (p *Painting) EncodeEntity() string {
	return "minecraft:painting"
}

// BBox ...
func (p *Painting) BBox() cube.BBox {
	// TODO: Axis calculations
	x, y := p.motive.Size()
	return cube.Box(0, 0, 0, float64(x), float64(y), 0)
}

// Rotation ...
func (p *Painting) Rotation() (float64, float64) {
	return float64(sliceutil.Index(cube.Directions(), p.direction) * 90), 0
}

// DecodeNBT ...
func (p *Painting) DecodeNBT(data map[string]any) any {
	motive := PaintingMotiveFromString(nbtconv.Map[string](data, "Motive"))
	direction := cube.Directions()[nbtconv.Map[byte](data, "Direction")]
	return NewPainting(motive, direction, nbtconv.MapVec3(data, "Pos"))
}

// EncodeNBT ...
func (p *Painting) EncodeNBT() map[string]any {
	return map[string]any{
		"UniqueID":  -rand.Int63(),
		"Direction": byte(sliceutil.Index(cube.Directions(), p.direction)),
		"Motive":    p.motive.String(),
		"Pos":       nbtconv.Vec3ToFloat32Slice(p.Position()),
	}
}
