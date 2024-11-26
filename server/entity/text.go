package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// NewText creates and returns a new Text entity with the text and position provided.
func NewText(text string, pos mgl64.Vec3) *world.EntityHandle {
	return world.EntitySpawnOpts{Position: pos, NameTag: text}.New(TextType, textConf)
}

var textConf StationaryBehaviourConfig

// TextType is a world.EntityType implementation for Text.
var TextType textType

type textType struct{}

func (t textType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}
func (textType) EncodeEntity() string        { return "dragonfly:text" }
func (textType) BBox(world.Entity) cube.BBox { return cube.BBox{} }
func (textType) NetworkEncodeEntity() string { return "minecraft:falling_block" }

func (textType) DecodeNBT(_ map[string]any, data *world.EntityData) { data.Data = textConf.New() }
func (textType) EncodeNBT(_ *world.EntityData) map[string]any       { return nil }
