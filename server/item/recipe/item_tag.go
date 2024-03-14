package recipe

import (
	_ "embed"
	"encoding/json"
)

var (
	//go:embed item_tags.json
	itemTagData []byte
	itemTags    = make(map[string][]string)
)

func init() {
	if err := json.Unmarshal(itemTagData, &itemTags); err != nil {
		panic(err)
	}
}

// ItemTag represents a recipe item that is identified by a tag, such as "minecraft:planks" or
// "minecraft:digger" and so on.
type ItemTag struct {
	tag   string
	count int

	items []string
}

// NewItemTag creates a new item tag with the tag and count passed.
func NewItemTag(tag string, count int) ItemTag {
	if count < 0 {
		count = 0
	}
	return ItemTag{tag: tag, count: count, items: itemTags[tag]}
}

// Count ...
func (i ItemTag) Count() int {
	return i.count
}

// Empty ...
func (i ItemTag) Empty() bool {
	return i.count == 0 || i.tag == ""
}

// Tag returns the tag of the item.
func (i ItemTag) Tag() string {
	return i.tag
}

// Contains returns true if the item tag contains the item with the name passed.
func (i ItemTag) Contains(name string) bool {
	for _, item := range i.items {
		if item == name {
			return true
		}
	}
	return false
}
