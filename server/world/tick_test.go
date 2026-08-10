package world

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// TestScheduledTickQueueAddDeduplicates verifies that adding a scheduled tick for a block that is already scheduled
// leaves one. Ticks are written to disk with their chunk and added back when it is loaded again, so a block scheduled
// while its chunk was unloaded would otherwise be ticked twice, which for a dropper means it dispenses twice.
func TestScheduledTickQueueAddDeduplicates(t *testing.T) {
	pos := cube.Pos{1, 2, 3}
	tests := []struct {
		name string
		add  []scheduledTick
		want int
	}{
		{name: "the same tick twice", add: []scheduledTick{{pos: pos, t: 10, bhash: 1}, {pos: pos, t: 10, bhash: 1}}, want: 1},
		{name: "a later tick for the same block", add: []scheduledTick{{pos: pos, t: 10, bhash: 1}, {pos: pos, t: 20, bhash: 1}}, want: 2},
		{name: "another block at the same position", add: []scheduledTick{{pos: pos, t: 10, bhash: 1}, {pos: pos, t: 10, bhash: 2}}, want: 2},
		{name: "another position", add: []scheduledTick{{pos: pos, t: 10, bhash: 1}, {pos: cube.Pos{4, 5, 6}, t: 10, bhash: 1}}, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &scheduledTickQueue{furthestTicks: map[scheduledTickIndex]int64{}}
			for _, tick := range tt.add {
				queue.add([]scheduledTick{tick})
			}
			if got := len(queue.ticks); got != tt.want {
				t.Errorf("queue holds %v scheduled ticks, want %v", got, tt.want)
			}
		})
	}
}
