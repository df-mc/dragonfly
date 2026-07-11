package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// Firework is an item (and entity) used for creating decorative explosions, boosting when flying with elytra, and
// loading into a crossbow as ammunition.
type Firework struct {
	// Duration is the flight duration of the firework.
	Duration time.Duration
	// Explosions is the list of explosions the firework should create when launched.
	Explosions []FireworkExplosion
}

// Use ...
func (f Firework) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	if g, ok := user.(interface {
		Gliding() bool
	}); !ok || !g.Gliding() {
		return false
	}

	pos := user.Position()

	tx.PlaySound(pos, sound.FireworkLaunch{})
	create := tx.World().EntityRegistry().Config().Firework
	opts := world.EntitySpawnOpts{Position: pos, Rotation: user.Rotation()}
	tx.AddEntity(create(opts, f, user, 1.15, 0.04, true))

	ctx.SubtractFromCount(1)
	return true
}

// UseOnBlock ...
func (f Firework) UseOnBlock(pos cube.Pos, _ cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	fpos := pos.Vec3().Add(clickPos)
	create := tx.World().EntityRegistry().Config().Firework
	opts := world.EntitySpawnOpts{Position: fpos, Rotation: cube.Rotation{rand.Float64() * 360, 90}}
	tx.AddEntity(create(opts, f, user, 1.15, 0.04, false))
	tx.PlaySound(fpos, sound.FireworkLaunch{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeNBT ...
func (f Firework) EncodeNBT() map[string]any {
	explosions := make([]any, 0, len(f.Explosions))
	for _, explosion := range f.Explosions {
		explosions = append(explosions, explosion.EncodeNBT())
	}
	return map[string]any{"Fireworks": map[string]any{
		"Explosions": explosions,
		"Flight":     uint8((f.Duration/10 - time.Millisecond*50).Milliseconds() / 50),
	}}
}

// DecodeNBT ...
func (f Firework) DecodeNBT(data map[string]any) any {
	if fireworks, ok := data["Fireworks"].(map[string]any); ok {
		if explosions, ok := fireworks["Explosions"].([]any); ok {
			f.Explosions = make([]FireworkExplosion, len(explosions))
			for i, explosion := range f.Explosions {
				f.Explosions[i] = explosion.DecodeNBT(explosions[i].(map[string]any)).(FireworkExplosion)
			}
		}
		if durationTicks, ok := fireworks["Flight"].(uint8); ok {
			f.Duration = (time.Duration(durationTicks)*time.Millisecond*50 + time.Millisecond*50) * 10
		}
	}
	return f
}

// RandomisedDuration returns the randomised flight duration of the firework.
func (f Firework) RandomisedDuration() time.Duration {
	return f.Duration + time.Duration(rand.IntN(int(time.Millisecond*600)))
}

// OffHand ...
func (Firework) OffHand() bool {
	return true
}

// EncodeItem ...
func (Firework) EncodeItem() (name string, meta int16) {
	return "minecraft:firework_rocket", 0
}
