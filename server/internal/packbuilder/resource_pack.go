package packbuilder

import (
	_ "embed"
	"github.com/rogpeppe/go-internal/dirhash"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"os"
)

//go:embed pack_icon.png
var packIcon []byte

// BuildResourcePack builds a resource pack based on custom features that have been registered to the server.
// It creates a UUID based on the hash of the directory so the client will only be prompted to download it
// once it is changed.
func BuildResourcePack() (*resource.Pack, bool) {
	dir, err := os.MkdirTemp("", "dragonfly_resource_pack-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	var assets int
	var lang []string

	itemCount, itemLang := buildItems(dir)
	assets += itemCount
	lang = append(lang, itemLang...)

	blockCount, blockLang := buildBlocks(dir)
	assets += blockCount
	lang = append(lang, blockLang...)

	if assets > 0 {
		buildLanguageFile(dir, lang)
		if err := os.WriteFile(dir+"/pack_icon.png", packIcon, 0666); err != nil {
			panic(err)
		}
		hash, err := dirhash.HashDir(dir, "", dirhash.Hash1)
		if err != nil {
			panic(err)
		}
		var header, module [16]byte
		copy(header[:], hash)
		copy(module[:], hash[16:])
		buildManifest(dir, header, module)
		return resource.MustReadPath(dir), true
	}
	return nil, false
}
