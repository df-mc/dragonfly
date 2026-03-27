package portal

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

var endEntryPlatformOrigin = cube.Pos{100, 49, 0}

const endPodiumPillarHeight = 4
const endPodiumInnerRadiusSq = 6
const endPodiumOuterRadiusSq = 12

// EnsureEndEntryFeatures restores the fixed vanilla End entry platform and the
// inactive central podium so older worlds don't strand players in the void.
func EnsureEndEntryFeatures(tx *world.Tx) {
	if tx.World().Dimension() != world.End {
		return
	}
	ensureEndEntryPlatform(tx, endEntryPlatformOrigin)
	ensureEndPodium(tx, cube.Pos{0, clampEndPodiumY(tx.HighestBlock(0, 0)), 0})
}

func ensureEndEntryPlatform(tx *world.Tx, origin cube.Pos) {
	for x := origin[0] - 2; x <= origin[0]+2; x++ {
		for z := origin[2] - 2; z <= origin[2]+2; z++ {
			tx.SetBlock(cube.Pos{x, origin[1] - 1, z}, mustEndFeatureBlock("minecraft:obsidian"), nil)
			for y := origin[1]; y <= origin[1]+3; y++ {
				tx.SetBlock(cube.Pos{x, y, z}, mustEndFeatureBlock("minecraft:air"), nil)
			}
		}
	}
}

func ensureEndPodium(tx *world.Tx, origin cube.Pos) {
	baseY := origin[1] - 1
	topY := origin[1]

	for dx := -3; dx <= 3; dx++ {
		for dz := -3; dz <= 3; dz++ {
			distSq := dx*dx + dz*dz
			if distSq > endPodiumOuterRadiusSq {
				continue
			}

			base := cube.Pos{origin[0] + dx, baseY, origin[2] + dz}
			if distSq <= endPodiumInnerRadiusSq {
				tx.SetBlock(base, mustEndFeatureBlock("minecraft:bedrock"), nil)
			} else {
				tx.SetBlock(base, mustEndFeatureBlock("minecraft:end_stone"), nil)
			}

			top := cube.Pos{origin[0] + dx, topY, origin[2] + dz}
			if distSq <= endPodiumInnerRadiusSq {
				tx.SetBlock(top, mustEndFeatureBlock("minecraft:air"), nil)
			} else {
				tx.SetBlock(top, mustEndFeatureBlock("minecraft:bedrock"), nil)
			}
		}
	}

	for dy := 0; dy < endPodiumPillarHeight; dy++ {
		tx.SetBlock(origin.Add(cube.Pos{0, dy, 0}), mustEndFeatureBlock("minecraft:bedrock"), nil)
	}

	for dy := 1; dy < endPodiumPillarHeight; dy++ {
		for dx := -2; dx <= 2; dx++ {
			for dz := -2; dz <= 2; dz++ {
				if dx == 0 && dz == 0 {
					continue
				}
				if dx*dx+dz*dz > endPodiumInnerRadiusSq {
					continue
				}
				tx.SetBlock(origin.Add(cube.Pos{dx, dy, dz}), mustEndFeatureBlock("minecraft:air"), nil)
			}
		}
	}
}

func clampEndPodiumY(y int) int {
	if y <= 0 {
		return 63
	}
	return y
}

func mustEndFeatureBlock(name string) world.Block {
	b, ok := world.BlockByName(name, nil)
	if !ok && name == "minecraft:bedrock" {
		b, ok = world.BlockByName(name, map[string]any{"infiniburn_bit": false})
	}
	if !ok {
		panic("missing block registry entry for " + name)
	}
	return b
}
