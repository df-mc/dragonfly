package input

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// Lock represents a client input lock that can be applied to a player to disable specific inputs such as
// camera rotation, movement, jumping, sneaking or mounting/dismounting entities.
type Lock struct {
	lock
}

type lock uint32

// Camera is the lock that disables all camera movement.
func Camera() Lock {
	return Lock{lock(packet.ClientInputLockCamera)}
}

// Movement is the lock that disables all player movement, including jumping and sneaking.
func Movement() Lock {
	return Lock{lock(packet.ClientInputLockMovement)}
}

// LateralMovement is the lock that disables all player movement excluding jumping and sneaking.
func LateralMovement() Lock {
	return Lock{lock(packet.ClientInputLockLateralMovement)}
}

// Sneak is the lock that disables the player from sneaking.
func Sneak() Lock {
	return Lock{lock(packet.ClientInputLockSneak)}
}

// Jump is the lock that disables the player from jumping.
func Jump() Lock {
	return Lock{lock(packet.ClientInputLockJump)}
}

// Mount is the lock that prevents the player from mounting entities.
func Mount() Lock {
	return Lock{lock(packet.ClientInputLockMount)}
}

// Dismount is the lock that prevents the player from dismounting entities.
func Dismount() Lock {
	return Lock{lock(packet.ClientInputLockDismount)}
}

// MoveForward is the lock that disables forward movement.
func MoveForward() Lock {
	return Lock{lock(packet.ClientInputLockMoveForward)}
}

// MoveBackward is the lock that disables backward movement.
func MoveBackward() Lock {
	return Lock{lock(packet.ClientInputLockMoveBackward)}
}

// MoveLeft is the lock that disables left strafe movement.
func MoveLeft() Lock {
	return Lock{lock(packet.ClientInputLockMoveLeft)}
}

// MoveRight is the lock that disables right strafe movement.
func MoveRight() Lock {
	return Lock{lock(packet.ClientInputLockMoveRight)}
}

// Uint32 returns the lock as a uint32.
func (l lock) Uint32() uint32 {
	return uint32(l)
}

// All returns all the input locks that are available to be applied to a player.
func All() []Lock {
	return []Lock{
		Camera(), Movement(), LateralMovement(), Sneak(), Jump(), Mount(), Dismount(),
		MoveForward(), MoveBackward(), MoveLeft(), MoveRight(),
	}
}
