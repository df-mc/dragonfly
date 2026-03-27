package vanilla

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

const (
	endPodiumInnerRadiusSq = 6
	endPodiumOuterRadiusSq = 12
	endPodiumPillarHeight  = 4
)

func (g Generator) decorateEndMainIsland(c *chunk.Chunk, chunkX, chunkZ, minY, maxY int) {
	if g.dimension != world.End {
		return
	}

	y := clamp(g.preliminarySurfaceLevelAt(0, 0, minY, maxY), minY+1, maxY-endPodiumPillarHeight)
	g.executeInactiveEndPodium(c, cube.Pos{0, y, 0}, chunkX, chunkZ, minY, maxY)
}

func (g Generator) executeInactiveEndPodium(c *chunk.Chunk, origin cube.Pos, chunkX, chunkZ, minY, maxY int) {
	baseY := origin[1] - 1
	topY := origin[1]

	for dx := -3; dx <= 3; dx++ {
		for dz := -3; dz <= 3; dz++ {
			distSq := dx*dx + dz*dz
			if distSq > endPodiumOuterRadiusSq {
				continue
			}

			base := cube.Pos{origin[0] + dx, baseY, origin[2] + dz}
			if g.positionInChunk(base, chunkX, chunkZ, minY, maxY) {
				state := gen.BlockState{Name: "end_stone"}
				if distSq <= endPodiumInnerRadiusSq {
					state = plainBedrockFeatureState()
				}
				_ = g.setBlockStateDirect(c, base, state)
			}

			top := cube.Pos{origin[0] + dx, topY, origin[2] + dz}
			if !g.positionInChunk(top, chunkX, chunkZ, minY, maxY) {
				continue
			}
			state := plainBedrockFeatureState()
			if distSq <= endPodiumInnerRadiusSq {
				state = gen.BlockState{Name: "air"}
			}
			_ = g.setBlockStateDirect(c, top, state)
		}
	}

	for dy := 0; dy < endPodiumPillarHeight; dy++ {
		current := origin.Add(cube.Pos{0, dy, 0})
		if g.positionInChunk(current, chunkX, chunkZ, minY, maxY) {
			_ = g.setBlockStateDirect(c, current, plainBedrockFeatureState())
		}
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
				current := origin.Add(cube.Pos{dx, dy, dz})
				if g.positionInChunk(current, chunkX, chunkZ, minY, maxY) {
					c.SetBlock(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0, g.airRID)
				}
			}
		}
	}
}

func plainBedrockFeatureState() gen.BlockState {
	return gen.BlockState{Name: "bedrock", Properties: map[string]string{"infiniburn_bit": "false"}}
}
