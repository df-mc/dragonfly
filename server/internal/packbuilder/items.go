package packbuilder

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"os"
	"path/filepath"
	"strings"
	_ "unsafe" // Imported for compiler directives.
)

// buildItems builds all the item-related files for the resource pack. This includes textures, language
// entries and item atlas.
func buildItems(dir string) (count int, lang []string) {
	if err := os.Mkdir(filepath.Join(dir, "items"), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "textures/items"), os.ModePerm); err != nil {
		panic(err)
	}

	textureData := make(map[string]any)
	for _, item := range world.CustomItems() {
		identifier, _ := item.EncodeItem()
		lang = append(lang, fmt.Sprintf("item.%s.name=%s", identifier, item.Name()))

		name := strings.Split(identifier, ":")[1]
		textureData[name] = map[string]string{"textures": fmt.Sprintf("textures/items/%s.png", name)}

		count++
	}

	buildItemAtlas(dir, map[string]any{
		"resource_pack_name": "vanilla",
		"texture_name":       "atlas.items",
		"texture_data":       textureData,
	})
	return
}

// buildItemAtlas creates the identifier to texture mapping and writes it to the pack.
func buildItemAtlas(dir string, atlas map[string]any) {
	b, err := json.Marshal(atlas)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "textures/item_texture.json"), b, 0666); err != nil {
		panic(err)
	}
}
