package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math"
	"time"
)

// Beacon is a block that projects a light beam skyward, and can provide status effects such as Speed, Jump
// Boost, Haste, Regeneration, Resistance, or Strength to nearby players.
type Beacon struct {
	solid
	transparent
	clicksAndSticks

	// Primary and Secondary are the primary and secondary effects broadcast to nearby entities by the
	// beacon.
	Primary, Secondary effect.LastingType
	// level is the amount of the pyramid's levels, it is defined by the mineral blocks which build up the
	// pyramid, and can be 0-4.
	level int
}

// BeaconSource represents a block which is capable of contributing to powering a beacon pyramid.
type BeaconSource interface {
	// PowersBeacon returns a bool which indicates whether this block can contribute to powering up a
	// beacon pyramid.
	PowersBeacon() bool
}

// BreakInfo ...
func (b Beacon) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, nothingEffective, oneOf(b))
}

// Activate manages the opening of a beacon by activating it.
func (b Beacon) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return true
}

// DecodeNBT ...
func (b Beacon) DecodeNBT(data map[string]any) any {
	b.level = int(nbtconv.Map[int32](data, "Levels"))
	if primary, ok := effect.ByID(int(nbtconv.Map[int32](data, "Primary"))); ok {
		b.Primary = primary.(effect.LastingType)
	}
	if secondary, ok := effect.ByID(int(nbtconv.Map[int32](data, "Secondary"))); ok {
		b.Secondary = secondary.(effect.LastingType)
	}
	return b
}

// EncodeNBT ...
func (b Beacon) EncodeNBT() map[string]any {
	m := map[string]any{
		"Levels": int32(b.level),
	}
	if primary, ok := effect.ID(b.Primary); ok {
		m["Primary"] = int32(primary)
	}
	if secondary, ok := effect.ID(b.Secondary); ok {
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
func (b Beacon) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
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
func (b Beacon) Tick(currentTick int64, pos cube.Pos, w *world.World) {
	if currentTick%80 == 0 {
		before := b.level
		// Recalculating pyramid level and powering up players in range once every 4 seconds.
		b.level = b.recalculateLevel(pos, w)
		if before != b.level {
			w.SetBlock(pos, b, nil)
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
func (b Beacon) recalculateLevel(pos cube.Pos, w *world.World) int {
	var lvl int
	iter := 1
	// This loop goes over all 4 possible pyramid levels.
	for y := pos.Y() - 1; y >= pos.Y()-4; y-- {
		for x := pos.X() - iter; x <= pos.X()+iter; x++ {
			for z := pos.Z() - iter; z <= pos.Z()+iter; z++ {
				if s, ok := w.Block(cube.Pos{x, y, z}).(BeaconSource); !ok || !s.PowersBeacon() {
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
func (b Beacon) obstructed(pos cube.Pos, w *world.World) bool {
	// Fast obstructed light calculation.
	if w.SkyLight(pos.Add(cube.Pos{0, 1})) == 15 {
		return false
	}
	// Slow obstructed light calculation, if the fast way out didn't suffice.
	return w.HighestLightBlocker(pos.X(), pos.Z()) > pos[1]
}

// broadcastBeaconEffects determines the entities in range which could receive the beacon's powers, and
// determines the powers (effects) that these entities could get. Afterwards, the entities in range that are
// beaconAffected get their according effect(s).
func (b Beacon) broadcastBeaconEffects(pos cube.Pos, w *world.World) {
	seconds := 9 + b.level*2
	if b.level == 4 {
		seconds--
	}
	dur := time.Duration(seconds) * time.Second

	// Establishing what effects are active with the current amount of beacon levels.
	primary, secondary := b.Primary, effect.LastingType(nil)
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
		// Accept all effects for primary, but leave secondary as nil.
	default:
		secondary = b.Secondary
	}
	var primaryEff, secondaryEff effect.Effect
	// Determining whether the primary power is set.
	if primary != nil {
		primaryEff = effect.NewAmbient(primary, 1, dur)
		// Secondary power can only be set if the primary power is set.
		if secondary != nil {
			// It is possible to select 2 primary powers if the beacon's level is 4. This then means that the effect
			// should get a level of 2.
			if primary == secondary {
				primaryEff = effect.NewAmbient(primary, 2, dur)
			} else {
				secondaryEff = effect.NewAmbient(secondary, 1, dur)
			}
		}
	}

	// Finding entities in range.
	r := 10 + (b.level * 10)
	entitiesInRange := w.EntitiesWithin(cube.Box(
		float64(pos.X()-r), -math.MaxFloat64, float64(pos.Z()-r),
		float64(pos.X()+r), math.MaxFloat64, float64(pos.Z()+r),
	), nil)
	for _, e := range entitiesInRange {
		if p, ok := e.(beaconAffected); ok {
			if primaryEff.Type() != nil {
				p.AddEffect(primaryEff)
			}
			if secondaryEff.Type() != nil {
				p.AddEffect(secondaryEff)
			}
		}
	}
}

// beaconAffected represents an entity that can be powered by a beacon. Only players will implement this.
type beaconAffected interface {
	// AddEffect adds a specific effect to the entity that implements this interface.
	AddEffect(e effect.Effect)
	// BeaconAffected returns whether this entity can be powered by a beacon.
	BeaconAffected() bool
}

// EncodeItem ...
func (Beacon) EncodeItem() (name string, meta int16) {
	return "minecraft:beacon", 0
}

// EncodeBlock ...
func (Beacon) EncodeBlock() (string, map[string]any) {
	return "minecraft:beacon", nil
}
