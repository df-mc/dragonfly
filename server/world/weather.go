package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// weather implements weather related methods for World. World embeds this struct, so any exported methods on weather
// are exported methods on World.
type weather struct{ w *World }

// StopWeatherCycle disables weather cycle of the World.
func (w weather) StopWeatherCycle() {
	w.enableWeatherCycle(false)
}

// StartWeatherCycle enables weather cycle of the World.
func (w weather) StartWeatherCycle() {
	w.enableWeatherCycle(true)
}

// SnowingAt checks if it is snowing at a specific cube.Pos in the World. True is returned if the temperature in the
// biome at that position is sufficiently low, if it is raining and if it's above the top-most obstructing block.
func (w weather) SnowingAt(pos cube.Pos) bool {
	if w.w == nil || !w.w.Dimension().WeatherCycle() {
		return false
	}
	if b := w.w.Biome(pos); b.Rainfall() == 0 || w.w.Temperature(pos) > 0.15 {
		return false
	}
	w.w.set.Lock()
	raining := w.w.set.Raining
	w.w.set.Unlock()
	return raining && w.w.highestObstructingBlock(pos[0], pos[2]) < pos[1]
}

// RainingAt checks if it is raining at a specific cube.Pos in the World. True is returned if it is raining, if the
// temperature is high enough in the biome for it not to be snow and if the block is above the top-most obstructing
// block.
func (w weather) RainingAt(pos cube.Pos) bool {
	if w.w == nil || !w.w.Dimension().WeatherCycle() {
		return false
	}
	if b := w.w.Biome(pos); b.Rainfall() == 0 || w.w.Temperature(pos) <= 0.15 {
		return false
	}
	w.w.set.Lock()
	a := w.w.set.Raining
	w.w.set.Unlock()
	return a && w.w.highestObstructingBlock(pos[0], pos[2]) < pos[1]
}

// ThunderingAt checks if it is thundering at a specific cube.Pos in the World. True is returned if RainingAt returns
// true and if it is thundering in the world.
func (w weather) ThunderingAt(pos cube.Pos) bool {
	raining := w.RainingAt(pos)
	w.w.set.Lock()
	a := w.w.set.Thundering && raining
	w.w.set.Unlock()
	return a && w.w.highestObstructingBlock(pos[0], pos[2]) < pos[1]
}

// StartRaining makes it rain in the World. The time.Duration passed will determine how long it will rain.
func (w weather) StartRaining(dur time.Duration) {
	w.w.set.Lock()
	defer w.w.set.Unlock()
	w.setRaining(true, dur)
}

// StopRaining makes it stop raining in the World.
func (w weather) StopRaining() {
	w.w.set.Lock()
	defer w.w.set.Unlock()

	if w.w.set.Raining {
		w.setRaining(false, time.Second*(time.Duration(w.w.r.Intn(8400)+600)))
		if w.w.set.Thundering {
			// Also reset thunder if it was previously thundering.
			w.setThunder(false, time.Second*(time.Duration(w.w.r.Intn(8400)+600)))
		}
	}
}

// StartThundering makes it thunder in the World. The time.Duration passed will determine how long it will thunder.
// StartThundering will also make it rain if it wasn't already raining. In this case the rain will, like the thunder,
// last for the time.Duration passed.
func (w weather) StartThundering(dur time.Duration) {
	w.w.set.Lock()
	defer w.w.set.Unlock()

	w.setThunder(true, dur)
	w.setRaining(true, dur)
}

// StopThundering makes it stop thundering in the current world.
func (w weather) StopThundering() {
	w.w.set.Lock()
	defer w.w.set.Unlock()
	if w.w.set.Thundering && w.w.set.Raining {
		w.setThunder(false, time.Second*(time.Duration(w.w.r.Intn(8400)+600)))
	}
}

// advanceWeather advances the weather counters of the World. Rain and thunder are stopped/started when the rain and
// thunder times reach 0.
func (w weather) advanceWeather() {
	w.w.set.RainTime--
	w.w.set.ThunderTime--

	if w.w.set.RainTime <= 0 {
		// Wiki: The rain counter counts down to zero, and each time it reaches zero, the rain is toggled on or off.
		// When the rain is turned on, the counter is reset to a value between 12,000-23,999 ticks (0.5-1 game days)
		// and when the rain is turned off it is reset to a value of 12,000-179,999 ticks (0.5-7.5 game days).
		if w.w.set.Raining {
			w.w.setRaining(false, time.Second*(time.Duration(w.w.r.Intn(8400)+600)))
		} else {
			w.w.setRaining(true, time.Second*time.Duration(w.w.r.Intn(600)+600))
		}
	}
	if w.w.set.ThunderTime <= 0 {
		// Wiki: the thunder counter toggles thunder on/off when it reaches zero, but clear weather overrides the
		// "on" state. When thunder is turned on, the thunder counter is reset to 3,600-15,999 ticks (3-13 minutes),
		// and when thunder is turned off the counter rests to 12,000-179,999 ticks (0.5-7.5 days).
		if w.w.set.Thundering {
			w.w.setThunder(false, time.Second*(time.Duration(w.w.r.Intn(8400)+600)))
		} else {
			w.w.setThunder(true, time.Second*time.Duration(w.w.r.Intn(620)+180))
		}
	}
}

