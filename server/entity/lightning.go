package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"sync/atomic"
	"time"
)

// Lightning is a lethal element to thunderstorms. Lightning momentarily increases the skylight's brightness to slightly greater than full daylight.
type Lightning struct {
	pos atomic.Value

	state    int
	liveTime int
}

// NewLightning creates a lightning entity. The lightning entity will be positioned at the position passed.
func NewLightning(pos mgl64.Vec3) *Lightning {
	li := &Lightning{
		state:    2,
		liveTime: rand.Intn(3) + 1,
	}
	li.pos.Store(pos)

	return li
}

// Position returns the current position of the lightning entity.
func (li *Lightning) Position() mgl64.Vec3 {
	return li.pos.Load().(mgl64.Vec3)
}

// World returns the world that the lightning entity is currently in, or nil if it is not added to a world.
func (li *Lightning) World() *world.World {
	w, _ := world.OfEntity(li)
	return w
}

// AABB ...
func (Lightning) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{})
}

// Close closes the lighting.
func (li *Lightning) Close() error {
	li.World().RemoveEntity(li)
	return nil
}

// OnGround ...
func (Lightning) OnGround() bool {
	return false
}

// Rotation ...
func (li *Lightning) Rotation() (yaw, pitch float64) {
	return 0, 0
}

// EncodeEntity ...
func (li *Lightning) EncodeEntity() string {
	return "minecraft:lightning_bolt"
}

// Name ...
func (li *Lightning) Name() string {
	return "Lightning Bolt"
}

// Tick ...
func (li *Lightning) Tick(_ int64) {
	pos, w := li.Position(), li.World()
	if li.state == 2 { // Init phase
		difficulty := w.Difficulty()
		_, normal := difficulty.(world.DifficultyNormal)
		_, hard := difficulty.(world.DifficultyHard)
		if normal || hard {
			li.spawnFire(pos, w, 4)
		}
		w.PlaySound(pos, sound.Thunder{})
		w.PlaySound(pos, sound.Explosion{})
	}
	if li.state--; li.state < 0 {
		if li.liveTime == 0 {
			_ = li.Close()
		} else if li.state < -rand.Intn(10) {
			li.liveTime--
			li.state = 1

			li.spawnFire(pos, w, 0)
		}
	} else {
		bb := li.AABB().Translate(pos).Grow(3).ExtendTowards(cube.FaceUp, 6)
		for _, e := range w.CollidingEntities(bb) {
			// Only damage entities that weren't already dead.
			if l, ok := e.(Living); ok && l.Health() > 0 {
				l.Hurt(5, damage.SourceLightning{})
				if f, ok := e.(Flammable); ok && f.OnFireDuration() < 8*20 {
					f.SetOnFire(time.Second * 8)
				}
			}
		}
	}
}

// spawnFire spawns fire at the position passed.
func (li *Lightning) spawnFire(pos mgl64.Vec3, w *world.World, extra int) {
	air := air()

	blockPos := cube.PosFromVec3(pos)
	if b := w.Block(blockPos); b == air && canFireSurvive(blockPos, w) {
		w.PlaceBlock(blockPos, fire())
	}

	for i := 0; i < extra; i++ {
		p := blockPos.Add(cube.Pos{rand.Intn(3) - 1, rand.Intn(3) - 1, rand.Intn(3) - 1})
		if b := w.Block(p); b == air && canFireSurvive(p, w) {
			w.PlaceBlock(p, fire())
		}
	}
}

// fire returns a fire block.
func fire() world.Block {
	f, ok := world.BlockByName("minecraft:fire", map[string]interface{}{"age": int32(0)})
	if !ok {
		panic("could not find fire block")
	}
	return f
}

// air returns an air block.
func air() world.Block {
	a, ok := world.BlockByName("minecraft:air", nil)
	if !ok {
		panic("could not find air block")
	}
	return a
}

// canFireSurvive returns whether a fire block can spawn in a specific block position.
func canFireSurvive(pos cube.Pos, w *world.World) bool {
	below := w.Block(pos.Side(cube.FaceDown))
	return below.Model().FaceSolid(pos, cube.FaceUp, w) || neighboursFlammable(pos, w)
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

var flammableBlock func(block world.Block) bool
