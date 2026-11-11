package item

import (
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Cushion is an item that places a cushion entity, which players may sit on.
type Cushion struct {
	// Colour is the colour of the cushion.
	Colour Colour
}

const cushionWidth, cushionHeight = 0.999, 0.249

// UseOnBlock places a cushion on the top face of a block, centred at the height clicked and facing the user,
// unless it would overlap another cushion.
func (c Cushion) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	if face != cube.FaceUp {
		return false
	}
	click := pos.Vec3().Add(clickPos)
	box := cube.Box(0, 0, 0, cushionWidth, cushionHeight, cushionWidth).Translate(mgl64.Vec3{float64(pos[0]), click[1], float64(pos[2])})
	for e := range tx.EntitiesWithin(box.Grow(1)) {
		if e.H().Type().EncodeEntity() != "minecraft:cushion" {
			continue
		}
		other := e.H().Type().BBox(e).Translate(e.Position())
		if other.IntersectsWith(box) || cushionClicked(other, click) {
			return false
		}
	}

	var name string
	if held, _ := user.HeldItems(); held.Item() == c {
		name = held.CustomName()
	}
	spawnPos := mgl64.Vec3{float64(pos[0]) + 0.5, click[1], float64(pos[2]) + 0.5}
	opts := world.EntitySpawnOpts{Position: spawnPos, Rotation: cube.Rotation{cushionYaw(user.Rotation().Yaw()), 0}, NameTag: name}
	tx.AddEntity(tx.World().EntityRegistry().Config().Cushion(opts, c.Colour))
	tx.PlaySound(spawnPos, sound.CushionPlace{})

	ctx.SubtractFromCount(1)
	return true
}

// cushionClicked reports if the point clicked lies within the box of another cushion.
func cushionClicked(box cube.BBox, click mgl64.Vec3) bool {
	lo, hi := box.Min(), box.Max()
	return click[0] > lo[0] && click[0] < hi[0] && click[2] > lo[2] && click[2] < hi[2] &&
		click[1]-0.001 < hi[1] && click[1]+0.001 > lo[1]
}

// cushionYaw returns the yaw of a placed cushion, facing the user rounded to 90 degrees like vanilla.
func cushionYaw(userYaw float64) float64 {
	w := math.Mod(userYaw, 360)
	if w < 0 {
		w += 360
	}
	yaw := math.Floor((w-135)/90) * 90
	if yaw >= 180 {
		yaw -= 360
	}
	return yaw
}

// MaxCount ...
func (Cushion) MaxCount() int {
	return 16
}

// FuelInfo ...
func (Cushion) FuelInfo() FuelInfo {
	return newFuelInfo(time.Second * 5 / 2)
}

// EncodeItem ...
func (c Cushion) EncodeItem() (name string, meta int16) {
	return "minecraft:" + c.Colour.String() + "_cushion", 0
}
