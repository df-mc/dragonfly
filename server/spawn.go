package server

import (
	"math"
	"strings"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type spawnCandidate struct {
	pos   cube.Pos
	score int
	dist  float64
}

func (srv *Server) adjustOverworldSpawn(w *world.World) {
	if w == nil || w.Dimension() != world.Overworld {
		return
	}

	current := w.Spawn()
	needsRelocation := current == (cube.Pos{})
	if !needsRelocation && (current[0] == 0 && current[2] == 0) {
		<-w.Exec(func(tx *world.Tx) {
			if candidate, ok := currentSpawnCandidate(tx, current); !ok || candidate.score < 40 {
				needsRelocation = true
			}
		})
	}
	if !needsRelocation {
		return
	}

	var best spawnCandidate
	<-w.Exec(func(tx *world.Tx) {
		best = findOverworldSpawnCandidate(tx)
	})
	if best.score > 0 {
		w.SetSpawn(best.pos)
	}
}

func (srv *Server) adjustOverworldSpawnHint(w *world.World, center world.ChunkPos) bool {
	if w == nil || w.Dimension() != world.Overworld {
		return false
	}

	var best spawnCandidate
	<-w.Exec(func(tx *world.Tx) {
		for radius := 0; radius <= 1; radius++ {
			for _, chunkPos := range chunkRingAround(center, radius) {
				column := loadSpawnChunk(tx, chunkPos)
				candidate, ok := bestSpawnCandidateInChunk(tx, column, chunkPos)
				if !ok {
					continue
				}
				if candidate.score > best.score || (candidate.score == best.score && candidate.dist < best.dist) {
					best = candidate
				}
			}
		}
	})
	if best.score <= 0 {
		return false
	}
	w.SetSpawn(best.pos)
	return true
}

func currentSpawnCandidate(tx *world.Tx, current cube.Pos) (spawnCandidate, bool) {
	chunkPos := world.ChunkPos{int32(current[0] >> 4), int32(current[2] >> 4)}
	column := loadSpawnChunk(tx, chunkPos)
	return bestSpawnCandidateInChunk(tx, column, chunkPos)
}

func findOverworldSpawnCandidate(tx *world.Tx) spawnCandidate {
	best := spawnCandidate{score: -1, dist: math.MaxFloat64}
	for radius := 0; radius <= 16; radius++ {
		for _, chunkPos := range chunkRing(radius, 1) {
			column := loadSpawnChunk(tx, chunkPos)
			candidate, ok := bestSpawnCandidateInChunk(tx, column, chunkPos)
			if !ok {
				continue
			}
			if candidate.score > best.score || (candidate.score == best.score && candidate.dist < best.dist) {
				best = candidate
				if best.score >= 70 {
					return best
				}
			}
		}
	}
	for radius := 20; radius <= 64; radius += 4 {
		for _, chunkPos := range chunkRing(radius, 4) {
			column := loadSpawnChunk(tx, chunkPos)
			candidate, ok := bestSpawnCandidateInChunk(tx, column, chunkPos)
			if !ok {
				continue
			}
			if candidate.score > best.score || (candidate.score == best.score && candidate.dist < best.dist) {
				best = candidate
				if best.score >= 70 {
					return best
				}
			}
		}
	}
	return best
}

func loadSpawnChunk(tx *world.Tx, pos world.ChunkPos) *world.Column {
	return tx.Chunk(pos)
}

func bestSpawnCandidateInChunk(tx *world.Tx, column *world.Column, chunkPos world.ChunkPos) (spawnCandidate, bool) {
	best := spawnCandidate{score: -1, dist: math.MaxFloat64}
	baseX := int(chunkPos[0]) * 16
	baseZ := int(chunkPos[1]) * 16

	for localX := 0; localX < 16; localX++ {
		for localZ := 0; localZ < 16; localZ++ {
			height := int(column.Chunk.HighestBlock(uint8(localX), uint8(localZ)))
			if height <= tx.Range()[0] || height >= tx.Range()[1] {
				continue
			}

			worldPos := cube.Pos{baseX + localX, height, baseZ + localZ}
			top := tx.Block(worldPos)
			switch top.(type) {
			case block.Air, block.Water, block.Lava, block.Log, block.Leaves:
				continue
			}
			if _, blocked := tx.Block(worldPos.Side(cube.FaceUp)).(block.Air); !blocked {
				continue
			}

			biomeRID := column.Chunk.Biome(uint8(localX), int16(height), uint8(localZ))
			biomeName := ""
			if biome, ok := world.BiomeByID(int(biomeRID)); ok {
				biomeName = biome.String()
			}
			score := spawnBiomeScore(biomeName)
			if score <= 0 {
				continue
			}

			dist := math.Hypot(float64(localX-8), float64(localZ-8))
			candidate := spawnCandidate{
				pos:   worldPos.Side(cube.FaceUp),
				score: score,
				dist:  dist,
			}
			if candidate.score > best.score || (candidate.score == best.score && candidate.dist < best.dist) {
				best = candidate
			}
		}
	}
	return best, best.score > 0
}

func chunkRing(radius, step int) []world.ChunkPos {
	if radius == 0 {
		return []world.ChunkPos{{}}
	}
	if step <= 0 {
		step = 1
	}

	out := make([]world.ChunkPos, 0, radius*8)
	for x := -radius; x <= radius; x += step {
		out = append(out, world.ChunkPos{int32(x), int32(-radius)})
		out = append(out, world.ChunkPos{int32(x), int32(radius)})
	}
	for z := -radius + step; z <= radius-step; z += step {
		out = append(out, world.ChunkPos{int32(-radius), int32(z)})
		out = append(out, world.ChunkPos{int32(radius), int32(z)})
	}
	return out
}

func chunkRingAround(center world.ChunkPos, radius int) []world.ChunkPos {
	if radius == 0 {
		return []world.ChunkPos{center}
	}

	out := make([]world.ChunkPos, 0, radius*8)
	for x := -radius; x <= radius; x++ {
		out = append(out, world.ChunkPos{center[0] + int32(x), center[1] - int32(radius)})
		out = append(out, world.ChunkPos{center[0] + int32(x), center[1] + int32(radius)})
	}
	for z := -radius + 1; z <= radius-1; z++ {
		out = append(out, world.ChunkPos{center[0] - int32(radius), center[1] + int32(z)})
		out = append(out, world.ChunkPos{center[0] + int32(radius), center[1] + int32(z)})
	}
	return out
}

func spawnBiomeScore(name string) int {
	switch name {
	case "forest", "birch_forest", "dark_forest", "flower_forest", "taiga", "snowy_taiga", "old_growth_pine_taiga", "old_growth_spruce_taiga", "old_growth_birch_forest", "jungle", "bamboo_jungle", "sparse_jungle", "swamp", "mangrove_swamp", "cherry_grove":
		return 100
	case "plains", "sunflower_plains", "savanna", "savanna_plateau", "windswept_savanna", "meadow":
		return 70
	case "beach", "snowy_beach", "stony_shore", "river", "frozen_river":
		return 0
	default:
		if strings.Contains(name, "ocean") {
			return 0
		}
		return 40
	}
}
