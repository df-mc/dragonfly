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

// enchantNames are names translated to the 'Standard Galactic Alphabet' client-side. The names generally have no meaning
// on the vanilla server implementation.
var enchantNames = []string{
	"dragonfly",
	"gophertunnel",
	"go raknet",
	"go lang",
	"sandertv",
	"t 14 raptor",
	"da pig guy",
	"potatoe train yt",
}

const (
	// enchantingInputSlot is the slot index of the input item in the enchanting table.
	enchantingInputSlot = 14
	// enchantingLapisSlot is the slot index of the lapis in the enchanting table.
	enchantingLapisSlot = 15
)

// handleEnchant handles the enchantment of an item using the CraftRecipe stack request action.
func (h *ItemStackRequestHandler) handleEnchant(a *protocol.CraftRecipeStackRequestAction, s *Session) error {
	if a.RecipeNetworkID > 2 {
		return fmt.Errorf("invalid recipe network id: %d", a.RecipeNetworkID)
	}

	input, err := s.ui.Item(enchantingInputSlot)
	if err != nil {
		return err
	}
	if input.Count() > 1 {
		return fmt.Errorf("enchanting tables only accept one item at a time")
	}

	allCosts, allEnchants := s.determineAvailableEnchantments(s.c.World(), s.openedPos.Load(), input)
	if len(allEnchants) == 0 {
		return fmt.Errorf("can't enchant non-enchantable item")
	}

	cost := int(a.RecipeNetworkID + 1)
	requirement := allCosts[a.RecipeNetworkID]
	enchants := allEnchants[a.RecipeNetworkID]

	if !s.c.GameMode().CreativeInventory() {
		if s.c.ExperienceLevel() < requirement {
			return fmt.Errorf("not enough levels to meet requirement")
		}
		if s.c.ExperienceLevel() < cost {
			return fmt.Errorf("not enough levels to meet cost")
		}

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

		s.c.SetExperienceLevel(s.c.ExperienceLevel() - cost)
		h.setItemInSlot(protocol.StackRequestSlotInfo{
			ContainerID: containerEnchantingTableLapis,
			Slot:        enchantingLapisSlot,
		}, lapis.Grow(-cost), s)
	}

	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerEnchantingTableInput,
		Slot:        enchantingInputSlot,
	}, item.Stack{}, s)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerOutput,
		Slot:        craftingResult,
	}, input.WithEnchantments(enchants...), s)
	s.c.ResetEnchantmentSeed()
	return nil
}

