package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/exp/slices"
	"math"
	"math/rand"
)

const (
	// enchantingInputSlot is the slot index of the input item in the enchanting table.
	enchantingInputSlot = 0x0e
	// enchantingLapisSlot is the slot index of the lapis in the enchanting table.
	enchantingLapisSlot = 0x0f
)

// handleEnchant handles the enchantment of an item using the CraftRecipe stack request action.
func (h *ItemStackRequestHandler) handleEnchant(a *protocol.CraftRecipeStackRequestAction, s *Session) error {
	// First ensure that the selected slot is not out of bounds.
	if a.RecipeNetworkID > 2 {
		return fmt.Errorf("invalid recipe network id: %d", a.RecipeNetworkID)
	}

	// Now ensure we have an input and only one input.
	input, err := s.ui.Item(enchantingInputSlot)
	if err != nil {
		return err
	}
	if input.Count() > 1 {
		return fmt.Errorf("enchanting tables only accept one item at a time")
	}

	// Determine the available enchantments using the session's enchantment seed.
	allCosts, allEnchants := s.determineAvailableEnchantments(s.c.World(), s.openedPos.Load(), input)
	if len(allEnchants) == 0 {
		return fmt.Errorf("can't enchant non-enchantable item")
	}

	// Use the slot plus one as the cost. The requirement and enchantments can be found in the results from
	// determineAvailableEnchantments using the same slot index.
	cost := int(a.RecipeNetworkID + 1)
	requirement := allCosts[a.RecipeNetworkID]
	enchants := allEnchants[a.RecipeNetworkID]

	// If we don't have infinite resources, we need to deduct Lapis Lazuli and experience.
	if !s.c.GameMode().CreativeInventory() {
		// First ensure that the experience level is both underneath the requirement and the cost.
		if s.c.ExperienceLevel() < requirement {
			return fmt.Errorf("not enough levels to meet requirement")
		}
		if s.c.ExperienceLevel() < cost {
			return fmt.Errorf("not enough levels to meet cost")
		}

		// Then ensure that the player has input Lapis Lazuli, and enough of it to meet the cost.
		lapis, err := s.ui.Item(enchantingLapisSlot)
		if err != nil {
			return err
		}
		if _, ok := lapis.Item().(item.LapisLazuli); !ok {
			return fmt.Errorf("lapis lazuli was not input")
		}
		if lapis.Count() < cost {
			return fmt.Errorf("not enough lapis lazuli to meet cost")
		}

		// Deduct the experience and Lapis Lazuli.
		s.c.SetExperienceLevel(s.c.ExperienceLevel() - cost)
		h.setItemInSlot(protocol.StackRequestSlotInfo{
			ContainerID: containerEnchantingTableLapis,
			Slot:        enchantingLapisSlot,
		}, lapis.Grow(-cost), s)
	}

	// Clear the existing input item, and apply the new item into the crafting result slot of the UI. The client will
	// automatically move the item into the input slot.
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerEnchantingTableInput,
		Slot:        enchantingInputSlot,
	}, item.Stack{}, s)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerOutput,
		Slot:        craftingResult,
	}, input.WithEnchantments(enchants...), s)

	// Reset the enchantment seed so different enchantments can be selected.
	s.c.ResetEnchantmentSeed()
	return nil
}

// sendEnchantmentOptions sends a list of available enchantments to the client based on the client's enchantment seed
// and nearby bookshelves.
func (s *Session) sendEnchantmentOptions(w *world.World, pos cube.Pos, stack item.Stack) {
	// First determine the available enchantments for the given item stack.
	selectedCosts, selectedEnchants := s.determineAvailableEnchantments(w, pos, stack)
	if len(selectedEnchants) == 0 {
		// No available enchantments.
		return
	}

	// Build the protocol variant of the enchantment options.
	options := make([]protocol.EnchantmentOption, 0, 3)
	for i := 0; i < 3; i++ {
		// First build the enchantment instances for each selected enchantment.
		enchants := make([]protocol.EnchantmentInstance, 0, len(selectedEnchants[i]))
		for _, enchant := range selectedEnchants[i] {
			id, _ := item.EnchantmentID(enchant.Type())
			enchants = append(enchants, protocol.EnchantmentInstance{
				Type:  byte(id),
				Level: byte(enchant.Level()),
			})
		}

		// Then build the enchantment option. We can use the slot as the RecipeNetworkID, since the IDs seem to be unique
		// to enchanting tables only. We also only need to set the middle index of Enchantments. The other two serve
		// an unknown purpose and can cause various unexpected issues.
		options = append(options, protocol.EnchantmentOption{
			Name:            enchantNames[rand.Intn(len(enchantNames))],
			Cost:            uint32(selectedCosts[i]),
			RecipeNetworkID: uint32(i),
			Enchantments: protocol.ItemEnchantments{
				Slot:         int32(i),
				Enchantments: [3][]protocol.EnchantmentInstance{1: enchants},
			},
		})
	}

	// Send the enchantment options to the client.
	s.writePacket(&packet.PlayerEnchantOptions{Options: options})
}

