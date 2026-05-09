package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EndCrystal is an item that can be placed on obsidian or bedrock to spawn an End crystal entity.
type EndCrystal struct{}

// UseOnBlock ...
func (e EndCrystal) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, tx *world.Tx, _ User, ctx *UseContext) bool {
	clickedBlock := blockName(tx.Block(pos))
	if clickedBlock != "minecraft:obsidian" && clickedBlock != "minecraft:bedrock" {
		return false
	}

	above, twoAbove := pos.Side(cube.FaceUp), pos.Side(cube.FaceUp).Side(cube.FaceUp)
	if above.OutOfBounds(tx.Range()) || twoAbove.OutOfBounds(tx.Range()) {
		return false
	}
	if blockName(tx.Block(above)) != "minecraft:air" || blockName(tx.Block(twoAbove)) != "minecraft:air" {
		return false
	}

	box := cube.Box(
		float64(pos[0]), float64(pos[1]), float64(pos[2]),
		float64(pos[0]+1), float64(pos[1]+2), float64(pos[2]+1),
	)
	for entity := range tx.EntitiesWithin(box.Grow(2)) {
		if entity.H().Type().BBox(entity).Translate(entity.Position()).IntersectsWith(box) {
			return false
		}
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

func blockName(b world.Block) string {
	name, _ := b.EncodeBlock()
	return name
}
