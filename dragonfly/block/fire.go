package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/fire"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/difficulty"
	"math/rand"
	"time"
)

// Fire is a non-solid block that can spread to nearby flammable blocks.
type Fire struct {
	noNBT
	replaceable
	transparent
	empty

	// Type is the type of fire.
	Type fire.Fire
	// Age affects how fire extinguishes. Newly placed fire starts at 0 and the value has a 1/3 chance of incrementing
	// each block tick.
	Age int
}

// FlammableBlock returns true if a block is flammable.
func FlammableBlock(block world.Block) bool {
	if flammable, ok := block.(Flammable); ok && flammable.FlammabilityInfo().Encouragement > 0 {
		return true
	}
	return false
}

// woodTypeFlammable returns whether a type of wood is flammable.
func woodTypeFlammable(w wood.Wood) bool {
	switch w {
	case wood.Crimson(), wood.Warped():
		return false
	default:
		return true
	}
}

// neighbourFlammable returns true if one a block adjacent to the passed position is flammable.
func neighbourFlammable(pos world.BlockPos, w *world.World) bool {
	for i := world.Face(0); i < 6; i++ {
		if FlammableBlock(w.Block(pos.Side(i))) {
			return true
		}
	}
	return false
}

// max ...
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// difficultyOffset ...
func difficultyOffset(d difficulty.Difficulty) int {
	switch d.(type) {
	case difficulty.Peaceful:
		return 0
	case difficulty.Easy:
		return 7
	case difficulty.Normal:
		return 14
	case difficulty.Hard:
		return 21
	}
	panic("unknown difficulty")
}

// infinitelyBurning returns true if fire can infinitely burn at the specified position.
func infinitelyBurning(pos world.BlockPos, w *world.World) bool {
	switch block := w.Block(pos.Side(world.FaceDown)).(type) {
	//TODO: Magma Block
	case Netherrack:
		return true
	case Bedrock:
		return block.InfiniteBurning
	}
	return false
}

// burn attempts to burn a block.
func (f Fire) burn(pos world.BlockPos, w *world.World, chanceBound int) {
	if flammable, ok := w.Block(pos).(Flammable); ok && rand.Intn(chanceBound) < flammable.FlammabilityInfo().Flammability {
		//TODO: Check if not raining
		if rand.Intn(f.Age+10) < 5 {
			age := f.Age + rand.Intn(5)/4
			if age > 15 {
				age = 15
			}
			w.PlaceBlock(pos, Fire{Type: f.Type, Age: age})
		} else {
			w.BreakBlockWithoutParticles(pos)
		}
		//TODO: Light TNT
	}
}

// tick ...
func (f Fire) tick(pos world.BlockPos, w *world.World) {
	if f.Type == fire.Normal() {
		infinitelyBurns := infinitelyBurning(pos, w)

		// TODO: !infinitelyBurning && raining && exposed to rain && 20 + age * 3% = extinguish & return

		if f.Age < 15 && rand.Intn(3) == 0 {
			f.Age++
			w.PlaceBlock(pos, f)
		}

		w.ScheduleBlockUpdate(pos, time.Duration(30+rand.Intn(10))*time.Second/20)

		if !infinitelyBurns {
			if !neighbourFlammable(pos, w) {
				if !w.Block(pos.Side(world.FaceDown)).Model().FaceSolid(pos, world.FaceUp, w) || f.Age > 3 {
					w.BreakBlockWithoutParticles(pos)
				}
				return
			}
			if !FlammableBlock(w.Block(pos.Side(world.FaceDown))) && f.Age == 15 && rand.Intn(4) == 0 {
				w.BreakBlockWithoutParticles(pos)
				return
			}
		}

		//TODO: If high humidity, chance should be subtracted by 50
		for face := world.Face(0); face < 6; face++ {
			if face == world.FaceUp || face == world.FaceDown {
				f.burn(pos.Side(face), w, 300)
			} else {
				f.burn(pos.Side(face), w, 250)
			}
		}

		for y := -1; y <= 4; y++ {
			randomBound := 100
			if y > 1 {
				randomBound += (y - 1) * 100
			}

			for x := -1; x <= 1; x++ {
				for z := -1; z <= 1; z++ {
					if x != 0 || y != 0 || z != 0 {
						blockPos := pos.Add(world.BlockPos{x, y, z})
						block := w.Block(blockPos)
						if _, ok := block.(Air); !ok {
							continue
						}

						encouragement := 0
						blockPos.Neighbours(func(neighbour world.BlockPos) {
							if flammable, ok := w.Block(neighbour).(Flammable); ok {
								encouragement = max(encouragement, flammable.FlammabilityInfo().Encouragement)
							}
						})
						if encouragement <= 0 {
							continue
						}

						//TODO: Divide chance by 2 in high humidity
						maxChance := (encouragement + 40 + difficultyOffset(w.Difficulty())) / (f.Age + 30)

						//TODO: Check if exposed to rain
						if maxChance > 0 && rand.Intn(randomBound) <= maxChance {
							age := f.Age + rand.Intn(5)/4
							if age > 15 {
								age = 15
							}
							w.PlaceBlock(blockPos, Fire{Type: f.Type, Age: age})
						}
					}
				}
			}
		}
	}
}

// EntityCollide ...
func (f Fire) EntityCollide(e world.Entity) {
	if flammable, ok := e.(entity.Flammable); ok {
		flammable.FireDamage(1)
		flammable.SetFireTicks(160)
	}
}

// ScheduledTick ...
func (f Fire) ScheduledTick(pos world.BlockPos, w *world.World) {
	f.tick(pos, w)
}

// RandomTick ...
func (f Fire) RandomTick(pos world.BlockPos, w *world.World, _ *rand.Rand) {
	f.tick(pos, w)
}

// NeighbourUpdateTick ...
func (f Fire) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	below := w.Block(pos.Side(world.FaceDown))
	if !below.Model().FaceSolid(pos, world.FaceUp, w) && (!neighbourFlammable(pos, w) || f.Type == fire.Soul()) {
		w.BreakBlockWithoutParticles(pos)
	} else {
		switch below.(type) {
		case SoulSand, SoulSoil:
			f.Type = fire.Soul()
			w.PlaceBlock(pos, f)
		default:
			if f.Type == fire.Soul() {
				w.BreakBlockWithoutParticles(pos)
				return
			}
		}
		w.ScheduleBlockUpdate(pos, time.Duration(30+rand.Intn(10))*time.Second/20)
	}
}

// HasLiquidDrops ...
func (f Fire) HasLiquidDrops() bool {
	return false
}

// LightEmissionLevel ...
func (f Fire) LightEmissionLevel() uint8 {
	return f.Type.LightLevel
}

// EncodeBlock ...
func (f Fire) EncodeBlock() (name string, properties map[string]interface{}) {
	switch f.Type {
	case fire.Normal():
		return "minecraft:fire", map[string]interface{}{"age": int32(f.Age)}
	case fire.Soul():
		return "minecraft:soul_fire", map[string]interface{}{"age": int32(f.Age)}
	}
	panic("unknown fire type")
}

// Hash ...
func (f Fire) Hash() uint64 {
	return hashFire | (uint64(f.Age) << 32) | (uint64(f.Type.Uint8()) << 36)
}

// allFire ...
func allFire() (b []world.Block) {
	for i := 0; i < 16; i++ {
		b = append(b, Fire{Age: i, Type: fire.Normal()})
		b = append(b, Fire{Age: i, Type: fire.Soul()})
	}
	return
}
