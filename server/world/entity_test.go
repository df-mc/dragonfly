package world

import (
	"testing"
	"time"

	"github.com/go-gl/mathgl/mgl64"
)

// TestEntityAgeRoundTrip verifies that the age of an entity survives being written to NBT and read back. The tag
// holds ticks, and the age drives the check that despawns dropped items, so losing it stops them from ever expiring.
func TestEntityAgeRoundTrip(t *testing.T) {
	for _, age := range []time.Duration{time.Second, time.Second * 30, time.Minute * 4} {
		h := EntitySpawnOpts{Position: mgl64.Vec3{1, 2, 3}}.New(testEntityType{}, testEntityConfig{})
		h.data.Age, h.data.FireDuration = age, age

		m := h.encodeNBT()
		got := EntitySpawnOpts{}.New(testEntityType{}, testEntityConfig{})
		got.decodeNBT(m)

		// The fire duration beside it uses the right conversion, and is the control for this test.
		if got.data.FireDuration != age {
			t.Fatalf("FireDuration = %v, want %v (tag %v)", got.data.FireDuration, age, m["Fire"])
		}
		if got.data.Age != age {
			t.Errorf("Age = %v, want %v (encoded tag %v, want %v ticks)", got.data.Age, age, m["Age"], int64(age/(time.Second/20)))
		}
	}
}
