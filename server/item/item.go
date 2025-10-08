package item

import (
	"encoding/binary"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"image/color"
	"time"
)

// MaxCounter represents an item that has a specific max count. By default, each item will be expected to have
// a maximum count of 64. MaxCounter may be implemented to change this behaviour.
type MaxCounter interface {
	// MaxCount returns the maximum number of items a stack may be composed of. The number returned must
	// be positive.
	MaxCount() int
}

// UsableOnBlock represents an item that may be used on a block. If an item implements this interface, the
// UseOnBlock method is called whenever the item is used on a block.
type UsableOnBlock interface {
	// UseOnBlock is called when an item is used on a block. The world passed is the world that the item was
	// used in. The user passed is the entity that used the item. Usually this entity is a player.
	// The position of the block that was clicked, along with the clicked face and the position clicked
	// relative to the corner of the block are passed.
	// UseOnBlock returns a bool indicating if the item was used successfully.
	UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool
}

// UsableOnEntity represents an item that may be used on an entity. If an item implements this interface, the
// UseOnEntity method is called whenever the item is used on an entity.
type UsableOnEntity interface {
	// UseOnEntity is called when an item is used on an entity. The world passed is the world that the item is
	// used in, and the entity clicked and the user of the item are also passed.
	// UseOnEntity returns a bool indicating if the item was used successfully.
	UseOnEntity(e world.Entity, tx *world.Tx, user User, ctx *UseContext) bool
}

// Usable represents an item that may be used 'in the air'. If an item implements this interface, the Use
// method is called whenever the item is used while pointing at the air. (For example, when throwing an egg.)
type Usable interface {
	// Use is called when the item is used in the air. The user that used the item and the world that the item
	// was used in are passed to the method.
	// Use returns a bool indicating if the item was used successfully.
	Use(tx *world.Tx, user User, ctx *UseContext) bool
}

// Throwable represents a custom item that can be thrown such as a projectile. This will only have an effect on
// non-vanilla items.
type Throwable interface {
	// SwingAnimation returns true if the client should cause the player's arm to swing when the item is thrown.
	SwingAnimation() bool
}

// OffHand represents an item that can be held in the off hand.
type OffHand interface {
	// OffHand returns true if the item can be held in the off hand.
	OffHand() bool
}

// Consumable represents an item that may be consumed by a player. If an item implements this interface, a player
// may use and hold the item to consume it.
type Consumable interface {
	// AlwaysConsumable specifies if the item is always consumable. Normal food can generally only be consumed
	// when the food bar is not full or when in creative mode. Returning true here means the item can always
	// be consumed, like golden apples or potions.
	AlwaysConsumable() bool
	// ConsumeDuration is the duration consuming the item takes. If the player is using the item for at least
	// this duration, the item will be consumed and have its Consume method called.
	ConsumeDuration() time.Duration
	// Consume consumes one item of the Stack that the Consumable is in. The Stack returned is added back to
	// the inventory after consuming the item. For potions, for example, an empty bottle is returned.
	Consume(tx *world.Tx, c Consumer) Stack
}

// Consumer represents a User that is able to consume Consumable items.
type Consumer interface {
	User
	// Saturate saturates the Consumer's food bar by the amount of food points passed and the saturation by
	// up to as many saturation points as passed. The final saturation will never exceed the final food level.
	Saturate(food int, saturation float64)
	// AddEffect adds an effect.Effect to the Consumer. If the effect is instant, it is applied to the Consumer
	// immediately. If not, the effect is applied to the consumer every time the Tick method is called.
	// AddEffect will overwrite any effects present if the level of the effect is higher than the existing one, or
	// if the effects' levels are equal and the new effect has a longer duration.
	AddEffect(e effect.Effect)
	// RemoveEffect removes any effect that might currently be active on the Consumer.
	RemoveEffect(e effect.Type)
	// Effects returns any effect currently applied to the Consumer. The returned effects are guaranteed not to have
	// expired when returned.
	Effects() []effect.Effect
}

// DefaultConsumeDuration is the default duration that consuming an item takes. Dried kelp takes half this
// time to be consumed.
const DefaultConsumeDuration = (time.Second * 161) / 100

// Drinkable represents a custom item that can be drunk. It is used to make the client show the correct drinking
// animation when a player is using an item. This will only have an effect on non-vanilla items.
type Drinkable interface {
	// Drinkable returns if the item can be drunk or not.
	Drinkable() bool
}

// Glinted represents a custom item that can have a permanent enchantment glint, this glint is purely cosmetic and
// will show regardless of whether it is actually enchanted. An example of this is the enchanted golden apple.
type Glinted interface {
	// Glinted returns whether the item has an enchantment glint.
	Glinted() bool
}