// setRaining toggles raining depending on the raining argument.
// This does not lock the world mutex as opposed to StartRaining and StopRaining.
func (w weather) setRaining(raining bool, x time.Duration) {
	w.w.set.Raining = raining
	w.w.set.RainTime = int64(x.Seconds() * 20)
}

// setThunder toggles thundering depending on the thundering argument.
// This does not lock the world mutex as opposed to StartThundering and StopThundering.
func (w weather) setThunder(thundering bool, x time.Duration) {
	w.w.set.Thundering = thundering
	w.w.set.ThunderTime = int64(x.Seconds() * 20)
}

// enableWeatherCycle either enables or disables the weather cycle of the World.
func (w weather) enableWeatherCycle(v bool) {
	if w.w == nil {
		return
	}
	w.w.set.Lock()
	defer w.w.set.Unlock()
	w.w.set.WeatherCycle = v
}

// tickLightning iterates over all loaded chunks in the World, striking lightning in each one with a 1/100,000 chance.
func (w weather) tickLightning() {
	w.w.chunkMu.Lock()
	positions := make([]ChunkPos, 0, len(w.w.chunks)/100000)
	for pos := range w.w.chunks {
		// Wiki: For each loaded chunk, every tick there is a 1⁄100,000 chance of an attempted lightning strike
		// during a thunderstorm
		if w.w.r.Intn(100000) == 0 {
			positions = append(positions, pos)
		}
	}
	w.w.chunkMu.Unlock()

	for _, pos := range positions {
		w.w.strikeLightning(pos)
	}
}

// strikeLightning attempts to strike lightning in the world at a specific ChunkPos. The final position is influenced by
// living entities that might be near the lightning strike. If there is no rain at the final position selected, the
// lightning strike will fail.
func (w weather) strikeLightning(c ChunkPos) {
	if pos := w.lightningPosition(c); w.ThunderingAt(cube.PosFromVec3(pos)) {
		e, _ := EntityByName("minecraft:lightning_bolt")
		w.w.AddEntity(e.(interface{ New(mgl64.Vec3) Entity }).New(pos))
	}
}

// lightningPosition finds a random position in the ChunkPos to strike lightning and adjusts the position to any of the
// living entities found in or above the position if any are found.
func (w weather) lightningPosition(c ChunkPos) mgl64.Vec3 {
	v := w.w.r.Int31()
	x, z := float64(c[0]<<4+(v&0xf)), float64(c[1]<<4+((v>>8)&0xf))

	vec := w.adjustPositionToEntities(mgl64.Vec3{x, float64(w.w.HighestBlock(int(x), int(z)) + 1), z})
	if pos := cube.PosFromVec3(vec); len(w.w.Block(pos).Model().BBox(pos, w.w)) != 0 {
		// If lightning is about to strike inside a block that is not fully transparent. In this case, move the
		// lightning up by one block so that it strikes above the block.
		return vec.Add(mgl64.Vec3{0, 1})
	}
	return vec
}

// adjustPositionToEntities adjusts the mgl64.Vec3 passed to the position of any entity found in the 3x3 column upwards
// from the mgl64.Vec3. If multiple entities are found, the position of one of the entities is selected randomly.
func (w weather) adjustPositionToEntities(vec mgl64.Vec3) mgl64.Vec3 {
	max := vec.Add(mgl64.Vec3{0, float64(w.w.Range().Max())})
	ent := w.w.EntitiesWithin(cube.Box(vec[0], vec[1], vec[2], max[0], max[1], max[2]).GrowVec3(mgl64.Vec3{3, 3, 3}), nil)

	list := make([]mgl64.Vec3, 0, len(ent)/3)
	for _, e := range ent {
		if h, ok := e.(interface{ Health() float64 }); ok && h.Health() > 0 {
			// Any (living) entity that is positioned higher than the highest block at its position is eligible to be
			// struck by lightning. We first save all entity positions where this is the case.
			pos := cube.PosFromVec3(e.Position())
			if w.w.HighestBlock(pos[0], pos[1]) < pos[2] {
				list = append(list, e.Position())
			}
		}
	}
	// We then select one of the positions of entities higher than the highest block and adjust the position of the
	// lightning to it, so that the entity is struck directly.
	if len(list) > 0 {
		vec = list[w.w.r.Intn(len(list))]
	}
	return vec
}
