package block

import (
	"math/rand"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Bush is a transparent plant block which can be used to obtain seeds and as decoration.
type TallDryGrass struct {
	replaceable
	transparent
	empty
}

// FlammabilityInfo ...
func (t TallDryGrass) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (t TallDryGrass) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, func(t item.Tool) bool {
		return t.ToolType() == item.TypeShears
	}, nothingEffective, oneOf(t))
}

// BoneMeal attempts to affect the block using a bone meal item.
// It picks a random horizontal side and attempts to place a short dry grass there.
func (t TallDryGrass) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	// Define the 4 horizontal faces.
	faces := []cube.Face{cube.FaceNorth, cube.FaceSouth, cube.FaceEast, cube.FaceWest}

	// Shuffle the faces to ensure randomness.
	rand.Shuffle(len(faces), func(i, j int) {
		faces[i], faces[j] = faces[j], faces[i]
	})

	// Iterate through shuffled faces, but return true as soon as ONE is successfully placed.
	for _, face := range faces {
		sidePos := pos.Side(face)

		// 1. Check if the block already there is a ShortDryGrass.
		// If it is, we skip this side so we don't "replace" it with itself and waste bone meal.
		if _, ok := tx.Block(sidePos).(ShortDryGrass); ok {
			continue
		}

		// 2. Check if the target position is replaceable (air, tall grass, etc.)
		if _, _, used := firstReplaceable(tx, sidePos, cube.FaceDown, t); !used {
			continue
		}

		// 3. Check if the block BELOW the target position is solid.
		if !tx.Block(sidePos.Side(cube.FaceDown)).Model().FaceSolid(sidePos.Side(cube.FaceDown), cube.FaceUp, tx) {
			continue
		}

		// 4. Place the block and return true (consuming the bone meal)
		tx.SetBlock(sidePos, ShortDryGrass{}, nil)
		return true
	}

	return false
}

// FuelInfo ...
func (t TallDryGrass) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 5)
}

// CompostChance ...
func (t TallDryGrass) CompostChance() float64 {
	return 0.3
}

// NeighbourUpdateTick ...
func (t TallDryGrass) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		breakBlock(t, pos, tx)
	}
}

// HasLiquidDrops ...
func (t TallDryGrass) HasLiquidDrops() bool {
	return false
}

// UseOnBlock ...
func (t TallDryGrass) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, t)
	if !used {
		return false
	}
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		return false
	}

	place(tx, pos, t, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (t TallDryGrass) EncodeItem() (name string, meta int16) {
	return "minecraft:tall_dry_grass", 0
}

// EncodeBlock ...
func (t TallDryGrass) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:tall_dry_grass", nil
}
