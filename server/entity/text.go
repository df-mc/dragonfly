package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
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

// Type returns TextType.
func (*Text) Type() world.EntityType {
	return TextType{}
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

// TextType is a world.EntityType implementation for Text.
type TextType struct{}

func (TextType) EncodeEntity() string        { return "dragonfly:text" }
func (TextType) BBox(world.Entity) cube.BBox { return cube.BBox{} }
func (TextType) NetworkEncodeEntity() string { return "minecraft:falling_block" }

func (TextType) DecodeNBT(m map[string]any) world.Entity {
	return NewText(nbtconv.String(m, "Text"), nbtconv.Vec3(m, "Pos"))
}

func (TextType) EncodeNBT(e world.Entity) map[string]any {
	t := e.(*Text)
	return map[string]any{
		"Pos":  nbtconv.Vec3ToFloat32Slice(t.Position()),
		"Text": t.text,
	}
}
