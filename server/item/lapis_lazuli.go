package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// LapisLazuli is a mineral used for enchanting and decoration.
type LapisLazuli struct{}

func (LapisLazuli) EncodeItem() (name string, meta int16) {
	return "minecraft:lapis_lazuli", 0
}

func (LapisLazuli) TrimMaterial() string {
	return "lapis"
}

func (LapisLazuli) MaterialColour() string {
	return text.Lapis
}
