package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Bamboo and BambooSapling must satisfy item.BoneMealAffected: the interface
// is discovered through a runtime type assertion, so a signature mismatch
// would silently disable bone meal on bamboo rather than fail to compile.
var (
	_ item.BoneMealAffected = Bamboo{}
	_ item.BoneMealAffected = BambooSapling{}
	_ Flammable             = Bamboo{}
	_ Flammable             = BambooSapling{}
)

// TestBambooSoilFor checks that bamboo and bamboo saplings may be placed on
// every block that supports them in vanilla: dirt, coarse dirt, grass,
// podzol, mud, muddy mangrove roots, sand, red sand and gravel.
func TestBambooSoilFor(t *testing.T) {
	soils := []world.Block{
		Dirt{}, Dirt{Coarse: true}, Grass{}, Podzol{}, Mud{},
		MuddyMangroveRoots{}, Sand{}, Sand{Red: true}, Gravel{},
	}
	for _, soil := range soils {
		for _, plant := range []world.Block{Bamboo{}, BambooSapling{}} {
			if !soil.(Soil).SoilFor(plant) {
				t.Errorf("%#v should be soil for %#v", soil, plant)
			}
		}
	}
}

// TestBambooGrowthLayout checks the leaf sizes and stalk thickness of the top
// blocks of a bamboo stalk at each height, matching vanilla behaviour: small
// leaves for two and three block stalks, thick stalks with large leaves from
// four blocks up.
func TestBambooGrowthLayout(t *testing.T) {
	thin, thick := Bamboo{}, Bamboo{Thick: true}
	smallThin := Bamboo{LeafSize: BambooSizeSmallLeaves()}
	smallThick := Bamboo{Thick: true, LeafSize: BambooSizeSmallLeaves()}
	largeThick := Bamboo{Thick: true, LeafSize: BambooSizeLargeLeaves()}

	for _, c := range []struct {
		newHeight, amount int
		layout            []world.Block
	}{
		{2, 1, []world.Block{smallThin}},
		{3, 1, []world.Block{smallThin, smallThin}},
		{4, 1, []world.Block{largeThick, smallThick, thick, thick}},
		{5, 1, []world.Block{largeThick, largeThick, smallThick, thick}},
		{6, 2, []world.Block{largeThick, largeThick, smallThick, thick, thick}},
		{16, 1, []world.Block{largeThick, largeThick, smallThick, thick}},
	} {
		layout := thin.growthLayout(c.newHeight, c.amount)
		if len(layout) != len(c.layout) {
			t.Fatalf("height %v: expected %v blocks, got %v", c.newHeight, len(c.layout), len(layout))
		}
		for i, block := range layout {
			if block != c.layout[i] {
				t.Errorf("height %v, block %v: expected %#v, got %#v", c.newHeight, i, c.layout[i], block)
			}
		}
	}
}

// TestBambooMaxHeight checks that the maximum height of a bamboo stalk is
// always between 12 and 16 and deterministic for a given position.
func TestBambooMaxHeight(t *testing.T) {
	seen := map[int]bool{}
	for x := -100; x <= 100; x += 3 {
		for z := -100; z <= 100; z += 3 {
			pos := cube.Pos{x, 0, z}
			h := Bamboo{}.maxHeight(pos)
			if h < 12 || h > 16 {
				t.Fatalf("max height at %v out of range: %v", pos, h)
			}
			if h2 := (Bamboo{}).maxHeight(pos); h2 != h {
				t.Fatalf("max height at %v not deterministic: %v != %v", pos, h, h2)
			}
			seen[h] = true
		}
	}
	if len(seen) < 2 {
		t.Errorf("max height should vary by position, only saw %v", seen)
	}
}
