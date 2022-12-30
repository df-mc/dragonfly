package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// NewArrow creates a new Arrow and returns it. It is equivalent to calling NewTippedArrow with `potion.Potion{}` as
// tip.
func NewArrow(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity) *Ent {
	return NewTippedArrowWithDamage(pos, yaw, pitch, 2.0, owner, potion.Potion{})
}

// NewArrowWithDamage creates a new Arrow with the given base damage, and returns it. It is equivalent to calling
// NewTippedArrowWithDamage with `potion.Potion{}` as tip.
func NewArrowWithDamage(pos mgl64.Vec3, yaw, pitch, damage float64, owner world.Entity) *Ent {
	return NewTippedArrowWithDamage(pos, yaw, pitch, damage, owner, potion.Potion{})
}

// NewTippedArrow creates a new Arrow with a potion effect added to an entity when hit.
func NewTippedArrow(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity, tip potion.Potion) *Ent {
	return NewTippedArrowWithDamage(pos, yaw, pitch, 2.0, owner, tip)
}

// NewTippedArrowWithDamage creates a new Arrow with a potion effect added to an entity when hit and, and returns it.
// It uses the given damage as the base damage.
func NewTippedArrowWithDamage(pos mgl64.Vec3, yaw, pitch, damage float64, owner world.Entity, tip potion.Potion) *Ent {
	conf := arrowConf
	conf.Damage = damage
	conf.Potion = tip
	a := Config{Behaviour: conf.New(owner)}.New(ArrowType{}, pos)
	a.rot = cube.Rotation{yaw, pitch}
	return a
}

var arrowConf = ProjectileBehaviourConfig{
	Gravity:               0.05,
	Drag:                  0.01,
	Damage:                2.0,
	Sound:                 sound.ArrowHit{},
	SurviveBlockCollision: true,
}

// boolByte returns 1 if the bool passed is true, or 0 if it is false.
func boolByte(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// ArrowType is a world.EntityType implementation for Arrow.
type ArrowType struct{}

func (ArrowType) EncodeEntity() string { return "minecraft:arrow" }
func (ArrowType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (ArrowType) DecodeNBT(m map[string]any) world.Entity {
	pot := potion.From(nbtconv.Int32(m, "auxValue") - 1)
	arr := NewTippedArrowWithDamage(nbtconv.Vec3(m, "Pos"), float64(nbtconv.Float32(m, "Yaw")), float64(nbtconv.Float32(m, "Pitch")), float64(nbtconv.Float32(m, "Damage")), nil, pot)
	b := arr.conf.Behaviour.(*ProjectileBehaviour)
	arr.vel = nbtconv.Vec3(m, "Motion")
	b.conf.DisablePickup = !nbtconv.Bool(m, "player")
	if !nbtconv.Bool(m, "isCreative") {
		b.conf.PickupItem = item.NewStack(item.Arrow{Tip: pot}, 1)
	}
	arr.fireDuration = time.Duration(nbtconv.Int16(m, "Fire")) * time.Second / 20
	b.conf.KnockBackAddend = (enchantment.Punch{}).KnockBackMultiplier() * float64(nbtconv.Uint8(m, "enchantPunch"))
	if _, ok := m["StuckToBlockPos"]; ok {
		b.collisionPos = nbtconv.Pos(m, "StuckToBlockPos")
		b.collided = true
	}
	return arr
}

func (ArrowType) EncodeNBT(e world.Entity) map[string]any {
	a := e.(*Ent)
	b := a.conf.Behaviour.(*ProjectileBehaviour)
	yaw, pitch := a.Rotation().Elem()
	data := map[string]any{
		"Pos":          nbtconv.Vec3ToFloat32Slice(a.Position()),
		"Yaw":          float32(yaw),
		"Pitch":        float32(pitch),
		"Motion":       nbtconv.Vec3ToFloat32Slice(a.Velocity()),
		"Damage":       float32(b.conf.Damage),
		"Fire":         int16(a.OnFireDuration() * 20),
		"enchantPunch": byte(b.conf.KnockBackAddend / (enchantment.Punch{}).KnockBackMultiplier()),
		"auxValue":     int32(b.conf.Potion.Uint8() + 1),
		"player":       boolByte(!b.conf.DisablePickup),
		"isCreative":   boolByte(b.conf.PickupItem.Empty()),
	}
	// TODO: Save critical flag if Minecraft ever saves it?
	if b.collided {
		data["StuckToBlockPos"] = nbtconv.PosToInt32Slice(b.collisionPos)
	}
	return data
}
