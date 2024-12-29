package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
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
func (m MelonSeeds) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		breakBlock(m, pos, tx)
	} else if m.Direction != cube.FaceDown {
		if _, ok := tx.Block(pos.Side(m.Direction)).(Melon); !ok {
			m.Direction = cube.FaceDown
			tx.SetBlock(pos, m, nil)
		}
	}
}

// RandomTick ...
func (m MelonSeeds) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if r.Float64() <= m.CalculateGrowthChance(pos, tx) && tx.Light(pos) >= 8 {
		if m.Growth < 7 {
			m.Growth++
			tx.SetBlock(pos, m, nil)
		} else {
			directions := cube.Directions()
			for _, i := range directions {
				if _, ok := tx.Block(pos.Side(i.Face())).(Melon); ok {
					return
				}
			}
			direction := directions[r.IntN(len(directions))].Face()
			stemPos := pos.Side(direction)
			if _, ok := tx.Block(stemPos).(Air); ok {
				switch tx.Block(stemPos.Side(cube.FaceDown)).(type) {
				case Farmland, Dirt, Grass:
					m.Direction = direction
					tx.SetBlock(pos, m, nil)
					tx.SetBlock(stemPos, Melon{}, nil)
				}
			}
		}
	}
}

// BoneMeal ...
func (m MelonSeeds) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	if m.Growth == 7 {
		return false
	}
	m.Growth = min(m.Growth+rand.IntN(4)+2, 7)
	tx.SetBlock(pos, m, nil)
	return true
}

// UseOnBlock ...
func (m MelonSeeds) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, m)
	if !used {
		return false
	}

	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		return false
	}

	place(tx, pos, m, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (m MelonSeeds) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(m))
}

// CompostChance ...
func (MelonSeeds) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (m MelonSeeds) EncodeItem() (name string, meta int16) {
	return "minecraft:melon_seeds", 0
}

// EncodeBlock ...
func (m MelonSeeds) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:melon_stem", map[string]any{"facing_direction": int32(m.Direction), "growth": int32(m.Growth)}
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
