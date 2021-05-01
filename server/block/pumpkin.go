package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/instrument"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Pumpkin is a crop block. Interacting with shears results in the carved variant.
type Pumpkin struct {
	solid

	// Carved is whether the pumpkin is carved.
	Carved bool
	// Facing is the direction the pumpkin is facing.
	Facing cube.Direction
}

// Instrument ...
func (p Pumpkin) Instrument() instrument.Instrument {
	if !p.Carved {
		return instrument.Didgeridoo()
	}
	return instrument.Piano()
}

// UseOnBlock ...
func (p Pumpkin) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, p)
	if !used {
		return
	}
	p.Facing = user.Facing().Opposite()

	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (p Pumpkin) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    1,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(p, 1)),
	}
}

// Carve ...
func (p Pumpkin) Carve(f cube.Face) (world.Block, bool) {
	return Pumpkin{Facing: f.Direction(), Carved: true}, !p.Carved
}

// EncodeItem ...
func (p Pumpkin) EncodeItem() (id int32, name string, meta int16) {
	if p.Carved {
		return -155, "minecraft:carved_pumpkin", 0
	}
	return 86, "minecraft:pumpkin", 0
}

// EncodeBlock ...
func (p Pumpkin) EncodeBlock() (name string, properties map[string]interface{}) {
	direction := 2
	switch p.Facing {
	case cube.South:
		direction = 0
	case cube.West:
		direction = 1
	case cube.East:
		direction = 3
	}

	if p.Carved {
		return "minecraft:carved_pumpkin", map[string]interface{}{"direction": int32(direction)}
	}
	return "minecraft:pumpkin", map[string]interface{}{"direction": int32(direction)}
}

func allPumpkins() (pumpkins []world.Block) {
	for i := cube.Direction(0); i <= 3; i++ {
		pumpkins = append(pumpkins, Pumpkin{Facing: i})
		pumpkins = append(pumpkins, Pumpkin{Facing: i, Carved: true})
	}
	return
}
