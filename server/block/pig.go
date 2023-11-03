package block

import (
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
func (p Pig) Textures() map[string]image.Image {
	return map[string]image.Image{
		"pig": p.Texture(),
	}
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

// Geometry ...
func (p Pig) Geometry() []byte {
	data, err := os.ReadFile("skull.geo.json")
	if err != nil {
		panic(err)
	}
	return data
}

// UseOnBlock ...
func (p Pig) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, p)
	if !used {
		return
	}

	p.Facing = user.Rotation().Direction()
	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (p Pig) EncodeItem() (name string, meta int16) {
	return "dragonfly:pig", 0
}

// EncodeBlock ...
func (p Pig) EncodeBlock() (string, map[string]any) {
	return "dragonfly:pig", map[string]any{"rotation": int32(p.Facing)}
}

// pigHash ...
var pigHash = NextHash()

// Hash ...
func (p Pig) Hash() uint64 {
	return pigHash | (uint64(p.Facing) << 8)
}

func (p Pig) Properties() customblock.Properties {
	return customblock.Properties{
		CollisionBox: cube.Box(0.25, 0, 0.25, 0.75, 0.5, 0.75),
		SelectionBox: cube.Box(0.25, 0, 0.25, 0.75, 0.5, 0.75),
		Geometry:     "geometry.skull",
		Textures: map[string]customblock.Material{
			"*": customblock.NewMaterial("pig", customblock.OpaqueRenderMethod()),
		},
	}
}

func (p Pig) States() map[string][]any {
	return map[string][]any{
		"rotation": {int32(0), int32(1), int32(2), int32(3)},
	}
}

func (p Pig) Permutations() []customblock.Permutation {
	return []customblock.Permutation{
		{
			Condition: "query.block_state('rotation') == 1",
			Properties: customblock.Properties{
				Rotation: cube.Pos{0, 3, 0},
			},
		},
		{
			Condition: "query.block_state('rotation') == 2",
			Properties: customblock.Properties{
				Rotation: cube.Pos{0, 2, 0},
			},
		},
		{
			Condition: "query.block_state('rotation') == 3",
			Properties: customblock.Properties{
				Rotation: cube.Pos{0, 1, 0},
			},
		},
	}
}
