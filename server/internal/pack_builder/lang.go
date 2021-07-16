package pack_builder

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func buildLanguageFile(dir string, lang []string) {
	if err := os.Mkdir(filepath.Join(dir, "texts"), os.ModePerm); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "texts/en_US.lang"), []byte(strings.Join(lang, "\n")), 0666); err != nil {
		panic(err)
	}
}
