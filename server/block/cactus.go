package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// Cactus is a plant block that generates naturally in dry areas and causes damage.
type Cactus struct {
	transparent

	// Age is the growth state of cactus. Values range from 0 to 15.
	Age int
}

// UseOnBlock handles making sure the neighbouring blocks are air.
func (c Cactus) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, c)
	if !used || !c.canGrowHere(pos, tx, true) {
		return false
	}

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (c Cactus) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !c.canGrowHere(pos, tx, true) {
		breakBlock(c, pos, tx)
	}
}

// RandomTick ...
func (c Cactus) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if c.Age < 15 {
		c.Age++
	} else if c.Age == 15 {
		c.Age = 0
		if c.canGrowHere(pos.Side(cube.FaceDown), tx, false) {
			for y := 1; y < 3; y++ {
				if _, ok := tx.Block(pos.Add(cube.Pos{0, y})).(Air); ok {
					tx.SetBlock(pos.Add(cube.Pos{0, y}), Cactus{Age: 0}, nil)
					break
				} else if _, ok := tx.Block(pos.Add(cube.Pos{0, y})).(Cactus); !ok {
					break
				}
			}
		}
	}
	tx.SetBlock(pos, c, nil)
}

// canGrowHere implements logic to check if cactus can live/grow here.
func (c Cactus) canGrowHere(pos cube.Pos, tx *world.Tx, recursive bool) bool {
	for _, face := range cube.HorizontalFaces() {
		if _, ok := tx.Block(pos.Side(face)).(Air); !ok {
			return false
		}
	}
	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Cactus); ok && recursive {
		return c.canGrowHere(pos.Side(cube.FaceDown), tx, recursive)
	}
	return supportsVegetation(c, tx.Block(pos.Sub(cube.Pos{0, 1})))
}

// EntityInside ...
func (c Cactus) EntityInside(_ cube.Pos, _ *world.Tx, e world.Entity) {
	if l, ok := e.(livingEntity); ok {
		l.Hurt(0.5, DamageSource{Block: c})
	}
}

// BreakInfo ...
func (c Cactus) BreakInfo() BreakInfo {
	return newBreakInfo(0.4, alwaysHarvestable, nothingEffective, oneOf(c))
}

// CompostChance ...
func (Cactus) CompostChance() float64 {
	return 0.5
}

// EncodeItem ...
func (c Cactus) EncodeItem() (name string, meta int16) {
	return "minecraft:cactus", 0
}

// EncodeBlock ...
func (c Cactus) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:cactus", map[string]any{"age": int32(c.Age)}
}

// Model ...
func (c Cactus) Model() world.BlockModel {
	return model.Cactus{}
}

// allCactus returns all possible states of a cactus block.
func allCactus() (b []world.Block) {
	for i := 0; i < 16; i++ {
		b = append(b, Cactus{Age: i})
	}
	return
}

// DamageSource is passed as world.DamageSource for damage caused by a block,
// such as a cactus or a falling anvil.
type DamageSource struct {
	// Block is the block that caused the damage.
	Block world.Block
}

func (DamageSource) ReducedByResistance() bool { return true }
func (DamageSource) ReducedByArmour() bool     { return true }
func (DamageSource) Fire() bool                { return false }
func (DamageSource) IgnoreTotem() bool         { return false }
