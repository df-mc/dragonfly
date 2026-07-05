package block

import (
	"image/color"
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

var redstoneOreParticleColour = color.RGBA{R: 255, A: 255}

// RedstoneOre is an ore block that lights up when disturbed.
type RedstoneOre struct {
	solid
	bassDrum

	// Type is the type of redstone ore.
	Type OreType
	// Lit is whether the redstone ore is glowing.
	Lit bool
}

// Activate lights the redstone ore unless the user is sneaking.
// It returns false so the held item may still be used.
func (r RedstoneOre) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if sneaking, ok := u.(interface{ Sneaking() bool }); ok && sneaking.Sneaking() {
		return false
	}
	r.light(pos, tx)
	return false
}

// Punch lights the redstone ore when mining starts.
func (r RedstoneOre) Punch(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User) {
	r.light(pos, tx)
}

// EntityStepOn lights the redstone ore when a non-sneaking entity steps on it.
func (r RedstoneOre) EntityStepOn(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if sneaking, ok := e.(interface{ Sneaking() bool }); ok && sneaking.Sneaking() {
		return
	}
	r.light(pos, tx)
}

// RandomTick turns lit redstone ore off again.
func (r RedstoneOre) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	r.fade(pos, tx)
}

// light emits particles and switches the ore to its lit state if it is not already lit.
func (r RedstoneOre) light(pos cube.Pos, tx *world.Tx) {
	r.emitParticles(pos, tx)
	if r.Lit {
		return
	}
	r.Lit = true
	tx.SetBlock(pos, r, nil)
}

// emitParticles emits the red particles produced when redstone ore is disturbed.
func (RedstoneOre) emitParticles(pos cube.Pos, tx *world.Tx) {
	for range 5 {
		tx.AddParticle(pos.Vec3().Add(redstoneOreParticleOffset()), particle.Dust{Colour: redstoneOreParticleColour})
	}
}

// redstoneOreParticleOffset returns a random offset on or just outside one face of the block.
func redstoneOreParticleOffset() mgl64.Vec3 {
	offset := mgl64.Vec3{rand.Float64(), rand.Float64(), rand.Float64()}
	const edge = 0.05
	switch rand.IntN(6) {
	case 0:
		offset[0] = -edge
	case 1:
		offset[0] = 1 + edge
	case 2:
		offset[1] = -edge
	case 3:
		offset[1] = 1 + edge
	case 4:
		offset[2] = -edge
	case 5:
		offset[2] = 1 + edge
	}
	return offset
}

// fade returns lit redstone ore to its unlit state.
func (r RedstoneOre) fade(pos cube.Pos, tx *world.Tx) {
	if !r.Lit {
		return
	}
	r.Lit = false
	tx.SetBlock(pos, r, nil)
}

// LightEmissionLevel ...
func (r RedstoneOre) LightEmissionLevel() uint8 {
	if r.Lit {
		return 9
	}
	return 0
}

// BreakInfo ...
func (r RedstoneOre) BreakInfo() BreakInfo {
	return newBreakInfo(r.Type.Hardness(), func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, discreteDrops(RedstoneWire{}, RedstoneOre{Type: r.Type}, 4, 5, 8)).withXPDropRange(1, 5).withBlastResistance(15)
}

// SmeltInfo ...
func (RedstoneOre) SmeltInfo() item.SmeltInfo {
	return newOreSmeltInfo(item.NewStack(RedstoneWire{}, 1), 0.7)
}

// EncodeItem ...
func (r RedstoneOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + r.Type.Prefix() + "redstone_ore", 0
}

// EncodeBlock ...
func (r RedstoneOre) EncodeBlock() (string, map[string]any) {
	lit := ""
	if r.Lit {
		lit = "lit_"
	}
	return "minecraft:" + lit + r.Type.Prefix() + "redstone_ore", nil
}
