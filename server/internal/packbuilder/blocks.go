package packbuilder

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	_ "unsafe" // Imported for compiler directives.
)

// buildBlocks builds all the block-related files for the resource pack. This includes textures, geometries, language
// entries and terrain texture atlas.
func buildBlocks(dir string) (count int, lang []string) {
	if err := os.MkdirAll(filepath.Join(dir, "models/blocks"), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "textures/blocks"), os.ModePerm); err != nil {
		panic(err)
	}

	textureData := make(map[string]any)
	for identifier, blk := range world.CustomBlocks() {
		b, ok := blk.(world.CustomBlockBuildable)
		if !ok {
			continue
		}

		name := strings.Split(identifier, ":")[1]
		lang = append(lang, fmt.Sprintf("tile.%s.name=%s", identifier, b.Name()))
		for name, texture := range b.Textures() {
			textureData[name] = map[string]string{"textures": "textures/blocks/" + name}
			buildBlockTexture(dir, name, texture)
		}
		if b.Geometry() != nil {
			if err := os.WriteFile(filepath.Join(dir, "models/blocks", fmt.Sprintf("%s.geo.json", name)), b.Geometry(), 0666); err != nil {
				panic(err)
			}
		}
		count++
	}

	buildBlockAtlas(dir, map[string]any{
		"resource_pack_name": "vanilla",
		"texture_name":       "atlas.terrain",
		"padding":            8,
		"num_mip_levels":     4,
		"texture_data":       textureData,
	})
	return
}

// buildBlockTexture creates a PNG file for the block from the provided image and name and writes it to the pack.
func buildBlockTexture(dir, name string, img image.Image) {
	texture, err := os.Create(filepath.Join(dir, fmt.Sprintf("textures/blocks/%s.png", name)))
	if err != nil {
		panic(err)
	}
	if err := png.Encode(texture, img); err != nil {
		_ = texture.Close()
		panic(err)
	}
	if err := texture.Close(); err != nil {
		panic(err)
	}
}

// buildBlockAtlas creates the identifier to texture mapping and writes it to the pack.
func buildBlockAtlas(dir string, atlas map[string]any) {
	b, err := json.Marshal(atlas)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "textures/terrain_texture.json"), b, 0666); err != nil {
		panic(err)
	}
}
