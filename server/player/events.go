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

type EventAttackEntity struct {
	Player   *Player
	Entity   world.Entity
	Force    *float64
	Height   *float64
	Critical *bool
	*event.Context
}

type EventBlockBreak struct {
	Player     *Player
	Position   cube.Pos
	Drops      *[]item.Stack
	Experience *int
	*event.Context
}

type EventBlockPick struct {
	Player   *Player
	Position cube.Pos
	Block    world.Block
	*event.Context
}

type EventBlockPlace struct {
	Player   *Player
	Position cube.Pos
	Block    world.Block
	*event.Context
}

type EventChangeWorld struct {
	Player *Player
	Before *world.World
	After  *world.World
}

type EventChat struct {
	Player  *Player
	Prefix  *string
	Message *string
	*event.Context
}

type EventCommandExecution struct {
	Player    *Player
	Command   cmd.Command
	Arguments []string
	*event.Context
}

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
