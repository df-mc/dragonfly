package world

import "github.com/df-mc/dragonfly/server/block/cube"

// Redstone returns a transaction-scoped handle for redstone engine operations.
func (tx *Tx) Redstone() RedstoneTransaction {
	return RedstoneTransaction{tx: tx}
}

// RedstoneTransaction provides access to redstone engine operations within a transaction.
type RedstoneTransaction struct {
	tx *Tx
}

// ScheduleUpdate marks pos and its neighbours for re-evaluation during the next redstone phase.
func (r RedstoneTransaction) ScheduleUpdate(pos cube.Pos) {
	r.tx.World().redstone.invalidateAround(pos, pos, RedstoneUpdateCauseScheduledTick, r.tx.Range())
}

// Torch returns a transaction-scoped handle for transient redstone torch state at pos.
func (r RedstoneTransaction) Torch(pos cube.Pos) RedstoneTorchTransaction {
	return RedstoneTorchTransaction{tx: r.tx, pos: pos}
}

// RedstoneTorchTransaction provides access to transient redstone torch state within a transaction.
type RedstoneTorchTransaction struct {
	tx  *Tx
	pos cube.Pos
}

// BurnoutStatus returns the transient burnout state for the redstone torch.
func (t RedstoneTorchTransaction) BurnoutStatus() (burnedOut, recoverable bool) {
	return t.tx.World().redstone.torchBurnoutStatus(t.pos, t.tx.CurrentTick())
}

// RecordTurnOff records that the redstone torch was forced off.
func (t RedstoneTorchTransaction) RecordTurnOff() (burnsOut bool) {
	return t.tx.World().redstone.recordTorchTurnOff(t.pos, t.tx.CurrentTick())
}

// MarkSelfTriggered records that the next turn-off was caused by the torch's own output.
func (t RedstoneTorchTransaction) MarkSelfTriggered() {
	t.tx.World().redstone.markTorchSelfTriggered(t.pos)
}

// ConsumeSelfTriggered reports and clears whether the next turn-off was self-triggered.
func (t RedstoneTorchTransaction) ConsumeSelfTriggered() bool {
	return t.tx.World().redstone.consumeTorchSelfTriggered(t.pos)
}

// ClearBurnout removes transient burnout state for the redstone torch.
func (t RedstoneTorchTransaction) ClearBurnout() {
	t.tx.World().redstone.clearTorchBurnout(t.pos)
}

// RedstonePower returns the strongest redstone power currently applied to the position passed. Custom redstone block
// implementations may use this method to query the transaction's current redstone state.
func (tx *Tx) RedstonePower(pos cube.Pos) int {
	return tx.World().redstone.powerTo(pos, tx)
}

// RedstoneDirectPower returns the strongest direct redstone power currently applied to the position passed, excluding
// power conducted through solid blocks. Custom redstone block implementations may use this method to query the
// transaction's current redstone state.
func (tx *Tx) RedstoneDirectPower(pos cube.Pos) int {
	return tx.World().redstone.directPower(pos, tx)
}

// RedstoneStrongPower returns the strongest strong redstone power currently applied to the position passed. Custom
// redstone block implementations may use this method to query the transaction's current redstone state.
func (tx *Tx) RedstoneStrongPower(pos cube.Pos) int {
	return tx.World().redstone.strongPower(pos, tx)
}

// RedstoneConductivePower returns the power held by pos as a conductive block, excluding direct component activation.
// Custom redstone block implementations may use this method to query the transaction's current redstone state.
func (tx *Tx) RedstoneConductivePower(pos cube.Pos) int {
	return tx.World().redstone.conductivePowerTo(pos, tx)
}

// RedstonePowerFrom returns the strongest redstone power reaching pos from the side passed. Custom redstone block
// implementations may use this method to query the transaction's current redstone state.
func (tx *Tx) RedstonePowerFrom(pos cube.Pos, face cube.Face) int {
	return tx.World().redstone.powerFrom(pos, tx, face)
}

// RedstoneDirectPowerFrom returns the strongest direct redstone power reaching pos from the side passed. Custom
// redstone block implementations may use this method to query the transaction's current redstone state.
func (tx *Tx) RedstoneDirectPowerFrom(pos cube.Pos, face cube.Face) int {
	return tx.World().redstone.directPowerFrom(pos, tx, face)
}

// RedstoneStrongPowerFrom returns the strongest strong redstone power reaching pos from the side passed. Custom
// redstone block implementations may use this method to query the transaction's current redstone state.
func (tx *Tx) RedstoneStrongPowerFrom(pos cube.Pos, face cube.Face) int {
	return tx.World().redstone.strongPowerFrom(pos, tx, face)
}
