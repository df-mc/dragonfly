package item

import (
	"image"
	"image/png"
	"os"
)

type Strawberry struct{}

func (Strawberry) EncodeItem() (id int32, meta int16) {
	return 1000, 0
}

func (Strawberry) Name() string {
	return "Strawberry"
}

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

func (Strawberry) Category() string {
	return "Nature"
}
