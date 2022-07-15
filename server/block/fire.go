package block

//lint:file-ignore ST1022 Exported variables in this package have compiler directives. These variables are not otherwise exposed to users.

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
	"time"
)

// Fire is a non-solid block that can spread to nearby flammable blocks.
type Fire struct {
	replaceable
	transparent
	empty

	// Type is the type of fire.
	Type FireType
	// Age affects how fire extinguishes. Newly placed fire starts at 0 and the value has a 1/3 chance of incrementing
	// each block tick.
	Age int
}

// flammableBlock returns true if a block is flammable.
func flammableBlock(block world.Block) bool {
	flammable, ok := block.(Flammable)
	return ok && flammable.FlammabilityInfo().Encouragement > 0
}

// neighboursFlammable returns true if one a block adjacent to the passed position is flammable.
func neighboursFlammable(pos cube.Pos, w *world.World) bool {
	for _, i := range cube.Faces() {
		if flammableBlock(w.Block(pos.Side(i))) {
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

// infinitelyBurning returns true if fire can infinitely burn at the specified position.
func infinitelyBurning(pos cube.Pos, w *world.World) bool {
	switch block := w.Block(pos.Side(cube.FaceDown)).(type) {
	//TODO: Magma Block
	case Netherrack:
		return true
	case Bedrock:
		return block.InfiniteBurning
	}
	return false
}

// burn attempts to burn a block.
func (f Fire) burn(from, to cube.Pos, w *world.World, r *rand.Rand, chanceBound int) {
	if flammable, ok := w.Block(to).(Flammable); ok && r.Intn(chanceBound) < flammable.FlammabilityInfo().Flammability {
		if r.Intn(f.Age+10) < 5 && !rainingAround(to, w) {
			f.spread(from, to, w, r)
		} else {
			w.SetBlock(to, nil, nil)
		}
		//TODO: Light TNT
	}
}

// rainingAround checks if it is raining either at the cube.Pos passed or at any of its horizontal neighbours.
func rainingAround(pos cube.Pos, w *world.World) bool {
	raining := w.RainingAt(pos)
	for _, face := range cube.HorizontalFaces() {
		if raining {
			break
		}
		raining = w.RainingAt(pos.Side(face))
	}
	return raining
}

// tick ...
func (f Fire) tick(pos cube.Pos, w *world.World, r *rand.Rand) {
	if f.Type == SoulFire() {
		return
	}
	infinitelyBurns := infinitelyBurning(pos, w)
	if !infinitelyBurns && (20+f.Age*3) > r.Intn(100) && rainingAround(pos, w) {
		// Fire is extinguished by the rain.
		w.SetBlock(pos, nil, nil)
		return
	}

	if f.Age < 15 && r.Intn(3) == 0 {
		f.Age++
		w.SetBlock(pos, f, nil)
	}

	w.ScheduleBlockUpdate(pos, time.Duration(30+r.Intn(10))*time.Second/20)

	if !infinitelyBurns {
		_, waterBelow := w.Block(pos.Side(cube.FaceDown)).(Water)
		if waterBelow {
			w.SetBlock(pos, nil, nil)
			return
		}
		if !neighboursFlammable(pos, w) {
			if !w.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos, cube.FaceUp, w) || f.Age > 3 {
				w.SetBlock(pos, nil, nil)
			}
			return
		}
		if !flammableBlock(w.Block(pos.Side(cube.FaceDown))) && f.Age == 15 && r.Intn(4) == 0 {
			w.SetBlock(pos, nil, nil)
			return
		}
	}

	//TODO: If high humidity, chance should be subtracted by 50
	for face := cube.Face(0); face < 6; face++ {
		if face == cube.FaceUp || face == cube.FaceDown {
			f.burn(pos, pos.Side(face), w, r, 300)
		} else {
			f.burn(pos, pos.Side(face), w, r, 250)
		}
	}

	for y := -1; y <= 4; y++ {
		randomBound := 100
		if y > 1 {
			randomBound += (y - 1) * 100
		}

		for x := -1; x <= 1; x++ {
			for z := -1; z <= 1; z++ {
				if x == 0 && y == 0 && z == 0 {
					continue
				}
				blockPos := pos.Add(cube.Pos{x, y, z})
				block := w.Block(blockPos)
				if _, ok := block.(Air); !ok {
					continue
				}

				encouragement := 0
				blockPos.Neighbours(func(neighbour cube.Pos) {
					if flammable, ok := w.Block(neighbour).(Flammable); ok {
						encouragement = max(encouragement, flammable.FlammabilityInfo().Encouragement)
					}
				}, w.Range())
				if encouragement <= 0 {
					continue
				}

				//TODO: Divide chance by 2 in high humidity
				maxChance := (encouragement + 40 + w.Difficulty().FireSpreadIncrease()) / (f.Age + 30)

				if maxChance > 0 && r.Intn(randomBound) <= maxChance && !rainingAround(blockPos, w) {
					f.spread(pos, blockPos, w, r)
				}
			}
		}
	}
}

// spread attempts to spread fire from a cube.Pos to another. If the block burn or fire spreading events are cancelled,
// this might end up not happening.
func (f Fire) spread(from, to cube.Pos, w *world.World, r *rand.Rand) {
	if _, air := w.Block(to).(Air); !air {
		ctx := event.C()
		if w.Handler().HandleBlockBurn(ctx, to); ctx.Cancelled() {
			return
		}
	}
	ctx := event.C()
	if w.Handler().HandleFireSpread(ctx, from, to); ctx.Cancelled() {
		return
	}
	w.SetBlock(to, Fire{Type: f.Type, Age: min(15, f.Age+r.Intn(5)/4)}, nil)
	w.ScheduleBlockUpdate(to, time.Duration(30+r.Intn(10))*time.Second/20)
}

// EntityInside ...
func (f Fire) EntityInside(_ cube.Pos, _ *world.World, e world.Entity) {
	if flammable, ok := e.(entity.Flammable); ok {
		if l, ok := e.(entity.Living); ok && !l.AttackImmune() {
			l.Hurt(f.Type.Damage(), damage.SourceFire{})
		}
		if flammable.OnFireDuration() < time.Second*8 {
			flammable.SetOnFire(8 * time.Second)
		}
	}
}

// ScheduledTick ...
func (f Fire) ScheduledTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	f.tick(pos, w, r)
}

// RandomTick ...
func (f Fire) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	f.tick(pos, w, r)
}

