package block

import "github.com/df-mc/dragonfly/server/world"

func allChorusFlowers() (b []world.Block) {
	for age := 0; age <= 5; age++ {
		b = append(b, ChorusFlower{Age: age})
	}
	return
}
