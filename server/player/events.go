package player

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"net"
	"time"
)

// EventAttackEntity occurs when a player is attacking an entity using the item
// held in its hand. Cancel() may be called to cancel an attack, which will
// cancel damage dealt to the entity and will stop the entity from being
// knocked back. The attacked entity may not be alive
// (implements entity.Living), in which case no damage will be dealt and the
// entity won't be knocked back. The entity attacked may also be immune when
// this method is called, in which case no damage or knock-back will be dealt.
// Pointers to the knock back force and height associated with a specific
// EventAttackEntity event are provided, which can be modified. The attack can
// be a critical attack, which would increase damage by a factor of 1.5 and
// spawn critical hit particles around the target entity. These particles will
// not be displayed if no damage is dealt.
type EventAttackEntity struct {
	Player   *Player
	Entity   world.Entity
	Force    *float64
	Height   *float64
	Critical *bool
	*event.Context
}

// EventBlockBreak occurs when a player finishes breaking a block. Cancel() may
// be called to prevent a block from being broken. A pointer to a slice of the
// block associated with this event's drops is provided, and may be altered to
// change what items will be dropped.
type EventBlockBreak struct {
	Player     *Player
	Position   cube.Pos
	Drops      *[]item.Stack
	Experience *int
	*event.Context
}

// EventBlockPick occurs when a player is picking a block. Cancel() may be
// called to prevent a block from being picked.
type EventBlockPick struct {
	Player   *Player
	Position cube.Pos
	Block    world.Block
	*event.Context
}

// EventBlockPlace occurs when a player attempts to place a block. Cancel() may
// be called to prevent a block being placed.
type EventBlockPlace struct {
	Player   *Player
	Position cube.Pos
	Block    world.Block
	*event.Context
}

// EventChangeWorld occurs when a player is added to a new world. Before may be
// nil. Before is nil when a player joins a world for the first time since
// being accepted to the server.
type EventChangeWorld struct {
	Player *Player
	Before *world.World
	After  *world.World
}

// EventChat occurs when a message is sent by a player. Cancel() may be called
// to prevent a message from being sent. Pointers to both the prefix and
// message strings that will appear in the chat are provided.
type EventChat struct {
	Player  *Player
	Prefix  *string
	Message *string
	*event.Context
}

// EventCommandExecution occurs when a player attempts to execute a command.
// Cancel() may be called to prevent an execution of a command.
type EventCommandExecution struct {
	Player    *Player
	Command   cmd.Command
	Arguments []string
	*event.Context
}

// EventDeath occurs when a player dies. A pointer to a boolean keepInventory
// is provided.
type EventDeath struct {
	Player        *Player
	Source        world.DamageSource
	KeepInventory *bool
}

type EventExperienceGain struct {
	Player *Player
	Amount *int
	*event.Context
}

type EventFoodLoss struct {
	Player *Player
	From   int
	To     *int
	*event.Context
}

type EventHeal struct {
	Player *Player
	Amount *float64
	Source world.HealingSource
	*event.Context
}

type EventHurt struct {
	Player         *Player
	Damage         *float64
	AttackImmunity *time.Duration
	Source         world.DamageSource
	*event.Context
}

type EventItemConsume struct {
	Player    *Player
	ItemStack item.Stack
	*event.Context
}

type EventItemDamage struct {
	Player    *Player
	ItemStack item.Stack
	Damage    int
	*event.Context
}

type EventItemDrop struct {
	Player *Player
	Item   *entity.Item
	*event.Context
}

type EventItemPickup struct {
	Player    *Player
	ItemStack item.Stack
	*event.Context
}

type EventItemUse struct {
	Player *Player
	*event.Context
}

type EventItemUseOnBlock struct {
	Player        *Player
	Position      cube.Pos
	Face          cube.Face
	ClickPosition mgl64.Vec3
	*event.Context
}

type EventItemUseOnEntity struct {
	Player *Player
	Entity world.Entity
	*event.Context
}

type EventJump struct {
	Player *Player
}

type EventMove struct {
	Player      *Player
	NewPosition mgl64.Vec3
	NewYaw      float64
	NewPitch    float64
	*event.Context
}

type EventPunchAir struct {
	Player *Player
	*event.Context
}

type EventQuit struct {
	Player *Player
}

type EventRespawn struct {
	Player   *Player
	Position *mgl64.Vec3
	World    **world.World
}

type EventSignEdit struct {
	Player  *Player
	Sign    block.Sign
	NewText *string
	*event.Context
}

type EventSkinChange struct {
	Player *Player
	Skin   *skin.Skin
	*event.Context
}

type EventStartBreak struct {
	Player   *Player
	Position cube.Pos
	*event.Context
}

type EventTeleport struct {
	Player      *Player
	NewPosition mgl64.Vec3
	*event.Context
}

type EventToggleSneak struct {
	Player     *Player
	IsSneaking bool
	*event.Context
}

type EventToggleSprint struct {
	Player      *Player
	IsSprinting bool
	*event.Context
}

type EventTransfer struct {
	Player  *Player
	Address *net.UDPAddr
	*event.Context
}
