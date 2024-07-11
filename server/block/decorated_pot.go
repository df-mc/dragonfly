package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// PotDecoration represents an item that can be used as a decoration on a pot.
type PotDecoration interface {
	world.Item
	PotDecoration() bool
}

// DecoratedPot is a decoration block that can be crafted from up to four pottery sherds, and bricks on the sides where
// no pattern should be displayed.
type DecoratedPot struct {
	// Facing is the direction the pot is facing. The first decoration will be facing opposite of this direction.
	Facing cube.Direction
	// Decorations are the four decorations displayed on the sides of the pot. If a decoration is a brick or nil,
	// the side will appear to be empty.
	Decorations [4]PotDecoration
}

// BreakInfo ...
func (p DecoratedPot) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(p))
}

// MaxCount ...
func (DecoratedPot) MaxCount() int {
	return 64
}

// EncodeItem ...
func (p DecoratedPot) EncodeItem() (name string, meta int16) {
	return "minecraft:decorated_pot", 0
}

// EncodeBlock ...
func (p DecoratedPot) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:decorated_pot", map[string]any{"direction": int32(horizontalDirection(p.Facing))}
}

// Model ...
func (p DecoratedPot) Model() world.BlockModel {
	return model.DecoratedPot{}
}

// UseOnBlock ...
func (p DecoratedPot) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, p)
	if !used {
		return
	}
	p.Facing = user.Rotation().Direction().Opposite()

	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// EncodeNBT ...
func (p DecoratedPot) EncodeNBT() map[string]any {
	var sherds []any
	for _, decoration := range p.Decorations {
		if decoration == nil {
			sherds = append(sherds, "minecraft:brick")
		} else {
			name, _ := decoration.EncodeItem()
			sherds = append(sherds, name)
		}
	}
	return map[string]any{
		"sherds": sherds,
		"id":     "DecoratedPot",
	}
}

// DecodeNBT ...
func (p DecoratedPot) DecodeNBT(data map[string]any) any {
	p.Decorations = [4]PotDecoration{}
	if sherds := nbtconv.Slice(data, "sherds"); sherds != nil {
		for i, name := range sherds {
			it, ok := world.ItemByName(name.(string), 0)
			if !ok {
				panic(fmt.Errorf("unknown item %s", name))
			}
			decoration, ok := it.(PotDecoration)
			if !ok {
				panic(fmt.Errorf("item %s is not a pot decoration", name))
			}
			p.Decorations[i] = decoration
		}
	}
	return p
}

// allDecoratedPots ...
func allDecoratedPots() (pots []world.Block) {
	for _, f := range cube.Directions() {
		pots = append(pots, DecoratedPot{Facing: f})
	}
	return
}
