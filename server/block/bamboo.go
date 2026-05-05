package block

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Bamboo is a versatile, fast-growing plant found primarily in jungles.
type Bamboo struct {
	transparent
	bass

	Ready    bool
	Thick    bool
	LeafSize BambooLeafSize
}

// FuelInfo ...
func (b Bamboo) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Millisecond * 2500)
}

// EncodeItem ...
func (b Bamboo) EncodeItem() (name string, meta int16) {
	return "minecraft:bamboo", 0
}

// BoneMeal ...
func (b Bamboo) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	top := b.top(pos, tx)
	return tx.Block(top).(Bamboo).grow(top, rand.IntN(2)+1, b.maxHeight(top), tx)
}

// BreakInfo ...
func (b Bamboo) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, axeEffective, oneOf(b))
}

// EncodeBlock ...
func (b Bamboo) EncodeBlock() (string, map[string]any) {
	thickness := "thin"
	if b.Thick {
		thickness = "thick"
	}
	return "minecraft:bamboo", map[string]any{
		"age_bit":                boolByte(b.Ready),
		"bamboo_leaf_size":       b.LeafSize.String(),
		"bamboo_stalk_thickness": thickness,
	}
}

// Model ...
func (b Bamboo) Model() world.BlockModel {
	return model.Bamboo{Thick: b.Thick}
}

// RandomTick ...
func (b Bamboo) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if b.Ready {
		if tx.Light(pos) < 9 || !b.grow(pos, 1, b.maxHeight(pos), tx) {
			b.Ready = false
			tx.SetBlock(pos, b, nil)
		}
	} else if replaceableWith(tx, pos.Side(cube.FaceUp), b) {
		b.Ready = true
		tx.SetBlock(pos, b, nil)
	}
}

// NeighbourUpdateTick ...
func (b Bamboo) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	down := tx.Block(pos.Side(cube.FaceDown))
	switch down.(type) {
	case BambooSapling, Bamboo:
		return
	}
	if supportsVegetation(b, down) {
		return
	}
	breakBlock(b, pos, tx)
}

// UseOnBlock ...
func (b Bamboo) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	if face == cube.FaceUp {
		switch x := tx.Block(pos).(type) {
		case Bamboo:
			top := x.top(pos, tx)
			return b.grow(top, 1, math.MaxInt, tx)
		case BambooSapling:
			return x.grow(pos, tx)
		default:
		}
	}

	pos, _, used := firstReplaceable(tx, pos, face, b)
	if !used {
		return false
	}
	s := BambooSapling{}
	if !supportsVegetation(s, tx.Block(pos.Sub(cube.Pos{0, 1}))) {
		return false
	}
	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// maxHeight ...
func (b Bamboo) maxHeight(pos cube.Pos) int {
	// TODO: The RNG algorithm does not match vanilla's.
	return 12 + int(rand.NewPCG(uint64(pos.X()), uint64(pos.Z())).Uint64()%5)
}

// top ...
func (b Bamboo) top(pos cube.Pos, tx *world.Tx) (top cube.Pos) {
	top = pos
	for {
		up := top.Side(cube.FaceUp)
		if _, ok := tx.Block(up).(Bamboo); !ok {
			return top
		}
		top = up
	}
}

// grow ...
func (b Bamboo) grow(pos cube.Pos, amount int, maxHeight int, tx *world.Tx) bool {
	if !replaceableWith(tx, pos.Side(cube.FaceUp), b) {
		return false
	}

	height := 1
	for {
		if _, ok := tx.Block(pos.Sub(cube.Pos{0, height})).(Bamboo); !ok {
			break
		}
		height++
		if height >= maxHeight {
			return false
		}
	}

	newHeight := height + amount
	stemBlock := Bamboo{Thick: b.Thick}
	if newHeight >= 4 && !stemBlock.Thick {
		stemBlock.Thick = true
	}
	smallLeavesBlock := Bamboo{Thick: stemBlock.Thick, LeafSize: BambooSizeSmallLeaves()}
	bigLeavesBlock := Bamboo{Thick: stemBlock.Thick, LeafSize: BambooSizeLargeLeaves()}

	var newBlocks []world.Block
	switch {
	case newHeight == 2:
		newBlocks = []world.Block{smallLeavesBlock}
	case newHeight == 3:
		newBlocks = []world.Block{smallLeavesBlock, smallLeavesBlock}
	case newHeight == 4:
		newBlocks = []world.Block{bigLeavesBlock, smallLeavesBlock, stemBlock, stemBlock}
	case newHeight > 4:
		newBlocks = []world.Block{bigLeavesBlock, bigLeavesBlock, smallLeavesBlock}
		for i, mx := 0, min(amount, newHeight-len(newBlocks)); i < mx; i++ {
			newBlocks = append(newBlocks, stemBlock)
		}
	}

	for i, b := range newBlocks {
		tx.SetBlock(pos.Sub(cube.Pos{0, i - amount}), b, nil)
	}

	return true
}

// allBamboos ...
func allBamboos() (bamboos []world.Block) {
	for _, thick := range []bool{false, true} {
		for _, ready := range []bool{false, true} {
			for _, leafSize := range BambooLeafSizes() {
				bamboos = append(bamboos, Bamboo{Thick: thick, Ready: ready, LeafSize: leafSize})
			}
		}
	}
	return
}
