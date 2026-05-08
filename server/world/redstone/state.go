package redstone

import (
	"slices"

	"github.com/df-mc/dragonfly/server/block/cube"
)

const (
	// torchBurnoutThreshold is the maximum number of state changes allowed before burnout occurs.
	torchBurnoutThreshold = 8
	// torchBurnoutWindowTicks is the time window during which state changes are counted.
	torchBurnoutWindowTicks = 60
)

// State holds transient redstone runtime state owned by a world.
type State struct {
	torchBurnout       map[cube.Pos]torchBurnout
	activeTorchUpdates map[cube.Pos]int
}

// torchBurnout holds the burnout state and state change history for a redstone torch.
type torchBurnout struct {
	expirationTicks []int64
	burnedOut       bool
	// selfTriggered is set when the next scheduled toggle was caused by this torch's own outgoing propagation.
	selfTriggered bool
}

// TorchBurnoutStatus returns the current transient burnout state for a redstone torch. Expired state-change
// history is pruned before the state is returned.
func (s *State) TorchBurnoutStatus(pos cube.Pos, currentTick int64) (burnedOut, recoverable bool) {
	data, ok := s.pruneTorchBurnoutData(pos, currentTick)
	if !ok {
		return false, false
	}
	return data.burnedOut, len(data.expirationTicks) < torchBurnoutThreshold
}

// PruneTorchBurnout removes idle redstone torch burnout state once all tracked state changes have expired.
func (s *State) PruneTorchBurnout(pos cube.Pos, currentTick int64) {
	data, ok := s.torchBurnout[pos]
	if !ok {
		return
	}
	remaining := data.removeExpired(currentTick)
	data.selfTriggered = false
	if remaining == 0 && !data.burnedOut {
		s.ClearTorchBurnout(pos)
		return
	}
	s.torchBurnout[pos] = data
}

// RecordTorchToggle records a self-triggered redstone torch toggle and reports whether it should burn out.
func (s *State) RecordTorchToggle(pos cube.Pos, currentTick int64) (burnsOut bool) {
	data, ok := s.torchBurnout[pos]
	if !ok {
		return false
	}
	data.removeExpired(currentTick)
	if data.selfTriggered {
		data.expirationTicks = append(data.expirationTicks, currentTick+torchBurnoutWindowTicks)
	}
	data.selfTriggered = false
	if len(data.expirationTicks) == 0 && !data.burnedOut {
		s.ClearTorchBurnout(pos)
		return false
	}
	s.torchBurnout[pos] = data
	return len(data.expirationTicks) > torchBurnoutThreshold
}

// BurnOutTorch marks a redstone torch as burned out until a later redstone update recovers it.
func (s *State) BurnOutTorch(pos cube.Pos) {
	data := s.torchBurnoutData(pos)
	data.burnedOut = true
	data.selfTriggered = false
	s.torchBurnout[pos] = data
}

// ClearTorchBurnout removes transient burnout state for a redstone torch.
func (s *State) ClearTorchBurnout(pos cube.Pos) {
	delete(s.torchBurnout, pos)
}

// WithActiveTorchUpdate marks redstone propagation as originating from the redstone torch at pos for the duration of fn.
func (s *State) WithActiveTorchUpdate(pos cube.Pos, fn func()) {
	if s.activeTorchUpdates == nil {
		s.activeTorchUpdates = make(map[cube.Pos]int)
	}
	s.activeTorchUpdates[pos]++
	defer func() {
		if s.activeTorchUpdates[pos] <= 1 {
			delete(s.activeTorchUpdates, pos)
			return
		}
		s.activeTorchUpdates[pos]--
	}()
	fn()
}

// MarkTorchSelfTriggeredIfActive marks the next scheduled tick for the redstone torch at pos as self-triggered if the
// current redstone propagation originated from that torch.
func (s *State) MarkTorchSelfTriggeredIfActive(pos cube.Pos) {
	if s.activeTorchUpdates[pos] == 0 {
		return
	}
	data := s.torchBurnoutData(pos)
	data.selfTriggered = true
	s.torchBurnout[pos] = data
}

// torchBurnoutData retrieves or creates burnout tracking data for the given torch position.
func (s *State) torchBurnoutData(pos cube.Pos) torchBurnout {
	if s.torchBurnout == nil {
		s.torchBurnout = make(map[cube.Pos]torchBurnout)
	}
	data, ok := s.torchBurnout[pos]
	if !ok {
		data.expirationTicks = make([]int64, 0, torchBurnoutThreshold+1)
	}
	return data
}

// pruneTorchBurnoutData removes expired burnout history and returns the remaining torch data if it is still relevant.
func (s *State) pruneTorchBurnoutData(pos cube.Pos, currentTick int64) (torchBurnout, bool) {
	data, ok := s.torchBurnout[pos]
	if !ok {
		return torchBurnout{}, false
	}
	data.removeExpired(currentTick)
	if len(data.expirationTicks) == 0 && !data.burnedOut && !data.selfTriggered {
		s.ClearTorchBurnout(pos)
		return torchBurnout{}, false
	}
	s.torchBurnout[pos] = data
	return data, true
}

// removeExpired removes expired state change entries and returns the number of entries that remain.
func (data *torchBurnout) removeExpired(currentTick int64) int {
	data.expirationTicks = slices.DeleteFunc(data.expirationTicks, func(t int64) bool {
		return t < currentTick
	})
	return len(data.expirationTicks)
}
