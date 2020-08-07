package item

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/effect"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// MaxCounter represents an item that has a specific max count. By default, each item will be expected to have
// a maximum count of 64. MaxCounter may be implemented to change this behaviour.
type MaxCounter interface {
	// MaxCount returns the maximum number of items that a stack may be composed of. The number returned must
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
	UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user User, ctx *UseContext) bool
}

// UsableOnEntity represents an item that may be used on an entity. If an item implements this interface, the
// UseOnEntity method is called whenever the item is used on an entity.
type UsableOnEntity interface {
	// UseOnEntity is called when an item is used on an entity. The world passed is the world that the item is
	// used in, and the entity clicked and the user of the item are also passed.
	// UseOnEntity returns a bool indicating if the item was used successfully.
	UseOnEntity(e world.Entity, w *world.World, user User, ctx *UseContext) bool
}

// Usable represents an item that may be used 'in the air'. If an item implements this interface, the Use
// method is called whenever the item is used while pointing at the air. (For example, when throwing an egg.)
type Usable interface {
	// Use is called when the item is used in the air. The user that used the item and the world that the item
	// was used in are passed to the method.
	// Use returns a bool indicating if the item was used successfully.
	Use(w *world.World, user User, ctx *UseContext) bool
}

// Consumable represents an item that may consumed by a player. If an item implements this interface, a player
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
	Consume(w *world.World, c Consumer) Stack
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
}

// DefaultConsumeDuration is the default duration that consuming an item takes. Dried kelp takes half this
// time to be consumed.
const DefaultConsumeDuration = (time.Second * 161) / 100

// UseContext is passed to every item Use methods. It may be used to subtract items or to deal damage to them
// after the action is complete.
type UseContext struct {
	Damage     int
	CountSub   int
	IgnoreAABB bool
	// NewItem is the item that is added after the item is used. If the player no longer has an item in the
	// hand, it'll be added there.
	NewItem Stack
	// NewItemSurvivalOnly will add any new items only in survival mode.
	NewItemSurvivalOnly bool
}

// DamageItem damages the item used by d points.
func (ctx *UseContext) DamageItem(d int) { ctx.Damage += d }

// SubtractFromCount subtracts d from the count of the item stack used.
func (ctx *UseContext) SubtractFromCount(d int) { ctx.CountSub += d }

// Weapon is an item that may be used as a weapon. It has an attack damage which may be different to the 2
// damage that attacking with an empty hand deals.
type Weapon interface {
	// AttackDamage returns the custom attack damage of the weapon. The damage returned must not be negative.
	AttackDamage() float64
}

// nameable represents a block that may be named. These are often containers such as chests, which have a
// name displayed in their interface.
type nameable interface {
	// WithName returns the block itself, except with a custom name applied to it.
	WithName(a ...interface{}) world.Item
}

// User represents an entity that is able to use an item in the world, typically entities such as players,
// which interact with the world using an item.
type User interface {
	// Facing returns the direction that the user is facing.
	Facing() world.Direction
	// Position returns the current position of the user in the world.
	Position() mgl64.Vec3
	// Yaw returns the yaw of the entity. This is horizontal rotation (rotation around the vertical axis), and
	// is 0 when the entity faces forward.
	Yaw() float64
	// Pitch returns the pitch of the entity. This is vertical rotation (rotation around the horizontal axis),
	// and is 0 when the entity faces forward.
	Pitch() float64
	HeldItems() (right, left Stack)
	SetHeldItems(right, left Stack)
}

// Collector represents an entity in the world that is able to collect an item, typically an entity such as
// a player or a zombie.
type Collector interface {
	world.Entity
	// Collect collects the stack passed. It is called if the Collector is standing near an item entity that
	// may be picked up.
	// The count of items collected from the stack n is returned.
	Collect(stack Stack) (n int)
}

// Carrier represents an entity that is able to carry an item.
type Carrier interface {
	// HeldItems returns the items currently held by the entity. Viewers of the entity will be able to see
	// these items.
	HeldItems() (mainHand, offHand Stack)
}

// BeaconPayment represents an item that may be used as payment for a beacon to select effects to be broadcast
// to surrounding players.
type BeaconPayment interface {
	PayableForBeacon() bool
}
