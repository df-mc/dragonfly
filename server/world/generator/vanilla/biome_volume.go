package vanilla

import (
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

const biomeCellSize = 4

type sourceBiomeVolume struct {
	startY int
	cellsY int
	data   []gen.Biome
}

func newSourceBiomeVolume(minY, maxY int) sourceBiomeVolume {
	startY := alignDown(minY, biomeCellSize)
	cellsY := (maxY-startY)/biomeCellSize + 1
	return sourceBiomeVolume{
		startY: startY,
		cellsY: cellsY,
		data:   make([]gen.Biome, 4*4*cellsY),
	}
}

func (v sourceBiomeVolume) cellIndex(localX, y, localZ int) int {
	cellX := clamp(localX>>2, 0, 3)
	cellZ := clamp(localZ>>2, 0, 3)
	cellY := clamp((y-v.startY)/biomeCellSize, 0, v.cellsY-1)
	return (cellY*4+cellZ)*4 + cellX
}

func (v sourceBiomeVolume) set(localX, y, localZ int, biome gen.Biome) {
	v.data[v.cellIndex(localX, y, localZ)] = biome
}

func (v sourceBiomeVolume) biomeAt(localX, y, localZ int) gen.Biome {
	return v.data[v.cellIndex(localX, y, localZ)]
}

func (g Generator) populateBiomeVolume(c *chunk.Chunk, chunkX, chunkZ, minY, maxY int) sourceBiomeVolume {
	startY := alignDown(minY, biomeCellSize)
	volume := newSourceBiomeVolume(minY, maxY)

	for baseX := 0; baseX < 16; baseX += biomeCellSize {
		worldX := chunkX*16 + baseX

		for baseZ := 0; baseZ < 16; baseZ += biomeCellSize {
			worldZ := chunkZ*16 + baseZ

			for baseY := startY; baseY <= maxY; baseY += biomeCellSize {
				biome := g.biomeSource.GetBiome(worldX, baseY, worldZ)
				volume.set(baseX, baseY, baseZ, biome)
				biomeRID := biomeRuntimeID(biome)

				fillFromY := baseY
				if fillFromY < minY {
					fillFromY = minY
				}
				fillToY := baseY + biomeCellSize - 1
				if fillToY > maxY {
					fillToY = maxY
				}

				for localY := fillFromY; localY <= fillToY; localY++ {
					for localX := baseX; localX < baseX+biomeCellSize; localX++ {
						for localZ := baseZ; localZ < baseZ+biomeCellSize; localZ++ {
							c.SetBiome(uint8(localX), int16(localY), uint8(localZ), biomeRID)
						}
					}
				}
			}
		}
	}
	return volume
}

func (g Generator) biomeAt(c *chunk.Chunk, localX, y, localZ int) gen.Biome {
	return biomeFromRuntimeID(c.Biome(uint8(localX), int16(y), uint8(localZ)))
}

func alignDown(value, multiple int) int {
	remainder := value % multiple
	if remainder < 0 {
		remainder += multiple
	}
	return value - remainder
}
