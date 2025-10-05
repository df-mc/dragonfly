package block

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// CopperChest is a container block which may be used to store items. Chests may also be paired to create a bigger
// single container.
// The empty value of CopperChest is not valid. It must be created using block.NewCopperChest().
type CopperChest struct {
	baseChest
	chest
	transparent
	bass
	sourceWaterDisplacer

	// Facing is the direction that the chest is facing.
	Facing cube.Direction
	// CustomName is the custom name of the chest. This name is displayed when the chest is opened, and may
	// include colour codes.
	CustomName string
	// Oxidation is the level of oxidation of the copper chest.
	Oxidation OxidationType
	// Waxed bool is whether the copper chest has been waxed with honeycomb.
	Waxed bool
}

// NewCopperChest creates a new initialised chest. The inventory is properly initialised.
func NewCopperChest() CopperChest {
	return CopperChest{
		baseChest: newBaseChest(),
	}
}

// Inventory returns the inventory of the chest. The size of the inventory will be 27 or 54, depending on
// whether the chest is single or double.
func (c CopperChest) Inventory(tx *world.Tx, pos cube.Pos) *inventory.Inventory {
	inv, _ := c.tryPair(tx, pos)
	return inv
}

// tryPair attempts to pair the inventories of this chest with a potential paired chest next to it.
func (c CopperChest) tryPair(tx *world.Tx, pos cube.Pos) (*inventory.Inventory, bool) {
	if c.paired {
		if c.pairInv == nil {
			if ch, pair, ok := c.pair(tx, pos, c.pairPos(pos)); ok {
				tx.SetBlock(pos, ch, nil)
				tx.SetBlock(c.pairPos(pos), pair, nil)
				return ch.pairInv, true
			}
			c.paired = false
			tx.SetBlock(pos, c, nil)
			return c.inventory, true
		}
		return c.pairInv, false
	}
	return c.inventory, false
}

// WithName returns the chest after applying a specific name to the block.
func (c CopperChest) WithName(a ...any) world.Item {
	c.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return c
}

// SideClosed ...
func (CopperChest) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// AddViewer adds a viewer to the chest, so that it is updated whenever the inventory of the chest is changed.
func (c CopperChest) AddViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	if _, changed := c.tryPair(tx, pos); changed {
		c = tx.Block(pos).(CopperChest)
	}
	c.baseChest.addViewer(v, tx, pos)
}

// RemoveViewer removes a viewer from the chest, so that slot updates in the inventory are no longer sent to it.
func (c CopperChest) RemoveViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	if _, changed := c.tryPair(tx, pos); changed {
		c = tx.Block(pos).(CopperChest)
	}
	c.baseChest.removeViewer(v, tx, pos)
}

// Activate ...
func (c CopperChest) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		if c.paired {
			if d, ok := tx.Block(c.pairPos(pos).Side(cube.FaceUp)).(LightDiffuser); !ok || d.LightDiffusionLevel() > 2 {
				return false
			}
		}
		if d, ok := tx.Block(pos.Side(cube.FaceUp)).(LightDiffuser); ok && d.LightDiffusionLevel() <= 2 {
			opener.OpenBlockContainer(pos, tx)
		}
		return true
	}
	return false
}

// UseOnBlock ...
func (c CopperChest) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, c)
	if !used {
		return
	}

	oxidation := c.Oxidation
	waxed := c.Waxed

	//noinspection GoAssignmentToReceiver
	c = NewCopperChest()
	c.Facing = user.Rotation().Direction().Opposite()
	c.Oxidation = oxidation
	c.Waxed = waxed

	// Check both sides of the chest to see if it is possible to pair with another chest.
	for _, dir := range []cube.Direction{c.Facing.RotateLeft(), c.Facing.RotateRight()} {
		if ch, pair, ok := c.pair(tx, pos, pos.Side(dir.Face())); ok {
			place(tx, pos, ch, user, ctx)
			tx.SetBlock(ch.pairPos(pos), pair, nil)
			return placed(ctx)
		}
	}

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (c CopperChest) BreakInfo() BreakInfo {
	return newBreakInfo(3, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierStone.HarvestLevel
	}, pickaxeEffective, oneOf(c)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		if c.paired {
			pairPos := c.pairPos(pos)
			if _, pair, ok := c.unpair(tx, pos); ok {
				c.paired = false
				tx.SetBlock(pairPos, pair, nil)
			}
		}

		for _, i := range c.Inventory(tx, pos).Clear() {
			dropItem(tx, i, pos.Vec3Centre())
		}
	})
}

// Paired returns whether the chest is paired with another chest.
func (c CopperChest) Paired() bool {
	return c.paired
}

