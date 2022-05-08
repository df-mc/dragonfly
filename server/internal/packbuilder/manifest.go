package packbuilder

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

// buildManifest creates a JSON manifest file for the client to be able to read the resource pack. It creates
// basic information and writes it to the pack.
func buildManifest(dir string, headerUUID, moduleUUID uuid.UUID) {
	m, err := json.Marshal(resource.Manifest{
		FormatVersion: 2,
		Header: resource.Header{
			Name:               "dragonfly auto-generated resource pack",
			Description:        "This resource pack contains auto-generated content from dragonfly",
			UUID:               headerUUID.String(),
			Version:            [3]int{0, 0, 1},
			MinimumGameVersion: parseVersion(protocol.CurrentVersion),
		},
		Modules: []resource.Module{
			{
				UUID:        moduleUUID.String(),
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

// parseVersion parses the version passed in the format of a.b.c as a [3]int.
func parseVersion(ver string) [3]int {
	frag := strings.Split(ver, ".")
	if len(frag) != 3 {
		panic("invalid version number " + ver)
	}
	a, _ := strconv.ParseInt(frag[0], 10, 64)
	b, _ := strconv.ParseInt(frag[1], 10, 64)
	c, _ := strconv.ParseInt(frag[2], 10, 64)
	return [3]int{int(a), int(b), int(c)}
}
