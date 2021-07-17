package pack_builder

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	_ "unsafe" // Imported for compiler directives.
)

// buildItems builds all of the item-related files for the resource pack. This includes textures, language
// entries and item atlas.
func buildItems(dir string) (count int, lang []string) {
	if err := os.Mkdir(filepath.Join(dir, "items"), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "textures/items"), os.ModePerm); err != nil {
		panic(err)
	}

	textureData := make(map[string]interface{})
	for _, item := range world_allCustomItems() {
		identifier, _ := item.EncodeItem()
		lang = append(lang, fmt.Sprintf("item.%s.name=%s", identifier, item.Name()))

		name := strings.Split(identifier, ":")[1]
		textureData[name] = map[string]string{"textures": fmt.Sprintf("textures/items/%s.png", name)}

		buildItemTexture(dir, name, item.Texture())
		buildItem(dir, identifier, name, item)

		count++
	}

	buildItemAtlas(dir, map[string]interface{}{
		"resource_pack_name": "vanilla",
		"texture_name":       "atlas.items",
		"texture_data":       textureData,
	})
	return
}

// buildItemTexture creates a PNG file for the item from the provided image and name and writes it to the pack.
func buildItemTexture(dir, name string, img image.Image) {
	texture, err := os.Create(filepath.Join(dir, "textures/items", name+".png"))
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

// buildItem creases an item JSON file for the provided item and its properties and writes it to the pack.
func buildItem(dir, identifier, name string, item world.CustomItem) {
	itemData, err := json.Marshal(map[string]interface{}{
		"format_version": formatVersion,
		"minecraft:item": map[string]interface{}{
			"description": map[string]interface{}{
				"identifier": identifier,
				"category":   item.Category().String(),
			},
			"components": map[string]interface{}{
				"minecraft:icon":           name,
				"minecraft:render_offsets": "tools",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "items", fmt.Sprintf("%s.json", name)), itemData, 0666); err != nil {
		panic(err)
	}
}

// buildItemAtlas creates the identifier to texture mapping and writes it to the pack.
func buildItemAtlas(dir string, atlas map[string]interface{}) {
	b, err := json.Marshal(atlas)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "textures/item_texture.json"), b, 0666); err != nil {
		panic(err)
	}
}

//go:linkname world_allCustomItems github.com/df-mc/dragonfly/server/world.allCustomItems
//noinspection ALL
func world_allCustomItems() []world.CustomItem