// determineAvailableEnchantments returns a list of pseudo-random enchantments for the given item stack.
func (s *Session) determineAvailableEnchantments(w *world.World, pos cube.Pos, stack item.Stack) ([]int, [][]item.Enchantment) {
	// First ensure that the item is enchantable and does not already have any enchantments.
	enchantable, ok := stack.Item().(item.Enchantable)
	if !ok {
		// We can't enchant this item.
		return nil, nil
	}
	if len(stack.Enchantments()) > 0 {
		// We can't enchant this item.
		return nil, nil
	}

	// Search for bookshelves around the enchanting table. Bookshelves help boost the value of the enchantments that
	// are selected, resulting in enchantments that are rarer but also more expensive.
	random := rand.New(rand.NewSource(s.c.EnchantmentSeed()))
	bookshelves := searchBookshelves(w, pos)
	value := enchantable.EnchantmentValue()

	// Calculate the base cost, used to calculate the upper, middle, and lower level costs.
	baseCost := random.Intn(8) + 1 + (bookshelves >> 1) + random.Intn(bookshelves+1)

	// Calculate the upper, middle, and lower level costs.
	upperLevelCost := max(baseCost/3, 1)
	middleLevelCost := baseCost*2/3 + 1
	lowerLevelCost := max(baseCost, bookshelves*2)

	// Create a list of available enchantments for each slot.
	return []int{
			upperLevelCost,
			middleLevelCost,
			lowerLevelCost,
		}, [][]item.Enchantment{
			createEnchantments(random, stack, value, upperLevelCost),
			createEnchantments(random, stack, value, middleLevelCost),
			createEnchantments(random, stack, value, lowerLevelCost),
		}
}

// treasureEnchantment represents an enchantment that may be a treasure enchantment.
type treasureEnchantment interface {
	item.EnchantmentType
	Treasure() bool
}

