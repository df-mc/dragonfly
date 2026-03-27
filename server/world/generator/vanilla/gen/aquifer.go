package gen

import "math"

const (
	aquiferYSpacing             = 12
	aquiferXRange               = 10
	aquiferYRange               = 9
	aquiferZRange               = 10
	aquiferSampleOffsetX        = -5
	aquiferSampleOffsetY        = 1
	aquiferSampleOffsetZ        = -5
	aquiferLavaThresholdY       = -10
	aquiferWayBelowMinY         = math.MinInt32 + 1
	aquiferErosionDeepDarkLimit = -0.225
	aquiferDepthDeepDarkLimit   = 0.9
)

var aquiferSurfaceSamplingOffsets = [13][2]int{
	{0, 0},
	{-2, -1},
	{-1, -1},
	{0, -1},
	{1, -1},
	{-3, 0},
	{-2, 0},
	{-1, 0},
	{1, 0},
	{-2, 1},
	{-1, 1},
	{0, 1},
	{1, 1},
}

type AquiferSubstance uint8

const (
	AquiferAir AquiferSubstance = iota
	AquiferWater
	AquiferLava
	AquiferBarrier
)

type FluidType uint8

const (
	FluidWater FluidType = iota
	FluidLava
)

type FluidStatus struct {
	FluidLevel int
	FluidType  FluidType
}

func (s FluidStatus) At(y int) AquiferSubstance {
	if y < s.FluidLevel {
		if s.FluidType == FluidLava {
			return AquiferLava
		}
		return AquiferWater
	}
	return AquiferAir
}

type FluidPicker interface {
	ComputeFluid(x, y, z int) FluidStatus
}

type OverworldFluidPicker struct {
	SeaLevel int
}

func (p OverworldFluidPicker) ComputeFluid(_ int, y int, _ int) FluidStatus {
	if y < -54 {
		return FluidStatus{FluidLevel: -54, FluidType: FluidLava}
	}
	return FluidStatus{FluidLevel: p.SeaLevel, FluidType: FluidWater}
}

type aquiferCellKey struct {
	x int
	y int
	z int
}

type aquiferLocation struct {
	x int
	y int
	z int
}

type NoiseBasedAquifer struct {
	graph            *Graph
	noises           NoiseSource
	mainFlat         *FlatCacheGrid
	mainChunkX       int
	mainChunkZ       int
	positionalRandom PositionalRandomFactory
	fluidPicker      FluidPicker
	scratch          *EvalScratch

	neighborGrids map[[2]int]*FlatCacheGrid
	columnCache   map[[2]int]*ColumnContext
	surfaceCache  map[[2]int]int
	cacheMin      aquiferCellKey
	cacheSizeX    int
	cacheSizeY    int
	cacheSizeZ    int
	locationSet   []bool
	locationCache []aquiferLocation
	statusSet     []bool
	statusCache   []FluidStatus

	skipSamplingAboveY int
}

