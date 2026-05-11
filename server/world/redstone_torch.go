package world

import (
	"slices"

	"github.com/df-mc/dragonfly/server/block/cube"
)

const (
	// redstoneTorchBurnoutThreshold is the maximum number of torch state changes allowed before burnout occurs.
	redstoneTorchBurnoutThreshold = 8
	// redstoneTorchBurnoutWindowTicks is the window during which torch state changes are counted.
	redstoneTorchBurnoutWindowTicks = 60
)

type redstoneTorchBurnout struct {
	expirationTicks []int64
	burnedOut       bool
}

func (e *redstoneEngine) redstoneTorchBurnoutStatus(pos cube.Pos, currentTick int64) (burnedOut, recoverable bool) {
	data, ok := e.pruneRedstoneTorchBurnoutData(pos, currentTick)
	if !ok {
		return false, false
	}
	return data.burnedOut, len(data.expirationTicks) < redstoneTorchBurnoutThreshold
}

func (e *redstoneEngine) pruneRedstoneTorchBurnout(pos cube.Pos, currentTick int64) {
	data, ok := e.torchBurnout[pos]
	if !ok {
		return
	}
	data.removeExpired(currentTick)
	if len(data.expirationTicks) == 0 && !data.burnedOut {
		e.clearRedstoneTorchBurnout(pos)
		return
	}
	e.torchBurnout[pos] = data
}

func (e *redstoneEngine) recordRedstoneTorchToggle(pos cube.Pos, currentTick int64) (burnsOut bool) {
	data := e.redstoneTorchBurnoutData(pos)
	data.removeExpired(currentTick)
	data.expirationTicks = append(data.expirationTicks, currentTick+redstoneTorchBurnoutWindowTicks)
	if len(data.expirationTicks) >= redstoneTorchBurnoutThreshold {
		data.burnedOut = true
		burnsOut = true
	}
	e.torchBurnout[pos] = data
	return burnsOut
}

func (e *redstoneEngine) burnOutRedstoneTorch(pos cube.Pos) {
	data := e.redstoneTorchBurnoutData(pos)
	data.burnedOut = true
	e.torchBurnout[pos] = data
}

func (e *redstoneEngine) clearRedstoneTorchBurnout(pos cube.Pos) {
	delete(e.torchBurnout, pos)
}

func (e *redstoneEngine) redstoneTorchBurnoutData(pos cube.Pos) redstoneTorchBurnout {
	if e.torchBurnout == nil {
		e.torchBurnout = make(map[cube.Pos]redstoneTorchBurnout)
	}
	data, ok := e.torchBurnout[pos]
	if !ok {
		data.expirationTicks = make([]int64, 0, redstoneTorchBurnoutThreshold+1)
	}
	return data
}

func (e *redstoneEngine) pruneRedstoneTorchBurnoutData(pos cube.Pos, currentTick int64) (redstoneTorchBurnout, bool) {
	data, ok := e.torchBurnout[pos]
	if !ok {
		return redstoneTorchBurnout{}, false
	}
	data.removeExpired(currentTick)
	if len(data.expirationTicks) == 0 && !data.burnedOut {
		e.clearRedstoneTorchBurnout(pos)
		return redstoneTorchBurnout{}, false
	}
	e.torchBurnout[pos] = data
	return data, true
}

func (data *redstoneTorchBurnout) removeExpired(currentTick int64) {
	data.expirationTicks = slices.DeleteFunc(data.expirationTicks, func(t int64) bool {
		return t < currentTick
	})
}
