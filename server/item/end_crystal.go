package item

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EndCrystal is an item that can be placed on obsidian or bedrock to spawn an End crystal entity.
type EndCrystal struct{}

// UseOnBlock ...
func (e EndCrystal) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, _ User, ctx *UseContext) bool {
	if face != cube.FaceUp {
		return false
	}
	clickedBlock, _ := tx.Block(pos).EncodeBlock()
	if clickedBlock != "minecraft:obsidian" && clickedBlock != "minecraft:bedrock" {
		return false
	}

	above, twoAbove := pos.Side(cube.FaceUp), pos.Side(cube.FaceUp).Side(cube.FaceUp)
	if above.OutOfBounds(tx.Range()) || twoAbove.OutOfBounds(tx.Range()) {
		return false
	}
	if !endCrystalPlacementReplaceable(tx.Block(above)) || !endCrystalPlacementReplaceable(tx.Block(twoAbove)) {
		return false
	}

	box := cube.Box(0, 0, 0, 1, 2, 1).Translate(pos.Vec3())
	for entity := range tx.EntitiesWithin(box.Grow(2)) {
		if entity.H().Type().BBox(entity).Translate(entity.Position()).IntersectsWith(box) {
			return false
		}
	}

	if tx.World().Dimension() == world.End {
		flame := fire()
		tx.SetBlock(above, flame, nil)
		tx.ScheduleBlockUpdate(above, flame, time.Duration(30+rand.IntN(10))*time.Second/20)
	}
	opts := world.EntitySpawnOpts{Position: pos.Vec3().Add(mgl64.Vec3{0.5, 1, 0.5})}
	tx.AddEntity(tx.World().EntityRegistry().Config().EndCrystal(opts))
	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (EndCrystal) EncodeItem() (name string, meta int16) {
	return "minecraft:end_crystal", 0
}

func endCrystalPlacementReplaceable(b world.Block) bool {
	replacement := air()
	if b == replacement {
		return true
	}
	return replaceableWith(b, replacement)
}
