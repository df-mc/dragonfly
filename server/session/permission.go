package session

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PermissionLevel is the permission level a client displays for itself. It only
// decides what the client shows and allows locally, such as the command button
// in the chat window: the server does not use it to authorise anything. Whether
// a Controllable may run a command is decided by cmd.Allower alone.
type PermissionLevel byte

const (
	// PermissionLevelMember is the level of an ordinary client. It is the level
	// a Controllable has unless another one is set.
	PermissionLevelMember PermissionLevel = iota
	// PermissionLevelOperator is the level of a client that vanilla considers
	// an operator.
	PermissionLevelOperator
)

// abilityPermissions returns the permission level and command permission level
// sent to a client for the PermissionLevel.
func (l PermissionLevel) abilityPermissions() (byte, byte) {
	if l == PermissionLevelOperator {
		return packet.PermissionLevelOperator, protocol.CommandPermissionLevelGameDirectors
	}
	return packet.PermissionLevelMember, protocol.CommandPermissionLevelAny
}