// pair pairs this chest with the given chest position.
func (c CopperChest) pair(tx *world.Tx, pos, pairPos cube.Pos) (ch, pair CopperChest, ok bool) {
	pair, ok = tx.Block(pairPos).(CopperChest)
	if !ok || c.Facing != pair.Facing || pair.paired && (pair.pairX != pos[0] || pair.pairZ != pos[2]) {
		return c, pair, false
	}

	left, right, double, mu, viewers := mergeInventories(c.inventory, pair.inventory, pos, pairPos, c.Facing)

	c.inventory, pair.inventory = left, right
	if pos.Side(c.Facing.RotateRight().Face()) == pairPos {
		c.inventory, pair.inventory = right, left
	}
	c.pairX, c.pairZ, c.paired = pairPos[0], pairPos[2], true
	pair.pairX, pair.pairZ, pair.paired = pos[0], pos[2], true
	c.viewerMu, pair.viewerMu = mu, mu
	c.viewers, pair.viewers = viewers, viewers
	c.pairInv, pair.pairInv = double, double
	return c, pair, true
}

// unpair unpairs this chest from the chest it is currently paired with.
func (c CopperChest) unpair(tx *world.Tx, pos cube.Pos) (ch, pair CopperChest, ok bool) {
	if !c.paired {
		return c, CopperChest{}, false
	}

	pair, ok = tx.Block(c.pairPos(pos)).(CopperChest)
	if !ok || c.Facing != pair.Facing || pair.paired && (pair.pairX != pos[0] || pair.pairZ != pos[2]) {
		return c, pair, false
	}

	unpairChests(&c.baseChest, tx, pos)
	unpairChests(&pair.baseChest, tx, pos)
	return c, pair, true
}

// Wax waxes the copper chest to stop it from oxidising further.
func (c CopperChest) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if c.Waxed {
		return c, false
	}
	c.Waxed = true
	return c, true
}

// Strip ...
func (c CopperChest) Strip() (world.Block, world.Sound, bool) {
	if c.Waxed {
		c.Waxed = false
		return c, sound.WaxRemoved{}, true
	} else if ot, ok := c.Oxidation.Decrease(); ok {
		c.Oxidation = ot
		return c, sound.CopperScraped{}, true
	}
	return c, nil, false
}

// CanOxidate ...
func (c CopperChest) CanOxidate() bool {
	return !c.Waxed
}

// OxidationLevel ...
func (c CopperChest) OxidationLevel() OxidationType {
	return c.Oxidation
}

// WithOxidationLevel ...
func (c CopperChest) WithOxidationLevel(o OxidationType) Oxidisable {
	c.Oxidation = o
	return c
}

// RandomTick ...
func (c CopperChest) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	attemptOxidation(pos, tx, r, c)
}

// DecodeNBT ...
func (c CopperChest) DecodeNBT(data map[string]any) any {
	facing := c.Facing
	oxidation := c.Oxidation
	waxed := c.Waxed

	//noinspection GoAssignmentToReceiver
	c = NewCopperChest()
	c.Facing = facing
	c.CustomName = nbtconv.String(data, "CustomName")
	c.Oxidation = oxidation
	c.Waxed = waxed

	pairX, ok := data["pairx"]
	pairZ, ok2 := data["pairz"]
	if ok && ok2 {
		pairX, ok := pairX.(int32)
		pairZ, ok2 := pairZ.(int32)
		if ok && ok2 {
			c.paired = true
			c.pairX, c.pairZ = int(pairX), int(pairZ)
		}
	}

	nbtconv.InvFromNBT(c.inventory, nbtconv.Slice(data, "Items"))
	return c
}

// EncodeNBT ...
func (c CopperChest) EncodeNBT() map[string]any {
	if c.inventory == nil {
		facing, customName, oxidation, waxed := c.Facing, c.CustomName, c.Oxidation, c.Waxed

		//noinspection GoAssignmentToReceiver
		c = NewCopperChest()
		c.Facing, c.CustomName, c.Oxidation, c.Waxed = facing, customName, oxidation, waxed
	}
	m := map[string]any{
		"Items": nbtconv.InvToNBT(c.inventory),
		"id":    "CopperChest",
	}
	if c.CustomName != "" {
		m["CustomName"] = c.CustomName
	}

	if c.paired {
		m["pairx"] = int32(c.pairX)
		m["pairz"] = int32(c.pairZ)
	}
	return m
}

// EncodeItem ...
func (c CopperChest) EncodeItem() (name string, meta int16) {
	return copperBlockName("copper_chest", c.Oxidation, c.Waxed), 0
}

// EncodeBlock ...
func (c CopperChest) EncodeBlock() (name string, properties map[string]any) {
	return copperBlockName("copper_chest", c.Oxidation, c.Waxed), map[string]any{"minecraft:cardinal_direction": c.Facing.String()}
}

// allCopperChests ...
func allCopperChests() (chests []world.Block) {
	f := func(waxed bool) {
		for _, o := range OxidationTypes() {
			for _, direction := range cube.Directions() {
				chests = append(chests, CopperChest{Facing: direction, Oxidation: o, Waxed: waxed})
			}
		}
	}
	f(true)
	f(false)
	return
}
