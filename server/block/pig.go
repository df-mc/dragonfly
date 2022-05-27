package block

import (
	"encoding/json"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/customblock"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/category"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"image"
	"image/png"
	"os"
)

// Pig represents the Pig block.
type Pig struct {
	empty
	transparent

	// Facing ...
	Facing cube.Direction
}

// Name ...
func (p Pig) Name() string {
	return "Pig Head"
}

// FlammabilityInfo ...
func (p Pig) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(5, 20, true)
}

// Geometries ...
func (p Pig) Geometries() (customblock.Geometries, bool) {
	b, err := os.ReadFile("skull.geo.json")
	if err != nil {
		panic(err)
	}
	var geometry customblock.Geometries
	err = json.Unmarshal(b, &geometry)
	if err != nil {
		panic(err)
	}
	return geometry, true
}

// Category ...
func (p Pig) Category() category.Category {
	return category.Nature()
}

// Textures ...
func (p Pig) Textures() (map[customblock.MaterialTarget]image.Image, customblock.RenderMethod) {
	return map[customblock.MaterialTarget]image.Image{
		customblock.MaterialTargetAll(): p.Texture(),
	}, customblock.AlphaTestRenderMethod()
}

// UseOnBlock ...
func (p Pig) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, p)
	if !used {
		return
	}

	p.Facing = user.Facing().Opposite()
	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// Texture ...
func (p Pig) Texture() image.Image {
	texture, err := os.OpenFile("pig.png", os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer texture.Close()
	img, err := png.Decode(texture)
	if err != nil {
		panic(err)
	}
	return img
}

// pigHash is the unique hash used for Pig blocks.
var pigHash = NextHash()

// Hash ...
func (p Pig) Hash() uint64 {
	return pigHash | uint64(p.Facing)<<8
}

// EncodeItem ...
func (p Pig) EncodeItem() (name string, meta int16) {
	return "dragonfly:pig", 0
}

// EncodeBlock ...
func (p Pig) EncodeBlock() (string, map[string]any) {
	return "dragonfly:pig", map[string]any{"direction": int32(p.Facing)}
}

// Rotation ...
func (p Pig) Rotation() (mgl64.Vec3, bool, map[string]mgl64.Vec3) {
	return mgl64.Vec3{}, false, map[string]mgl64.Vec3{
		"query.block_property('direction') == 0": {0, 180, 0},
		"query.block_property('direction') == 2": {0, 270, 0},
		"query.block_property('direction') == 3": {0, 90, 0},
	}
}
