package player

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/event"
	"github.com/sandertv/gophertunnel/minecraft/cmd"
	"net"
)

// Handle handles events that are called by a player. Implementations of Handler may be used to listen to
// specific events such as when a player chats or moves.
type Handler interface {
	// HandleChat handles a message sent in the chat by a player. ctx.Cancel() may be called to cancel the
	// message being sent in chat.
	HandleChat(ctx *event.Context, message string)
	// HandleTransfer handles a player being transferred to another server. ctx.Cancel() may be called to
	// cancel the transfer.
	HandleTransfer(ctx *event.Context, addr *net.UDPAddr)
	// HandleCommandExecution handles the command execution of a player, who wrote a command in the chat.
	// ctx.Cancel() may be called to cancel the command execution.
	HandleCommandExecution(ctx *event.Context, command cmd.Command, args []string)
}

// NopHandler implements the Handler interface but does not execute any code when an event is called. The
// default handler of players is set to NopHandler.
// Users may use type aliases to overwrite methods of NopHandler to avoid having to implement each method.
type NopHandler struct{}

// HandleCommandExecution ...
func (n NopHandler) HandleCommandExecution(ctx *event.Context, command cmd.Command, args []string) {}

// HandleTransfer ...
func (n NopHandler) HandleTransfer(ctx *event.Context, addr *net.UDPAddr) {}

// HandleChat ...
func (n NopHandler) HandleChat(ctx *event.Context, message string) {}
