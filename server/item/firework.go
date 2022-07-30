package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
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

// UseOnBlock ...
func (f Firework) UseOnBlock(blockPos cube.Pos, _ cube.Face, clickPos mgl64.Vec3, w *world.World, user User, ctx *UseContext) bool {
	firework, ok := world.EntityByName("minecraft:fireworks_rocket")
	if !ok {
		return false
	}

	p, ok := firework.(interface {
		New(pos mgl64.Vec3, yaw, pitch float64, firework Firework, owner world.Entity) world.Entity
	})
	if !ok {
		return false
	}
	pos := blockPos.Vec3().Add(clickPos)

	w.PlaySound(pos, sound.FireworkLaunch{})
	w.AddEntity(p.New(pos, rand.Float64()*360, 90, f, user))

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
		"Flight":     uint8(f.Duration.Milliseconds() / 50),
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
			f.Duration = time.Duration(durationTicks) * 50 * time.Millisecond
		}
	}
	return f
}

// RandomisedDuration returns the randomised flight duration of the firework.
func (f Firework) RandomisedDuration() time.Duration {
	definite := f.Duration + time.Millisecond*50
	randomness := time.Duration(rand.Intn(int(time.Millisecond * 600)))
	return definite*10 + randomness
}

// EncodeItem ...
func (Firework) EncodeItem() (name string, meta int16) {
	return "minecraft:firework_rocket", 0
}
