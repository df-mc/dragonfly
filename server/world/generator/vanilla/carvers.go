package vanilla

import (
	"hash/fnv"
	"math"
	"sort"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

const carverSearchRadius = 8

func (g Generator) carveTerrain(c *chunk.Chunk, biomes sourceBiomeVolume, chunkX, chunkZ, minY, maxY int, aquifer *gen.NoiseBasedAquifer) {
	if g.carvers == nil {
		return
	}

	chunkBiomes := g.collectChunkBiomes(c, biomes, minY, maxY, false)
	if len(chunkBiomes) == 0 {
		return
	}

	carverSet := make(map[string]struct{})
	for _, biome := range chunkBiomes {
		for _, carverName := range g.biomeGeneration.carverNames[biome] {
			if carverName != "" {
				carverSet[carverName] = struct{}{}
			}
		}
	}
	if len(carverSet) == 0 {
		return
	}

	carverNames := make([]string, 0, len(carverSet))
	for carverName := range carverSet {
		carverNames = append(carverNames, carverName)
	}
	sort.Strings(carverNames)

	for _, carverName := range carverNames {
		g.runCarver(c, chunkX, chunkZ, minY, maxY, carverName, aquifer)
	}
}

func (g Generator) runCarver(c *chunk.Chunk, chunkX, chunkZ, minY, maxY int, carverName string, aquifer *gen.NoiseBasedAquifer) {
	configured, err := g.carvers.Configured(carverName)
	if err != nil {
		return
	}

	switch configured.Type {
	case "cave", "nether_cave":
		cfg, err := configured.Cave()
		if err != nil {
			return
		}
		g.carveCaveSystem(c, chunkX, chunkZ, minY, maxY, carverName, cfg, aquifer)
	case "canyon":
		cfg, err := configured.Canyon()
		if err != nil {
			return
		}
		g.carveCanyonSystem(c, chunkX, chunkZ, minY, maxY, carverName, cfg, aquifer)
	}
}

func (g Generator) carveCaveSystem(c *chunk.Chunk, chunkX, chunkZ, minY, maxY int, carverName string, cfg gen.CaveCarverConfig, aquifer *gen.NoiseBasedAquifer) {
	lavaLevel := clamp(g.anchorY(cfg.LavaLevel, minY, maxY), minY, maxY)

	for dx := -carverSearchRadius; dx <= carverSearchRadius; dx++ {
		for dz := -carverSearchRadius; dz <= carverSearchRadius; dz++ {
			startChunkX := chunkX + dx
			startChunkZ := chunkZ + dz
			rng := g.carverRNG(startChunkX, startChunkZ, carverName)
			if rng.NextDouble() >= cfg.Probability {
				continue
			}

			caveCount := int(rng.NextInt(15)) + 1
			for i := 0; i < caveCount; i++ {
				startX := float64(startChunkX*16) + rng.NextDouble()*16.0
				startY := float64(g.sampleHeightProvider(cfg.Y, minY, maxY, &rng))
				startZ := float64(startChunkZ*16) + rng.NextDouble()*16.0
				yScale := maxFloat(0.1, g.sampleFloatProvider(cfg.YScale, &rng))
				hRadius := maxFloat(0.5, g.sampleFloatProvider(cfg.HorizontalRadiusMultiplier, &rng)*(1.0+rng.NextDouble()*3.0))
				vRadius := maxFloat(0.35, g.sampleFloatProvider(cfg.VerticalRadiusMultiplier, &rng)*(0.75+rng.NextDouble()*1.5)*yScale)
				floorLevel := g.sampleFloatProvider(cfg.FloorLevel, &rng)
				yaw := rng.NextDouble() * math.Pi * 2.0
				pitch := (rng.NextDouble() - 0.5) * math.Pi / 4.0
				branchCount := int(rng.NextInt(40)) + 40

				g.carveCaveBranch(c, chunkX, chunkZ, minY, maxY, &rng, startX, startY, startZ, hRadius, vRadius, floorLevel, yaw, pitch, branchCount, lavaLevel, aquifer)
			}
		}
	}
}

func (g Generator) carveCaveBranch(c *chunk.Chunk, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128, startX, startY, startZ, hRadius, vRadius, floorLevel, yaw, pitch float64, branchCount, lavaLevel int, aquifer *gen.NoiseBasedAquifer) {
	chunkCenterX := float64(chunkX*16) + 8.0
	chunkCenterZ := float64(chunkZ*16) + 8.0

	x, y, z := startX, startY, startZ
	currentYaw := yaw
	currentPitch := pitch
	currentHRadius := hRadius
	currentVRadius := vRadius

	for i := 0; i < branchCount; i++ {
		progress := float64(i) / float64(max(1, branchCount))
		radiusModifier := 1.0 + math.Sin(progress*math.Pi)*0.75
		hRad := currentHRadius * radiusModifier
		vRad := currentVRadius * radiusModifier

		cosPitch := math.Cos(currentPitch)
		sinPitch := math.Sin(currentPitch)

		x += math.Cos(currentYaw) * cosPitch
		y += sinPitch
		z += math.Sin(currentYaw) * cosPitch

		distX := x - chunkCenterX
		distZ := z - chunkCenterZ
		if distX*distX+distZ*distZ < math.Pow(16.0+hRad*2.0, 2) {
			g.carveCaveEllipsoid(c, chunkX, chunkZ, minY, maxY, x, y, z, hRad, vRad, floorLevel, lavaLevel, aquifer)
		}

		if rng.NextDouble() < 0.02 {
			currentYaw += (rng.NextDouble() - 0.5) * math.Pi * 0.25
		}
		if rng.NextDouble() < 0.05 {
			currentPitch += (rng.NextDouble() - 0.5) * 0.5
		}
		currentPitch = clampFloat(currentPitch*0.7, -math.Pi/4.0, math.Pi/4.0)
		if rng.NextDouble() < 0.25 {
			currentHRadius *= 0.9 + rng.NextDouble()*0.2
			currentVRadius *= 0.9 + rng.NextDouble()*0.2
		}
	}
}

func (g Generator) carveCaveEllipsoid(c *chunk.Chunk, chunkX, chunkZ, minY, maxY int, centerX, centerY, centerZ, hRadius, vRadius, floorLevel float64, lavaLevel int, aquifer *gen.NoiseBasedAquifer) {
	chunkMinX := chunkX * 16
	chunkMinZ := chunkZ * 16

	minX := max(int(math.Floor(centerX-hRadius)), chunkMinX)
	maxX := min(int(math.Ceil(centerX+hRadius)), chunkMinX+15)
	minCarveY := max(int(math.Floor(centerY-vRadius)), minY+1)
	maxCarveY := min(int(math.Ceil(centerY+vRadius)), maxY)
	minZ := max(int(math.Floor(centerZ-hRadius)), chunkMinZ)
	maxZ := min(int(math.Ceil(centerZ+hRadius)), chunkMinZ+15)

	floorY := int(centerY + vRadius*floorLevel)
	if floorY > maxCarveY {
		return
	}
	if floorY < minCarveY {
		floorY = minCarveY
	}

	for worldX := minX; worldX <= maxX; worldX++ {
		localX := worldX - chunkMinX
		dx := (float64(worldX) + 0.5 - centerX) / hRadius

		for worldZ := minZ; worldZ <= maxZ; worldZ++ {
			localZ := worldZ - chunkMinZ
			dz := (float64(worldZ) + 0.5 - centerZ) / hRadius
			hDistSq := dx*dx + dz*dz
			if hDistSq >= 1.0 {
				continue
			}

			for worldY := floorY; worldY <= maxCarveY; worldY++ {
				dy := (float64(worldY) + 0.5 - centerY) / vRadius
				if hDistSq+dy*dy < 1.0 {
					g.carveBlock(c, localX, worldY, localZ, worldX, worldZ, minY, lavaLevel, aquifer)
				}
			}
		}
	}
}

func (g Generator) carveCanyonSystem(c *chunk.Chunk, chunkX, chunkZ, minY, maxY int, carverName string, cfg gen.CanyonCarverConfig, aquifer *gen.NoiseBasedAquifer) {
	lavaLevel := clamp(g.anchorY(cfg.LavaLevel, minY, maxY), minY, maxY)

	for dx := -carverSearchRadius; dx <= carverSearchRadius; dx++ {
		for dz := -carverSearchRadius; dz <= carverSearchRadius; dz++ {
			startChunkX := chunkX + dx
			startChunkZ := chunkZ + dz
			rng := g.carverRNG(startChunkX, startChunkZ, carverName)
			if rng.NextDouble() >= cfg.Probability {
				continue
			}

			startX := float64(startChunkX*16) + rng.NextDouble()*16.0
			startY := float64(g.sampleHeightProvider(cfg.Y, minY, maxY, &rng))
			startZ := float64(startChunkZ*16) + rng.NextDouble()*16.0
			length := max(24, int(float64(50+int(rng.NextInt(40)))*maxFloat(0.5, g.sampleFloatProvider(cfg.Shape.DistanceFactor, &rng))))
			yScale := maxFloat(0.6, g.sampleFloatProvider(cfg.YScale, &rng))
			thickness := maxFloat(0.8, g.sampleFloatProvider(cfg.Shape.Thickness, &rng))
			horizontalFactor := maxFloat(0.6, g.sampleFloatProvider(cfg.Shape.HorizontalRadiusFactor, &rng))
			verticalRotation := g.sampleFloatProvider(cfg.VerticalRotation, &rng)

			g.carveCanyonPath(c, chunkX, chunkZ, minY, maxY, &rng, startX, startY, startZ, length, yScale, thickness, horizontalFactor, verticalRotation, cfg, lavaLevel, aquifer)
		}
	}
}

func (g Generator) carveCanyonPath(c *chunk.Chunk, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128, startX, startY, startZ float64, length int, yScale, thickness, horizontalFactor, verticalRotation float64, cfg gen.CanyonCarverConfig, lavaLevel int, aquifer *gen.NoiseBasedAquifer) {
	chunkCenterX := float64(chunkX*16) + 8.0
	chunkCenterZ := float64(chunkZ*16) + 8.0

	x, y, z := startX, startY, startZ
	yaw := rng.NextDouble() * math.Pi * 2.0
	pitch := (rng.NextDouble() - 0.5) * maxFloat(0.125, math.Abs(verticalRotation)+0.125)
	baseWidth := (rng.NextDouble()*2.0 + 2.0) * thickness

	for i := 0; i < length; i++ {
		progress := float64(i) / float64(max(1, length))
		widthFactor := 1.0 - math.Abs(progress*2.0-1.0)
		width := maxFloat(1.5, baseWidth*(0.5+widthFactor)*horizontalFactor)
		heightFactor := maxFloat(0.5, cfg.Shape.VerticalRadiusDefaultFactor+cfg.Shape.VerticalRadiusCenterFactor*widthFactor)
		height := maxFloat(2.0, width*yScale*heightFactor)

		cosPitch := math.Cos(pitch)
		sinPitch := math.Sin(pitch)

		x += math.Cos(yaw) * cosPitch
		y += sinPitch * yScale
		z += math.Sin(yaw) * cosPitch

		yaw += (rng.NextDouble() - 0.5) * 0.2
		pitch = clampFloat(pitch*0.9+(rng.NextDouble()-0.5)*maxFloat(0.05, math.Abs(verticalRotation)+0.05), -0.5, 0.5)

		distX := x - chunkCenterX
		distZ := z - chunkCenterZ
		if distX*distX+distZ*distZ < math.Pow(16.0+width*2.0, 2) {
			g.carveCanyonEllipsoid(c, chunkX, chunkZ, minY, maxY, x, y, z, width, height, lavaLevel, aquifer)
		}
	}
}

func (g Generator) carveCanyonEllipsoid(c *chunk.Chunk, chunkX, chunkZ, minY, maxY int, centerX, centerY, centerZ, hRadius, vRadius float64, lavaLevel int, aquifer *gen.NoiseBasedAquifer) {
	chunkMinX := chunkX * 16
	chunkMinZ := chunkZ * 16

	minX := max(int(math.Floor(centerX-hRadius)), chunkMinX)
	maxX := min(int(math.Ceil(centerX+hRadius)), chunkMinX+15)
	minCarveY := max(int(math.Floor(centerY-vRadius)), minY+1)
	maxCarveY := min(int(math.Ceil(centerY+vRadius)), maxY)
	minZ := max(int(math.Floor(centerZ-hRadius)), chunkMinZ)
	maxZ := min(int(math.Ceil(centerZ+hRadius)), chunkMinZ+15)

	for worldX := minX; worldX <= maxX; worldX++ {
		localX := worldX - chunkMinX
		dx := (float64(worldX) + 0.5 - centerX) / hRadius

		for worldZ := minZ; worldZ <= maxZ; worldZ++ {
			localZ := worldZ - chunkMinZ
			dz := (float64(worldZ) + 0.5 - centerZ) / hRadius
			hDistSq := dx*dx + dz*dz
			if hDistSq >= 1.0 {
				continue
			}

			for worldY := maxCarveY; worldY >= minCarveY; worldY-- {
				dy := (float64(worldY) + 0.5 - centerY) / vRadius
				if hDistSq+(dy*dy)/6.0 < 1.0 {
					g.carveBlock(c, localX, worldY, localZ, worldX, worldZ, minY, lavaLevel, aquifer)
				}
			}
		}
	}
}

func (g Generator) carveBlock(c *chunk.Chunk, localX, worldY, localZ, worldX, worldZ, minY, lavaLevel int, aquifer *gen.NoiseBasedAquifer) bool {
	if worldY <= minY {
		return false
	}

	rid := c.Block(uint8(localX), int16(worldY), uint8(localZ), 0)
	if !g.isCarverReplaceableRID(rid) {
		return false
	}

	if aquifer == nil {
		c.SetBlock(uint8(localX), int16(worldY), uint8(localZ), 0, g.airRID)
		return true
	}

	substance := aquifer.ComputeSubstance(
		gen.FunctionContext{BlockX: worldX, BlockY: worldY, BlockZ: worldZ},
		0,
	)
	if worldY <= lavaLevel {
		substance = gen.AquiferLava
	}

	switch substance {
	case gen.AquiferBarrier:
		return false
	case gen.AquiferWater:
		c.SetBlock(uint8(localX), int16(worldY), uint8(localZ), 0, g.waterRID)
	case gen.AquiferLava:
		c.SetBlock(uint8(localX), int16(worldY), uint8(localZ), 0, g.lavaRID)
	default:
		c.SetBlock(uint8(localX), int16(worldY), uint8(localZ), 0, g.airRID)
	}
	return true
}

func (g Generator) isCarverReplaceableRID(rid uint32) bool {
	if rid == g.defaultBlockRID {
		return true
	}
	return g.dimension == world.Overworld && rid == g.deepRID
}

func (g Generator) sampleFloatProvider(provider gen.FloatProvider, rng *gen.Xoroshiro128) float64 {
	switch provider.Kind {
	case "constant":
		if provider.Constant != nil {
			return *provider.Constant
		}
		return provider.Min
	case "uniform":
		if provider.Max <= provider.Min {
			return provider.Min
		}
		return provider.Min + rng.NextDouble()*(provider.Max-provider.Min)
	case "trapezoid":
		if provider.Max <= provider.Min {
			return provider.Min
		}
		span := provider.Max - provider.Min
		if provider.Plateau <= 0 {
			return provider.Min + (rng.NextDouble()+rng.NextDouble())*span*0.5
		}
		if provider.Plateau >= span {
			return provider.Min + rng.NextDouble()*span
		}

		sideWidth := (span - provider.Plateau) / 2.0
		height := 1.0 / (provider.Plateau + sideWidth)
		leftArea := 0.5 * sideWidth * height
		plateauArea := provider.Plateau * height
		u := rng.NextDouble()

		switch {
		case u < leftArea:
			return provider.Min + math.Sqrt(2.0*sideWidth*u/height)
		case u < leftArea+plateauArea:
			return provider.Min + sideWidth + (u-leftArea)/height
		default:
			return provider.Max - math.Sqrt(2.0*sideWidth*(1.0-u)/height)
		}
	default:
		if provider.Constant != nil {
			return *provider.Constant
		}
		return provider.Min
	}
}

func (g Generator) carverRNG(chunkX, chunkZ int, carverName string) gen.Xoroshiro128 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(carverName))
	seed := int64(h.Sum64()) ^ g.seed ^ int64(chunkX)*341873128712 ^ int64(chunkZ)*132897987541
	return gen.NewXoroshiro128FromSeed(seed)
}

func clampFloat(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