func NewNoiseBasedAquifer(
	graph *Graph,
	chunkX, chunkZ int,
	minY, maxY int,
	noises NoiseSource,
	flat *FlatCacheGrid,
	seed int64,
	fluidPicker FluidPicker,
) *NoiseBasedAquifer {
	minGridX := aquiferGridX(chunkX*16 + aquiferSampleOffsetX)
	maxGridX := aquiferGridX(chunkX*16+15+aquiferSampleOffsetX) + 1
	minGridY := aquiferGridY(minY+aquiferSampleOffsetY) - 1
	maxGridY := aquiferGridY(maxY+aquiferSampleOffsetY) + 1
	minGridZ := aquiferGridZ(chunkZ*16 + aquiferSampleOffsetZ)
	maxGridZ := aquiferGridZ(chunkZ*16+15+aquiferSampleOffsetZ) + 1
	cacheSizeX := maxGridX - minGridX + 1
	cacheSizeY := maxGridY - minGridY + 1
	cacheSizeZ := maxGridZ - minGridZ + 1
	cacheLen := cacheSizeX * cacheSizeY * cacheSizeZ

	return &NoiseBasedAquifer{
		graph:              graph,
		noises:             noises,
		mainFlat:           flat,
		mainChunkX:         chunkX,
		mainChunkZ:         chunkZ,
		positionalRandom:   NewPositionalRandomFactory(seed).ForkAquiferRandom(),
		fluidPicker:        fluidPicker,
		scratch:            NewEvalScratch(graph),
		neighborGrids:      make(map[[2]int]*FlatCacheGrid, 16),
		columnCache:        make(map[[2]int]*ColumnContext, 128),
		surfaceCache:       make(map[[2]int]int, 128),
		cacheMin:           aquiferCellKey{x: minGridX, y: minGridY, z: minGridZ},
		cacheSizeX:         cacheSizeX,
		cacheSizeY:         cacheSizeY,
		cacheSizeZ:         cacheSizeZ,
		locationSet:        make([]bool, cacheLen),
		locationCache:      make([]aquiferLocation, cacheLen),
		statusSet:          make([]bool, cacheLen),
		statusCache:        make([]FluidStatus, cacheLen),
		skipSamplingAboveY: math.MaxInt32,
	}
}

func (a *NoiseBasedAquifer) ComputeSubstance(ctx FunctionContext, density float64) AquiferSubstance {
	if density > 0 {
		return AquiferBarrier
	}

	globalFluid := a.fluidPicker.ComputeFluid(ctx.BlockX, ctx.BlockY, ctx.BlockZ)
	if ctx.BlockY > a.skipSamplingAboveY {
		return globalFluid.At(ctx.BlockY)
	}
	if globalFluid.At(ctx.BlockY) == AquiferLava {
		return AquiferLava
	}

	gx := aquiferGridX(ctx.BlockX + aquiferSampleOffsetX)
	gy := aquiferGridY(ctx.BlockY + aquiferSampleOffsetY)
	gz := aquiferGridZ(ctx.BlockZ + aquiferSampleOffsetZ)

	var (
		closestKeys  [3]aquiferCellKey
		closestDists = [3]int{math.MaxInt32, math.MaxInt32, math.MaxInt32}
	)

	for dx := 0; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			for dz := 0; dz <= 1; dz++ {
				key := aquiferCellKey{x: gx + dx, y: gy + dy, z: gz + dz}
				pos := a.getAquiferLocation(key)
				distSq := squaredDistance(pos.x-ctx.BlockX, pos.y-ctx.BlockY, pos.z-ctx.BlockZ)
				insertAquiferNearest(&closestKeys, &closestDists, key, distSq)
			}
		}
	}

	fluid1 := a.getAquiferStatus(closestKeys[0])
	blockAtY := fluid1.At(ctx.BlockY)
	similarity := aquiferSimilarity(closestDists[0], closestDists[1])
	flowingUpdateSimilarity := aquiferSimilarity(100, 144)

	if similarity <= 0 {
		return blockAtY
	}

	if blockAtY == AquiferWater {
		belowFluid := a.fluidPicker.ComputeFluid(ctx.BlockX, ctx.BlockY-1, ctx.BlockZ)
		if belowFluid.At(ctx.BlockY-1) == AquiferLava {
			return blockAtY
		}
	}

	col := a.getColumnContext(ctx.BlockX, ctx.BlockZ)
	fluid2 := a.getAquiferStatus(closestKeys[1])
	barrierValue := math.NaN()
	pressure1 := similarity * a.calculatePressure(ctx, col, &barrierValue, fluid1, fluid2)
	if density+pressure1 > 0 {
		return AquiferBarrier
	}

	fluid3 := a.getAquiferStatus(closestKeys[2])
	similarity2 := aquiferSimilarity(closestDists[0], closestDists[2])
	if similarity2 > 0 {
		pressure2 := similarity * similarity2 * a.calculatePressure(ctx, col, &barrierValue, fluid1, fluid3)
		if density+pressure2 > 0 {
			return AquiferBarrier
		}
	}

	similarity3 := aquiferSimilarity(closestDists[1], closestDists[2])
	if similarity3 > 0 {
		pressure3 := similarity * similarity3 * a.calculatePressure(ctx, col, &barrierValue, fluid2, fluid3)
		if density+pressure3 > 0 {
			return AquiferBarrier
		}
	}

	_ = flowingUpdateSimilarity
	return blockAtY
}

