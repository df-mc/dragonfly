package lang

import (
	"embed"
	"encoding/json"
	"github.com/df-mc/dragonfly/server/world"
	"golang.org/x/text/language"
)

// itemHash is a combination of an item's name and metadata. It is used as a key in hash maps.
type itemHash struct {
	name string
	meta int16
}

var (
	//go:embed names/*
	namesFS embed.FS
	// names is a mapping from language.Tag to an item->display name mapping.
	names = make(map[language.Tag]map[itemHash]string)
)

// DisplayName returns the display name of the item as shown in game in the language passed.
func DisplayName(item world.Item, locale language.Tag) (string, bool) {
	id, meta := item.EncodeItem()
	h := itemHash{name: id, meta: meta}

	if _, ok := names[locale]; !ok && !load(locale) {
		// Language not supported, default to American English.
		return names[language.AmericanEnglish][h], false
	}

	name, ok := names[locale][h]
	return name, ok
}

// load loads the locale for the item display names.
func load(locale language.Tag) bool {
	b, err := namesFS.ReadFile("names/" + locale.String() + ".json")
	if err != nil {
		return false
	}

	var entries []struct {
		ID   string `json:"id"`
		Meta int16  `json:"meta,omitempty"`
		Name string `json:"name"`
	}
	err = json.Unmarshal(b, &entries)
	if err != nil {
		panic(err)
	}

	names[locale] = make(map[itemHash]string, len(entries))
	for _, entry := range entries {
		h := itemHash{name: entry.ID, meta: entry.Meta}
		names[locale][h] = entry.Name
	}
	return true
}
