package vanilla

import (
	"sort"
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

const noWaterHeight = -1 << 31
const maxSurfaceColumnHeight = 384

type surfaceColumn struct {
	surfaceDepth     int
	surfaceSecondary float64
	minSurfaceLevel  int
	steep            bool
}

type surfaceColumnScratch struct {
	stoneDepthAbove [maxSurfaceColumnHeight]int
	stoneDepthBelow [maxSurfaceColumnHeight]int
	waterHeights    [maxSurfaceColumnHeight]int
}

func (g Generator) applySurfaceAndBiomes(c *chunk.Chunk, biomes sourceBiomeVolume, chunkX, chunkZ, minY, maxY int) {
	heightMap := c.HeightMap()
	var columns [16][16]surfaceColumn
	var scratch surfaceColumnScratch

	for localX := 0; localX < 16; localX++ {
		for localZ := 0; localZ < 16; localZ++ {
			worldX := chunkX*16 + localX
			worldZ := chunkZ*16 + localZ
			surfaceY := int(heightMap.At(uint8(localX), uint8(localZ))) - 1
			if surfaceY < minY {
				surfaceY = minY
			}

			surfaceDepth := g.surface.SurfaceDepth(worldX, worldZ)

			columns[localX][localZ] = surfaceColumn{
				surfaceDepth:     surfaceDepth,
				surfaceSecondary: g.surface.SurfaceSecondary(worldX, worldZ),
				minSurfaceLevel:  surfaceY - surfaceDepth,
				steep:            g.isSteepColumn(heightMap, uint8(localX), uint8(localZ)),
			}
		}
	}

	for localX := 0; localX < 16; localX++ {
		for localZ := 0; localZ < 16; localZ++ {
			worldX := chunkX*16 + localX
			worldZ := chunkZ*16 + localZ
			g.applySurfaceColumn(c, biomes, localX, localZ, worldX, worldZ, minY, maxY, columns[localX][localZ], &scratch)
		}
	}
}

func (g Generator) isSteepColumn(heightMap chunk.HeightMap, x, z uint8) bool {
	center := heightMap.At(x, z)
	if x > 0 {
		if diff := center - heightMap.At(x-1, z); diff >= 3 || diff <= -3 {
			return true
		}
	}
	if x < 15 {
		if diff := center - heightMap.At(x+1, z); diff >= 3 || diff <= -3 {
			return true
		}
	}
	if z > 0 {
		if diff := center - heightMap.At(x, z-1); diff >= 3 || diff <= -3 {
			return true
		}
	}
	if z < 15 {
		if diff := center - heightMap.At(x, z+1); diff >= 3 || diff <= -3 {
			return true
		}
	}
	return false
}

func (g Generator) applySurfaceColumn(c *chunk.Chunk, biomes sourceBiomeVolume, localX, localZ, worldX, worldZ, minY, maxY int, column surfaceColumn, scratch *surfaceColumnScratch) {
	columnHeight := maxY - minY + 1
	stoneDepthAbove, stoneDepthBelow, waterHeights := scratch.slices(columnHeight)
	clear(stoneDepthAbove)
	clear(stoneDepthBelow)
	for i := range waterHeights {
		waterHeights[i] = noWaterHeight
	}

	currentStoneDepthAbove := 0
	currentWaterHeight := noWaterHeight
	inStoneAbove := false

	for y := maxY; y >= minY; y-- {
		idx := y - minY
		rid := c.Block(uint8(localX), int16(y), uint8(localZ), 0)
		switch rid {
		case g.airRID, g.lavaRID:
			currentStoneDepthAbove = 0
			inStoneAbove = false
			continue
		case g.waterRID:
			if currentWaterHeight == noWaterHeight {
				currentWaterHeight = y + 1
			}
			waterHeights[idx] = currentWaterHeight
			currentStoneDepthAbove = 0
			inStoneAbove = false
			continue
		}

		if !g.isSolidRID(rid) {
			currentStoneDepthAbove = 0
			inStoneAbove = false
			continue
		}

		if !inStoneAbove {
			inStoneAbove = true
			currentStoneDepthAbove = 0
		}

		stoneDepthAbove[idx] = currentStoneDepthAbove
		waterHeights[idx] = currentWaterHeight
		currentStoneDepthAbove++
	}

	currentStoneDepthBelow := 0
	inStoneBelow := false
	for y := minY; y <= maxY; y++ {
		idx := y - minY
		rid := c.Block(uint8(localX), int16(y), uint8(localZ), 0)
		if !g.isSolidRID(rid) {
			currentStoneDepthBelow = 0
			inStoneBelow = false
			continue
		}

		if !inStoneBelow {
			inStoneBelow = true
			currentStoneDepthBelow = 0
		}
		stoneDepthBelow[idx] = currentStoneDepthBelow
		currentStoneDepthBelow++
	}

	for y := maxY; y >= minY; y-- {
		idx := y - minY
		rid := c.Block(uint8(localX), int16(y), uint8(localZ), 0)
		if !g.isSurfaceBaseRID(rid) {
			continue
		}

		ctx := gen.SurfaceContext{
			BlockX:           worldX,
			BlockY:           y,
			BlockZ:           worldZ,
			SurfaceDepth:     column.surfaceDepth,
			SurfaceSecondary: column.surfaceSecondary,
			WaterHeight:      waterHeights[idx],
			StoneDepthAbove:  stoneDepthAbove[idx],
			StoneDepthBelow:  stoneDepthBelow[idx],
			Steep:            column.steep,
			// Reuse the logical source-biome volume materialized for this chunk
			// instead of recomputing the preset biome source for every block.
			Biome:           biomes.biomeAt(localX, y, localZ),
			MinSurfaceLevel: column.minSurfaceLevel,
			MinY:            minY,
			MaxY:            maxY,
		}

		if replacement, ok := g.surface.TryApply(ctx, g.lookupSurfaceBlock); ok && replacement != rid {
			c.SetBlock(uint8(localX), int16(y), uint8(localZ), 0, replacement)
		}
	}
}

func (s *surfaceColumnScratch) slices(columnHeight int) ([]int, []int, []int) {
	if columnHeight <= maxSurfaceColumnHeight {
		return s.stoneDepthAbove[:columnHeight], s.stoneDepthBelow[:columnHeight], s.waterHeights[:columnHeight]
	}
	return make([]int, columnHeight), make([]int, columnHeight), make([]int, columnHeight)
}

func (g Generator) lookupSurfaceBlock(name string, properties map[string]string) uint32 {
	key := surfaceBlockCacheKey(name, properties)
	if rid, ok := g.surfaceBlockCache.Lookup(key); ok {
		return rid
	}

	switch name {
	case "minecraft:air":
		g.surfaceBlockCache.Store(key, g.airRID)
		return g.airRID
	case "minecraft:water":
		g.surfaceBlockCache.Store(key, g.waterRID)
		return g.waterRID
	case "minecraft:lava":
		g.surfaceBlockCache.Store(key, g.lavaRID)
		return g.lavaRID
	}

	if rid, ok := g.lookupRegisteredSurfaceBlock(name, properties); ok {
		g.surfaceBlockCache.Store(key, rid)
		return rid
	}
	if len(properties) != 0 {
		if rid, ok := g.lookupRegisteredSurfaceBlock(name, nil); ok {
			g.surfaceBlockCache.Store(key, rid)
			return rid
		}
	}

	g.surfaceBlockCache.Store(key, g.airRID)
	return g.airRID
}

func (g Generator) lookupRegisteredSurfaceBlock(name string, properties map[string]string) (uint32, bool) {
	blockProps := make(map[string]any, len(properties))
	for key, value := range properties {
		switch value {
		case "true":
			blockProps[key] = true
		case "false":
			blockProps[key] = false
		default:
			if n, err := strconv.ParseInt(value, 10, 32); err == nil {
				blockProps[key] = int32(n)
			} else {
				blockProps[key] = value
			}
		}
	}
	if len(blockProps) == 0 {
		blockProps = nil
	}

	b, ok := world.BlockByName(name, blockProps)
	if !ok {
		return 0, false
	}
	return world.BlockRuntimeID(b), true
}

func surfaceBlockCacheKey(name string, properties map[string]string) string {
	if len(properties) == 0 {
		return name
	}

	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.Grow(len(name) + len(keys)*16)
	b.WriteString(name)
	for _, key := range keys {
		b.WriteByte('|')
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(properties[key])
	}
	return b.String()
}

func (g Generator) isSurfaceBaseRID(rid uint32) bool {
	if rid == g.defaultBlockRID {
		return true
	}
	return g.dimension == world.Overworld && rid == g.deepRID
}
