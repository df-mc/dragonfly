package vanilla

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

func (g Generator) GenerateColumn(pos world.ChunkPos, col *chunk.Column) {
	if col == nil || col.Chunk == nil {
		return
	}
	g.GenerateChunk(pos, col.Chunk)
	g.populateStructureMetadata(pos, col)
}

func (g Generator) populateStructureMetadata(pos world.ChunkPos, col *chunk.Column) {
	if col == nil || col.Chunk == nil || g.structureStarts == nil || len(g.structurePlanners) == 0 {
		return
	}
	chunkX := int(pos[0])
	chunkZ := int(pos[1])
	minY := col.Chunk.Range().Min()
	maxY := col.Chunk.Range().Max()

	starts := make([]chunk.StructureStart, 0, 4)
	refs := make([]chunk.StructureReference, 0, 8)
	startSeen := map[structureStartKey]struct{}{}
	refSeen := map[structureStartKey]struct{}{}

	for _, planner := range g.structurePlanners {
		startMinChunkX := chunkX - planner.maxBackreachX
		startMaxChunkX := chunkX + planner.maxBackreachX
		startMinChunkZ := chunkZ - planner.maxBackreachZ
		startMaxChunkZ := chunkZ + planner.maxBackreachZ

		minGridX := randomSpreadMinGrid(startMinChunkX, planner.placement.Spacing, planner.placement.Separation)
		maxGridX := floorDiv(startMaxChunkX, planner.placement.Spacing)
		minGridZ := randomSpreadMinGrid(startMinChunkZ, planner.placement.Spacing, planner.placement.Separation)
		maxGridZ := floorDiv(startMaxChunkZ, planner.placement.Spacing)

		for gridX := minGridX; gridX <= maxGridX; gridX++ {
			for gridZ := minGridZ; gridZ <= maxGridZ; gridZ++ {
				startChunk := randomSpreadPotentialChunk(g.seed, planner.placement, gridX, gridZ)
				if int(startChunk[0]) < startMinChunkX || int(startChunk[0]) > startMaxChunkX || int(startChunk[1]) < startMinChunkZ || int(startChunk[1]) > startMaxChunkZ {
					continue
				}
				start, ok := g.planStructureStart(planner, startChunk, minY, maxY)
				if !ok || !structureIntersectsChunk(start, chunkX, chunkZ, minY, maxY) {
					continue
				}

				refKey := structureStartKey{setName: start.structureName, chunkX: start.startChunk[0], chunkZ: start.startChunk[1]}
				if _, ok := refSeen[refKey]; !ok {
					refSeen[refKey] = struct{}{}
					refs = append(refs, chunk.StructureReference{
						StructureSet: start.setName,
						Structure:    start.structureName,
						StartChunkX:  start.startChunk[0],
						StartChunkZ:  start.startChunk[1],
					})
				}

				if start.startChunk != pos {
					continue
				}
				if _, ok := startSeen[refKey]; ok {
					continue
				}
				startSeen[refKey] = struct{}{}
				starts = append(starts, chunk.StructureStart{
					StructureReference: chunk.StructureReference{
						StructureSet: start.setName,
						Structure:    start.structureName,
						StartChunkX:  start.startChunk[0],
						StartChunkZ:  start.startChunk[1],
					},
					Template: start.templateName,
					OriginX:  int32(start.origin.X()),
					OriginY:  int32(start.origin.Y()),
					OriginZ:  int32(start.origin.Z()),
					SizeX:    int32(start.size[0]),
					SizeY:    int32(start.size[1]),
					SizeZ:    int32(start.size[2]),
				})
			}
		}
	}

	col.StructureStarts = starts
	col.StructureRefs = refs
}
