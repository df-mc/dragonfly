package entity

import (
	"github.com/df-mc/dragonfly/server/entity/physics"
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

// AABB returns an empty physics.AABB so that players cannot interact with the entity.
func (t *Text) AABB() physics.AABB {
	return physics.AABB{}
}

// Immobile always returns true.
func (t *Text) Immobile() bool {
	return true
}

// NameTag returns the text passed to NewText.
func (t *Text) NameTag() string {
	return t.text
}

// DecodeNBT decodes the data passed to create and return a new Text entity.
func (t *Text) DecodeNBT(data map[string]interface{}) interface{} {
	return NewText(nbtconv.MapString(data, "Text"), nbtconv.MapVec3(data, "Pos"))
}

// EncodeNBT encodes the Text entity to a map representation that can be encoded to NBT.
func (t *Text) EncodeNBT() map[string]interface{} {
	pos := t.Position()
	return map[string]interface{}{
		"Pos":  pos[:],
		"Text": t.text,
	}
}
