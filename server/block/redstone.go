package block

import (
	"time"
)

func redstonePower(power int) int {
	return min(max(power, 0), 15)
}

func redstoneTicks(ticks int) time.Duration {
	return time.Duration(max(ticks, 1)) * time.Second / 10
}
