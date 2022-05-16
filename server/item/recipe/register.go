package recipe

import (
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"golang.org/x/exp/slices"
	"sort"
	"strconv"
	"strings"
)

var (
	// recipes is a list of each recipe.
	recipes []Recipe
	// index maps an input hash to output stacks for each block.
	index = make(map[world.Block]map[string][]item.Stack)
)

// Recipes returns each recipe in a slice.
func Recipes() []Recipe {
	return slices.Clone(recipes)
}

// Register registers a new recipe.
func Register(recipe Recipe) {
	recipes = append(recipes, recipe)
	block, hash := recipe.Block(), hashItems(sliceutil.Map(recipe.Input(), func(st item.Stack) world.Item {
		return st.Item()
	}))
	if index[block] == nil {
		index[block] = make(map[string][]item.Stack)
	}
	index[block][hash] = recipe.Output()
}

// Perform performs the recipe with the given block and inputs and returns the outputs. If the inputs do not map to
// any outputs, false is returned for the second return value.
func Perform(block world.Block, inputs []world.Item) ([]item.Stack, bool) {
	blockInd, ok := index[block]
	if !ok {
		// Block specific index didn't exist.
		return nil, false
	}
	outputs, ok := blockInd[hashItems(inputs)]
	return outputs, ok
}

// hashItems hashes the given list of item types and returns it.
func hashItems(items []world.Item) string {
	items = sliceutil.Filter(items, func(it world.Item) bool {
		return it != nil
	})
	sort.Slice(items, func(i, j int) bool {
		nameOne, metaOne := items[i].EncodeItem()
		nameTwo, metaTwo := items[j].EncodeItem()
		if nameOne == nameTwo {
			return metaOne < metaTwo
		}
		return nameOne < nameTwo
	})

	var b strings.Builder
	for _, it := range items {
		name, meta := it.EncodeItem()
		b.WriteString(name + strconv.Itoa(int(meta)))
	}
	return b.String()
}
