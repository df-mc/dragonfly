package recipe

import (
	"slices"
	"sort"
	"strings"
	"unsafe"

	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// recipes is a list of each recipe.
var (
	recipes []Recipe
	// index maps an input hash to output stacks for each PotionContainerChange and Potion recipe.
	index = make(map[string]map[string]Recipe)
	// reagent maps the item name and an item.Stack.
	reagent = make(map[string]item.Stack)
)

// Recipes returns each recipe in a slice.
func Recipes() []Recipe {
	return slices.Clone(recipes)
}

// Register registers a new recipe.
func Register(recipe Recipe) {
	recipes = append(recipes, recipe)

	_, ok := recipe.(PotionContainerChange)
	p, okTwo := recipe.(Potion)

	if okTwo {
		stack := p.Input()[1].(item.Stack)
		name, _ := stack.Item().EncodeItem()
		reagent[name] = stack
	}

	if ok || okTwo {
		input := make([]world.Item, len(recipe.Input()))
		for i, stack := range recipe.Input() {
			if s, ok := stack.(item.Stack); ok {
				input[i] = s.Item()
			}
		}
		hash := hashItems(input, !ok)

		block := recipe.Block()
		if index[block] == nil {
			index[block] = make(map[string]Recipe)
		}
		index[block][hash] = recipe
	}
}

// Perform performs the recipe with the given block and inputs and returns the outputs. If the inputs do not map to
// any outputs, false is returned for the second return value.
func Perform(block string, input ...world.Item) (output []world.ItemStack, ok bool) {
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

// ValidBrewingReagent checks if the world.Item is a brewing reagent.
func ValidBrewingReagent(i world.Item) bool {
	name, _ := i.EncodeItem()
	_, exists := reagent[name]
	return exists
}
