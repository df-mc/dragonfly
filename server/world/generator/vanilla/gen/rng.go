package gen

type Xoroshiro128 struct {
	low  uint64
	high uint64
}

type PositionalRandomFactory struct {
	seedLo int64
	seedHi int64
}

const (
	aquiferHashLo int64 = 8913007134489619686
	aquiferHashHi int64 = 854934872360429201
	oreHashLo     int64 = -7239516275217526419
	oreHashHi     int64 = 3091299299654006625
)

func NewXoroshiro128FromSeed(seed int64) Xoroshiro128 {
	const xl uint64 = 0x9e3779b97f4a7c15
	const xh uint64 = 0x6a09e667f3bcc909
	const a uint64 = 0xbf58476d1ce4e5b9
	const b uint64 = 0x94d049bb133111eb

	s := uint64(seed)
	l := s ^ xh
	h := l + xl

	l = (l ^ (l >> 30)) * a
	h = (h ^ (h >> 30)) * a
	l = (l ^ (l >> 27)) * b
	h = (h ^ (h >> 27)) * b
	l ^= l >> 31
	h ^= h >> 31

	return Xoroshiro128{low: l, high: h}
}

func NewXoroshiro128FromState(low, high uint64) Xoroshiro128 {
	return Xoroshiro128{low: low, high: high}
}

func NewPositionalRandomFactory(seed int64) PositionalRandomFactory {
	rng := NewXoroshiro128FromSeed(seed)
	return PositionalRandomFactory{
		seedLo: int64(rng.NextLong()),
		seedHi: int64(rng.NextLong()),
	}
}

func NewPositionalRandomFactoryFromSeeds(seedLo, seedHi int64) PositionalRandomFactory {
	return PositionalRandomFactory{seedLo: seedLo, seedHi: seedHi}
}

func (f PositionalRandomFactory) At(x, y, z int) Xoroshiro128 {
	seed := positionalSeed(x, y, z)
	return NewXoroshiro128FromState(uint64(seed^f.seedLo), uint64(f.seedHi))
}

func (f PositionalRandomFactory) ForkAquiferRandom() PositionalRandomFactory {
	rng := f.fromHashOf(aquiferHashLo, aquiferHashHi)
	return PositionalRandomFactory{
		seedLo: int64(rng.NextLong()),
		seedHi: int64(rng.NextLong()),
	}
}

func (f PositionalRandomFactory) ForkOreRandom() PositionalRandomFactory {
	rng := f.fromHashOf(oreHashLo, oreHashHi)
	return PositionalRandomFactory{
		seedLo: int64(rng.NextLong()),
		seedHi: int64(rng.NextLong()),
	}
}

func (f PositionalRandomFactory) fromHashOf(hashLo, hashHi int64) Xoroshiro128 {
	return NewXoroshiro128FromState(uint64(hashLo^f.seedLo), uint64(hashHi^f.seedHi))
}

func (x *Xoroshiro128) NextLong() uint64 {
	l := x.low
	h := x.high
	n := bitsRotateLeft64(l+h, 17) + l
	xor := h ^ l

	x.low = bitsRotateLeft64(l, 49) ^ xor ^ (xor << 21)
	x.high = bitsRotateLeft64(xor, 28)
	return n
}

func (x *Xoroshiro128) NextInt(bound uint32) uint32 {
	r := uint64(uint32(x.NextLong())) * uint64(bound)
	if uint32(r) < bound {
		threshold := (^bound + 1) % bound
		for uint32(r) < threshold {
			r = uint64(uint32(x.NextLong())) * uint64(bound)
		}
	}
	return uint32(r >> 32)
}

func (x *Xoroshiro128) NextDouble() float64 {
	return float64(x.NextLong()>>11) * 1.1102230246251565e-16
}

func bitsRotateLeft64(v uint64, k int) uint64 {
	return (v << k) | (v >> (64 - k))
}

func positionalSeed(x, y, z int) int64 {
	l := int64(x)*3129871 ^ int64(z)*116129781 ^ int64(y)
	l = l*l*42317861 + l*11
	return l >> 16
}
