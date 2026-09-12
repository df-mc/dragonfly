package recipe

import (
	"time"

	"github.com/df-mc/dragonfly/server/item"
)

// FireworkRocketRecipe is a dynamic recipe for crafting firework rockets.
// The output depends on the firework stars, gunpowder, and paper used.
// Pattern: firework star(s) + paper + gunpowder (1-3 gunpowder for duration)
// Firework stars determine the explosion effects.
type FireworkRocketRecipe struct {
	block string
}

// NewFireworkRocketRecipe creates a new firework rocket recipe.
func NewFireworkRocketRecipe() FireworkRocketRecipe {
	return FireworkRocketRecipe{block: "crafting_table"}
}

// Match checks if the given input items match the firework rocket recipe pattern.
// Pattern: firework star(s) + paper + gunpowder (1-3 gunpowder for flight duration)
// Returns the crafted firework rocket with the appropriate explosions and duration.
func (r FireworkRocketRecipe) Match(input []Item) (output []item.Stack, ok bool) {
	// 3x3 crafting grid = 9 slots
	if len(input) != 9 {
		return nil, false
	}

	var stars []item.FireworkStar
	var paperCount int
	var gunpowderCount int

	for _, it := range input {
		if it.Empty() {
			continue
		}

		// Check item type
		stack, ok := it.(item.Stack)
		if !ok {
			return nil, false
		}

		itemName, _ := stack.Item().EncodeItem()

		switch itemName {
		case "minecraft:firework_star":
			// Firework star - extract explosion data
			if star, ok := stack.Item().(item.FireworkStar); ok {
				stars = append(stars, star)
			}
		case "minecraft:paper":
			paperCount += stack.Count()
		case "minecraft:gunpowder":
			gunpowderCount += stack.Count()
		default:
			return nil, false
		}
	}

	// Must have at least one firework star, one paper, and 1-3 gunpowder
	if len(stars) == 0 || paperCount < 1 || gunpowderCount < 1 || gunpowderCount > 3 {
		return nil, false
	}

	// No other items allowed
	if len(stars)+paperCount+gunpowderCount > 9 {
		return nil, false
	}

	// Build firework rocket with the stars and duration based on gunpowder
	// Duration: 1 gunpowder = 1, 2 = 2, 3 = 3 (flight duration)
	flightDuration := byte(gunpowderCount)

	explosions := make([]item.FireworkExplosion, len(stars))
	for i, star := range stars {
		explosions[i] = star.FireworkExplosion
	}

	firework := item.Firework{
		Duration:   time.Duration(flightDuration) * 10 * time.Second, // 10 seconds per gunpowder
		Explosions: explosions,
	}

	return []item.Stack{item.NewStack(firework, 1)}, true
}

// Block returns the block used to craft this recipe.
func (r FireworkRocketRecipe) Block() string {
	return r.block
}
