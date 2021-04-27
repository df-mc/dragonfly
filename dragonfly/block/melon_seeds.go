package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// MelonSeeds grow melon blocks.
type MelonSeeds struct {
	crop

	// direction is the direction from the stem to the melon.
	Direction cube.Face
}

// SameCrop ...
func (MelonSeeds) SameCrop(c Crop) bool {
	_, ok := c.(MelonSeeds)
	return ok
}

// NeighbourUpdateTick ...
func (m MelonSeeds) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		w.BreakBlock(pos)
	} else if m.Direction != cube.FaceDown {
		if _, ok := w.Block(pos.Side(m.Direction)).(Melon); !ok {
			m.Direction = cube.FaceDown
			w.PlaceBlock(pos, m)
		}
	}
}

// RandomTick ...
func (m MelonSeeds) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	if r.Float64() <= m.CalculateGrowthChance(pos, w) && w.Light(pos) >= 8 {
		if m.Growth < 7 {
			m.Growth++
			w.PlaceBlock(pos, m)
		} else {
			directions := cube.Directions()
			for _, i := range directions {
				if _, ok := w.Block(pos.Side(i.Face())).(Melon); ok {
					return
				}
			}
			direction := directions[r.Intn(len(directions))].Face()
			stemPos := pos.Side(direction)
			if _, ok := w.Block(stemPos).(Air); ok {
				switch w.Block(stemPos.Side(cube.FaceDown)).(type) {
				case Farmland, Dirt, Grass:
					m.Direction = direction
					w.PlaceBlock(pos, m)
					w.PlaceBlock(stemPos, Melon{})
				}
			}
		}
	}
}

// BoneMeal ...
func (m MelonSeeds) BoneMeal(pos cube.Pos, w *world.World) bool {
	if m.Growth == 7 {
		return false
	}
	m.Growth = min(m.Growth+rand.Intn(4)+2, 7)
	w.PlaceBlock(pos, m)
	return true
}

// UseOnBlock ...
func (m MelonSeeds) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, m)
	if !used {
		return false
	}

	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
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

// allMelonStems ...
func allMelonStems() (stems []world.Block) {
	for i := 0; i <= 7; i++ {
		for j := cube.Face(0); j <= 5; j++ {
			stems = append(stems, MelonSeeds{crop: crop{Growth: i}, Direction: j})
		}
	}
	return
}