// NeighbourUpdateTick ...
func (f Fire) NeighbourUpdateTick(pos, neighbour cube.Pos, w *world.World) {
	below := w.Block(pos.Side(cube.FaceDown))
	if diffuser, ok := below.(LightDiffuser); (ok && diffuser.LightDiffusionLevel() != 15) && (!neighboursFlammable(pos, w) || f.Type == SoulFire()) {
		w.SetBlock(pos, nil, nil)
		return
	}
	switch below.(type) {
	case SoulSand, SoulSoil:
		f.Type = SoulFire()
		w.SetBlock(pos, f, nil)
	case Water:
		if neighbour == pos {
			w.SetBlock(pos, nil, nil)
		}
	default:
		if f.Type == SoulFire() {
			w.SetBlock(pos, nil, nil)
			return
		}
	}
}

// HasLiquidDrops ...
func (f Fire) HasLiquidDrops() bool {
	return false
}

// LightEmissionLevel ...
func (f Fire) LightEmissionLevel() uint8 {
	return f.Type.LightLevel()
}

// EncodeBlock ...
func (f Fire) EncodeBlock() (name string, properties map[string]any) {
	switch f.Type {
	case NormalFire():
		return "minecraft:fire", map[string]any{"age": int32(f.Age)}
	case SoulFire():
		return "minecraft:soul_fire", map[string]any{"age": int32(f.Age)}
	}
	panic("unknown fire type")
}

// Start starts a fire at a position in the world. The position passed must be either air or tall grass and conditions
// for a fire to be present must be present.
func (f Fire) Start(w *world.World, pos cube.Pos) {
	b := w.Block(pos)
	_, isAir := b.(Air)
	_, isTallGrass := b.(TallGrass)
	if isAir || isTallGrass {
		below := w.Block(pos.Side(cube.FaceDown))
		if below.Model().FaceSolid(pos, cube.FaceUp, w) || neighboursFlammable(pos, w) {
			w.SetBlock(pos, Fire{}, nil)
		}
	}
}

// allFire ...
func allFire() (b []world.Block) {
	for i := 0; i < 16; i++ {
		b = append(b, Fire{Age: i, Type: NormalFire()})
		b = append(b, Fire{Age: i, Type: SoulFire()})
	}
	return
}