// HandEquipped represents an item that can be 'hand equipped'. This means the item will show up in third person like
// a tool, sword or stick would, giving them a different orientation in the hand and making them slightly bigger.
type HandEquipped interface {
	// HandEquipped returns whether the item is hand equipped.
	HandEquipped() bool
}

// Weapon is an item that may be used as a weapon. It has an attack damage which may be different to the 2
// damage that attacking with an empty hand deals.
type Weapon interface {
	// AttackDamage returns the custom attack damage to the weapon. The damage returned must not be negative.
	AttackDamage() float64
}

// Cooldown represents an item that has a cooldown.
type Cooldown interface {
	// Cooldown is the duration of the cooldown.
	Cooldown() time.Duration
}

// nameable represents a block that may be named. These are often containers such as chests, which have a
// name displayed in their interface.
type nameable interface {
	// WithName returns the block itself, except with a custom name applied to it.
	WithName(a ...any) world.Item
}

// Releaser represents an entity that can release items, such as bows.
type Releaser interface {
	User
	// GameMode returns the gamemode of the releaser.
	GameMode() world.GameMode
	// PlaySound plays a world.Sound that only this Releaser can hear.
	PlaySound(sound world.Sound)
}

// Releasable represents an item that can be released.
type Releasable interface {
	// Release is called when an item is released.
	Release(releaser Releaser, tx *world.Tx, ctx *UseContext, duration time.Duration)
	// Requirements returns the required items to release this item.
	Requirements() []Stack
}

// Chargeable represents an item that can be charged.
type Chargeable interface {
	// Charge is called when an item is being used.
	Charge(releaser Releaser, tx *world.Tx, ctx *UseContext, duration time.Duration) bool
	// ContinueCharge continues the charge.
	ContinueCharge(releaser Releaser, tx *world.Tx, ctx *UseContext, duration time.Duration)
	// ReleaseCharge is called when an item is being released.
	ReleaseCharge(releaser Releaser, tx *world.Tx, ctx *UseContext) bool
}

// User represents an entity that is able to use an item in the world, typically entities such as players,
// which interact with the world using an item.
type User interface {
	Carrier
	SetHeldItems(mainHand, offHand Stack)

	UsingItem() bool
	ReleaseItem()
	UseItem()
}

// Carrier represents an entity that is able to carry an item.
type Carrier interface {
	world.Entity
	// HeldItems returns the items currently held by the entity. Viewers of the entity will be able to see
	// these items.
	HeldItems() (mainHand, offHand Stack)
}

// BeaconPayment represents an item that may be used as payment for a beacon to select effects to be broadcast
// to surrounding players.
type BeaconPayment interface {
	PayableForBeacon() bool
}

// Compostable represents an item that may be used to fill up a composter.
type Compostable interface {
	// CompostChance returns the chance the item will produce a layer of compost in the range of 0-1.
	CompostChance() float64
}

// nopReleasable represents a releasable item that does nothing.
type nopReleasable struct{}

func (nopReleasable) Release(Releaser, *world.Tx, *UseContext, time.Duration) {}
func (nopReleasable) Requirements() []Stack {
	return []Stack{}
}

// defaultFood represents a consumable item with a default consumption duration.
type defaultFood struct{}

// AlwaysConsumable ...
func (defaultFood) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (d defaultFood) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// eyePosition returns the position of the eyes of the entity if the entity implements entity.Eyed, or the
// actual position if it doesn't.
func eyePosition(e world.Entity) mgl64.Vec3 {
	pos := e.Position()
	if eyed, ok := e.(interface{ EyeHeight() float64 }); ok {
		pos = pos.Add(mgl64.Vec3{0, eyed.EyeHeight()})
	}
	return pos
}

// torsoPosition returns the position of the torso of the entity if the entity implements entity.Torsoed, or the
// actual position if it doesn't.
func torsoPosition(e world.Entity) mgl64.Vec3 {
	pos := e.Position()
	if torso, ok := e.(interface{ TorsoHeight() float64 }); ok {
		pos = pos.Add(mgl64.Vec3{0, torso.TorsoHeight()})
	}
	return pos
}

// Int32FromRGBA converts a color.RGBA into an int32. These int32s are present in things such as signs and dyed leather armour.
func int32FromRGBA(x color.RGBA) int32 {
	if x.R == 0 && x.G == 0 && x.B == 0 {
		// Default to black colour. The default (0x000000) is a transparent colour. Text with this colour will not show
		// up on the sign.
		return int32(-0x1000000)
	}
	return int32(binary.BigEndian.Uint32([]byte{x.A, x.R, x.G, x.B}))
}

// rgbaFromInt32 converts an int32 into a color.RGBA. These int32s are present in things such as signs and dyed leather armour.
func rgbaFromInt32(x int32) color.RGBA {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(x))

	return color.RGBA{A: b[0], R: b[1], G: b[2], B: b[3]}
}

// boolByte returns 1 if the bool passed is true, or 0 if it is false.
func boolByte(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
