package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// PumpkinSeeds grow pumpkin blocks.
type PumpkinSeeds struct {
	crop

	// Direction is the direction from the stem to the pumpkin.
	Direction world.Face
}

// NeighbourUpdateTick ...
func (p PumpkinSeeds) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if _, ok := w.Block(pos.Side(world.FaceDown)).(Farmland); !ok {
		w.BreakBlock(pos)
	} else if p.Direction != world.FaceDown {
		if pumpkin, ok := w.Block(pos.Side(p.Direction)).(Pumpkin); !ok || pumpkin.Carved {
			p.Direction = world.FaceDown
			w.PlaceBlock(pos, p)
		}
	}
}

// RandomTick ...
func (p PumpkinSeeds) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	if rand.Float64() <= p.CalculateGrowthChance(pos, w) && w.Light(pos) >= 8 {
		if p.Growth < 7 {
			p.Growth++
			w.PlaceBlock(pos, p)
		} else {
			directions := []world.Direction{world.North, world.South, world.West, world.East}
			for _, i := range directions {
				if _, ok := w.Block(pos.Side(i.Face())).(Pumpkin); ok {
					return
				}
			}
			direction := directions[rand.Intn(len(directions))].Face()
			stemPos := pos.Side(direction)
			if _, ok := w.Block(stemPos).(Air); ok {
				switch w.Block(stemPos.Side(world.FaceDown)).(type) {
				case Farmland:
				case Dirt:
				case Grass:
					p.Direction = direction
					w.PlaceBlock(pos, p)
					w.PlaceBlock(stemPos, Pumpkin{})
				}
			}
		}
	}
}

// Bonemeal ...
func (p PumpkinSeeds) Bonemeal(pos world.BlockPos, w *world.World) bool {
	if p.Growth == 7 {
		return false
	}
	p.Growth = min(p.Growth+rand.Intn(4)+2, 7)
	w.PlaceBlock(pos, p)
	return true
}

// UseOnBlock ...
func (p PumpkinSeeds) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, p)
	if !used {
		return false
	}

	if _, ok := w.Block(pos.Side(world.FaceDown)).(Farmland); !ok {
		return false
	}

	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (p PumpkinSeeds) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(p, 1)),
	}
}

// EncodeItem ...
func (p PumpkinSeeds) EncodeItem() (id int32, meta int16) {
	return 361, 0
}

// EncodeBlock ...
func (p PumpkinSeeds) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:pumpkin_stem", map[string]interface{}{"facing_direction": int32(p.Direction), "growth": int32(p.Growth)}
}

// Hash ...
func (p PumpkinSeeds) Hash() uint64 {
	return hashPumpkinStem | (uint64(p.Growth) << 32) | (uint64(p.Direction) << 35)
}

// allPumpkinStems
func allPumpkinStems() (stems []world.Block) {
	for i := 0; i <= 7; i++ {
		for j := world.Face(0); j <= 5; j++ {
			stems = append(stems, PumpkinSeeds{Direction: j, crop: crop{Growth: i}})
		}
	}
	return
}
