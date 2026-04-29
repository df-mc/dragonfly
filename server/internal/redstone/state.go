package redstone

import "github.com/df-mc/dragonfly/server/block/cube"

const (
	// torchBurnoutThreshold is the maximum number of state changes allowed before burnout occurs.
	torchBurnoutThreshold = 8
	// torchBurnoutWindowTicks is the time window during which state changes are counted.
	torchBurnoutWindowTicks = 60
)

// State holds transient redstone runtime state owned by a world.
type State struct {
	torchBurnout map[cube.Pos]torchBurnout
}

// torchBurnout holds the burnout state and state change history for a redstone torch.
type torchBurnout struct {
	expirationTicks []int64
	burnedOut       bool
}

// TorchBurnoutStatus returns the current transient burnout state for a redstone torch. Expired state-change
// history is pruned before the state is returned.
func (s *State) TorchBurnoutStatus(pos cube.Pos, currentTick int64) (exists, burnedOut, recoverable bool) {
	data, ok := s.torchBurnout[pos]
	if !ok {
		return false, false, false
	}
	remaining := data.removeExpired(currentTick)
	if remaining == 0 && !data.burnedOut {
		s.ClearTorchBurnout(pos)
		return false, false, false
	}
	s.torchBurnout[pos] = data
	return true, data.burnedOut, remaining < torchBurnoutThreshold
}

// TorchBurnoutScheduledTick prepares burnout state for a redstone torch scheduled tick and reports whether the torch is
// currently burned out.
func (s *State) TorchBurnoutScheduledTick(pos cube.Pos, currentTick int64) (burnedOut bool) {
	data := s.torchBurnoutData(pos)
	data.removeExpired(currentTick)
	s.torchBurnout[pos] = data
	return data.burnedOut
}

// PruneTorchBurnout removes idle redstone torch burnout state once all tracked state changes have expired.
func (s *State) PruneTorchBurnout(pos cube.Pos, currentTick int64) {
	data, ok := s.torchBurnout[pos]
	if !ok {
		return
	}
	remaining := data.removeExpired(currentTick)
	if remaining == 0 && !data.burnedOut {
		s.ClearTorchBurnout(pos)
		return
	}
	s.torchBurnout[pos] = data
}

// RecordTorchStateChange records a redstone torch state change and reports whether it should burn out.
func (s *State) RecordTorchStateChange(pos cube.Pos, currentTick int64) (burnsOut bool) {
	data := s.torchBurnoutData(pos)
	data.removeExpired(currentTick)
	data.expirationTicks = append(data.expirationTicks, currentTick+torchBurnoutWindowTicks)
	s.torchBurnout[pos] = data
	return data.counter(currentTick) > torchBurnoutThreshold
}

// BurnOutTorch marks a redstone torch as burned out until a later redstone update recovers it.
func (s *State) BurnOutTorch(pos cube.Pos) {
	data := s.torchBurnoutData(pos)
	data.burnedOut = true
	s.torchBurnout[pos] = data
}

// ClearTorchBurnout removes transient burnout state for a redstone torch.
func (s *State) ClearTorchBurnout(pos cube.Pos) {
	delete(s.torchBurnout, pos)
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

// removeExpired removes expired state change entries and returns the number of entries that remain.
func (data *torchBurnout) removeExpired(currentTick int64) int {
	expirationTicks := data.expirationTicks[:0]
	for _, expirationTick := range data.expirationTicks {
		if expirationTick >= currentTick {
			expirationTicks = append(expirationTicks, expirationTick)
		}
	}
	data.expirationTicks = expirationTicks
	return len(data.expirationTicks)
}

// counter returns the number of active non-expired state changes.
func (data torchBurnout) counter(currentTick int64) int {
	count := 0
	for _, expirationTick := range data.expirationTicks {
		if expirationTick >= currentTick {
			count++
		}
	}
	return count
}
