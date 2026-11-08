package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// ArmourStandBehaviourConfig holds optional parameters for
// ArmourStandBehaviour.
type ArmourStandBehaviourConfig struct {
	Armour *inventory.Armour
	// MainHand is the item equipped in the main hand slot of the armour stand.
	MainHand item.Stack
	// OffHand is the item equipped in the offhand slot of the armour stand.
	OffHand item.Stack
	// PoseIndex is the pose index of the armour stand. Possible values range
	// from 0 to 12 (inclusive).
	PoseIndex int
}

// Apply ...
func (conf ArmourStandBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
}

// New creates an ArmourStandBehaviour using the optional parameters in conf.
func (conf ArmourStandBehaviourConfig) New() *ArmourStandBehaviour {
	a := &ArmourStandBehaviour{
		conf:         conf,
		armours:      conf.Armour,
		lastOnGround: true,
	}
	a.passive = PassiveBehaviourConfig{
		Gravity: 0.04,
		Drag:    0.02,
		Tick:    a.tick,
	}.New()
	return a
}

// ArmourStandBehaviour implements the behaviour for armour stand entities.
type ArmourStandBehaviour struct {
	conf ArmourStandBehaviourConfig

	passive *PassiveBehaviour
	hurt    time.Duration
	armours *inventory.Armour

	lastOnGround bool
}

// armourStandDropOffset returns the position offset at which an item should be
// dropped from an armour stand, based on the type of item.
func armourStandDropOffset(stack item.Stack) mgl64.Vec3 {
	var offset mgl64.Vec3
	switch stack.Item().(type) {
	case item.Helmet:
		offset[1] = 1.8
	case item.Chestplate:
		offset[1] = 1.4
	case item.Leggings:
		offset[1] = 0.6
	case item.Boots:
		offset[1] = 0.2
	default:
		offset[1] = 1.4
	}
	return offset
}

// Tick ...
func (a *ArmourStandBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	return a.passive.Tick(e, tx)
}

// tick ...
func (a *ArmourStandBehaviour) tick(e *Ent, tx *world.Tx) {
	if a.hurt > 0 {
		a.hurt -= time.Millisecond * 50
		if a.hurt < 0 {
			a.hurt = 0
		}
		a.updateState(e)
	}

	if a.lastOnGround != a.passive.mc.OnGround() {
		if !a.lastOnGround {
			tx.PlaySound(e.Position(), sound.ArmourStandLand{})
		}
		a.lastOnGround = a.passive.mc.OnGround()
	}
}

// Explode ...
func (a *ArmourStandBehaviour) Explode(e *Ent, src mgl64.Vec3, impact float64, config block.ExplosionConfig) {
	a.passive.Explode(e, src, impact, config)
}

// AcceptItem ...
func (a *ArmourStandBehaviour) AcceptItem(e *Ent, from world.Entity, _ *world.Tx, ctx *item.UseContext) bool {
	if sneaker, ok := from.(interface {
		Sneaking() bool
	}); ok && sneaker.Sneaking() {
		a.SetPoseIndex(e, (a.PoseIndex()+1)%13)
		return false
	}

	heldItems, ok := from.(interface {
		HeldItems() (mainHand, offHand item.Stack)
	})
	if !ok {
		return false
	}
	mainHand, _ := heldItems.HeldItems()
	if mainHand.Empty() {
		return false
	}
	var (
		dropItem item.Stack
		add      = mainHand.Grow(-mainHand.Count() + 1)
	)
	i := add.Item()
	inv := a.armours
	if _, isArmour := i.(item.Armour); isArmour {
		switch i.(type) {
		case item.Helmet:
			dropItem = inv.Helmet()
			inv.SetHelmet(add)
		case item.Chestplate:
			dropItem = inv.Chestplate()
			inv.SetChestplate(add)
		case item.Leggings:
			dropItem = inv.Leggings()
			inv.SetLeggings(add)
		case item.Boots:
			dropItem = inv.Boots()
			inv.SetBoots(add)
		}
		a.updateArmours(e)
	} else {
		it, left := a.HeldItems()
		dropItem = it
		a.SetHeldItems(e, add, left)
	}

	ctx.SubtractFromCount(1)
	ctx.NewItem, ctx.NewItemSurvivalOnly = dropItem, true
	return true
}

