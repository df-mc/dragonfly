package block

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
	_ "unsafe" // For compiler directives.
)

// Beacon is a block that projects a light beam skyward, and can provide status effects such as Speed, Jump
// Boost, Haste, Regeneration, Resistance, or Strength to nearby players.
type Beacon struct {
	// level is the amount of the pyramid's levels, it is defined by the mineral blocks which build up the
	// pyramid, and can be 0-4.
	level int
	// Primary and Secondary powers, in other words, effects of the beacon.
	Primary, Secondary entity.Effect
	// Active is a bool which indicates whether the beacon is currently properly powered.
	Active bool
	// tick is the beacon's current tick, this can be 0-80. This value is incremented every tick, and when
	// it reaches 80, the state of the beacon pyramid is recalculated and nearby players get beacon's powers.
	tick int
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

// DecodeNBT ...
func (b Beacon) DecodeNBT(data map[string]interface{}) interface{} {
	b.level = readInt(data, "Levels")
	if primary, ok := effect_effectByID(readInt(data, "Primary")); ok {
		b.Primary = primary
	}
	if secondary, ok := effect_effectByID(readInt(data, "Secondary")); ok {
		b.Secondary = secondary
	}
	return b
}

// EncodeNBT ...
func (b Beacon) EncodeNBT() map[string]interface{} {
	m := map[string]interface{}{
		"Levels": b.level,
	}
	if primary, ok := effect_idByEffect(b.Primary); ok {
		m["Primary"] = primary
	}
	if secondary, ok := effect_idByEffect(b.Secondary); ok {
		m["Secondary"] = secondary
	}
	return m
}

// TODO: Beacon UI.
// TODO: Assigning Primary & Secondary powers via Beacon UI handling.

// CanDisplace ...
func (b Beacon) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

// SideClosed ...
func (b Beacon) SideClosed(pos, side world.BlockPos, w *world.World) bool {
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
	b.tick++

	// Recalculating pyramid level and powering up players in range once every 4 seconds.
	if b.tick == 80 {
		b.tick = 0
		b.recalculateLevel(pos, w)
		b.recalculateActive(pos, w)
		if b.Active {
			b.preparePowerPlayers(pos, w)
		}
	}
}

// recalculateLevel recalculates the level of the beacon's pyramid. The level can be 0-4.
func (b Beacon) recalculateLevel(pos world.BlockPos, w *world.World) {
	newLevel := 0
	iter := 1
	// This loop goes over all 4 possible pyramid levels.
	for y := pos.Y() - 1; y > pos.Y()-4; y-- {
		for x := pos.X() - iter; x <= pos.X()+iter; x++ {
			for z := pos.Z() - iter; z <= pos.Z()+iter; z++ {
				if _, ok := w.Block(world.BlockPos{x, y, z}).(BeaconSource); !ok {
					b.level = newLevel
					return
				}
			}
		}
		iter++
		newLevel++
	}
	b.level = newLevel
}

// recalculateActive determines whether the beacon can power up nearby players at the moment.
func (b Beacon) recalculateActive(pos world.BlockPos, w *world.World) {
	obstructed := true
	// Fast obstructed light calculation.
	if w.SkyLight(pos.Add(world.BlockPos{0, 1})) == 15 {
		obstructed = false
		// Slow obstructed light calculation, if the fast way out failed.
	} else if world_highestLightBlocker(w, pos.X(), pos.Z()) == uint8(pos.Y()) {
		obstructed = false
	}

	b.Active = b.level > 0 && !obstructed && b.Primary != nil
}

// preparePowerPlayers determines the entities in range which could receive the beacon's powers,
// and determines the powers (effects) that these entities could get. Afterwards, the entities
// in range that are beaconAffected get their according power(s).
func (b Beacon) preparePowerPlayers(pos world.BlockPos, w *world.World) {
	// Finding entities in range.
	halfRange := 10 + ((b.level - 1) * 5)
	entitiesInRange := w.EntitiesWithin(physics.NewAABB(
		mgl64.Vec3{
			float64(pos.X() - halfRange), float64(pos.Y() - halfRange), float64(pos.Z() - halfRange),
		},
		mgl64.Vec3{
			float64(pos.X() + halfRange), float64(pos.Y() + halfRange), float64(pos.Z() + halfRange),
		}),
	)

	var effs []entity.Effect
	dur := int64(9+(b.level*2)) * time.Second.Nanoseconds()

	// Determining whether the primary power is set.
	if b.Primary != nil {
		primary := b.Primary.WithDuration(time.Duration(dur))
		var secondary entity.Effect = nil
		// Secondary power can only be set if the primary power is set.
		if b.Secondary != nil {
			// It is possible to select 2 primary powers if the beacon's level is 4.
			if effect_idByEffect(b.Primary) == effect_idByEffect(b.Secondary) {
				// TODO: Increment primary effect level by 1
			} else {
				secondary = b.Secondary.WithDuration(time.Duration(dur))
			}
		}
		effs = append(effs, primary)
		if secondary != nil {
			effs = append(effs, secondary)
		}
	}

	b.powerPlayers(entitiesInRange, effs)
}

// powerPlayers gives beacon powers to any beaconAffected entity in the entities array passed.
// Usually the only beaconAffected entity is a player. The effects array passed holds the
// beacon's power(s).
func (Beacon) powerPlayers(entities []world.Entity, effects []entity.Effect) {
	for _, e := range entities {
		if p, ok := e.(beaconAffected); ok {
			for _, eff := range effects {
				p.AddEffect(eff)
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

//go:linkname effect_effectByID github.com/df-mc/dragonfly/dragonfly/entity/effect.effectByID
//noinspection ALL
func effect_effectByID(id int) (entity.Effect, bool)

//go:linkname effect_idByEffect github.com/df-mc/dragonfly/dragonfly/entity/effect.idByEffect
//noinspection ALL
func effect_idByEffect(e entity.Effect) (int, bool)

//go:linkname world_highestLightBlocker github.com/df-mc/dragonfly/dragonfly/world.highestLightBlocker
//noinspection ALL
func world_highestLightBlocker(w *world.World, x, z int) uint8
