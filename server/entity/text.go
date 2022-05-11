package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/go-gl/mathgl/mgl64"
)

// Text is an entity that only displays floating text. The entity is otherwise invisible and cannot be moved.
type Text struct {
	transform
	text string
}

// NewText creates and returns a new Text entity with the text and position provided.
func NewText(text string, pos mgl64.Vec3) *Text {
	t := &Text{text: text}
	t.transform = newTransform(t, pos)
	return t
}

// Name returns the name of the text entity, including the text written on it.
func (t *Text) Name() string {
	return "Text('" + t.text + "')"
}

// EncodeEntity returns the ID for falling blocks.
func (t *Text) EncodeEntity() string {
	return "dragonfly:text"
}

// BBox returns an empty physics.BBox so that players cannot interact with the entity.
func (t *Text) BBox() cube.BBox {
	return cube.BBox{}
}

// Immobile always returns true.
func (t *Text) Immobile() bool {
	return true
}

// SetText updates the designated text of the entity.
func (t *Text) SetText(text string) {
	t.mu.Lock()
	t.text = text
	t.mu.Unlock()

	for _, v := range t.World().Viewers(t.Position()) {
		v.ViewEntityState(t)
	}
}

// Text returns the designated text of the entity.
func (t *Text) Text() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.text
}

// NameTag returns the designated text of the entity. It is an alias for the Text function.
func (t *Text) NameTag() string {
	return t.Text()
}

// DecodeNBT decodes the data passed to create and return a new Text entity.
func (t *Text) DecodeNBT(data map[string]any) any {
	return NewText(nbtconv.Map[string](data, "Text"), nbtconv.MapVec3(data, "Pos"))
}

// EncodeNBT encodes the Text entity to a map representation that can be encoded to NBT.
func (t *Text) EncodeNBT() map[string]any {
	return map[string]any{
		"Pos":  nbtconv.Vec3ToFloat32Slice(t.Position()),
		"Text": t.text,
	}
}
