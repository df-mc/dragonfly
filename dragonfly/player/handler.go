package player

import (
	"github.com/df-mc/dragonfly/dragonfly/cmd"
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"github.com/df-mc/dragonfly/dragonfly/entity/healing"
	"github.com/df-mc/dragonfly/dragonfly/event"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"net"
)

// Handler handles events that are called by a player. Implementations of Handler may be used to listen to
// specific events such as when a player chats or moves.
type Handler interface {
	// HandleMove handles the movement of a player. ctx.Cancel() may be called to cancel the movement event.
	// The new position, yaw and pitch are passed.
	HandleMove(ctx *event.Context, newPos mgl64.Vec3, newYaw, newPitch float64)
	// HandleTeleport handles the teleportation of a player. ctx.Cancel() may be called to cancel it.
	HandleTeleport(ctx *event.Context, pos mgl64.Vec3)
	// HandleChat handles a message sent in the chat by a player. ctx.Cancel() may be called to cancel the
	// message being sent in chat.
	// The message may be changed by assigning to *message.
	HandleChat(ctx *event.Context, message *string)
	// HandleFoodLoss handles the food bar of a player depleting naturally, for example because the player was
	// sprinting and jumping. ctx.Cancel() may be called to cancel the food points being lost.
	HandleFoodLoss(ctx *event.Context, from, to int)
	// HandleHeal handles the player being healed by a healing source. ctx.Cancel() may be called to cancel
	// the healing.
	// The health added may be changed by assigning to *health.
	HandleHeal(ctx *event.Context, health *float64, src healing.Source)
	// HandleHurt handles the player being hurt by any damage source. ctx.Cancel() may be called to cancel the
	// damage being dealt to the player.
	// The damage dealt to the player may be changed by assigning to *damage.
	HandleHurt(ctx *event.Context, damage *float64, src damage.Source)
	// HandleDeath handles the player dying to a particular damage cause.
	HandleDeath(src damage.Source)
	// HandleRespawn handles the respawning of the player in the world. The spawn position passed may be
	// changed by assigning to *pos.
	HandleRespawn(pos *mgl64.Vec3)
	// HandleStartBreak handles the player starting to break a block at the position passed. ctx.Cancel() may
	// be called to stop the player from breaking the block completely.
	HandleStartBreak(ctx *event.Context, pos world.BlockPos)
	// HandleBlockBreak handles a block that is being broken by a player. ctx.Cancel() may be called to cancel
	// the block being broken.
	HandleBlockBreak(ctx *event.Context, pos world.BlockPos)
	// HandleBlockPlace handles the player placing a specific block at a position in its world. ctx.Cancel()
	// may be called to cancel the block being placed.
	HandleBlockPlace(ctx *event.Context, pos world.BlockPos, b world.Block)
	// HandleItemUse handles the player using an item in the air. It is called for each item, although most
	// will not actually do anything. Items such as snowballs may be thrown if HandleItemUse does not cancel
	// the context using ctx.Cancel(). It is not called if the player is holding no item.
	HandleItemUse(ctx *event.Context)
	// HandleItemUseOnBlock handles the player using the item held in its main hand on a block at the block
	// position passed. The face of the block clicked is also passed, along with the relative click position.
	// The click position has X, Y and Z values which are all in the range 0.0-1.0. It is also called if the
	// player is holding no item.
	HandleItemUseOnBlock(ctx *event.Context, pos world.BlockPos, face world.Face, clickPos mgl64.Vec3)
	// HandleItemUseOnEntity handles the player using the item held in its main hand on an entity passed to
	// the method.
	// HandleItemUseOnEntity is always called when a player uses an item on an entity, regardless of whether
	// the item actually does anything when used on an entity. It is also called if the player is holding no
	// item.
	HandleItemUseOnEntity(ctx *event.Context, e world.Entity)
	// HandleAttackEntity handles the player attacking an entity using the item held in its hand. ctx.Cancel()
	// may be called to cancel the attack, which will cancel damage dealt to the target and will stop the
	// entity from being knocked back.
	// The entity attacked may not be alive (implements entity.Living), in which case no damage will be dealt
	// and the target won't be knocked back.
	// The entity attacked may also be immune when this method is called, in which case no damage and knock-
	// back will be dealt.
	HandleAttackEntity(ctx *event.Context, e world.Entity)
	// HandleItemDamage handles the event wherein the item either held by the player or as armour takes
	// damage through usage.
	// The type of the item may be checked to determine whether it was armour or a tool used. The damage to
	// the item is passed.
	HandleItemDamage(ctx *event.Context, i item.Stack, damage int)
	// HandleItemPickup handles the player picking up an item from the ground. The item stack laying on the
	// ground is passed. ctx.Cancel() may be called to prevent the player from picking up the item.
	HandleItemPickup(ctx *event.Context, i item.Stack)
	// HandleTransfer handles a player being transferred to another server. ctx.Cancel() may be called to
	// cancel the transfer.
	HandleTransfer(ctx *event.Context, addr *net.UDPAddr)
	// HandleCommandExecution handles the command execution of a player, who wrote a command in the chat.
	// ctx.Cancel() may be called to cancel the command execution.
	HandleCommandExecution(ctx *event.Context, command cmd.Command, args []string)
	// HandleQuit handles the closing of a player. It is always called when the player is disconnected,
	// regardless of the reason.
	HandleQuit()
}

