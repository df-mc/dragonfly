package lang

import (
	"embed"
	"encoding/json"
	"github.com/df-mc/dragonfly/server/world"
	"golang.org/x/text/language"
	"strings"
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
	if _, ok := names[locale]; !ok {
		// Language not supported.
		return "", false
	}

	id, meta := item.EncodeItem()
	h := itemHash{name: id, meta: meta}

	name, ok := names[locale][h]
	return name, ok
}

// init ...
func init() {
	localizations, err := namesFS.ReadDir("names")
	if err != nil {
		panic(err)
	}
	for _, locale := range localizations {
		if locale.IsDir() {
			continue
		}
		tag, err := language.Parse(strings.Replace(strings.TrimSuffix(locale.Name(), ".json"), "_", "-", 1))
		if err != nil {
			panic(err)
		}
		b, err := namesFS.ReadFile("names/" + locale.Name())
		if err != nil {
			panic(err)
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

		names[tag] = make(map[itemHash]string, len(entries))
		for _, entry := range entries {
			h := itemHash{name: entry.ID, meta: entry.Meta}
			names[tag][h] = entry.Name
		}
	}
}