// createEnchantments creates a list of enchantments for the given item stack and returns them.
func createEnchantments(random *rand.Rand, stack item.Stack, value, level int) []item.Enchantment {
	// Calculate the "random bonus" for this level. This factor is used in calculating the enchantment cost, used
	// during the selection of enchantments.
	randomBonus := (random.Float64() + random.Float64() - 1.0) * 0.15

	// Calculate the enchantment cost and clamp it to ensure it is always at least one with triangular distribution.
	cost := level + 1 + random.Intn(value/4+1) + random.Intn(value/4+1)
	cost = clamp(int(math.Round(float64(cost)+float64(cost)*randomBonus)), 1, math.MaxInt32)

	// Books are applicable to all enchantments, so make sure we have a flag for them here.
	it := stack.Item()
	_, book := it.(item.Book)

	// Now that we have our enchantment cost, we need to select the available enchantments. First, we iterate through
	// each possible enchantment.
	availableEnchants := make([]item.Enchantment, 0, len(item.Enchantments()))
	for _, enchant := range item.Enchantments() {
		if t, ok := enchant.(treasureEnchantment); ok && t.Treasure() {
			// We then have to ensure that the enchantment is not a treasure enchantment, as those cannot be selected through
			// the enchanting table.
			continue
		}
		if !book && !enchant.CompatibleWithItem(it) {
			// The enchantment is not compatible with the item.
			continue
		}

		// Now iterate through each possible level of the enchantment.
		for i := enchant.MaxLevel(); i > 0; i-- {
			// Use the level to calculate the minimum and maximum costs for this enchantment.
			if minCost, maxCost := enchant.Cost(i); cost >= minCost && cost <= maxCost {
				// If the cost is within the bounds, add the enchantment to the list of available enchantments.
				availableEnchants = append(availableEnchants, item.NewEnchantment(enchant, i))
				break
			}
		}
	}
	if len(availableEnchants) == 0 {
		// No available enchantments, so we can't really do much here.
		return nil
	}

	// Now we need to select the enchantments.
	selectedEnchants := make([]item.Enchantment, 0, len(availableEnchants))

	// Select the first enchantment using a weighted random algorithm, favouring enchantments that have a higher weight.
	// These weights are based on the enchantment's rarity, with common and uncommon enchantments having a higher weight
	// than rare and very rare enchantments.
	enchant := weightedRandomEnchantment(random, availableEnchants)
	selectedEnchants = append(selectedEnchants, enchant)

	// Remove the selected enchantment from the list of available enchantments, so we don't select it again.
	ind := sliceutil.Index(availableEnchants, enchant)
	availableEnchants = slices.Delete(availableEnchants, ind, ind+1)

	// Based on the cost, select a random amount of additional enchantments.
	for random.Intn(50) <= cost {
		// Ensure that we don't have any conflicting enchantments. If so, remove them from the list of available
		// enchantments.
		lastEnchant := selectedEnchants[len(selectedEnchants)-1]
		if availableEnchants = sliceutil.Filter(availableEnchants, func(enchant item.Enchantment) bool {
			return lastEnchant.Type().CompatibleWithEnchantment(enchant.Type())
		}); len(availableEnchants) == 0 {
			// We've exhausted all available enchantments.
			break
		}

		// Select another enchantment using the same weighted random algorithm.
		enchant = weightedRandomEnchantment(random, availableEnchants)
		selectedEnchants = append(selectedEnchants, enchant)

		// Remove the selected enchantment from the list of available enchantments, so we don't select it again.
		ind = sliceutil.Index(availableEnchants, enchant)
		availableEnchants = slices.Delete(availableEnchants, ind, ind+1)

		// Halve the cost, so we have a lower chance of selecting another enchantment.
		cost /= 2
	}
	return selectedEnchants
}

// searchBookshelves searches for nearby bookshelves around the position passed, and returns the amount found.
func searchBookshelves(w *world.World, pos cube.Pos) (shelves int) {
	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			for y := 0; y <= 1; y++ {
				if x == 0 && z == 0 {
					// Ignore the center block.
					continue
				}
				if _, ok := w.Block(pos.Add(cube.Pos{x, y, z})).(block.Air); !ok {
					// There must be a one block space between the bookshelf and the player.
					continue
				}

				// Check for a bookshelf two blocks away.
				if _, ok := w.Block(pos.Add(cube.Pos{x * 2, y, z * 2})).(block.Bookshelf); ok {
					shelves++
				}
				if x != 0 && z != 0 {
					// Check for a bookshelf two blocks away on the X axis.
					if _, ok := w.Block(pos.Add(cube.Pos{x * 2, y, z})).(block.Bookshelf); ok {
						shelves++
					}
					// Check for a bookshelf two blocks away on the Z axis.
					if _, ok := w.Block(pos.Add(cube.Pos{x, y, z * 2})).(block.Bookshelf); ok {
						shelves++
					}
				}

				if shelves >= 15 {
					// We've found enough bookshelves.
					return 15
				}
			}
		}
	}
	return shelves
}

// weightedRandomEnchantment returns a random enchantment from the given list of enchantments using the rarity weight of
// each enchantment.
func weightedRandomEnchantment(rs *rand.Rand, enchants []item.Enchantment) item.Enchantment {
	var totalWeight int
	for _, e := range enchants {
		totalWeight += e.Type().Rarity().Weight()
	}
	r := rs.Intn(totalWeight)
	for _, e := range enchants {
		r -= e.Type().Rarity().Weight()
		if r < 0 {
			return e
		}
	}
	panic("should never happen")
}

// clamp clamps a value into the given range.
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
