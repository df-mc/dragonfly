package gen

import "math"

type NoiseRegistry struct {
	noises       []DoublePerlinNoise
	blendedNoise BlendedNoise
	endIslands   EndIslandDensity
}

func NewNoiseRegistry(seed int64) *NoiseRegistry {
	noises := make([]DoublePerlinNoise, len(NoiseParams))
	for i, params := range NoiseParams {
		saltedSeed := seed*31 + int64(i)
		rng := NewXoroshiro128FromSeed(saltedSeed)
		noises[i] = NewDoublePerlinNoise(&rng, params.Amplitudes, params.FirstOctave)
	}

	blendedSeed := seed*31 + 1000
	blendedRng := NewXoroshiro128FromSeed(blendedSeed)
	return &NoiseRegistry{
		noises: noises,
		blendedNoise: NewBlendedNoise(
			&blendedRng,
			0.25,
			0.125,
			80.0,
			160.0,
			8.0,
		),
		endIslands: NewEndIslandDensity(seed),
	}
}

func (n *NoiseRegistry) Sample(noise NoiseRef, x, y, z float64) float64 {
	idx := int(noise)
	if idx < 0 || idx >= len(n.noises) {
		return 0
	}
	return n.noises[idx].Sample(x, y, z)
}

func (n *NoiseRegistry) SampleBlendedNoise(x, y, z, _xzScale, _yScale, _xzFactor, _yFactor, _smearScaleMultiplier float64) float64 {
	return n.blendedNoise.Sample(x, y, z)
}

func (n *NoiseRegistry) SampleEndIslands(blockX, blockZ int) float64 {
	return n.endIslands.Sample(blockX, blockZ)
}

type PerlinNoise struct {
	d          [257]int32
	a          float64
	b          float64
	c          float64
	amplitude  float64
	lacunarity float64
	h2         int32
	d2         float64
	t2         float64
}

func NewPerlinNoise(rng *Xoroshiro128) PerlinNoise {
	a := rng.NextDouble() * 256
	b := rng.NextDouble() * 256
	c := rng.NextDouble() * 256

	var d [257]int32
	for i := 0; i < 256; i++ {
		d[i] = int32(i)
	}
	for i := 0; i < 256; i++ {
		j := int(rng.NextInt(uint32(256-i))) + i
		d[i], d[j] = d[j], d[i]
	}
	d[256] = d[0]

	i2 := math.Floor(b)
	d2 := b - i2
	h2 := int32(i2) & 255
	t2 := smoothstep(d2)

	return PerlinNoise{
		d:          d,
		a:          a,
		b:          b,
		c:          c,
		amplitude:  1,
		lacunarity: 1,
		h2:         h2,
		d2:         d2,
		t2:         t2,
	}
}

func (p *PerlinNoise) Sample(x, y, z float64) float64 {
	var d2, t2 float64
	var h2 int32
	if y == 0 {
		d2, h2, t2 = p.d2, p.h2, p.t2
	} else {
		yv := y + p.b
		i2 := math.Floor(yv)
		d2 = yv - i2
		h2 = int32(i2) & 255
		t2 = smoothstep(d2)
	}

	d1 := x + p.a
	d3 := z + p.c
	i1 := math.Floor(d1)
	i3 := math.Floor(d3)
	d1 -= i1
	d3 -= i3

	h1 := int32(i1) & 255
	h3 := int32(i3) & 255
	t1 := smoothstep(d1)
	t3 := smoothstep(d3)

	idx := p.d[:]
	a1 := (idx[h1] + h2) & 255
	b1 := (idx[(h1+1)&255] + h2) & 255
	a2 := (idx[a1] + h3) & 255
	a3 := (idx[(a1+1)&255] + h3) & 255
	b2 := (idx[b1] + h3) & 255
	b3 := (idx[(b1+1)&255] + h3) & 255

	l1 := indexedLerp(idx[a2]&15, d1, d2, d3)
	l2 := indexedLerp(idx[b2]&15, d1-1, d2, d3)
	l3 := indexedLerp(idx[a3]&15, d1, d2-1, d3)
	l4 := indexedLerp(idx[b3]&15, d1-1, d2-1, d3)
	l5 := indexedLerp(idx[(a2+1)&255]&15, d1, d2, d3-1)
	l6 := indexedLerp(idx[(b2+1)&255]&15, d1-1, d2, d3-1)
	l7 := indexedLerp(idx[(a3+1)&255]&15, d1, d2-1, d3-1)
	l8 := indexedLerp(idx[(b3+1)&255]&15, d1-1, d2-1, d3-1)

	l1 = lerp(t1, l1, l2)
	l3 = lerp(t1, l3, l4)
	l5 = lerp(t1, l5, l6)
	l7 = lerp(t1, l7, l8)
	l1 = lerp(t2, l1, l3)
	l5 = lerp(t2, l5, l7)
	return lerp(t3, l1, l5)
}

