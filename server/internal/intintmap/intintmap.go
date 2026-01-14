// Package intintmap implements a fast int64 key -> int64 value map.
// The implementation is based on http://java-performance.info/implementing-world-fastest-java-int-to-int-hash-map/.
package intintmap

import "math"

// INT_PHI is for scrambling the keys.
const INT_PHI = 0x9E3779B9

// FREE_KEY is the 'free' key.
const FREE_KEY = 0

func phiMix(x int64) int64 {
	h := x * INT_PHI
	return h ^ (h >> 16)
}

// Map is a map-like data structure for int64s.
type Map struct {
	data       []int64 // interleaved keys and values
	fillFactor float64
	threshold  int // map resizes when reaching this size
	size       int

	mask  int64 // mask to calculate the original position
	mask2 int64

	hasFreeKey bool  // indicates if FREE_KEY is present
	freeVal    int64 // value of FREE_KEY
}

func nextPowerOf2(x uint32) uint32 {
	if x == math.MaxUint32 {
		return x
	}
	if x == 0 {
		return 1
	}
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return x + 1
}

func arraySize(exp int, fill float64) int {
	s := nextPowerOf2(uint32(math.Ceil(float64(exp) / fill)))
	if s < 2 {
		s = 2
	}
	return int(s)
}

// New returns a map initialised with size slots and the provided fillFactor.
func New(size int, fillFactor float64) *Map {
	if fillFactor <= 0 || fillFactor >= 1 {
		panic("fillFactor must be in (0, 1)")
	}
	if size <= 0 {
		panic("size must be positive")
	}
	capacity := arraySize(size, fillFactor)
	return &Map{
		data:       make([]int64, 2*capacity),
		fillFactor: fillFactor,
		threshold:  int(math.Floor(float64(capacity) * fillFactor)),
		mask:       int64(capacity - 1),
		mask2:      int64(2*capacity - 1),
	}
}

// Get returns the value if the key is found.
func (m *Map) Get(key int64) (int64, bool) {
	if key == FREE_KEY {
		if m.hasFreeKey {
			return m.freeVal, true
		}
		return 0, false
	}
	ptr := (phiMix(key) & m.mask) << 1
	if ptr < 0 || ptr >= int64(len(m.data)) {
		return 0, false
	}
	k := m.data[ptr]
	if k == FREE_KEY {
		return 0, false
	}
	if k == key {
		return m.data[ptr+1], true
	}
	for {
		ptr = (ptr + 2) & m.mask2
		k = m.data[ptr]
		if k == FREE_KEY {
			return 0, false
		}
		if k == key {
			return m.data[ptr+1], true
		}
	}
}

// Put adds or updates key with value val.
func (m *Map) Put(key, val int64) {
	if key == FREE_KEY {
		if !m.hasFreeKey {
			m.size++
		}
		m.hasFreeKey = true
		m.freeVal = val
		return
	}
	ptr := (phiMix(key) & m.mask) << 1
	k := m.data[ptr]
	switch k {
	case FREE_KEY:
		m.data[ptr] = key
		m.data[ptr+1] = val
		if m.size >= m.threshold {
			m.rehash()
		} else {
			m.size++
		}
		return
	case key:
		m.data[ptr+1] = val
		return
	}
	for {
		ptr = (ptr + 2) & m.mask2
		k = m.data[ptr]
		switch k {
		case FREE_KEY:
			m.data[ptr] = key
			m.data[ptr+1] = val
			if m.size >= m.threshold {
				m.rehash()
			} else {
				m.size++
			}
			return
		case key:
			m.data[ptr+1] = val
			return
		}
	}
}

// Del deletes a key and its value.
func (m *Map) Del(key int64) {
	if key == FREE_KEY {
		if m.hasFreeKey {
			m.hasFreeKey = false
			m.size--
		}
		return
	}
	ptr := (phiMix(key) & m.mask) << 1
	k := m.data[ptr]
	switch k {
	case key:
		m.shiftKeys(ptr)
		m.size--
		return
	case FREE_KEY:
		return
	}
	for {
		ptr = (ptr + 2) & m.mask2
		k = m.data[ptr]
		switch k {
		case key:
			m.shiftKeys(ptr)
			m.size--
			return
		case FREE_KEY:
			return
		}
	}
}

func (m *Map) shiftKeys(pos int64) int64 {
	for {
		last := pos
		pos = (pos + 2) & m.mask2
		for {
			k := m.data[pos]
			if k == FREE_KEY {
				m.data[last] = FREE_KEY
				return last
			}
			slot := (phiMix(k) & m.mask) << 1
			if last <= pos {
				if last >= slot || slot > pos {
					break
				}
			} else {
				if last >= slot && slot > pos {
					break
				}
			}
			pos = (pos + 2) & m.mask2
		}
		m.data[last] = m.data[pos]
		m.data[last+1] = m.data[pos+1]
	}
}

func (m *Map) rehash() {
	oldCapacity := len(m.data)
	newCapacity := oldCapacity << 1
	temp := make([]int64, newCapacity)
	copy(temp, m.data)
	m.data = make([]int64, newCapacity)
	m.mask = int64(newCapacity/2 - 1)
	m.mask2 = int64(newCapacity - 1)
	m.threshold <<= 1
	m.size = 0
	if m.hasFreeKey {
		m.size++
	}
	for i := 0; i < oldCapacity; i += 2 {
		if temp[i] != FREE_KEY {
			m.Put(temp[i], temp[i+1])
		}
	}
}
