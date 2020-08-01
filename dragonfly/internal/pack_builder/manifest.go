package pack_builder

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"io/ioutil"
	"path/filepath"
)

func buildManifest(dir string) {
	m, err := json.Marshal(resource.Manifest{
		FormatVersion: 2,
		Header: resource.Header{
			Name:               "dragonfly auto-generated resource pack",
			Description:        "This resource pack contains auto-generated content from dragonfly",
			UUID:               uuid.New().String(),
			Version:            [3]int{0, 0, 1},
			MinimumGameVersion: [3]int{1, 16, 0},
		},
		Modules: []resource.Module{
			{
				UUID:        uuid.New().String(),
				Description: "This resource pack contains auto-generated content from dragonfly",
				Type:        "resources",
				Version:     [3]int{0, 0, 1},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "manifest.json"), m, 0666); err != nil {
		panic(err)
	}
}