func (p *PerlinNoise) SampleSmeared(x, y, z, yScale, yOrig float64) float64 {
	d1 := x + p.a
	d2Raw := y + p.b
	d3 := z + p.c

	i1 := math.Floor(d1)
	i2 := math.Floor(d2Raw)
	i3 := math.Floor(d3)

	d1 -= i1
	d2 := d2Raw - i2
	d3 -= i3

	s := 0.0
	if yScale != 0 {
		r := d2
		if yOrig >= 0 && yOrig < d2 {
			r = yOrig
		}
		s = math.Floor(r/yScale+1.0e-7) * yScale
	}
	d2Smeared := d2 - s

	h1 := int32(i1) & 255
	h2 := int32(i2) & 255
	h3 := int32(i3) & 255

	t1 := smoothstep(d1)
	t2 := smoothstep(d2)
	t3 := smoothstep(d3)

	idx := p.d[:]
	a1 := (idx[h1] + h2) & 255
	b1 := (idx[(h1+1)&255] + h2) & 255
	a2 := (idx[a1] + h3) & 255
	a3 := (idx[(a1+1)&255] + h3) & 255
	b2 := (idx[b1] + h3) & 255
	b3 := (idx[(b1+1)&255] + h3) & 255

	l1 := indexedLerp(idx[a2]&15, d1, d2Smeared, d3)
	l2 := indexedLerp(idx[b2]&15, d1-1, d2Smeared, d3)
	l3 := indexedLerp(idx[a3]&15, d1, d2Smeared-1, d3)
	l4 := indexedLerp(idx[b3]&15, d1-1, d2Smeared-1, d3)
	l5 := indexedLerp(idx[(a2+1)&255]&15, d1, d2Smeared, d3-1)
	l6 := indexedLerp(idx[(b2+1)&255]&15, d1-1, d2Smeared, d3-1)
	l7 := indexedLerp(idx[(a3+1)&255]&15, d1, d2Smeared-1, d3-1)
	l8 := indexedLerp(idx[(b3+1)&255]&15, d1-1, d2Smeared-1, d3-1)

	l1 = lerp(t1, l1, l2)
	l3 = lerp(t1, l3, l4)
	l5 = lerp(t1, l5, l6)
	l7 = lerp(t1, l7, l8)
	l1 = lerp(t2, l1, l3)
	l5 = lerp(t2, l5, l7)
	return lerp(t3, l1, l5)
}

type OctaveNoise struct {
	octaves []PerlinNoise
}

var md5OctaveN = [13][2]uint64{
	{0xb198de63a8012672, 0x7b84cad43ef7b5a8},
	{0x0fd787bfbc403ec3, 0x74a4a31ca21b48b8},
	{0x36d326eed40efeb2, 0x5be9ce18223c636a},
	{0x082fe255f8be6631, 0x4e96119e22dedc81},
	{0x0ef68ec68504005e, 0x48b6bf93a2789640},
	{0xf11268128982754f, 0x257a1d670430b0aa},
	{0xe51c98ce7d1de664, 0x5f9478a733040c45},
	{0x6d7b49e7e429850a, 0x2e3063c622a24777},
	{0xbd90d5377ba1b762, 0xc07317d419a7548d},
	{0x53d39c6752dac858, 0xbcd1c5a80ab65b3e},
	{0xb4a24d7a84e7677b, 0x023ff9668e89b5c4},
	{0xdffa22b534c5f608, 0xb9b67517d3665ca9},
	{0xd50708086cef4d7c, 0x6e1651ecc7f43309},
}

