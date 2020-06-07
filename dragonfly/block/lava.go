package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/event"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/sound"
	"time"
)

// Lava is a light-emitting fluid block that causes fire damage.
type Lava struct {
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

// ReplaceableBy ...
func (Lava) ReplaceableBy(world.Block) bool {
	return true
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
func (l Lava) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if !l.Harden(pos, w, nil) {
		w.ScheduleBlockUpdate(pos, time.Second*3/2)
	}
}

// ScheduledTick ...
func (l Lava) ScheduledTick(pos world.BlockPos, w *world.World) {
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

// LiquidType returns "lava" as a unique identifier for the lava liquid.
func (Lava) LiquidType() string {
	return "lava"
}

// Harden handles the hardening logic of lava.
func (l Lava) Harden(pos world.BlockPos, w *world.World, flownIntoBy *world.BlockPos) bool {
	var ok bool
	var water, b world.Block

	if flownIntoBy == nil {
		var water, b world.Block
		pos.Neighbours(func(neighbour world.BlockPos) {
			if b != nil || neighbour[1] == pos[1]-1 {
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
		})
		if b != nil {
			ctx := event.C()
			w.Handler().HandleLiquidHarden(ctx, pos, l, water, b)
			ctx.Continue(func() {
				w.PlaySound(pos.Vec3Centre(), sound.Fizz{})
				w.PlaceBlock(pos, Cobblestone{})
			})
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
	w.Handler().HandleLiquidHarden(ctx, pos, l, water, b)
	ctx.Continue(func() {
		w.PlaceBlock(pos, b)
		w.PlaySound(pos.Vec3Centre(), sound.Fizz{})
	})
	return true
}

// EncodeBlock ...
func (l Lava) EncodeBlock() (name string, properties map[string]interface{}) {
	if l.Depth < 1 || l.Depth > 8 {
		panic("invalid lava depth, must be between 1 and 8")
	}
	v := 8 - l.Depth
	if l.Falling {
		v += 8
	}
	if l.Still {
		return "minecraft:lava", map[string]interface{}{"liquid_depth": int32(v)}
	}
	return "minecraft:flowing_lava", map[string]interface{}{"liquid_depth": int32(v)}
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
