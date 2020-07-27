package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"math/rand"
)

// Glowstone is commonly found on the ceiling of the nether dimension.
type Glowstone struct {
	noNBT
	solid
}

// BreakInfo ...
func (g Glowstone) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.3,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(item.GlowstoneDust{}, rand.Intn(3)+2)),
	}
}

// EncodeItem ...
func (g Glowstone) EncodeItem() (id int32, meta int16) {
	return 89, 0
}

// EncodeBlock ...
func (g Glowstone) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:glowstone", nil
}

// LightEmissionLevel returns 15.
func (Glowstone) LightEmissionLevel() uint8 {
	return 15
}

// Hash ...
func (Glowstone) Hash() uint64 {
	return hashGlowstone
}