func NewOctaveNoise(rng *Xoroshiro128, amplitudes []float64, omin int) OctaveNoise {
	persistIni := []float64{
		0.0,
		1.0,
		0.6666666666666666,
		0.5714285714285714,
		0.5333333333333333,
		0.5161290322580645,
		0.5079365079365079,
		0.503937007874016,
		0.5019607843137255,
		0.5009775171065493,
		0.50048828125,
	}

	lacuna := 1.0
	persist := 0.0
	if len(amplitudes) < len(persistIni) {
		persist = persistIni[len(amplitudes)]
	} else {
		pow := 1 << len(amplitudes)
		persist = float64(pow) / float64(pow-1)
	}

	xLo := rng.NextLong()
	xHi := rng.NextLong()
	octaves := make([]PerlinNoise, 0, len(amplitudes))

	for i, amp := range amplitudes {
		if amp != 0 {
			octaveIdx := 12 + omin + i
			if octaveIdx >= 0 && octaveIdx < len(md5OctaveN) {
				md5 := md5OctaveN[octaveIdx]
				pxr := NewXoroshiro128FromState(xLo^md5[0], xHi^md5[1])
				noise := NewPerlinNoise(&pxr)
				noise.amplitude = amp * persist
				noise.lacunarity = lacuna
				octaves = append(octaves, noise)
			}
		}
		lacuna *= 2
		persist *= 0.5
	}

	return OctaveNoise{octaves: octaves}
}

func (o *OctaveNoise) Sample(x, y, z float64) float64 {
	value := 0.0
	for i := range o.octaves {
		octave := &o.octaves[i]
		lf := octave.lacunarity
		value += octave.amplitude * octave.Sample(x*lf, y*lf, z*lf)
	}
	return value
}

type DoublePerlinNoise struct {
	amplitude float64
	frequency float64
	octA      OctaveNoise
	octB      OctaveNoise
}

func NewDoublePerlinNoise(rng *Xoroshiro128, amplitudes []float64, omin int) DoublePerlinNoise {
	octA := NewOctaveNoise(rng, amplitudes, omin)
	octB := NewOctaveNoise(rng, amplitudes, omin)

	length := len(amplitudes)
	for i := len(amplitudes) - 1; i >= 0; i-- {
		if amplitudes[i] != 0 {
			break
		}
		length--
	}
	for _, amp := range amplitudes {
		if amp != 0 {
			break
		}
		length--
	}

	ampIni := []float64{
		0.0,
		0.8333333333333334,
		1.1111111111111112,
		1.25,
		1.3333333333333333,
		1.3888888888888888,
		1.4285714285714286,
		1.4583333333333333,
		1.4814814814814814,
		1.5,
		1.5151515151515151,
	}

	amplitude := 0.0
	if length >= 0 && length < len(ampIni) {
		amplitude = ampIni[length]
	} else if length > 0 {
		amplitude = (5.0 / 3.0) * float64(length) / float64(length+1)
	}

	return DoublePerlinNoise{
		amplitude: amplitude,
		frequency: math.Pow(2, float64(omin)),
		octA:      octA,
		octB:      octB,
	}
}

func (d *DoublePerlinNoise) Sample(x, y, z float64) float64 {
	const factor = 337.0 / 331.0
	nx := x * d.frequency
	ny := y * d.frequency
	nz := z * d.frequency
	return (d.octA.Sample(nx, ny, nz) + d.octB.Sample(nx*factor, ny*factor, nz*factor)) * d.amplitude
}

