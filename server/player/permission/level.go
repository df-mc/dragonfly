package permission

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Level represents the permission level a client displays for itself. It only decides what the client shows
// and allows locally, such as the command button in the chat window: the server does not use it to authorise
// anything. Whether a player may run a command is decided by cmd.Allower alone.
type Level struct {
	level
}

type level uint8

// Member is the level of an ordinary client. It is the level a player has unless another one is set.
func Member() Level {
	return Level{0}
}

// Operator is the level of a client that vanilla considers an operator.
func Operator() Level {
	return Level{1}
}

// Permissions returns the player permission level and command permission level sent to a client for the
// Level.
func (l level) Permissions() (byte, byte) {
	if l == 1 {
		return packet.PermissionLevelOperator, protocol.CommandPermissionLevelGameDirectors
	}
	return packet.PermissionLevelMember, protocol.CommandPermissionLevelAny
}
