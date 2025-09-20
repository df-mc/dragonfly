package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"math/rand/v2"
)

// SeaLantern is an underwater light sources that appear in ocean monuments and underwater ruins.
type SeaLantern struct {
	transparent
	solid
	clicksAndSticks
}

func (SeaLantern) LightEmissionLevel() uint8 {
	return 15
}

func (s SeaLantern) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, silkTouchDrop(item.NewStack(item.PrismarineCrystals{}, rand.IntN(2)+2), item.NewStack(s, 1)))
}

func (SeaLantern) EncodeItem() (name string, meta int16) {
	return "minecraft:sea_lantern", 0
}

func (SeaLantern) EncodeBlock() (string, map[string]any) {
	return "minecraft:sea_lantern", nil
}