func (a *NoiseBasedAquifer) getAquiferLocation(key aquiferCellKey) aquiferLocation {
	if index, ok := a.cacheIndex(key); ok {
		if a.locationSet[index] {
			return a.locationCache[index]
		}
		rng := a.positionalRandom.At(key.x, key.y, key.z)
		pos := aquiferLocation{
			x: aquiferFromGridX(key.x, int(rng.NextInt(aquiferXRange))),
			y: aquiferFromGridY(key.y, int(rng.NextInt(aquiferYRange))),
			z: aquiferFromGridZ(key.z, int(rng.NextInt(aquiferZRange))),
		}
		a.locationSet[index] = true
		a.locationCache[index] = pos
		return pos
	}
	rng := a.positionalRandom.At(key.x, key.y, key.z)
	return aquiferLocation{
		x: aquiferFromGridX(key.x, int(rng.NextInt(aquiferXRange))),
		y: aquiferFromGridY(key.y, int(rng.NextInt(aquiferYRange))),
		z: aquiferFromGridZ(key.z, int(rng.NextInt(aquiferZRange))),
	}
}

func (a *NoiseBasedAquifer) getAquiferStatus(key aquiferCellKey) FluidStatus {
	if index, ok := a.cacheIndex(key); ok {
		if a.statusSet[index] {
			return a.statusCache[index]
		}
		pos := a.getAquiferLocation(key)
		status := a.computeFluid(pos.x, pos.y, pos.z)
		a.statusSet[index] = true
		a.statusCache[index] = status
		return status
	}
	pos := a.getAquiferLocation(key)
	return a.computeFluid(pos.x, pos.y, pos.z)
}

func (a *NoiseBasedAquifer) cacheIndex(key aquiferCellKey) (int, bool) {
	relX := key.x - a.cacheMin.x
	relY := key.y - a.cacheMin.y
	relZ := key.z - a.cacheMin.z
	if relX < 0 || relX >= a.cacheSizeX || relY < 0 || relY >= a.cacheSizeY || relZ < 0 || relZ >= a.cacheSizeZ {
		return 0, false
	}
	return ((relY*a.cacheSizeZ)+relZ)*a.cacheSizeX + relX, true
}

func (a *NoiseBasedAquifer) computeFluid(x, y, z int) FluidStatus {
	globalFluid := a.fluidPicker.ComputeFluid(x, y, z)
	minSurfaceRaw := math.MaxInt32
	yUpper := y + 12
	yLower := y - 12
	isBelowSurfaceWithFluid := false

	for i, offset := range aquiferSurfaceSamplingOffsets {
		sampleX := x + offset[0]*16
		sampleZ := z + offset[1]*16
		quartX := (sampleX >> 2) << 2
		quartZ := (sampleZ >> 2) << 2

		rawSurface, ok := a.surfaceCache[[2]int{quartX, quartZ}]
		if !ok {
			col := a.getColumnContext(quartX, quartZ)
			grid := a.gridForBlock(quartX, quartZ)
			rawSurface = int(math.Floor(ComputePreliminarySurfaceLevel(
				FunctionContext{BlockX: quartX, BlockY: 0, BlockZ: quartZ},
				a.noises,
				grid,
				col,
			)))
			a.surfaceCache[[2]int{quartX, quartZ}] = rawSurface
		}

		adjustedSurface := rawSurface + 8
		isAtOurPosition := i == 0
		if isAtOurPosition && yLower > adjustedSurface {
			return globalFluid
		}

		isAboveAdjustedSurface := yUpper > adjustedSurface
		if isAboveAdjustedSurface || isAtOurPosition {
			surfaceFluid := a.fluidPicker.ComputeFluid(sampleX, adjustedSurface, sampleZ)
			if surfaceFluid.At(adjustedSurface) != AquiferAir {
				if isAtOurPosition {
					isBelowSurfaceWithFluid = true
				}
				if isAboveAdjustedSurface {
					return surfaceFluid
				}
			}
		}

		if rawSurface < minSurfaceRaw {
			minSurfaceRaw = rawSurface
		}
	}

	fluidLevel := a.computeSurfaceLevel(x, y, z, globalFluid, minSurfaceRaw, isBelowSurfaceWithFluid)
	return FluidStatus{
		FluidLevel: fluidLevel,
		FluidType:  a.computeFluidType(x, y, z, globalFluid, fluidLevel),
	}
}

