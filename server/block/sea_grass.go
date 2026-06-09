package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

var seaFlora = []world.Block{
	Coral{Type: TubeCoral()},
	Coral{Type: BrainCoral()},
	Coral{Type: BubbleCoral()},
	Coral{Type: FireCoral()},
	Coral{Type: HornCoral()},
	// TODO: coral fans.
}

type SeaGrass struct {
	empty
	replaceable
	transparent
	sourceOrFallingWaterDisplacer

	// Type is the type of the seagrass.
	Type SeaGrassType
}

func (s SeaGrass) HasLiquidDrops() bool {
	return false
}

func (s SeaGrass) SideClosed(_, _ cube.Pos, _ *world.Tx) bool {
	return false
}

func (s SeaGrass) BoneMeal(pos cube.Pos, tx *world.Tx) item.BoneMealResult {
	if liquid, ok := tx.Liquid(pos.Side(cube.FaceUp)); !ok || !s.CanDisplace(liquid) {
		return item.BoneMealResultNone
	}
	top := SeaGrass{Type: DoubleTopSeaGrass()}
	if replaceableWith(tx, pos.Side(cube.FaceUp), top) && s.Type == DefaultSeaGrass() {
		tx.SetBlock(pos, SeaGrass{Type: DoubleBottomSeaGrass()}, nil)
		tx.SetBlock(pos.Side(cube.FaceUp), top, nil)
		return item.BoneMealResultSmall
	}

	return item.BoneMealResultNone
}

// UseOnBlock ...
func (s SeaGrass) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, s)
	if !used {
		return false
	}
	if liquid, ok := tx.Liquid(pos); !ok || liquid.LiquidDepth() != 8 {
		return false
	}
	if !canSeaGrassStay(pos.Side(cube.FaceDown), tx) {
		return false
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

func (s SeaGrass) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if liquid, ok := tx.Liquid(pos); !ok || liquid.LiquidDepth() != 8 {
		breakBlockNoDrops(s, pos, tx)
		return
	}

	if s.Type == DoubleTopSeaGrass() {
		if bottom, ok := tx.Block(pos.Side(cube.FaceDown)).(SeaGrass); !ok || bottom.Type != DoubleBottomSeaGrass() {
			breakBlockNoDrops(s, pos, tx)
		}
		return
	} else if s.Type == DoubleBottomSeaGrass() {
		if upper, ok := tx.Block(pos.Side(cube.FaceUp)).(SeaGrass); !ok || upper.Type != DoubleTopSeaGrass() {
			breakBlockNoDrops(s, pos, tx)
			return
		}
	}
	if !canSeaGrassStay(pos.Side(cube.FaceDown), tx) {
		breakBlockNoDrops(s, pos, tx)
	}
}

func (s SeaGrass) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(tool item.Tool, enchantments []item.Enchantment) []item.Stack {
		if tool.ToolType() == item.TypeShears {
			return []item.Stack{item.NewStack(SeaGrass{Type: DefaultSeaGrass()}, 1)}
		}
		return nil
	})
}

func (s SeaGrass) CompostChance() float64 {
	return 0.3
}

func (s SeaGrass) EncodeItem() (name string, meta int16) {
	return "minecraft:seagrass", 0
}

func (s SeaGrass) EncodeBlock() (string, map[string]any) {
	return "minecraft:seagrass", map[string]any{"sea_grass_type": s.Type.String()}
}

func canSeaGrassStay(pos cube.Pos, tx *world.Tx) bool {
	block := tx.Block(pos)
	switch block.(type) {
	case SoulSand, Leaves:
		return false
	}
	return block.Model().FaceSolid(pos, cube.FaceUp, tx)
}

func trySpreadSeaFlora(pos cube.Pos, tx *world.Tx) (result item.BoneMealResult) {
	result = item.BoneMealResultNone
	for range 16 {
		dy := (rand.IntN(3) - 1) * rand.IntN(3) / 2
		candidate := pos.Add(cube.Pos{rand.IntN(3) - 1, dy, rand.IntN(3) - 1})
		above := candidate.Side(cube.FaceUp)

		switch tx.Block(candidate).(type) {
		case Dirt, Sand, Gravel, Clay:
		default:
			continue
		}

		if _, ok := tx.Block(above).(Water); !ok {
			continue
		}

		aboveAbove := above.Side(cube.FaceUp)
		existing := tx.Block(above)
		if _, ok := existing.(SeaGrass); ok {
			continue
		}
		result = item.BoneMealResultArea

		if rand.IntN(8) == 0 {
			if liquid, ok := tx.Liquid(aboveAbove); !ok || liquid.LiquidDepth() != 8 {
				continue
			}
			tx.SetBlock(above, SeaGrass{Type: DoubleBottomSeaGrass()}, nil)
			tx.SetBlock(aboveAbove, SeaGrass{Type: DoubleTopSeaGrass()}, nil)
			continue
		}

		if rand.IntN(3) == 0 {
			tx.SetBlock(above, seaFlora[rand.IntN(len(seaFlora))], nil)
			continue
		}
		tx.SetBlock(above, SeaGrass{Type: DefaultSeaGrass()}, nil)
	}
	return
}

// allSeaGrass ...
func allSeaGrass() (b []world.Block) {
	for _, s := range SeaGrassTypes() {
		b = append(b, SeaGrass{Type: s})
	}
	return
}
