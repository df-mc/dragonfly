package packbuilder

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// buildLanguageFile creates a lang file and writes all of the language entries to the pack.
func buildLanguageFile(dir string, lang []string) {
	if err := os.Mkdir(filepath.Join(dir, "texts"), os.ModePerm); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "texts/en_US.lang"), []byte(strings.Join(lang, "\n")), 0666); err != nil {
		panic(err)
	}
}