func (a *NoiseBasedAquifer) computeSurfaceLevel(
	x, y, z int,
	globalFluid FluidStatus,
	minSurfaceRaw int,
	isBelowSurfaceWithFluid bool,
) int {
	ctx := FunctionContext{BlockX: x, BlockY: y, BlockZ: z}
	col := a.getColumnContext(x, z)
	grid := a.gridForBlock(x, z)
	if a.isDeepDarkRegion(ctx, grid, col) {
		return aquiferWayBelowMinY
	}

	distFromSurface := minSurfaceRaw + 8 - y
	surfaceProximity := 0.0
	if isBelowSurfaceWithFluid {
		clamped := clampFloat(float64(distFromSurface), 0, 64)
		surfaceProximity = 1.0 - clamped/64.0
	}

	floodednessNoise := clampFloat(a.evalSimpleRoot(OverworldRootFluidLevelFloodedness, ctx), -1, 1)
	thresholdH := 0.8 - 1.1*surfaceProximity
	thresholdO := 0.4 - 1.2*surfaceProximity
	d := floodednessNoise - thresholdO
	e := floodednessNoise - thresholdH

	if e > 0 {
		return globalFluid.FluidLevel
	}
	if d > 0 {
		return a.computeRandomizedFluidLevel(x, y, z, minSurfaceRaw)
	}
	return aquiferWayBelowMinY
}

func (a *NoiseBasedAquifer) computeRandomizedFluidLevel(x, y, z, minSurface int) int {
	ctx := FunctionContext{
		BlockX: floorDivInt(x, 16),
		BlockY: floorDivInt(y, 40),
		BlockZ: floorDivInt(z, 16),
	}
	baseLevel := ctx.BlockY*40 + 20
	spread := a.evalSimpleRoot(OverworldRootFluidLevelSpread, ctx) * 10.0
	quantized := int(math.Floor(spread/3.0)) * 3
	level := baseLevel + quantized
	if level > minSurface {
		return minSurface
	}
	return level
}

func (a *NoiseBasedAquifer) computeFluidType(x, y, z int, globalFluid FluidStatus, fluidLevel int) FluidType {
	if globalFluid.FluidType == FluidLava {
		return FluidLava
	}
	if fluidLevel <= aquiferLavaThresholdY && fluidLevel != aquiferWayBelowMinY {
		ctx := FunctionContext{
			BlockX: floorDivInt(x, 64),
			BlockY: floorDivInt(y, 40),
			BlockZ: floorDivInt(z, 64),
		}
		if math.Abs(a.evalSimpleRoot(OverworldRootLava, ctx)) > 0.3 {
			return FluidLava
		}
	}
	return FluidWater
}

func (a *NoiseBasedAquifer) isDeepDarkRegion(ctx FunctionContext, grid *FlatCacheGrid, col *ColumnContext) bool {
	erosion := a.graph.Eval(OverworldRootErosion, ctx, a.noises, grid, col, a.scratch)
	depth := a.graph.Eval(OverworldRootDepth, ctx, a.noises, grid, col, a.scratch)
	return erosion < aquiferErosionDeepDarkLimit && depth > aquiferDepthDeepDarkLimit
}

