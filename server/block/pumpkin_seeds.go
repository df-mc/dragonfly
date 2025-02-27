package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// PumpkinSeeds grow pumpkin blocks.
type PumpkinSeeds struct {
	crop

	// Direction is the direction from the stem to the pumpkin.
	Direction cube.Face
}

// SameCrop ...
func (PumpkinSeeds) SameCrop(c Crop) bool {
	_, ok := c.(PumpkinSeeds)
	return ok
}

// NeighbourUpdateTick ...
func (p PumpkinSeeds) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		breakBlock(p, pos, tx)
	} else if p.Direction != cube.FaceDown {
		if pumpkin, ok := tx.Block(pos.Side(p.Direction)).(Pumpkin); !ok || pumpkin.Carved {
			p.Direction = cube.FaceDown
			tx.SetBlock(pos, p, nil)
		}
	}
}

// RandomTick ...
func (p PumpkinSeeds) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if r.Float64() <= p.CalculateGrowthChance(pos, tx) && tx.Light(pos) >= 8 {
		if p.Growth < 7 {
			p.Growth++
			tx.SetBlock(pos, p, nil)
		} else {
			directions := []cube.Direction{cube.North, cube.South, cube.West, cube.East}
			for _, i := range directions {
				if _, ok := tx.Block(pos.Side(i.Face())).(Pumpkin); ok {
					return
				}
			}
			direction := directions[r.IntN(len(directions))].Face()
			stemPos := pos.Side(direction)
			if _, ok := tx.Block(stemPos).(Air); ok {
				switch tx.Block(stemPos.Side(cube.FaceDown)).(type) {
				case Farmland, Dirt, Grass:
					p.Direction = direction
					tx.SetBlock(pos, p, nil)
					tx.SetBlock(stemPos, Pumpkin{}, nil)
				}
			}
		}
	}
}

// BoneMeal ...
func (p PumpkinSeeds) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	if p.Growth == 7 {
		return false
	}
	p.Growth = min(p.Growth+rand.IntN(4)+2, 7)
	tx.SetBlock(pos, p, nil)
	return true
}

// UseOnBlock ...
func (p PumpkinSeeds) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, p)
	if !used {
		return false
	}

	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		return false
	}

	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (p PumpkinSeeds) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(p))
}

// CompostChance ...
func (PumpkinSeeds) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (p PumpkinSeeds) EncodeItem() (name string, meta int16) {
	return "minecraft:pumpkin_seeds", 0
}

// EncodeBlock ...
func (p PumpkinSeeds) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:pumpkin_stem", map[string]any{"facing_direction": int32(p.Direction), "growth": int32(p.Growth)}
}

// allPumpkinStems
func allPumpkinStems() (stems []world.Block) {
	for i := 0; i <= 7; i++ {
		for j := cube.Face(0); j <= 5; j++ {
			stems = append(stems, PumpkinSeeds{Direction: j, crop: crop{Growth: i}})
		}
	}
	return
}