// Attack ...
func (a *ArmourStandBehaviour) Attack(e *Ent, _ world.Entity, tx *world.Tx) {
	if a.hurt > 0 {
		a.destroy(e, tx)
		return
	}
	tx.PlaySound(e.Position(), sound.ArmourStandPlace{})
	a.hurt = time.Millisecond * 300
	a.updateState(e)
}

// destroy destroys the armour stand, dropping all equipped items.
func (a *ArmourStandBehaviour) destroy(e *Ent, tx *world.Tx) {
	tx.PlaySound(e.Position(), sound.ArmourStandBreak{})
	a.dropAll(e, tx)
	_ = e.Close()
}

// HurtDuration returns the remaining hurt duration of the armour stand.
func (a *ArmourStandBehaviour) HurtDuration() time.Duration {
	return a.hurt
}

// dropAll drops all equipped items of the armour stand.
func (a *ArmourStandBehaviour) dropAll(e *Ent, tx *world.Tx) {
	dropPos := e.Position()
	drop := func(stack item.Stack) {
		if stack.Empty() {
			return
		}
		tx.AddEntity(NewItem(world.EntitySpawnOpts{Position: dropPos.Add(armourStandDropOffset(stack))}, stack))
	}
	for _, i := range a.armours.Items() {
		drop(i)
	}
	mainHand, offHand := a.HeldItems()
	drop(mainHand)
	drop(offHand)
}

// Armour returns the armour equipped on the armour stand.
func (a *ArmourStandBehaviour) Armour() *inventory.Armour {
	return a.armours
}

// HeldItems returns the items equipped in the main hand and offhand slots of the armour stand.
func (a *ArmourStandBehaviour) HeldItems() (mainHand, offHand item.Stack) {
	return a.conf.MainHand, a.conf.OffHand
}

// PoseIndex returns the pose index of the armour stand. Possible values range
// from 0 to 12 (inclusive).
func (a *ArmourStandBehaviour) PoseIndex() int {
	return a.conf.PoseIndex
}

// SetHeldItems sets the items equipped in the main hand and offhand slots of the armour stand.
func (a *ArmourStandBehaviour) SetHeldItems(e *Ent, mainHand, offHand item.Stack) {
	a.conf.MainHand = mainHand
	a.conf.OffHand = offHand
	a.updateHeldItems(e)
}

// SetPoseIndex sets the pose index of the armour stand. Possible values range
// from 0 to 12 (inclusive).
func (a *ArmourStandBehaviour) SetPoseIndex(e *Ent, poseIndex int) {
	if poseIndex < 0 || poseIndex > 12 {
		panic("pose index must be between 0 and 12")
	}
	a.conf.PoseIndex = poseIndex
	a.updateState(e)
}

// updateArmours updates the armour stand's equipped items for all viewers.
func (a *ArmourStandBehaviour) updateArmours(e *Ent) {
	for _, v := range e.tx.Viewers(e.Position()) {
		v.ViewEntityArmour(e)
	}
}

// updateHeldItems updates the armour stand's held items for all viewers.
func (a *ArmourStandBehaviour) updateHeldItems(e *Ent) {
	for _, v := range e.tx.Viewers(e.Position()) {
		v.ViewEntityItems(e)
	}
}

// updateState updates the armour stand's state for all viewers.
func (a *ArmourStandBehaviour) updateState(e *Ent) {
	for _, v := range e.tx.Viewers(e.data.Pos) {
		v.ViewEntityState(e)
	}
}
