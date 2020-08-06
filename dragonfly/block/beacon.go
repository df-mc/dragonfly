package block

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/effect"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/internal/block_internal"
	"github.com/df-mc/dragonfly/dragonfly/internal/world_internal"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"time"
	_ "unsafe" // For compiler directives.
)

// Beacon is a block that projects a light beam skyward, and can provide status effects such as Speed, Jump
// Boost, Haste, Regeneration, Resistance, or Strength to nearby players.
type Beacon struct {
	nbt
	solid
	transparent

	// Primary and Secondary are the primary and secondary effects broadcast to nearby entities by the
	// beacon.
	Primary, Secondary effect.Effect
	// level is the amount of the pyramid's levels, it is defined by the mineral blocks which build up the
	// pyramid, and can be 0-4.
	level int
}

// BreakInfo ...
func (b Beacon) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(b, 1)),
	}
}

// UseOnBlock ...
func (b Beacon) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, b)
	if !used {
		return
	}

	place(w, pos, b, user, ctx)
	return placed(ctx)
}

// Activate manages the opening of a beacon by activating it.
func (b Beacon) Activate(pos world.BlockPos, _ world.Face, _ *world.World, u item.User) {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
	}
}

// DecodeNBT ...
func (b Beacon) DecodeNBT(data map[string]interface{}) interface{} {
	b.level = int(readInt32(data, "Levels"))
	if primary, ok := effect_effectByID(int(readInt32(data, "Primary"))); ok {
		b.Primary = primary
	}
	if secondary, ok := effect_effectByID(int(readInt32(data, "Secondary"))); ok {
		b.Secondary = secondary
	}
	return b
}

// EncodeNBT ...
func (b Beacon) EncodeNBT() map[string]interface{} {
	m := map[string]interface{}{
		"Levels": int32(b.level),
	}
	if primary, ok := effect_idByEffect(b.Primary); ok {
		m["Primary"] = int32(primary)
	}
	if secondary, ok := effect_idByEffect(b.Secondary); ok {
		m["Secondary"] = int32(secondary)
	}
	return m
}

// CanDisplace ...
func (b Beacon) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

// SideClosed ...
func (b Beacon) SideClosed(world.BlockPos, world.BlockPos, *world.World) bool {
	return false
}

// LightEmissionLevel ...
func (Beacon) LightEmissionLevel() uint8 {
	return 15
}

// Level returns an integer 0-4 which defines the current pyramid level of the beacon.
func (b Beacon) Level() int {
	return b.level
}

// Tick recalculates level, recalculates the active state of the beacon, and powers players,
// once every 80 ticks (4 seconds).
func (b Beacon) Tick(currentTick int64, pos world.BlockPos, w *world.World) {
	if currentTick%80 == 0 {
		before := b.level
		// Recalculating pyramid level and powering up players in range once every 4 seconds.
		b.level = b.recalculateLevel(pos, w)
		if before != b.level {
			w.SetBlock(pos, b)
		}
		if b.level == 0 {
			return
		}
		if !b.obstructed(pos, w) {
			b.broadcastBeaconEffects(pos, w)
		}
	}
}

// recalculateLevel recalculates the level of the beacon's pyramid and returns it. The level can be 0-4.
func (b Beacon) recalculateLevel(pos world.BlockPos, w *world.World) int {
	var lvl int
	iter := 1
	// This loop goes over all 4 possible pyramid levels.
	for y := pos.Y() - 1; y >= pos.Y()-4; y-- {
		for x := pos.X() - iter; x <= pos.X()+iter; x++ {
			for z := pos.Z() - iter; z <= pos.Z()+iter; z++ {
				if src, ok := world_internal.BeaconSource[block_internal.World_runtimeID(w, world.BlockPos{x, y, z})]; !ok || !src {
					return lvl
				}
			}
		}
		iter++
		lvl++
	}
	return lvl
}

// obstructed determines whether the beacon is currently obstructed.
func (b Beacon) obstructed(pos world.BlockPos, w *world.World) bool {
	// Fast obstructed light calculation.
	if w.SkyLight(pos.Add(world.BlockPos{0, 1})) == 15 {
		return false
		// Slow obstructed light calculation, if the fast way out failed.
	} else if world_highestLightBlocker(w, pos.X(), pos.Z()) <= uint8(pos.Y()) {
		return false
	}
	return true
}

// broadcastBeaconEffects determines the entities in range which could receive the beacon's powers, and
// determines the powers (effects) that these entities could get. Afterwards, the entities in range that are
// beaconAffected get their according effect(s).
func (b Beacon) broadcastBeaconEffects(pos world.BlockPos, w *world.World) {
	seconds := 9 + b.level*2
	if b.level == 4 {
		seconds--
	}
	dur := time.Duration(seconds) * time.Second

	// Establishing what effects are active with the current amount of beacon levels.
	primary, secondary := b.Primary, effect.Effect(nil)
	switch b.level {
	case 0:
		primary = nil
	case 1:
		switch primary.(type) {
		case effect.Resistance, effect.JumpBoost, effect.Strength:
			primary = nil
		}
	case 2:
		if _, ok := primary.(effect.Strength); ok {
			primary = nil
		}
	case 3:
	default:
		secondary = b.Secondary
	}
	// Determining whether the primary power is set.
	if primary != nil {
		primary = primary.WithSettings(dur, 1, true)
		// Secondary power can only be set if the primary power is set.
		if secondary != nil {
			// It is possible to select 2 primary powers if the beacon's level is 4.
			pId, pOk := effect_idByEffect(primary)
			sId, sOk := effect_idByEffect(secondary)
			if pOk && sOk && pId == sId {
				primary = primary.WithSettings(dur, 2, true)
				secondary = nil
			} else {
				secondary = secondary.WithSettings(dur, 1, true)
			}
		}
	}

	// Finding entities in range.
	r := 10 + (b.level * 10)
	entitiesInRange := w.EntitiesWithin(physics.NewAABB(
		mgl64.Vec3{float64(pos.X() - r), -math.MaxFloat64, float64(pos.Z() - r)},
		mgl64.Vec3{float64(pos.X() + r), math.MaxFloat64, float64(pos.Z() + r)},
	))
	for _, e := range entitiesInRange {
		if p, ok := e.(beaconAffected); ok {
			if primary != nil {
				p.AddEffect(primary)
			}
			if secondary != nil {
				p.AddEffect(secondary)
			}
		}
	}
}

// EncodeItem ...
func (Beacon) EncodeItem() (id int32, meta int16) {
	return 138, 0
}

// EncodeBlock ...
func (Beacon) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:beacon", nil
}

// Hash ...
func (Beacon) Hash() uint64 {
	return hashBeacon
}

//go:linkname effect_effectByID github.com/df-mc/dragonfly/dragonfly/entity/effect.effectByID
//noinspection ALL
func effect_effectByID(id int) (effect.Effect, bool)

//go:linkname effect_idByEffect github.com/df-mc/dragonfly/dragonfly/entity/effect.idByEffect
//noinspection ALL
func effect_idByEffect(e effect.Effect) (int, bool)

//go:linkname world_highestLightBlocker github.com/df-mc/dragonfly/dragonfly/world.highestLightBlocker
//noinspection ALL
func world_highestLightBlocker(w *world.World, x, z int) uint8
