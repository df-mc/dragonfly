package block

import (
	"image/color"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/item"
)

// TestSignNBTRoundTrip checks that a fully populated Sign survives an
// EncodeNBT -> DecodeNBT round trip, which is what happens when a world is
// saved and loaded again.
func TestSignNBTRoundTrip(t *testing.T) {
	s := Sign{
		Wood:  OakWood(),
		Waxed: true,
		Front: SignText{
			Text:       "front text",
			BaseColour: color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff},
			Glowing:    true,
			Owner:      "1234567890",
		},
		Back: SignText{
			Text:       "back text",
			BaseColour: color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff},
			Glowing:    true,
			Owner:      "0987654321",
		},
	}

	got := Sign{Wood: OakWood()}.DecodeNBT(s.EncodeNBT()).(Sign)

	if got.Front.BaseColour != s.Front.BaseColour {
		t.Errorf("Front.BaseColour = %#v, want %#v", got.Front.BaseColour, s.Front.BaseColour)
	}
	if got.Front.Glowing != s.Front.Glowing {
		t.Errorf("Front.Glowing = %v, want %v", got.Front.Glowing, s.Front.Glowing)
	}
	if got.Front.Owner != s.Front.Owner {
		t.Errorf("Front.Owner = %q, want %q", got.Front.Owner, s.Front.Owner)
	}
	if got.Back.BaseColour != s.Back.BaseColour {
		t.Errorf("Back.BaseColour = %#v, want %#v", got.Back.BaseColour, s.Back.BaseColour)
	}
	if got.Back.Glowing != s.Back.Glowing {
		t.Errorf("Back.Glowing = %v, want %v", got.Back.Glowing, s.Back.Glowing)
	}
	if got.Back.Owner != s.Back.Owner {
		t.Errorf("Back.Owner = %q, want %q", got.Back.Owner, s.Back.Owner)
	}
	if got.Waxed != s.Waxed {
		t.Errorf("Waxed = %v, want %v", got.Waxed, s.Waxed)
	}
	if got.Front.Text != s.Front.Text {
		t.Errorf("Front.Text = %q, want %q", got.Front.Text, s.Front.Text)
	}
}

// TestCampfireNBTRoundTrip checks that the cook timers of items on a
// campfire survive a save/load round trip.
func TestCampfireNBTRoundTrip(t *testing.T) {
	c := Campfire{Type: NormalFire()}
	c.Items[0] = CampfireItem{Item: item.NewStack(item.Porkchop{}, 1), Time: time.Second * 20}

	got := Campfire{Type: NormalFire()}.DecodeNBT(c.EncodeNBT()).(Campfire)

	if got.Items[0].Item.Empty() {
		t.Fatalf("Items[0].Item is empty, want raw porkchop")
	}
	if got.Items[0].Time != c.Items[0].Time {
		t.Errorf("Items[0].Time = %v, want %v", got.Items[0].Time, c.Items[0].Time)
	}

	// A campfire cooks for 30 seconds, which is 600 ticks. The tag holding that is an int in the format, and the
	// value has to survive being written as one.
	full := Campfire{Type: NormalFire()}
	full.Items[0] = CampfireItem{Item: item.NewStack(item.Porkchop{}, 1), Time: time.Second * 30}
	if got := full.EncodeNBT()["ItemTime1"]; got != any(int32(600)) {
		t.Errorf("encoded ItemTime1 = %#v (%T), want int32(600)", got, got)
	}
}
