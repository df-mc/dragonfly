package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"math/rand"
	"time"
)

// Air is the block present in otherwise empty space.
type Air struct {
	empty
	replaceable
	transparent
}

// HasLiquidDrops ...
func (Air) HasLiquidDrops() bool {
	return false
}

// Light ...
func (Air) Light(pos cube.Pos, w *world.World) bool {
	if _, ok := w.Block(pos.Sub(cube.Pos{0, 1})).(Air); !ok {
		w.PlaySound(pos.Vec3Centre(), sound.Ignite{})
		w.SetBlock(pos, Fire{}, nil)
		w.ScheduleBlockUpdate(pos, time.Duration(30+rand.Intn(10))*time.Second/20)
		return true
	}
	return false
}

// EncodeItem ...
func (Air) EncodeItem() (name string, meta int16) {
	return "minecraft:air", 0
}

// EncodeBlock ...
func (Air) EncodeBlock() (string, map[string]any) {
	return "minecraft:air", nil
}
