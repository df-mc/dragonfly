package block

import (
	"encoding/json"
	"fmt"
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

// PHP represents the PHP block.
type PHP struct {
	solid

	// Facing ...
	Facing cube.Direction
}

// Name ...
func (p PHP) Name() string {
	return "PHP Elephant"
}

// Rotation ...
func (p PHP) Rotation() cube.Direction {
	return p.Facing
}

// Geometries ...
func (p PHP) Geometries() (geometries customblock.Geometries) {
	b, err := os.ReadFile("php.geo.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &geometries)
	if err != nil {
		panic(err)
	}
	return
}

// Category ...
func (p PHP) Category() category.Category {
	return category.Construction()
}

// Textures ...
func (p PHP) Textures() map[customblock.MaterialTarget]image.Image {
	return map[customblock.MaterialTarget]image.Image{
		customblock.MaterialTargetAll(): p.Texture(),
	}
}

// UseOnBlock ...
func (p PHP) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, p)
	if !used {
		return
	}

	fmt.Println("a")
	p.Facing = cube.South
	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// Texture ...
func (p PHP) Texture() image.Image {
	texture, err := os.OpenFile("php.png", os.O_RDONLY, os.ModePerm)
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

// EncodeItem ...
func (p PHP) EncodeItem() (name string, meta int16) {
	return "dragonfly:php", 0
}

// EncodeBlock ...
func (p PHP) EncodeBlock() (string, map[string]any) {
	return "dragonfly:php", map[string]any{"direction": int32(p.Facing.Face())}
}

// phpHash ...
var phpHash = NextHash()

// Hash ...
func (p PHP) Hash() uint64 {
	return phpHash | uint64(p.Facing)<<8
}