func (a *NoiseBasedAquifer) calculatePressure(
	ctx FunctionContext,
	_ *ColumnContext,
	barrierValue *float64,
	fluid1, fluid2 FluidStatus,
) float64 {
	block1 := fluid1.At(ctx.BlockY)
	block2 := fluid2.At(ctx.BlockY)
	if (block1 == AquiferLava && block2 == AquiferWater) || (block1 == AquiferWater && block2 == AquiferLava) {
		return 2.0
	}

	levelDiff := int(math.Abs(float64(fluid1.FluidLevel - fluid2.FluidLevel)))
	if levelDiff == 0 {
		return 0
	}

	avgLevel := float64(fluid1.FluidLevel+fluid2.FluidLevel) * 0.5
	signedOffset := float64(ctx.BlockY) + 0.5 - avgLevel
	halfDiff := float64(levelDiff) * 0.5
	o := halfDiff - math.Abs(signedOffset)

	var q float64
	if signedOffset > 0 {
		if o > 0 {
			q = o / 1.5
		} else {
			q = o / 2.5
		}
	} else {
		p := 3.0 + o
		if p > 0 {
			q = p / 3.0
		} else {
			q = p / 10.0
		}
	}

	barrier := 0.0
	if q >= -2.0 && q <= 2.0 {
		if math.IsNaN(*barrierValue) {
			*barrierValue = a.evalSimpleRoot(OverworldRootBarrier, ctx)
		}
		barrier = *barrierValue
	}
	return 2.0 * (barrier + q)
}

func (a *NoiseBasedAquifer) getColumnContext(x, z int) *ColumnContext {
	key := [2]int{x, z}
	if col, ok := a.columnCache[key]; ok {
		return col
	}
	grid := a.gridForBlock(x, z)
	col := a.graph.NewColumnContext(x, z, a.noises, grid)
	a.columnCache[key] = col
	return col
}

func (a *NoiseBasedAquifer) gridForBlock(x, z int) *FlatCacheGrid {
	chunkX := x >> 4
	chunkZ := z >> 4
	if chunkX == a.mainChunkX && chunkZ == a.mainChunkZ {
		return a.mainFlat
	}
	key := [2]int{chunkX, chunkZ}
	if grid, ok := a.neighborGrids[key]; ok {
		return grid
	}
	grid := a.graph.NewFlatCacheGrid(chunkX, chunkZ, a.noises)
	a.neighborGrids[key] = grid
	return grid
}

func (a *NoiseBasedAquifer) evalSimpleRoot(root int, ctx FunctionContext) float64 {
	return a.graph.Eval(root, ctx, a.noises, nil, nil, a.scratch)
}

func aquiferGridX(x int) int {
	return x >> 4
}

func aquiferGridY(y int) int {
	return floorDivInt(y, aquiferYSpacing)
}

func aquiferGridZ(z int) int {
	return z >> 4
}

func aquiferFromGridX(gridX, offset int) int {
	return (gridX << 4) + offset
}

func aquiferFromGridY(gridY, offset int) int {
	return gridY*aquiferYSpacing + offset
}

func aquiferFromGridZ(gridZ, offset int) int {
	return (gridZ << 4) + offset
}

func aquiferSimilarity(dist1Sq, dist2Sq int) float64 {
	return 1.0 - float64(dist2Sq-dist1Sq)/25.0
}

func insertAquiferNearest(keys *[3]aquiferCellKey, dists *[3]int, key aquiferCellKey, dist int) {
	if dists[0] >= dist {
		keys[2], keys[1], keys[0] = keys[1], keys[0], key
		dists[2], dists[1], dists[0] = dists[1], dists[0], dist
		return
	}
	if dists[1] >= dist {
		keys[2], keys[1] = keys[1], key
		dists[2], dists[1] = dists[1], dist
		return
	}
	if dists[2] >= dist {
		keys[2] = key
		dists[2] = dist
	}
}

func squaredDistance(x, y, z int) int {
	return x*x + y*y + z*z
}

func floorDivInt(v, d int) int {
	if v >= 0 {
		return v / d
	}
	return -((-v + d - 1) / d)
}
