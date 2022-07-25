package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	// loomInputSlot is the slot index of the input item in the loom table.
	loomInputSlot = 0x09
	// loomDyeSlot is the slot index of the dye item in the loom table.
	loomDyeSlot = 0x0A
	// loomPatternSlot is the slot index of the pattern item in the loom table.
	loomPatternSlot = 0x0B
)

// handleLoomCraft handles a CraftLoomRecipe stack request action made using a loom table.
func (h *ItemStackRequestHandler) handleLoomCraft(a *protocol.CraftLoomRecipeStackRequestAction, s *Session) error {
	// First check if there actually is a loom opened.
	if !s.containerOpened.Load() {
		return fmt.Errorf("no loom container opened")
	}
	if _, ok := s.c.World().Block(s.openedPos.Load()).(block.Loom); !ok {
		return fmt.Errorf("no loom container opened")
	}

	input, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerLoomInput,
		Slot:        loomInputSlot,
	}, s)
	if input.Empty() {
		return fmt.Errorf("input item is empty")
	}
	b, ok := input.Item().(block.Banner)
	if !ok {
		return fmt.Errorf("input item is not a banner")
	}

	dye, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerLoomDye,
		Slot:        loomDyeSlot,
	}, s)
	if dye.Empty() {
		return fmt.Errorf("dye item is empty")
	}
	d, ok := dye.Item().(item.Dye)
	if !ok {
		return fmt.Errorf("dye item is not a dye")
	}

	expectedPattern, ok := block.BannerPatternByID(a.Pattern)
	if !ok {
		return fmt.Errorf("pattern %v is not a valid banner pattern", a.Pattern)
	}

	pattern, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerLoomPattern,
		Slot:        loomPatternSlot,
	}, s)
	if expectedPatternItem, hasPatternItem := expectedPattern.Item(); hasPatternItem {
		if pattern.Empty() {
			return fmt.Errorf("pattern item is empty but the pattern is required")
		}
		p, ok := pattern.Item().(item.BannerPattern)
		if !ok {
			return fmt.Errorf("pattern item is not a banner pattern")
		}
		if expectedPatternItem != p.Pattern {
			return fmt.Errorf("pattern item does not match the expected pattern")
		}
	}

	b.Patterns = append(b.Patterns, block.BannerPatternLayer{
		BannerPatternType: expectedPattern,
		Colour:            d.Colour,
	})
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerLoomInput,
		Slot:        loomInputSlot,
	}, input.Grow(-1), s)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerLoomDye,
		Slot:        loomDyeSlot,
	}, dye.Grow(-1), s)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerOutput,
		Slot:        craftingResult,
	}, duplicateStack(input, b), s)
	return nil
}
