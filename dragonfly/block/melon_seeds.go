package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// MelonSeeds grow melon blocks.
type MelonSeeds struct {
	crop

	// Direction is the direction from the stem to the melon.
	Direction world.Face
}

// NeighbourUpdateTick ...
func (m MelonSeeds) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if _, ok := w.Block(pos.Side(world.FaceDown)).(Farmland); !ok {
		w.BreakBlock(pos)
	} else if m.Direction != world.FaceDown {
		if _, ok := w.Block(pos.Side(m.Direction)).(Melon); !ok {
			m.Direction = world.FaceDown
			w.PlaceBlock(pos, m)
		}
	}
}

// RandomTick ...
func (m MelonSeeds) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	if rand.Float64() <= m.CalculateGrowthChance(pos, w) && w.Light(pos) >= 8 {
		if m.Growth < 7 {
			m.Growth++
			w.PlaceBlock(pos, m)
		} else {
			directions := []world.Direction{world.North, world.South, world.West, world.East}
			for _, i := range directions {
				if _, ok := w.Block(pos.Side(i.Face())).(Melon); ok {
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
					m.Direction = direction
					w.PlaceBlock(pos, m)
					w.PlaceBlock(stemPos, Melon{})
				}
			}
		}
	}
}

// Bonemeal ...
func (m MelonSeeds) Bonemeal(pos world.BlockPos, w *world.World) bool {
	if m.Growth == 7 {
		return false
	}
	m.Growth = min(m.Growth+rand.Intn(4)+2, 7)
	w.PlaceBlock(pos, m)
	return true
}

// UseOnBlock ...
func (m MelonSeeds) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, m)
	if !used {
		return false
	}

	if _, ok := w.Block(pos.Side(world.FaceDown)).(Farmland); !ok {
		return false
	}

	place(w, pos, m, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (m MelonSeeds) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(m, 1)),
	}
}

// EncodeItem ...
func (m MelonSeeds) EncodeItem() (id int32, meta int16) {
	return 362, 0
}

// EncodeBlock ...
func (m MelonSeeds) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:melon_stem", map[string]interface{}{"facing_direction": int32(m.Direction), "growth": int32(m.Growth)}
}

// Hash ...
func (m MelonSeeds) Hash() uint64 {
	return hashMelonStem | (uint64(m.Growth) << 32) | (uint64(m.Direction) << 35)
}

// allMelonStems
func allMelonStems() (stems []world.Block) {
	for i := 0; i <= 7; i++ {
		for j := world.Face(0); j <= 5; j++ {
			stems = append(stems, MelonSeeds{Direction: j, crop: crop{Growth: i}})
		}
	}
	return
}
