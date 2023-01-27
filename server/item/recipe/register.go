package recipe

import (
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"golang.org/x/exp/slices"
	"sort"
	"strings"
	"unsafe"
)

var (
	// recipes is a list of each recipe.
	recipes []Recipe
	// index maps an input hash to output stacks for each recipe.
	index = make(map[string]map[string]Recipe)
)

// Recipes returns each recipe in a slice.
func Recipes() []Recipe {
	return slices.Clone(recipes)
}

// Register registers a new recipe.
func Register(r Recipe) {
	recipes = append(recipes, r)

	_, containerChange := r.(PotionContainerChange)
	hash := hashItems(sliceutil.Map(r.Input(), func(st item.Stack) world.Item {
		return st.Item()
	}), !containerChange)

	block := r.Block()
	if index[block] == nil {
		index[block] = make(map[string]Recipe)
	}
	index[block][hash] = r
}

// Perform performs the recipe with the given block and inputs and returns the outputs. If the inputs do not map to
// any outputs, false is returned for the second return value.
func Perform(block string, input ...world.Item) (output []item.Stack, ok bool) {
	blockInd, ok := index[block]
	if !ok {
		// Block specific index didn't exist.
		return nil, false
	}
	r, ok := blockInd[hashItems(input, true)]
	if !ok {
		r, ok = blockInd[hashItems(input, false)]
		if !ok {
			return nil, false
		}
	}
	_, containerChange := r.(PotionContainerChange)
	for ind, it := range r.Output() {
		if containerChange {
			name, _ := it.Item().EncodeItem()
			_, meta := input[ind].EncodeItem()
			if i, ok := world.ItemByName(name, meta); ok {
				it = item.NewStack(i, it.Count())
			}
		}
		output = append(output, it)
	}
	return output, ok
}

// hashItems hashes the given list of item types and returns it.
func hashItems(items []world.Item, useMeta bool) string {
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
		b.WriteString(name)
		if useMeta {
			a := *(*[2]byte)(unsafe.Pointer(&meta))
			b.Write(a[:])
		}
	}
	return b.String()
}
