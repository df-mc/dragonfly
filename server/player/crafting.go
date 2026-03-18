package player

import (
	"fmt"
	"math"
	"slices"

	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
)

// craftInventorySource identifies an inventory consulted while building an auto-crafting plan.
type craftInventorySource struct {
	// inventory is the source inventory searched for matching crafting ingredients.
	inventory *inventory.Inventory
}

// CraftItem calculates a crafting plan for a recipe crafted directly from the player's current crafting grid.
func (p *Player) CraftItem(craft recipe.Recipe, times int) (recipe.CraftingPlan, error) {
	if err := validateCraftingRecipe(craft); err != nil {
		return recipe.CraftingPlan{}, err
	}
	if times < 1 {
		return recipe.CraftingPlan{}, fmt.Errorf("times crafted must be at least 1")
	}

	size := int(p.session().CraftingGridSize())
	offset := int(p.session().CraftingGridOffset())
	consumed := make([]bool, size)
	plan := recipe.CraftingPlan{
		Inputs:  make([]item.Stack, 0, len(craft.Input())),
		Results: repeatCraftStacks(craft.Output(), times),
		Changes: make([]recipe.CraftingSlotChange, 0, len(craft.Input())),
	}

	for _, expected := range craft.Input() {
		var processed bool
		for slot := offset; slot < offset+size; slot++ {
			if consumed[slot-offset] {
				continue
			}
			has, _ := p.ui.Item(slot)
			if has.Empty() != expected.Empty() || has.Count() < expected.Count()*times {
				continue
			}
			if !matchingCraftItems(has, expected) {
				continue
			}

			processed, consumed[slot-offset] = true, true
			if removal := expected.Count() * times; removal > 0 {
				plan.Inputs = append(plan.Inputs, has.Grow(removal-has.Count()))
				plan.Changes = append(plan.Changes, recipe.CraftingSlotChange{
					Inventory: p.ui,
					Slot:      slot,
					Stack:     has.Grow(-removal),
				})
			}
			break
		}
		if !processed {
			return recipe.CraftingPlan{}, fmt.Errorf("recipe could not consume expected item: %v", expected)
		}
	}
	return p.approveCraftingPlan(plan)
}

// AutoCraftItem calculates a crafting plan for a recipe crafted from the player's crafting grid and inventory.
func (p *Player) AutoCraftItem(craft recipe.Recipe, times int) (recipe.CraftingPlan, error) {
	if err := validateCraftingRecipe(craft); err != nil {
		return recipe.CraftingPlan{}, err
	}
	if times < 1 {
		return recipe.CraftingPlan{}, fmt.Errorf("times crafted must be at least 1")
	}

	flattenedInputs := make([]recipe.Item, 0, len(craft.Input()))
	for _, it := range craft.Input() {
		if it.Empty() {
			continue
		}
		if index := slices.IndexFunc(flattenedInputs, func(other recipe.Item) bool {
			return matchingCraftItems(other, it)
		}); index >= 0 {
			flattenedInputs[index] = growRecipeItem(it, flattenedInputs[index].Count())
			continue
		}
		flattenedInputs = append(flattenedInputs, it)
	}

	plan := recipe.CraftingPlan{
		Inputs:  make([]item.Stack, 0, len(flattenedInputs)),
		Results: repeatCraftStacks(craft.Output(), times),
		Changes: make([]recipe.CraftingSlotChange, 0, len(flattenedInputs)),
	}
	sources := []craftInventorySource{
		{inventory: p.ui},
		{inventory: p.inv},
	}

	for _, expected := range flattenedInputs {
		remaining := expected.Count() * times

		for _, source := range sources {
			for slot, has := range source.inventory.Slots() {
				if has.Empty() || !matchingCraftItems(has, expected) {
					continue
				}

				removal := min(remaining, has.Count())
				remaining -= removal

				plan.Inputs = append(plan.Inputs, has.Grow(removal-has.Count()))
				plan.Changes = append(plan.Changes, recipe.CraftingSlotChange{
					Inventory: source.inventory,
					Slot:      slot,
					Stack:     has.Grow(-removal),
				})
				if remaining == 0 {
					break
				}
			}
			if remaining == 0 {
				break
			}
		}
		if remaining != 0 {
			return recipe.CraftingPlan{}, fmt.Errorf("recipe could not consume expected item: %v", expected)
		}
	}
	return p.approveCraftingPlan(plan)
}

