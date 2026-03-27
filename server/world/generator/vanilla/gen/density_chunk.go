package gen

type FinalDensityChunk struct {
	minY       int
	baseX      int
	baseZ      int
	cellCountY int
	corners    [5][5][49]float64
}

func NewFinalDensityChunk(graph *Graph, chunkX, chunkZ, minY, maxY int, noises NoiseSource, flat *FlatCacheGrid) *FinalDensityChunk {
	return NewFinalDensityChunkWithEvaluator(
		graph,
		OverworldRootFinalDensity,
		chunkX,
		chunkZ,
		minY,
		maxY,
		noises,
		flat,
		ComputeFinalDensity,
		ComputeFinalDensity4,
	)
}

type DensityScalarEvaluator func(FunctionContext, NoiseSource, *FlatCacheGrid, *ColumnContext) float64
type DensityVectorEvaluator func(FunctionContext4, NoiseSource, *FlatCacheGrid, *ColumnContext) [4]float64

func NewFinalDensityChunkFromRoot(graph *Graph, root, chunkX, chunkZ, minY, maxY int, noises NoiseSource, flat *FlatCacheGrid) *FinalDensityChunk {
	return NewFinalDensityChunkWithEvaluator(graph, root, chunkX, chunkZ, minY, maxY, noises, flat, nil, nil)
}

func NewFinalDensityChunkWithEvaluator(graph *Graph, root, chunkX, chunkZ, minY, maxY int, noises NoiseSource, flat *FlatCacheGrid, scalar DensityScalarEvaluator, vector DensityVectorEvaluator) *FinalDensityChunk {
	chunk := &FinalDensityChunk{
		minY:       minY,
		baseX:      chunkX * 16,
		baseZ:      chunkZ * 16,
		cellCountY: (maxY - minY + 1) / 8,
	}

	for cornerX := 0; cornerX <= 4; cornerX++ {
		worldX := chunk.baseX + cornerX*4
		for cornerZ := 0; cornerZ <= 4; cornerZ++ {
			worldZ := chunk.baseZ + cornerZ*4
			column := graph.NewColumnContext(worldX, worldZ, noises, flat)
			for cornerY := 0; cornerY <= chunk.cellCountY; {
				if cornerY+3 <= chunk.cellCountY {
					values := evalDensityVector(
						graph,
						root,
						FunctionContext4{
							BlockX: worldX,
							BlockY: [4]int{
								minY + cornerY*8,
								minY + (cornerY+1)*8,
								minY + (cornerY+2)*8,
								minY + (cornerY+3)*8,
							},
							BlockZ: worldZ,
						},
						noises,
						flat,
						column,
						scalar,
						vector,
					)
					chunk.corners[cornerX][cornerZ][cornerY] = values[0]
					chunk.corners[cornerX][cornerZ][cornerY+1] = values[1]
					chunk.corners[cornerX][cornerZ][cornerY+2] = values[2]
					chunk.corners[cornerX][cornerZ][cornerY+3] = values[3]
					cornerY += 4
					continue
				}

				worldY := minY + cornerY*8
				chunk.corners[cornerX][cornerZ][cornerY] = evalDensityScalar(
					graph,
					root,
					FunctionContext{BlockX: worldX, BlockY: worldY, BlockZ: worldZ},
					noises,
					flat,
					column,
					scalar,
				)
				cornerY++
			}
		}
	}
	return chunk
}

func (c *FinalDensityChunk) Density(localX, blockY, localZ int) float64 {
	cellX := localX >> 2
	cellZ := localZ >> 2
	cellY := (blockY - c.minY) >> 3

	inCellX := localX & 3
	inCellZ := localZ & 3
	inCellY := (blockY - c.minY) & 7

	tx := float64(inCellX) / 4.0
	ty := float64(inCellY) / 8.0
	tz := float64(inCellZ) / 4.0

	return lerp3(
		tx,
		ty,
		tz,
		c.corners[cellX][cellZ][cellY],
		c.corners[cellX+1][cellZ][cellY],
		c.corners[cellX][cellZ][cellY+1],
		c.corners[cellX+1][cellZ][cellY+1],
		c.corners[cellX][cellZ+1][cellY],
		c.corners[cellX+1][cellZ+1][cellY],
		c.corners[cellX][cellZ+1][cellY+1],
		c.corners[cellX+1][cellZ+1][cellY+1],
	)
}

func lerp3(tx, ty, tz, c000, c100, c010, c110, c001, c101, c011, c111 float64) float64 {
	x00 := densityLerp(tx, c000, c100)
	x10 := densityLerp(tx, c010, c110)
	x01 := densityLerp(tx, c001, c101)
	x11 := densityLerp(tx, c011, c111)
	y0 := densityLerp(ty, x00, x10)
	y1 := densityLerp(ty, x01, x11)
	return densityLerp(tz, y0, y1)
}

func densityLerp(t, a, b float64) float64 {
	return a + t*(b-a)
}

func evalDensityScalar(graph *Graph, root int, ctx FunctionContext, noises NoiseSource, flat *FlatCacheGrid, col *ColumnContext, scalar DensityScalarEvaluator) float64 {
	if scalar != nil {
		return scalar(ctx, noises, flat, col)
	}
	if graph == nil {
		return 0
	}
	return graph.Eval(root, ctx, noises, flat, col, nil)
}

func EvalDensityScalar(graph *Graph, root int, ctx FunctionContext, noises NoiseSource, flat *FlatCacheGrid, col *ColumnContext, scalar DensityScalarEvaluator) float64 {
	return evalDensityScalar(graph, root, ctx, noises, flat, col, scalar)
}

func evalDensityVector(graph *Graph, root int, ctx FunctionContext4, noises NoiseSource, flat *FlatCacheGrid, col *ColumnContext, scalar DensityScalarEvaluator, vector DensityVectorEvaluator) [4]float64 {
	if vector != nil {
		return vector(ctx, noises, flat, col)
	}
	return [4]float64{
		evalDensityScalar(graph, root, FunctionContext{BlockX: ctx.BlockX, BlockY: ctx.BlockY[0], BlockZ: ctx.BlockZ}, noises, flat, col, scalar),
		evalDensityScalar(graph, root, FunctionContext{BlockX: ctx.BlockX, BlockY: ctx.BlockY[1], BlockZ: ctx.BlockZ}, noises, flat, col, scalar),
		evalDensityScalar(graph, root, FunctionContext{BlockX: ctx.BlockX, BlockY: ctx.BlockY[2], BlockZ: ctx.BlockZ}, noises, flat, col, scalar),
		evalDensityScalar(graph, root, FunctionContext{BlockX: ctx.BlockX, BlockY: ctx.BlockY[3], BlockZ: ctx.BlockZ}, noises, flat, col, scalar),
	}
}
