package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math/rand"
)

// handleCraftRecipeOptional ...
func (h *ItemStackRequestHandler) handleCraftRecipeOptional(a *protocol.CraftRecipeOptionalStackRequestAction, s *Session, filterStrings []string) error {
	// First check if there actually is an anvil opened.
	if !s.containerOpened.Load() {
		return fmt.Errorf("no anvil container opened")
	}
	pos := s.openedPos.Load()
	w := s.c.World()
	anvil, ok := w.Block(pos).(block.Anvil)
	if !ok {
		return fmt.Errorf("no anvil container opened")
	}
	if len(filterStrings) < int(a.FilterStringIndex) {
		// Invalid filter string index.
		return nil
	}

	first, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerAnvilInput,
		Slot:        1,
	}, s)
	if first.Empty() {
		// First anvil slot is empty, can't result in anything.
		return nil
	}
	result := first

	second, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerAnvilMaterial,
		Slot:        0x2,
	}, s)

	cost := first.RepairCost()
	if !second.Empty() {
		cost += second.RepairCost()
	}

	var repairCount int
	if !second.Empty() {
		if repairable, ok := first.Item().(item.Repairable); ok && repairable.RepairableBy(second) {
			d := min(first.MaxDurability()-first.Durability(), first.MaxDurability()/4)
			if d <= 0 {
				return nil
			}

			for ; d > 0 && repairCount < second.Count(); repairCount, d = repairCount+1, min(result.MaxDurability()-result.Durability(), result.MaxDurability()/4) {
				result = result.WithDurability(result.Durability() + d)
				cost++
			}
		} else {
			_, book := second.Item().(item.EnchantedBook)
			_, durable := first.Item().(item.Durable)

			enchant := book && len(second.Enchantments()) > 0
			if !enchant && (first.Item() != second.Item() || !durable) {
				return nil
			}
			if durable && !enchant {
				d := first.MaxDurability() - (first.Durability() + (second.Durability() + first.MaxDurability()*12/100))
				if d < 0 {
					d = 0
				}
				if d < first.MaxDurability()-first.Durability() {
					result = result.WithDurability(d)
					cost += 2
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
						cost++
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
				cost += rarityCost * resultLevel
				if first.Count() > 1 {
					cost = 40
				}
			}
			if hasIncompatible && !hasCompatible {
				// We have no compatible enchantments, but we have incompatible ones.
				return nil
			}
		}
	}

	newName := filterStrings[int(a.FilterStringIndex)]
	existingName := item.DisplayName(first.Item(), s.c.Locale())
	if customName := first.CustomName(); len(customName) > 0 {
		existingName = customName
	}
	if existingName != newName {
		result = result.WithCustomName(newName)
		cost += 1
	}

	if cost == 0 {
		// No action was performed.
		return nil
	}

	c := s.c.GameMode().CreativeInventory()
	if cost >= 40 && !c {
		// Impossible repair/rename.
		return nil
	}

	if !result.Empty() {
		i := result.RepairCost()
		if !second.Empty() && i < second.RepairCost() {
			i = second.RepairCost()
		}
		if cost != 1 {
			i = i*2 + 1
		}
		result = result.WithRepairCost(i)
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