type BlendedNoise struct {
	minLimit        OctaveNoise
	maxLimit        OctaveNoise
	main            OctaveNoise
	xzMultiplier    float64
	yMultiplier     float64
	xzFactor        float64
	yFactor         float64
	limitSmearScale float64
	mainSmearScale  float64
}

func NewBlendedNoise(rng *Xoroshiro128, xzScale, yScale, xzFactor, yFactor, smearScaleMultiplier float64) BlendedNoise {
	const baseScale = 684.412
	minLimit := createLegacyOctaves(rng, 16)
	maxLimit := createLegacyOctaves(rng, 16)
	main := createLegacyOctaves(rng, 8)
	limitSmearScale := baseScale * yScale * smearScaleMultiplier

	return BlendedNoise{
		minLimit:        minLimit,
		maxLimit:        maxLimit,
		main:            main,
		xzMultiplier:    baseScale * xzScale,
		yMultiplier:     baseScale * yScale,
		xzFactor:        xzFactor,
		yFactor:         yFactor,
		limitSmearScale: limitSmearScale,
		mainSmearScale:  limitSmearScale / yFactor,
	}
}

func createLegacyOctaves(rng *Xoroshiro128, count int) OctaveNoise {
	octaves := make([]PerlinNoise, 0, count)
	for i := 0; i < count; i++ {
		noise := NewPerlinNoise(rng)
		octaves = append(octaves, noise)
	}
	return OctaveNoise{octaves: octaves}
}

func (b *BlendedNoise) Sample(x, y, z float64) float64 {
	dx := x * b.xzMultiplier
	dy := y * b.yMultiplier
	dz := z * b.xzMultiplier

	gx := dx / b.xzFactor
	gy := dy / b.yFactor
	gz := dz / b.xzFactor

	n := 0.0
	o := 1.0
	for i := 0; i < 8; i++ {
		if i < len(b.main.octaves) {
			n += b.main.octaves[i].SampleSmeared(
				wrapCoord(gx*o),
				wrapCoord(gy*o),
				wrapCoord(gz*o),
				b.mainSmearScale*o,
				gy*o,
			) / o
		}
		o /= 2
	}

	q := (n/10.0 + 1.0) / 2.0
	useMaxOnly := q >= 1
	useMinOnly := q <= 0

	l := 0.0
	m := 0.0
	o = 1.0
	for i := 0; i < 16; i++ {
		if !useMaxOnly && i < len(b.minLimit.octaves) {
			l += b.minLimit.octaves[i].SampleSmeared(
				wrapCoord(dx*o),
				wrapCoord(dy*o),
				wrapCoord(dz*o),
				b.limitSmearScale*o,
				dy*o,
			) / o
		}
		if !useMinOnly && i < len(b.maxLimit.octaves) {
			m += b.maxLimit.octaves[i].SampleSmeared(
				wrapCoord(dx*o),
				wrapCoord(dy*o),
				wrapCoord(dz*o),
				b.limitSmearScale*o,
				dy*o,
			) / o
		}
		o /= 2
	}

	return clampedLerp(q, l/512.0, m/512.0) / 128.0
}

func wrapCoord(value float64) float64 {
	const coordRange = 33554432.0
	return value - math.Floor(value/coordRange)*coordRange
}

func clampedLerp(t, a, b float64) float64 {
	t = clampFloat(t, 0, 1)
	return a + t*(b-a)
}

func smoothstep(d float64) float64 {
	return d * d * d * (d*(d*6-15) + 10)
}

func indexedLerp(idx int32, x, y, z float64) float64 {
	u := y
	if idx < 8 {
		u = x
	}
	var v float64
	if idx < 4 {
		v = y
	} else if idx == 12 || idx == 14 {
		v = x
	} else {
		v = z
	}
	if idx&1 != 0 {
		u = -u
	}
	if idx&2 != 0 {
		v = -v
	}
	return u + v
}

func lerp(t, a, b float64) float64 {
	return a + t*(b-a)
}
