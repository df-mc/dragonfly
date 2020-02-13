package player

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/damage"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/event"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/cmd"
	"net"
)

// Handle handles events that are called by a player. Implementations of Handler may be used to listen to
// specific events such as when a player chats or moves.
type Handler interface {
	// HandleMove handles the movement of a player. ctx.Cancel() may be called to cancel the movement event.
	// The new position, yaw and pitch are passed.
	HandleMove(ctx *event.Context, newPos mgl32.Vec3, newYaw, newPitch float32)
	// HandleTeleport handles the teleportation of a player. ctx.Cancel() may be called to cancel it.
	HandleTeleport(ctx *event.Context, pos mgl32.Vec3)
	// HandleChat handles a message sent in the chat by a player. ctx.Cancel() may be called to cancel the
	// message being sent in chat.
	// The message may be changed by assigning to *message.
	HandleChat(ctx *event.Context, message *string)
	// HandleHurt handles the player being hurt by any damage source. ctx.Cancel() may be called to cancel the
	// damage being dealt to the player.
	// The damage dealt to the player may be changed by assigning to *damage.
	HandleHurt(ctx *event.Context, damage *float32, src damage.Source)
	// HandleBlockBreak handles a block that is being broken by a player. ctx.Cancel() may be called to cancel
	// the block being broken.
	HandleBlockBreak(ctx *event.Context, pos block.Position)
	// HandleItemUse handles the player using an item in the air. It is called for each item, although most
	// will not actually do anything. Items such as snowballs may be thrown if HandleItemUse does not cancel
	// the context using ctx.Cancel(). It is also called if the player is holding no item.
	HandleItemUse(ctx *event.Context)
	// HandleItemUseOnBlock handles the player using the item held in its main hand on a block at the block
	// position passed. The face of the block clicked is also passed, along with the relative click position.
	// The click position has X, Y and Z values which are all in the range 0.0-1.0. It is also called if the
	// player is holding no item.
	HandleItemUseOnBlock(ctx *event.Context, pos block.Position, face block.Face, clickPos mgl32.Vec3)
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
	HandleAttackEntity(ctx *event.Context, e world.Entity)
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

// HandleMove ...
func (NopHandler) HandleMove(ctx *event.Context, newPos mgl32.Vec3, newYaw, newPitch float32) {}

// HandleTeleport ...
func (NopHandler) HandleTeleport(ctx *event.Context, pos mgl32.Vec3) {}

// HandleCommandExecution ...
func (NopHandler) HandleCommandExecution(ctx *event.Context, command cmd.Command, args []string) {}

// HandleTransfer ...
func (NopHandler) HandleTransfer(ctx *event.Context, addr *net.UDPAddr) {}

// HandleChat ...
func (NopHandler) HandleChat(ctx *event.Context, message *string) {}

// HandleBlockBreak ...
func (NopHandler) HandleBlockBreak(ctx *event.Context, pos block.Position) {}

// HandleItemUse ...
func (NopHandler) HandleItemUse(ctx *event.Context) {}

// HandleItemUseOnBlock ...
func (NopHandler) HandleItemUseOnBlock(ctx *event.Context, pos block.Position, face block.Face, clickPos mgl32.Vec3) {
}

// HandleItemUseOnEntity ...
func (NopHandler) HandleItemUseOnEntity(ctx *event.Context, e world.Entity) {}

// HandleHurt ...
func (NopHandler) HandleHurt(ctx *event.Context, damage *float32, src damage.Source) {}

// HandleAttackEntity ...
func (NopHandler) HandleAttackEntity(ctx *event.Context, e world.Entity) {}

// HandleQuit ...
func (NopHandler) HandleQuit() {}
