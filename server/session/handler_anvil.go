package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math/rand"
)

// handleCraftRecipeOptional handles the CraftRecipeOptional request action, sent when taking a result from an anvil
// menu. It also contains information such as the new name of the item and the multi-recipe network ID.
func (h *ItemStackRequestHandler) handleCraftRecipeOptional(a *protocol.CraftRecipeOptionalStackRequestAction, s *Session, filterStrings []string) error {
	// First check if there actually is an anvil opened.
	if !s.containerOpened.Load() {
		return fmt.Errorf("no anvil container opened")
	}

	w := s.c.World()
	pos := s.openedPos.Load()
	anvil, ok := w.Block(pos).(block.Anvil)
	if !ok {
		return fmt.Errorf("no anvil container opened")
	}
	if len(filterStrings) < int(a.FilterStringIndex) {
		return fmt.Errorf("filter string index %v is out of bounds", a.FilterStringIndex)
	}

	first, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerAnvilInput,
		Slot:        1,
	}, s)
	if first.Empty() {
		return fmt.Errorf("no item in first input slot")
	}
	result := first

	second, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerAnvilMaterial,
		Slot:        0x2,
	}, s)

	j := first.RepairCost()
	if !second.Empty() {
		j += second.RepairCost()
	}

	var i, k int
	var repairCount int
	if !second.Empty() {
		if repairable, ok := first.Item().(item.Repairable); ok && repairable.RepairableBy(second) {
			d := min(first.MaxDurability()-first.Durability(), first.MaxDurability()/4)
			if d <= 0 {
				return fmt.Errorf("first item is already fully repaired")
			}

			for ; d > 0 && repairCount < second.Count(); repairCount, d = repairCount+1, min(result.MaxDurability()-result.Durability(), result.MaxDurability()/4) {
				result = result.WithDurability(result.Durability() + d)
				i++
			}
		} else {
			_, book := second.Item().(item.EnchantedBook)
			_, durable := first.Item().(item.Durable)

			enchant := book && len(second.Enchantments()) > 0
			if !enchant && (first.Item() != second.Item() || !durable) {
				return fmt.Errorf("first item is not repairable or second item is not an enchanted book")
			}
			if durable && !enchant {
				d := first.MaxDurability() - (first.Durability() + (second.Durability() + first.MaxDurability()*12/100))
				if d < 0 {
					d = 0
				}
				if d < first.MaxDurability()-first.Durability() {
					result = result.WithDurability(d)
					i += 2
				}
			}

			var hasCompatible, hasIncompatible bool
			for _, e := range second.Enchantments() {
				t := e.Type()

				var firstLevel int
				if firstEnchant, ok := first.Enchantment(t); ok {
					firstLevel = firstEnchant.Level()
				}

				resultLevel := max(firstLevel, e.Level())
				if firstLevel == e.Level() {
					resultLevel = firstLevel + 1
				}

				compatible := t.CompatibleWithItem(first.Item())
				if _, ok := first.Item().(item.EnchantedBook); ok {
					compatible = true
				}

				for _, e2 := range first.Enchantments() {
					if t != e2.Type() && !t.CompatibleWithOther(e2.Type()) {
						compatible = false
						i++
					}
				}

				if !compatible {
					hasIncompatible = true
					continue
				}
				hasCompatible = true

				if resultLevel > t.MaxLevel() {
					resultLevel = t.MaxLevel()
				}
				rarityCost := t.Rarity().ApplyCost
				if enchant {
					rarityCost = max(1, rarityCost/2)
				}

				result = result.WithEnchantments(item.NewEnchantment(t, resultLevel))
				i += rarityCost * resultLevel
				if first.Count() > 1 {
					i = 40
				}
			}
			if hasIncompatible && !hasCompatible {
				return fmt.Errorf("no compatible enchantments but have incompatible ones")
			}
		}
	}

	newName := filterStrings[int(a.FilterStringIndex)]
	existingName := item.DisplayName(first.Item(), s.c.Locale())
	if customName := first.CustomName(); len(customName) > 0 {
		existingName = customName
	}
	if existingName != newName {
		k = 1
		i += k
		result = result.WithCustomName(newName)
	}

	cost := i + j
	if cost <= 0 {
		return fmt.Errorf("no action was taken")
	}

	if k == i && k > 0 && cost >= 40 {
		cost = 39
	}

	c := s.c.GameMode().CreativeInventory()
	if cost >= 40 && !c {
		return fmt.Errorf("impossible cost")
	}

	if !result.Empty() {
		k2 := result.RepairCost()
		if !second.Empty() && k2 < second.RepairCost() {
			k2 = second.RepairCost()
		}
		if k != i || k == 0 {
			k2 = k2*2 + 1
		}
		result = result.WithRepairCost(k2)
	}

	level := s.c.ExperienceLevel()
	if level < cost && !c {
		// Not enough experience.
		return nil
	} else if !c {
		s.c.SetExperienceLevel(level - cost)
	}

	if !c && rand.Float64() < 0.12 {
		damaged := anvil.Break()
		if _, ok := damaged.(block.Air); ok {
			w.PlaySound(pos.Vec3Centre(), sound.AnvilBreak{})
		} else {
			w.PlaySound(pos.Vec3Centre(), sound.AnvilUse{})
		}
		defer w.SetBlock(pos, damaged, nil)
	} else {
		w.PlaySound(pos.Vec3Centre(), sound.AnvilUse{})
	}

	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerAnvilInput,
		Slot:        1,
	}, item.Stack{}, s)
	if repairCount > 0 {
		h.setItemInSlot(protocol.StackRequestSlotInfo{
			ContainerID: containerAnvilMaterial,
			Slot:        2,
		}, second.Grow(-repairCount), s)
	} else {
		h.setItemInSlot(protocol.StackRequestSlotInfo{
			ContainerID: containerAnvilMaterial,
			Slot:        2,
		}, item.Stack{}, s)
	}
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerOutput,
		Slot:        50,
	}, result, s)
	return nil
}

// max returns the max of two integers.
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// min returns the min of two integers.
func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}
