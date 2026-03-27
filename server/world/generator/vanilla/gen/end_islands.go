package gen

import "math"

type EndIslandDensity struct {
	noise SimplexNoise
}

func NewEndIslandDensity(seed int64) EndIslandDensity {
	rng := NewLegacyRandom(seed)
	rng.ConsumeCount(17292)
	return EndIslandDensity{noise: NewSimplexNoise(&rng)}
}

func (d EndIslandDensity) Sample(blockX, blockZ int) float64 {
	sectionX := blockX / 8
	sectionZ := blockZ / 8
	return float64(endIslandHeightValue(d.noise, sectionX, sectionZ)-8.0) / 128.0
}

func endIslandHeightValue(noise SimplexNoise, sectionX, sectionZ int) float64 {
	chunkX := sectionX / 2
	chunkZ := sectionZ / 2
	subSectionX := sectionX % 2
	subSectionZ := sectionZ % 2

	doffs := 100.0 - math.Sqrt(float64(sectionX*sectionX+sectionZ*sectionZ))*8.0
	doffs = clampFloat(doffs, -100.0, 80.0)

	for xo := -12; xo <= 12; xo++ {
		for zo := -12; zo <= 12; zo++ {
			totalChunkX := int64(chunkX + xo)
			totalChunkZ := int64(chunkZ + zo)
			if totalChunkX*totalChunkX+totalChunkZ*totalChunkZ <= 4096 {
				continue
			}
			if noise.Sample2D(float64(totalChunkX), float64(totalChunkZ)) >= -0.9 {
				continue
			}

			islandSize := math.Mod(math.Abs(float64(totalChunkX))*3439.0+math.Abs(float64(totalChunkZ))*147.0, 13.0) + 9.0
			xd := float64(subSectionX - xo*2)
			zd := float64(subSectionZ - zo*2)
			newDoffs := 100.0 - math.Sqrt(xd*xd+zd*zd)*islandSize
			newDoffs = clampFloat(newDoffs, -100.0, 80.0)
			doffs = maxFloat(doffs, newDoffs)
		}
	}
	return doffs
}

type LegacyRandom struct {
	seed uint64
}

const (
	legacyMask       uint64 = (1 << 48) - 1
	legacyMultiplier uint64 = 25214903917
	legacyIncrement  uint64 = 11
)

func NewLegacyRandom(seed int64) LegacyRandom {
	return LegacyRandom{seed: (uint64(seed) ^ legacyMultiplier) & legacyMask}
}

func (r *LegacyRandom) next(bits int) int {
	r.seed = (r.seed*legacyMultiplier + legacyIncrement) & legacyMask
	return int(r.seed >> (48 - bits))
}

func (r *LegacyRandom) NextInt(bound int) int {
	if bound <= 0 {
		return 0
	}
	if bound&(bound-1) == 0 {
		return int((int64(bound) * int64(r.next(31))) >> 31)
	}
	for {
		bits := r.next(31)
		value := bits % bound
		if bits-value+(bound-1) >= 0 {
			return value
		}
	}
}

func (r *LegacyRandom) NextDouble() float64 {
	return float64((int64(r.next(26))<<27)+int64(r.next(27))) * 1.1102230246251565e-16
}

func (r *LegacyRandom) ConsumeCount(rounds int) {
	for i := 0; i < rounds; i++ {
		r.next(32)
	}
}

type SimplexNoise struct {
	p          [256]int
	xo, yo, zo float64
}

var simplexGradient = [16][3]int{
	{1, 1, 0},
	{-1, 1, 0},
	{1, -1, 0},
	{-1, -1, 0},
	{1, 0, 1},
	{-1, 0, 1},
	{1, 0, -1},
	{-1, 0, -1},
	{0, 1, 1},
	{0, -1, 1},
	{0, 1, -1},
	{0, -1, -1},
	{1, 1, 0},
	{0, -1, 1},
	{-1, 1, 0},
	{0, -1, -1},
}

func NewSimplexNoise(rng *LegacyRandom) SimplexNoise {
	s := SimplexNoise{
		xo: rng.NextDouble() * 256.0,
		yo: rng.NextDouble() * 256.0,
		zo: rng.NextDouble() * 256.0,
	}
	for i := range s.p {
		s.p[i] = i
	}
	for i := 0; i < 256; i++ {
		offset := rng.NextInt(256 - i)
		s.p[i], s.p[offset+i] = s.p[offset+i], s.p[i]
	}
	return s
}

func (s SimplexNoise) perm(x int) int {
	return s.p[x&0xFF]
}

func simplexDot(g [3]int, x, y, z float64) float64 {
	return float64(g[0])*x + float64(g[1])*y + float64(g[2])*z
}

func (s SimplexNoise) cornerNoise3D(index int, x, y, z, base float64) float64 {
	t0 := base - x*x - y*y - z*z
	if t0 < 0 {
		return 0
	}
	t0 *= t0
	return t0 * t0 * simplexDot(simplexGradient[index], x, y, z)
}

func (s SimplexNoise) Sample2D(xin, yin float64) float64 {
	const f2 = 0.3660254037844386
	const g2 = 0.21132486540518713

	skew := (xin + yin) * f2
	i := int(math.Floor(xin + skew))
	j := int(math.Floor(yin + skew))
	unskew := float64(i+j) * g2
	x0 := xin - (float64(i) - unskew)
	y0 := yin - (float64(j) - unskew)

	i1, j1 := 0, 1
	if x0 > y0 {
		i1, j1 = 1, 0
	}

	x1 := x0 - float64(i1) + g2
	y1 := y0 - float64(j1) + g2
	x2 := x0 - 1.0 + 2.0*g2
	y2 := y0 - 1.0 + 2.0*g2

	ii := i & 0xFF
	jj := j & 0xFF
	gi0 := s.perm(ii+s.perm(jj)) % 12
	gi1 := s.perm(ii+i1+s.perm(jj+j1)) % 12
	gi2 := s.perm(ii+1+s.perm(jj+1)) % 12

	n0 := s.cornerNoise3D(gi0, x0, y0, 0.0, 0.5)
	n1 := s.cornerNoise3D(gi1, x1, y1, 0.0, 0.5)
	n2 := s.cornerNoise3D(gi2, x2, y2, 0.0, 0.5)
	return 70.0 * (n0 + n1 + n2)
}
