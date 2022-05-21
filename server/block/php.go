package block

import (
	"encoding/json"
	"github.com/df-mc/dragonfly/server/block/customblock"
	"github.com/df-mc/dragonfly/server/item/category"
	"image"
	"image/png"
	"math"
	"os"
)

// PHP represents the PHP block.
type PHP struct {
	solid
}

// Name ...
func (p PHP) Name() string {
	return "PHP Elephant"
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
func (p PHP) Textures() map[string]image.Image {
	return map[string]image.Image{
		customblock.MaterialTargetAll: p.Texture(),
	}
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
	return "dragonfly:php", nil
}

// Hash ...
func (p PHP) Hash() uint64 {
	return math.MaxUint64 // TODO
}