// DynamicCraftItem calculates a crafting plan for the first matching server-side dynamic recipe in the player's grid.
func (p *Player) DynamicCraftItem(times int) (recipe.CraftingPlan, error) {
	if times < 1 {
		return recipe.CraftingPlan{}, fmt.Errorf("times crafted must be at least 1")
	}

	size := int(p.session().CraftingGridSize())
	offset := int(p.session().CraftingGridOffset())
	input := make([]recipe.Item, size)
	for i := 0; i < size; i++ {
		slot := offset + i
		stack, _ := p.ui.Item(slot)
		if stack.Empty() {
			input[i] = item.Stack{}
			continue
		}
		input[i] = stack
	}

	for _, dynamicRecipe := range recipe.DynamicRecipes() {
		if dynamicRecipe.Block() != "crafting_table" {
			continue
		}

		output, ok := dynamicRecipe.Match(input)
		if !ok {
			continue
		}

		minStackCount := math.MaxInt
		for i := 0; i < size; i++ {
			slot := offset + i
			stack, _ := p.ui.Item(slot)
			if !stack.Empty() && stack.Count() < minStackCount {
				minStackCount = stack.Count()
			}
		}
		if minStackCount < times {
			times = minStackCount
		}

		plan := recipe.CraftingPlan{
			Inputs:  make([]item.Stack, 0, size),
			Results: repeatCraftStacks(output, times),
			Changes: make([]recipe.CraftingSlotChange, 0, size),
		}
		for i := 0; i < size; i++ {
			slot := offset + i
			stack, _ := p.ui.Item(slot)
			if stack.Empty() {
				continue
			}
			plan.Inputs = append(plan.Inputs, stack.Grow(times-stack.Count()))
			plan.Changes = append(plan.Changes, recipe.CraftingSlotChange{
				Inventory: p.ui,
				Slot:      slot,
				Stack:     stack.Grow(-times),
			})
		}
		return p.approveCraftingPlan(plan)
	}

	return recipe.CraftingPlan{}, fmt.Errorf("no matching recipe found for crafting grid")
}

// approveCraftingPlan runs the player's craft handlers for a plan and returns the plan if it is allowed.
func (p *Player) approveCraftingPlan(plan recipe.CraftingPlan) (recipe.CraftingPlan, error) {
	ctx := event.C(p)
	if p.Handler().HandleCraftItem(ctx, slices.Clone(plan.Inputs), slices.Clone(plan.Results)); ctx.Cancelled() {
		return recipe.CraftingPlan{}, fmt.Errorf("craft item was cancelled")
	}
	return plan, nil
}

// validateCraftingRecipe validates that the recipe is a normal crafting-table recipe supported by the player grid.
func validateCraftingRecipe(craft recipe.Recipe) error {
	_, shaped := craft.(recipe.Shaped)
	_, shapeless := craft.(recipe.Shapeless)
	if !shaped && !shapeless {
		return fmt.Errorf("recipe is not a shaped or shapeless recipe")
	}
	if craft.Block() != "crafting_table" {
		return fmt.Errorf("recipe is not a crafting table recipe")
	}
	return nil
}

// matchingCraftItems reports whether the two recipe items represent the same crafting ingredient.
func matchingCraftItems(has, expected recipe.Item) bool {
	switch expected := expected.(type) {
	case item.Stack:
		switch has := has.(type) {
		case recipe.ItemTag:
			name, _ := expected.Item().EncodeItem()
			return has.Contains(name)
		case item.Stack:
			_, variants := expected.Value("variants")
			if !variants {
				return has.Comparable(expected)
			}
			nameOne, _ := has.Item().EncodeItem()
			nameTwo, _ := expected.Item().EncodeItem()
			return nameOne == nameTwo
		}
		panic(fmt.Errorf("client has unexpected recipe item %T", has))
	case recipe.ItemTag:
		switch has := has.(type) {
		case item.Stack:
			name, _ := has.Item().EncodeItem()
			return expected.Contains(name)
		case recipe.ItemTag:
			return has.Tag() == expected.Tag()
		}
		panic(fmt.Errorf("client has unexpected recipe item %T", has))
	}
	panic(fmt.Errorf("tried to match with unexpected recipe item %T", expected))
}

// repeatCraftStacks multiplies output stacks by the repetition count and splits them to respect max stack size.
func repeatCraftStacks(items []item.Stack, repetitions int) []item.Stack {
	output := make([]item.Stack, 0, len(items))
	for _, stack := range items {
		count, maxCount := stack.Count(), stack.MaxCount()
		total := count * repetitions

		stacks := int(math.Ceil(float64(total) / float64(maxCount)))
		for i := 0; i < stacks; i++ {
			increase := min(total, maxCount)
			total -= increase
			output = append(output, stack.Grow(increase-count))
		}
	}
	return output
}

// growRecipeItem increases the count stored in a recipe item while preserving its concrete recipe item type.
func growRecipeItem(it recipe.Item, count int) recipe.Item {
	switch it := it.(type) {
	case item.Stack:
		return it.Grow(count)
	case recipe.ItemTag:
		return recipe.NewItemTag(it.Tag(), it.Count()+count)
	}
	panic(fmt.Errorf("unexpected recipe item %T", it))
}
