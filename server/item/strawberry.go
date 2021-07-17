package item

import (
	"github.com/df-mc/dragonfly/server/item/category"
	"image"
	"image/png"
	"os"
)

// Strawberry is a custom item used only to test the custom item functionality. This will not remain as a
// dragonfly feature in the future.
type Strawberry struct{}

// EncodeItem ...
func (Strawberry) EncodeItem() (name string, meta int16) {
	return "dragonfly:strawberry", 0
}

// Name ...
func (Strawberry) Name() string {
	return "Strawberry"
}

// Texture ...
func (Strawberry) Texture() image.Image {
	texture, err := os.OpenFile("./resources/strawberry.png", os.O_RDONLY, os.ModePerm)
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

// Category ...
func (Strawberry) Category() category.Category {
	return category.Nature()
}
