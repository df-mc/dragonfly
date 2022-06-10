package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"math/rand"
	"time"
)

// Lava is a light-emitting fluid block that causes fire damage.
type Lava struct {
	empty
	replaceable

	// Still makes the lava not spread whenever it is updated. Still lava cannot be acquired in the game
	// without world editing.
	Still bool
	// Depth is the depth of the water. This is a number from 1-8, where 8 is a source block and 1 is the
	// smallest possible lava block.
	Depth int
	// Falling specifies if the lava is falling. Falling lava will always appear as a source block, but its
	// behaviour differs when it starts spreading.
	Falling bool
}

// neighboursLavaFlammable returns true if one a block adjacent to the passed position is flammable.
func neighboursLavaFlammable(pos cube.Pos, w *world.World) bool {
	for i := cube.Face(0); i < 6; i++ {
		if flammable, ok := w.Block(pos.Side(i)).(Flammable); ok && flammable.FlammabilityInfo().LavaFlammable {
			return true
		}
	}
	return false
}

// EntityInside ...
func (l Lava) EntityInside(_ cube.Pos, _ *world.World, e world.Entity) {
	if fallEntity, ok := e.(fallDistanceEntity); ok {
		fallEntity.ResetFallDistance()
	}
	if flammable, ok := e.(entity.Flammable); ok {
		if l, ok := e.(entity.Living); ok && !l.AttackImmune() {
			l.Hurt(4, damage.SourceLava{})
		}
		flammable.SetOnFire(15 * time.Second)
	}
}

// RandomTick ...
func (l Lava) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	i := r.Intn(3)
	if i > 0 {
		for j := 0; j < i; j++ {
			pos = pos.Add(cube.Pos{r.Intn(3) - 1, 1, r.Intn(3) - 1})
			if _, ok := w.Block(pos).(Air); ok {
				if neighboursLavaFlammable(pos, w) {
					w.SetBlock(pos, Fire{}, nil)
				}
			}
		}
	} else {
		for j := 0; j < 3; j++ {
			pos = pos.Add(cube.Pos{r.Intn(3) - 1, 0, r.Intn(3) - 1})
			if _, ok := w.Block(pos.Side(cube.FaceUp)).(Air); ok {
				if flammable, ok := w.Block(pos).(Flammable); ok && flammable.FlammabilityInfo().LavaFlammable && flammable.FlammabilityInfo().Encouragement > 0 {
					w.SetBlock(pos, Fire{}, nil)
				}
			}
		}
	}
}

// HasLiquidDrops ...
func (Lava) HasLiquidDrops() bool {
	return false
}

// LightDiffusionLevel always returns 2.
func (Lava) LightDiffusionLevel() uint8 {
	return 2
}

// LightEmissionLevel returns 15.
func (Lava) LightEmissionLevel() uint8 {
	return 15
}

// NeighbourUpdateTick ...
func (l Lava) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !l.Harden(pos, w, nil) {
		w.ScheduleBlockUpdate(pos, w.Dimension().LavaSpreadDuration())
	}
}

// ScheduledTick ...
func (l Lava) ScheduledTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if !l.Harden(pos, w, nil) {
		tickLiquid(l, pos, w)
	}
}

// LiquidDepth returns the depth of the lava.
func (l Lava) LiquidDepth() int {
	return l.Depth
}

// SpreadDecay always returns 2.
func (Lava) SpreadDecay() int {
	return 2
}

// WithDepth returns a new Lava block with the depth passed and falling if set to true.
func (l Lava) WithDepth(depth int, falling bool) world.Liquid {
	l.Depth = depth
	l.Falling = falling
	l.Still = false
	return l
}

// LiquidFalling checks if the lava is falling.
func (l Lava) LiquidFalling() bool {
	return l.Falling
}

// LiquidType returns 10 as a unique identifier for the lava liquid.
func (Lava) LiquidType() string {
	return "lava"
}

// Harden handles the hardening logic of lava.
func (l Lava) Harden(pos cube.Pos, w *world.World, flownIntoBy *cube.Pos) bool {
	var ok bool
	var water, b world.Block

	if flownIntoBy == nil {
		var water, b world.Block
		_, soulSoilFound := w.Block(pos.Side(cube.FaceDown)).(SoulSoil)
		pos.Neighbours(func(neighbour cube.Pos) {
			if b != nil || neighbour[1] == pos[1]-1 {
				return
			}
			if _, ok := w.Block(neighbour).(BlueIce); ok {
				if soulSoilFound {
					b = Basalt{}
				}
				return
			}
			if waterBlock, ok := w.Block(neighbour).(Water); ok {
				water = waterBlock
				if l.Depth == 8 && !l.Falling {
					b = Obsidian{}
					return
				}
				b = Cobblestone{}
			}
		}, w.Range())
		if b != nil {
			ctx := event.C()
			if w.Handler().HandleLiquidHarden(ctx, pos, l, water, b); ctx.Cancelled() {
				return false
			}
			w.PlaySound(pos.Vec3Centre(), sound.Fizz{})
			w.SetBlock(pos, b, nil)
			return true
		}
		return false
	}
	water, ok = w.Block(*flownIntoBy).(Water)
	if !ok {
		return false
	}

	if l.Depth == 8 && !l.Falling {
		b = Obsidian{}
	} else {
		b = Cobblestone{}
	}
	ctx := event.C()
	if w.Handler().HandleLiquidHarden(ctx, pos, l, water, b); ctx.Cancelled() {
		return false
	}
	w.SetBlock(pos, b, nil)
	w.PlaySound(pos.Vec3Centre(), sound.Fizz{})
	return true
}

// EncodeBlock ...
func (l Lava) EncodeBlock() (name string, properties map[string]any) {
	if l.Depth < 1 || l.Depth > 8 {
		panic("invalid lava depth, must be between 1 and 8")
	}
	v := 8 - l.Depth
	if l.Falling {
		v += 8
	}
	if l.Still {
		return "minecraft:lava", map[string]any{"liquid_depth": int32(v)}
	}
	return "minecraft:flowing_lava", map[string]any{"liquid_depth": int32(v)}
}

// allLava returns a list of all lava states.
func allLava() (b []world.Block) {
	f := func(still, falling bool) {
		b = append(b, Lava{Still: still, Falling: falling, Depth: 8})
		b = append(b, Lava{Still: still, Falling: falling, Depth: 7})
		b = append(b, Lava{Still: still, Falling: falling, Depth: 6})
		b = append(b, Lava{Still: still, Falling: falling, Depth: 5})
		b = append(b, Lava{Still: still, Falling: falling, Depth: 4})
		b = append(b, Lava{Still: still, Falling: falling, Depth: 3})
		b = append(b, Lava{Still: still, Falling: falling, Depth: 2})
		b = append(b, Lava{Still: still, Falling: falling, Depth: 1})
	}
	f(true, true)
	f(true, false)
	f(false, false)
	f(false, true)
	return
}
