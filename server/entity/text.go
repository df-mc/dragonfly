package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// NewText creates and returns a new Text entity with the text and position provided.
func NewText(text string, pos mgl64.Vec3) *world.EntityHandle {
	return world.EntitySpawnOpts{Position: pos, NameTag: text}.New(TextType{}, textConf)
}

var textConf StationaryBehaviourConfig

// TextType is a world.EntityType implementation for Text.
type TextType struct{}

func (t TextType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}
func (TextType) EncodeEntity() string        { return "dragonfly:text" }
func (TextType) BBox(world.Entity) cube.BBox { return cube.BBox{} }
func (TextType) NetworkEncodeEntity() string { return "minecraft:falling_block" }

func (TextType) DecodeNBT(_ map[string]any, data *world.EntityData) { data.Data = textConf.New() }
func (TextType) EncodeNBT(data *world.EntityData) map[string]any    { return nil }
