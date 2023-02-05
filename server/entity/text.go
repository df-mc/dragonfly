package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// NewText creates and returns a new Text entity with the text and position provided.
func NewText(text string, pos mgl64.Vec3) *Ent {
	e := Config{Behaviour: textConf.New()}.New(TextType{}, pos)
	e.SetNameTag(text)
	return e
}

var textConf = StationaryBehaviourConfig{}

// TextType is a world.EntityType implementation for Text.
type TextType struct{}

func (TextType) EncodeEntity() string        { return "dragonfly:text" }
func (TextType) BBox(world.Entity) cube.BBox { return cube.BBox{} }
func (TextType) NetworkEncodeEntity() string { return "minecraft:falling_block" }

func (TextType) DecodeNBT(m map[string]any) world.Entity {
	return NewText(nbtconv.String(m, "Text"), nbtconv.Vec3(m, "Pos"))
}

func (TextType) EncodeNBT(e world.Entity) map[string]any {
	t := e.(*Ent)
	return map[string]any{
		"Pos":  nbtconv.Vec3ToFloat32Slice(t.Position()),
		"Text": t.NameTag(),
	}
}
