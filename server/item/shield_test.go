package item

import (
	"reflect"
	"testing"
)

func TestShieldBannerNBTRoundTrip(t *testing.T) {
	want := Shield{Banner: &ShieldBanner{
		BaseColour: ColourRed(),
		Patterns: []ShieldPattern{
			{Pattern: "cre", Colour: ColourBlack()},
			{Pattern: "flo", Colour: ColourWhite()},
		},
		Illager: true,
	}}
	got := Shield{}.DecodeNBT(want.EncodeNBT()).(Shield)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("shield banner round trip = %#v, want %#v", got, want)
	}
}

func TestPlainShieldHasNoBannerNBT(t *testing.T) {
	if got := (Shield{}).EncodeNBT(); got != nil {
		t.Fatalf("plain shield NBT = %#v, want nil", got)
	}
	if got := (Shield{}).DecodeNBT(map[string]any{}).(Shield); got.Banner != nil {
		t.Fatalf("plain decoded shield banner = %#v, want nil", got.Banner)
	}
}

func TestShieldIsHandEquipped(t *testing.T) {
	if !(Shield{}).HandEquipped() {
		t.Fatal("shield must use its equipped hand transform")
	}
}
