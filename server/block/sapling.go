package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Sapling is a transparent plant block that can grow into a tree.
type Sapling struct {
	replaceable
	transparent
	empty

	// Wood is the type of wood of the sapling.
	Wood WoodType
	// Stage is the current growth stage of a sapling. Only non-mangrove saplings use this field.
	Stage int
	// Age is the propagule stage of a mangrove propagule. Values range from 0-4.
	Age int
	// Hanging specifies if a mangrove propagule is hanging from mangrove leaves.
	Hanging bool
}

// RandomTick advances sapling growth or attempts tree generation during a random tick.
func (s Sapling) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if !s.canSurvive(pos, tx) {
		breakBlock(s, pos, tx)
		return
	}
	if s.Wood == MangroveWood() {
		if s.Hanging {
			if s.Age < 4 {
				s.Age++
				tx.SetBlock(pos, s, nil)
			}
			return
		}
		if r.IntN(7) == 0 {
			s.advanceTree(pos, tx, r)
		}
		return
	}
	if tx.Light(pos.Side(cube.FaceUp)) >= 9 && r.IntN(7) == 0 {
		s.advanceTree(pos, tx, r)
	}
}

// BoneMeal advances the sapling or propagule using bone meal.
func (s Sapling) BoneMeal(pos cube.Pos, creative bool, tx *world.Tx) bool {
	if s.Wood == MangroveWood() && s.Hanging {
		if s.Age >= 4 {
			return false
		}
		s.Age++
		tx.SetBlock(pos, s, nil)
		return true
	}
	if !creative && rand.Float64() >= 0.45 {
		return true
	}
	r := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	if creative && s.Wood != MangroveWood() && s.Stage == 0 {
		return s.growTree(pos, tx, r)
	}
	return s.advanceTree(pos, tx, r)
}

// NeighbourUpdateTick breaks the sapling when its support block is removed.
func (s Sapling) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !s.canSurvive(pos, tx) {
		breakBlock(s, pos, tx)
	}
}

// UseOnBlock places the sapling and normalises its initial growth state.
func (s Sapling) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, s)
	if !used {
		return false
	}
	s = s.placedState()
	if !s.canSurvive(pos, tx) {
		return false
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// HasLiquidDrops reports that saplings drop their item when replaced by liquid.
func (Sapling) HasLiquidDrops() bool {
	return true
}

// FlammabilityInfo returns the fire behaviour of a sapling.
func (s Sapling) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, true)
}

// BreakInfo returns the breaking behaviour and drops of a sapling.
func (s Sapling) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(s.placedState()))
}

// CompostChance returns the chance of a sapling increasing a composter's level.
func (Sapling) CompostChance() float64 {
	return 0.3
}

// EncodeItem returns the item identifier used for the sapling.
func (s Sapling) EncodeItem() (name string, meta int16) {
	switch s.Wood {
	case MangroveWood():
		return "minecraft:mangrove_propagule", 0
	default:
		return "minecraft:" + s.Wood.String() + "_sapling", 0
	}
}

// EncodeBlock returns the Bedrock runtime state used for the sapling block.
func (s Sapling) EncodeBlock() (name string, properties map[string]any) {
	switch s.Wood {
	case MangroveWood():
		return "minecraft:mangrove_propagule", map[string]any{"hanging": boolByte(s.Hanging), "propagule_stage": int32(s.Age)}
	default:
		return "minecraft:" + s.Wood.String() + "_sapling", map[string]any{"age_bit": uint8(s.Stage)}
	}
}

// placedState returns the normal in-world state used when a sapling item is placed.
func (s Sapling) placedState() Sapling {
	s.Stage = 0
	s.Hanging = false
	if s.Wood == MangroveWood() {
		s.Age = 4
	} else {
		s.Age = 0
	}
	return s
}

// canSurvive checks if the sapling is still supported at its current position.
func (s Sapling) canSurvive(pos cube.Pos, tx *world.Tx) bool {
	if s.Wood == MangroveWood() && s.Hanging {
		leaves, ok := tx.Block(pos.Side(cube.FaceUp)).(Leaves)
		return ok && leaves.Wood == MangroveWood()
	}
	if s.Wood == MangroveWood() {
		if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Clay); ok {
			return true
		}
	}
	return supportsVegetation(s, tx.Block(pos.Side(cube.FaceDown)))
}

// advanceTree moves the sapling to its next stage or attempts to grow a tree.
func (s Sapling) advanceTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if s.Wood != MangroveWood() && s.Stage == 0 {
		s.Stage = 1
		tx.SetBlock(pos, s, nil)
		return true
	}
	return s.growTree(pos, tx, r)
}

// allSaplings returns all sapling variants with valid Bedrock states.
func allSaplings() (saplings []world.Block) {
	for _, w := range WoodTypes() {
		switch w {
		case CrimsonWood(), WarpedWood():
			continue
		case MangroveWood():
			for age := 0; age <= 4; age++ {
				saplings = append(saplings, Sapling{Wood: w, Age: age})
				saplings = append(saplings, Sapling{Wood: w, Age: age, Hanging: true})
			}
		default:
			saplings = append(saplings, Sapling{Wood: w})
			saplings = append(saplings, Sapling{Wood: w, Stage: 1})
		}
	}
	return
}
