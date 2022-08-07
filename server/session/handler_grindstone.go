package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	// grindstoneFirstInputSlot is the slot index of the first input item in the grindstone.
	grindstoneFirstInputSlot = 0x10
	// grindstoneSecondInputSlot is the slot index of the second input item in the grindstone.
	grindstoneSecondInputSlot = 0x11
)

// handleGrindstoneCraft handles a CraftGrindstoneRecipe stack request action made using a grindstone.
func (h *ItemStackRequestHandler) handleGrindstoneCraft(a *protocol.CraftGrindstoneRecipeStackRequestAction, s *Session) error {
	// First check if there actually is a grindstone opened.
	if !s.containerOpened.Load() {
		return fmt.Errorf("no grindstone container opened")
	}
	if _, ok := s.c.World().Block(s.openedPos.Load()).(block.Grindstone); !ok {
		return fmt.Errorf("no grindstone container opened")
	}

	// TODO: Implement proper support for this.
	return fmt.Errorf("not implemented")
}
