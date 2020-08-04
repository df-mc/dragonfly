package pack_builder

import (
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/sger/go-hashdir"
	"io/ioutil"
	"os"
)

func BuildResourcePack() *resource.Pack {
	dir, err := ioutil.TempDir("", "dragonfly_resource_pack-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	var assets int
	var lang []string

	itemCount, itemLang := buildItems(dir)
	assets += itemCount
	lang = append(lang, itemLang...)

	if assets > 0 {
		buildLanguageFile(dir, lang)
		hash, err := hashdir.Create(dir, "md5")
		if err != nil {
			panic(err)
		}
		var data [16]byte
		copy(data[:], hash)
		buildManifest(dir, data)
		return resource.MustCompile(dir)
	}
	return nil
}
