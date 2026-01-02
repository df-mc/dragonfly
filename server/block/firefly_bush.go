package block

import (
	"math/rand"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Bush is a transparent plant block which can be used to obtain seeds and as decoration.
type FireflyBush struct {
	transparent
	empty
}

// FlammabilityInfo ...
func (f FireflyBush) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (f FireflyBush) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, nothingEffective, oneOf(f))
}

// LightEmissionLevel ...
func (f FireflyBush) LightEmissionLevel() uint8 {
	return 2
}

// BoneMeal attempts to affect the block using a bone meal item.
// It picks a random horizontal side and attempts to place another bush there.
func (f FireflyBush) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	// Define the 4 horizontal faces.
	faces := []cube.Face{cube.FaceNorth, cube.FaceSouth, cube.FaceEast, cube.FaceWest}

	// Shuffle the faces to ensure randomness.
	rand.Shuffle(len(faces), func(i, j int) {
		faces[i], faces[j] = faces[j], faces[i]
	})

	// Iterate through shuffled faces, but return true as soon as ONE is successfully placed.
	for _, face := range faces {
		sidePos := pos.Side(face)

		// 1. Check if the target position is replaceable (air, tall grass, etc.)
		if _, _, used := firstReplaceable(tx, sidePos, cube.FaceDown, f); !used {
			continue // Try the next random face if this one is blocked
		}

		// 2. Check if the block BELOW the target position is solid (supports the bush)
		// This uses the same logic found in your UseOnBlock and NeighbourUpdateTick.
		if !tx.Block(sidePos.Side(cube.FaceDown)).Model().FaceSolid(sidePos.Side(cube.FaceDown), cube.FaceUp, tx) {
			continue // Try the next random face if the ground isn't solid
		}

		// 3. Place the block and return true (consuming the bone meal)
		tx.SetBlock(sidePos, f, nil)
		return true
	}

	return false
}

// CompostChance ...
func (f FireflyBush) CompostChance() float64 {
	return 0.3
}

// NeighbourUpdateTick ...
func (f FireflyBush) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		breakBlock(f, pos, tx)
	}
}

// HasLiquidDrops ...
func (f FireflyBush) HasLiquidDrops() bool {
	return true
}

// UseOnBlock ...
func (f FireflyBush) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, f)
	if !used {
		return false
	}
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		return false
	}

	place(tx, pos, f, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (f FireflyBush) EncodeItem() (name string, meta int16) {
	return "minecraft:firefly_bush", 0
}

// EncodeBlock ...
func (f FireflyBush) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:firefly_bush", nil
}
