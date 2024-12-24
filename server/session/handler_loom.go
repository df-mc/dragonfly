package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	// loomInputSlot is the slot index of the input item in the loom table.
	loomInputSlot = 0x09
	// loomDyeSlot is the slot index of the dye item in the loom table.
	loomDyeSlot = 0x0a
	// loomPatternSlot is the slot index of the pattern item in the loom table.
	loomPatternSlot = 0x0b
)

// handleLoomCraft handles a CraftLoomRecipe stack request action made using a loom table.
func (h *ItemStackRequestHandler) handleLoomCraft(a *protocol.CraftLoomRecipeStackRequestAction, s *Session, tx *world.Tx) error {
	// First check if there actually is a loom opened.
	if _, ok := tx.Block(*s.openedPos.Load()).(block.Loom); !ok || !s.containerOpened.Load() {
		return fmt.Errorf("no loom container opened")
	}
	timesCrafted := int(a.TimesCrafted)
	if timesCrafted < 1 {
		return fmt.Errorf("times crafted must be least 1")
	}

	// Next, check if the input slot has a valid banner item.
	input, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerLoomInput},
		Slot:      loomInputSlot,
	}, s, tx)
	if input.Count() < timesCrafted {
		return fmt.Errorf("input item count is less than times crafted")
	}
	b, ok := input.Item().(block.Banner)
	if !ok {
		return fmt.Errorf("input item is not a banner")
	}
	if b.Illager {
		return fmt.Errorf("input item is an illager banner")
	}

	// Do the same with the input dye.
	dye, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerLoomDye},
		Slot:      loomDyeSlot,
	}, s, tx)
	if dye.Count() < timesCrafted {
		return fmt.Errorf("dye item count is less than times crafted")
	}
	d, ok := dye.Item().(item.Dye)
	if !ok {
		return fmt.Errorf("dye item is not a dye")
	}

	// The action contains the pattern that the client wanted to apply, so parse the ID and check if it is a valid
	// pattern.
	expectedPattern := block.BannerPatternByID(a.Pattern)

	// Some banner patterns have equivalent banner pattern items that are required to craft the pattern. If the expected
	// pattern has a pattern item, check if the player input the correct pattern item.
	pattern, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerLoomMaterial},
		Slot:      loomPatternSlot,
	}, s, tx)
	if expectedPatternItem, hasPatternItem := expectedPattern.Item(); hasPatternItem {
		if pattern.Empty() {
			return fmt.Errorf("pattern item is empty but the pattern is required")
		}
		p, ok := pattern.Item().(item.BannerPattern)
		if !ok {
			return fmt.Errorf("pattern item is not a banner pattern")
		}
		if expectedPatternItem != p.Type {
			return fmt.Errorf("pattern item does not match the expected pattern")
		}
	}

	// Add a new pattern layer onto the banner, and create the result.
	b.Patterns = append(b.Patterns, block.BannerPatternLayer{
		Type:   expectedPattern,
		Colour: d.Colour,
	})
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerLoomInput},
		Slot:      loomInputSlot,
	}, input.Grow(-timesCrafted), s, tx)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerLoomDye},
		Slot:      loomDyeSlot,
	}, dye.Grow(-timesCrafted), s, tx)
	return h.createResults(s, tx, input.WithItem(b))
}
