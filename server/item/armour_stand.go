package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// ArmourStand is an armour stand item. It can be placed to create an armour stand
// entity that can hold and display armour and other items.
type ArmourStand struct{}

// UseOnBlock ...
func (ArmourStand) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	spawnPos := pos.Side(face).Vec3Middle()
	opts := world.EntitySpawnOpts{Position: spawnPos, Rotation: user.Rotation().Neg()}
	create := tx.World().EntityRegistry().Config().ArmourStand
	tx.AddEntity(create(opts))
	ctx.SubtractFromCount(1)
	tx.PlaySound(spawnPos, sound.ArmourStandPlace{})
	return true
}

// EncodeItem ...
func (ArmourStand) EncodeItem() (name string, meta int16) {
	return "minecraft:armor_stand", 0
}