// sendEnchantmentOptions sends a list of available enchantments to the client based on the client's enchantment seed
// and nearby bookshelves.
func (s *Session) sendEnchantmentOptions(w *world.World, pos cube.Pos, stack item.Stack) {
	selectedCosts, selectedEnchants := s.determineAvailableEnchantments(w, pos, stack)
	if len(selectedEnchants) == 0 {
		// No available enchantments.
		return
	}

	options := make([]protocol.EnchantmentOption, 0, 3)
	for i := 0; i < 3; i++ {
		enchants := make([]protocol.EnchantmentInstance, 0, len(selectedEnchants[i]))
		for _, enchant := range selectedEnchants[i] {
			id, _ := item.EnchantmentID(enchant.Type())
			enchants = append(enchants, protocol.EnchantmentInstance{
				Type:  byte(id),
				Level: byte(enchant.Level()),
			})
		}

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

	s.writePacket(&packet.PlayerEnchantOptions{Options: options})
}

// determineAvailableEnchantments returns a list of pseudo-random enchantments for the given item stack.
func (s *Session) determineAvailableEnchantments(w *world.World, pos cube.Pos, stack item.Stack) ([]int, [][]item.Enchantment) {
	enchantable, ok := stack.Item().(item.Enchantable)
	if !ok {
		// We can't enchant this item.
		return nil, nil
	}
	if len(stack.Enchantments()) > 0 {
		// We can't enchant this item.
		return nil, nil
	}

	random := rand.New(rand.NewSource(s.c.EnchantmentSeed()))
	bookshelves := searchBookshelves(w, pos)
	value := enchantable.EnchantmentValue()

	base := random.Intn(8) + 1 + (bookshelves >> 1) + random.Intn(bookshelves+1)

	upperLevel := max(base/3, 1)
	middleLevel := base*2/3 + 1
	lowerLevel := max(base, bookshelves*2)

	return []int{
			upperLevel,
			middleLevel,
			lowerLevel,
		}, [][]item.Enchantment{
			createEnchantments(random, stack, value, upperLevel),
			createEnchantments(random, stack, value, middleLevel),
			createEnchantments(random, stack, value, lowerLevel),
		}
}

// createEnchantments creates a list of enchantments for the given item stack and returns them.
func createEnchantments(random *rand.Rand, stack item.Stack, value, level int) []item.Enchantment {
	it := stack.Item()
	f := (random.Float64() + random.Float64() - 1.0) * 0.15

	useLevel := level + 1 + random.Intn(value/4+1) + random.Intn(value/4+1)
	useLevel = clamp(int(math.Round(float64(useLevel)+float64(useLevel)*f)), 1, math.MaxInt32)

	_, book := it.(item.Book)
	availableEnchants := make([]item.Enchantment, 0, len(item.Enchantments()))
	for _, enchant := range item.Enchantments() {
		if book || enchant.CompatibleWithItem(it) {
			for i := enchant.MaxLevel(); i > 0; i-- {
				if useLevel >= enchant.MinCost(i) && useLevel <= enchant.MaxCost(i) {
					availableEnchants = append(availableEnchants, item.NewEnchantment(enchant, i))
					break
				}
			}
		}
	}
	if len(availableEnchants) == 0 {
		// No available enchantments.
		return nil
	}

	selectedEnchants := make([]item.Enchantment, 0, len(availableEnchants))

	enchant := weightedRandomEnchantment(random, availableEnchants)
	selectedEnchants = append(selectedEnchants, enchant)

	ind := sliceutil.Index(availableEnchants, enchant)
	availableEnchants = slices.Delete(availableEnchants, ind, ind+1)

	for random.Intn(50) <= useLevel {
		lastEnchant := selectedEnchants[len(selectedEnchants)-1]
		if availableEnchants = sliceutil.Filter(availableEnchants, func(enchant item.Enchantment) bool {
			return lastEnchant.Type().CompatibleWithOther(enchant.Type())
		}); len(availableEnchants) == 0 {
			// We've exhausted all available enchantments.
			break
		}

		enchant = weightedRandomEnchantment(random, availableEnchants)
		selectedEnchants = append(selectedEnchants, enchant)

		ind = sliceutil.Index(availableEnchants, enchant)
		availableEnchants = slices.Delete(availableEnchants, ind, ind+1)

		useLevel /= 2
	}
	return selectedEnchants
}

// searchBookshelves searches for nearby bookshelves around the position passed, and returns the amount found.
func searchBookshelves(w *world.World, pos cube.Pos) int {
	var foundShelves int
	for z := -1; z <= 1; z++ {
		for x := -1; x <= 1; x++ {
			if z != 0 || x != 0 {
				if _, ok := w.Block(pos.Add(cube.Pos{x, 0, z})).(block.Air); !ok {
					continue
				}
				if _, ok := w.Block(pos.Add(cube.Pos{x, 1, z})).(block.Air); !ok {
					continue
				}

				if _, ok := w.Block(pos.Add(cube.Pos{x * 2, 0, z * 2})).(block.Bookshelf); ok {
					foundShelves++
				}
				if _, ok := w.Block(pos.Add(cube.Pos{x * 2, 1, z * 2})).(block.Bookshelf); ok {
					foundShelves++
				}

				if x != 0 && z != 0 {
					if _, ok := w.Block(pos.Add(cube.Pos{x * 2, 0, z})).(block.Bookshelf); ok {
						foundShelves++
					}
					if _, ok := w.Block(pos.Add(cube.Pos{x * 2, 1, z})).(block.Bookshelf); ok {
						foundShelves++
					}

					if _, ok := w.Block(pos.Add(cube.Pos{x, 0, z * 2})).(block.Bookshelf); ok {
						foundShelves++
					}
					if _, ok := w.Block(pos.Add(cube.Pos{x, 1, z * 2})).(block.Bookshelf); ok {
						foundShelves++
					}
				}

				if foundShelves >= 15 {
					return foundShelves
				}
			}
		}
	}
	return foundShelves
}

// weightedRandomEnchantment returns a random enchantment from the given list of enchantments using the rarity weight of
// each enchantment.
func weightedRandomEnchantment(rs *rand.Rand, enchants []item.Enchantment) item.Enchantment {
	var totalWeight int
	for _, e := range enchants {
		totalWeight += e.Type().Rarity().Weight
	}
	r := rs.Intn(totalWeight)
	for _, e := range enchants {
		r -= e.Type().Rarity().Weight
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
