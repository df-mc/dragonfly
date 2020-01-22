package player

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/block"
	"github.com/dragonfly-tech/dragonfly/dragonfly/event"
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
	HandleChat(ctx *event.Context, message *string)
	// HandleBlockBreak handles a block that is being broken by a player. ctx.Cancel() may be called to cancel
	// the block being broken.
	HandleBlockBreak(ctx *event.Context, pos block.Position)
	// HandleItemUseOnBlock handles the player using the item held in its main hand on a block at the block
	// position passed. The face of the block clicked is also passed, along with the relative click position.
	// The click position has X, Y and Z values which are all in the range 0.0-1.0.
	HandleItemUseOnBlock(ctx *event.Context, pos block.Position, face block.Face, clickPos mgl32.Vec3)
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
func (n NopHandler) HandleMove(ctx *event.Context, newPos mgl32.Vec3, newYaw, newPitch float32) {}

// HandleTeleport ...
func (n NopHandler) HandleTeleport(ctx *event.Context, pos mgl32.Vec3) {}

// HandleCommandExecution ...
func (n NopHandler) HandleCommandExecution(ctx *event.Context, command cmd.Command, args []string) {}

// HandleTransfer ...
func (n NopHandler) HandleTransfer(ctx *event.Context, addr *net.UDPAddr) {}

// HandleChat ...
func (n NopHandler) HandleChat(ctx *event.Context, message *string) {}

// HandleBlockBreak ...
func (n NopHandler) HandleBlockBreak(ctx *event.Context, pos block.Position) {}

// HandleItemUseOnBlock ...
func (n NopHandler) HandleItemUseOnBlock(ctx *event.Context, pos block.Position, face block.Face, clickPos mgl32.Vec3) {
}

// HandleQuit ...
func (n NopHandler) HandleQuit() {}
