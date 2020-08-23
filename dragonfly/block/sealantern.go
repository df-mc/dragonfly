package block

import (
	"math/rand"

	"github.com/df-mc/dragonfly/dragonfly/item"
)

// Sealantern is commonly found on Ocean Monuments as Lightsource.
type Sealantern struct {
	noNBT
	solid
}

// BreakInfo ...
func (s Sealantern) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.3,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(item.PrismarineCrystals{}, rand.Intn(3)+2)),
	}
}

// EncodeItem ...
func (s Sealantern) EncodeItem() (id int32, meta int16) {
	return 169, 0
}

// EncodeBlock ...
func (s Sealantern) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:seaLantern", nil
}

// LightEmissionLevel returns 15.
func (Sealantern) LightEmissionLevel() uint8 {
	return 15
}

// Hash ...
func (Sealantern) Hash() uint64 {
	return hashSealantern
}
