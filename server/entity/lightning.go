package entity

import (
	"github.com/df-mc/dragonfly/server/block"
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
		liveTime: nextInt(3) + 1,
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

// State ...
func (li *Lightning) State() int {
	return li.state
}

// Close closes the lighting.
func (li *Lightning) Close() error {
	if li.World() != nil {
		li.World().RemoveEntity(li)
	}

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
	if li.state == 2 { // Init phase
		li.World().PlaySound(li.Position(), sound.Thunder{})
		li.World().PlaySound(li.Position(), sound.Explosion{})

		li.setBlocksOnFire()
	}

	li.state--

	if li.state < 0 {
		if li.liveTime == 0 {
			_ = li.Close()
			return
		} else if li.state < -nextInt(10) {
			li.liveTime--
			li.state = 1

			li.setBlocksOnFire()
		}
	}

	// I am not entirely sure how many times we should damage the player, since this nearly insta-kills it.
	// Default MC behavior is "li.state >= 0", but this definitely insta-kills the player if it's colliding with the AABB of the lightning bolt
	// due to the state being reset ~5 times (depends on rand.Intn), dealing 5 * 5 damage, and 4 * 5 is already enough to insta-kill it.
	if li.state == 0 {
		bb := li.AABB().Grow(9).Extend(mgl64.Vec3{6})

		for _, e := range li.World().CollidingEntities(bb) {
			if l, ok := e.(Living); ok && l.Health() > 0 { // there's no point in damaging entities that are already dead
				l.Hurt(5, damage.SourceLightning{})
				if f, ok := e.(Flammable); ok && f.OnFireDuration() < 8*20 {
					f.SetOnFire(time.Second * 8)
				}
				// TODO Check if the entity is a creeper, if so, make it an supercharged creeper
			}
		}
	}
}

func nextInt(max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max)
}

func (li *Lightning) setBlocksOnFire() {
	_, isNormal := li.World().Difficulty().(world.DifficultyNormal)
	_, isHard := li.World().Difficulty().(world.DifficultyHard)
	if isNormal || isHard { // difficulty >= 2
		lPos := li.Position()
		bPos := cube.Pos{int(lPos.X()), int(lPos.Y()), int(lPos.Z())}
		b := li.World().Block(bPos)

		_, isAir := b.(block.Air)
		_, isTallGrass := b.(block.TallGrass)
		if isAir || isTallGrass {
			below := li.World().Block(bPos.Side(cube.FaceDown))
			if below.Model().FaceSolid(bPos, cube.FaceUp, li.World()) || block.NeighboursFlammable(bPos, li.World()) {
				li.World().PlaceBlock(bPos, block.Fire{})
			}
		}
	}
}