// NopHandler implements the Handler interface but does not execute any code when an event is called. The
// default handler of players is set to NopHandler.
// Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = (*NopHandler)(nil)

// HandleMove ...
func (NopHandler) HandleMove(*event.Context, mgl64.Vec3, float64, float64) {}

// HandleTeleport ...
func (NopHandler) HandleTeleport(*event.Context, mgl64.Vec3) {}

// HandleCommandExecution ...
func (NopHandler) HandleCommandExecution(*event.Context, cmd.Command, []string) {}

// HandleTransfer ...
func (NopHandler) HandleTransfer(*event.Context, *net.UDPAddr) {}

// HandleChat ...
func (NopHandler) HandleChat(*event.Context, *string) {}

// HandleStartBreak ...
func (NopHandler) HandleStartBreak(*event.Context, world.BlockPos) {}

// HandleBlockBreak ...
func (NopHandler) HandleBlockBreak(*event.Context, world.BlockPos) {}

// HandleBlockPlace ...
func (NopHandler) HandleBlockPlace(*event.Context, world.BlockPos, world.Block) {}

// HandleItemPickup ...
func (NopHandler) HandleItemPickup(*event.Context, item.Stack) {}

// HandleItemUse ...
func (NopHandler) HandleItemUse(*event.Context) {}

// HandleItemUseOnBlock ...
func (NopHandler) HandleItemUseOnBlock(*event.Context, world.BlockPos, world.Face, mgl64.Vec3) {
}

// HandleItemUseOnEntity ...
func (NopHandler) HandleItemUseOnEntity(*event.Context, world.Entity) {}

// HandleItemDamage ...
func (NopHandler) HandleItemDamage(*event.Context, item.Stack, int) {}

// HandleAttackEntity ...
func (NopHandler) HandleAttackEntity(*event.Context, world.Entity) {}

// HandleHurt ...
func (NopHandler) HandleHurt(*event.Context, *float64, damage.Source) {}

// HandleHeal ...
func (NopHandler) HandleHeal(*event.Context, *float64, healing.Source) {}

// HandleFoodLoss ...
func (NopHandler) HandleFoodLoss(*event.Context, int, int) {}

// HandleDeath ...
func (NopHandler) HandleDeath(damage.Source) {}

// HandleRespawn ...
func (NopHandler) HandleRespawn(*mgl64.Vec3) {}

// HandleQuit ...
func (NopHandler) HandleQuit() {}
