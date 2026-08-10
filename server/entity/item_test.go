package entity

import (
	"math"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// TestItemPickupDelayRoundTrip verifies that the delay before a dropped item may be picked up survives a save and
// load. The tag holds ticks, so both sides have to divide by the length of a tick.
func TestItemPickupDelayRoundTrip(t *testing.T) {
	for _, ticks := range []int{10, 40, 200, math.MaxInt16} {
		conf := itemConf
		conf.Item = item.NewStack(item.Diamond{}, 1)
		conf.PickupDelay = time.Duration(ticks) * (time.Second / 20)

		var data world.EntityData
		conf.Apply(&data)

		m := ItemType.EncodeNBT(&data)
		var decoded world.EntityData
		ItemType.DecodeNBT(m, &decoded)

		before := data.Data.(*ItemBehaviour).pickupDelay
		after := decoded.Data.(*ItemBehaviour).pickupDelay
		if before != after {
			t.Errorf("pickup delay of %v ticks not preserved across an NBT round trip: %v -> %v (encoded PickupDelay tag = %v, expected %v)",
				ticks, before, after, m["PickupDelay"], ticks)
		}
	}
}

// TestTNTFuseRoundTrip verifies that the fuse of primed TNT survives a save and load. The tag it is written to has to
// hold the whole fuse: NewTNT accepts any duration, and a fuse longer than the tag can express detonates early.
func TestTNTFuseRoundTrip(t *testing.T) {
	for _, fuse := range []time.Duration{time.Second * 4, time.Second * 20} {
		conf := tntConf
		conf.ExistenceDuration = fuse

		var data world.EntityData
		conf.Apply(&data)
		m := TNTType.EncodeNBT(&data)
		if got, want := m["Fuse"], int16(fuse/(time.Second/20)); got != want {
			t.Fatalf("encoded Fuse = %v (%T), want %v (%T) ticks", got, got, want, want)
		}

		var decoded world.EntityData
		TNTType.DecodeNBT(m, &decoded)
		if got := decoded.Data.(*PassiveBehaviour).Fuse(); got != fuse {
			t.Errorf("Fuse = %v, want %v", got, fuse)
		}
	}
}

// TestAreaEffectCloudRoundTrip verifies that the durations of an area effect cloud survive a save and load. They are
// written to int32 tags, which cannot hold a duration in nanoseconds, and each has to be read back from its own tag.
func TestAreaEffectCloudRoundTrip(t *testing.T) {
	conf := areaEffectCloudConf
	conf.Duration = time.Second * 30
	conf.DurationUseGrowth = time.Second * 2
	conf.ReapplicationDelay = time.Second

	var data world.EntityData
	conf.Apply(&data)

	var decoded world.EntityData
	AreaEffectCloudType.DecodeNBT(AreaEffectCloudType.EncodeNBT(&data), &decoded)

	before, after := data.Data.(*AreaEffectCloudBehaviour), decoded.Data.(*AreaEffectCloudBehaviour)
	if before.duration != after.duration {
		t.Errorf("Duration = %v, want %v", after.duration, before.duration)
	}
	if got := after.duration; got < 0 {
		t.Errorf("Duration = %v, want a duration that is not negative", got)
	}
}
