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

// Category ...
func (p Pig) Category() category.Category {
	return category.Nature()
}

// FlammabilityInfo ...
func (p Pig) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(5, 20, true)
}

// BreakInfo ...
func (p Pig) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, nothingEffective, oneOf(Pig{}))
}

// Textures ...
func (p Pig) Textures() (map[customblock.Target]image.Image, map[string]map[customblock.Target]image.Image, customblock.Method) {
	return map[customblock.Target]image.Image{
			customblock.MaterialTargetAll(): p.Texture(),
		}, map[string]map[customblock.Target]image.Image{
			"query.block_property('direction') == 0": {
				customblock.MaterialTargetAll(): p.Texture(),
			},
		}, customblock.AlphaTestRenderMethod()
}

// UseOnBlock ...
func (p Pig) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, p)
	if !used {
		return
	}

	p.Facing = user.Facing()
	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// Geometries ...
func (p Pig) Geometries() (customblock.Geometry, bool) {
	b, err := os.ReadFile("skull.geo.json")
	if err != nil {
		panic(err)
	}
	var geometry customblock.Geometries
	err = json.Unmarshal(b, &geometry)
	if err != nil {
		panic(err)
	}
	return geometry.Geometry[0], true
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

// Rotation ...
func (p Pig) Rotation() (mgl64.Vec3, bool, map[string]mgl64.Vec3) {
	return mgl64.Vec3{}, false, map[string]mgl64.Vec3{
		"query.block_property('direction') == 1": {0, 180, 0},
		"query.block_property('direction') == 2": {0, 90, 0},
		"query.block_property('direction') == 3": {0, 270, 0},
	}
}

// EncodeItem ...
func (p Pig) EncodeItem() (name string, meta int16) {
	return "dragonfly:pig", 0
}

// EncodeBlock ...
func (p Pig) EncodeBlock() (string, map[string]any) {
	return "dragonfly:pig", map[string]any{"direction": int32(p.Facing)}
}

// pigHash ...
var pigHash = NextHash()

// Hash ...
func (p Pig) Hash() uint64 {
	return pigHash | uint64(p.Facing)<<8
}
