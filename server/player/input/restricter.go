package input

// Restricter represents an interface that can manage input locks for a player.
type Restricter interface {
	// LockInput applies an input lock to the player, disabling the specified input.
	LockInput(l Lock)
	// UnlockInput removes an input lock from the player, re-enabling the specified input.
	UnlockInput(l Lock)
	// ClearInputLocks removes all input locks from the player, re-enabling all inputs.
	ClearInputLocks()
	// InputLocked checks if a specific input lock is currently applied to the player.
	InputLocked(l Lock) bool
}
