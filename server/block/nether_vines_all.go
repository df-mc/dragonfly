package block

import "github.com/df-mc/dragonfly/server/world"

func allNetherVines() (b []world.Block) {
	for age := 0; age <= 25; age++ {
		b = append(b, NetherVines{Twisting: true, Age: age})
		b = append(b, NetherVines{Twisting: false, Age: age})
	}
	return
}
