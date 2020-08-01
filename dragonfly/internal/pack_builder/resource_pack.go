package pack_builder

import (
	"github.com/sandertv/gophertunnel/minecraft/resource"
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
		buildManifest(dir)
		buildLanguageFile(dir, lang)
		return resource.MustCompile(dir)
	}
	return nil
}
