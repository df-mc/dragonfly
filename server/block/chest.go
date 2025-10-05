package block

import (
	"fmt"
	"strings"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Chest is a container block which may be used to store items. Chests may also be paired to create a bigger
// single container.
// The empty value of Chest is not valid. It must be created using block.NewChest().
type Chest struct {
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
}

// NewChest creates a new initialised chest. The inventory is properly initialised.
func NewChest() Chest {
	return Chest{
		baseChest: newBaseChest(),
	}
}

// Inventory returns the inventory of the chest. The size of the inventory will be 27 or 54, depending on
// whether the chest is single or double.
func (c Chest) Inventory(tx *world.Tx, pos cube.Pos) *inventory.Inventory {
	inv, _ := c.tryPair(tx, pos)
	return inv
}

// tryPair attempts to pair the inventories of this chest with a potential paired chest next to it.
func (c Chest) tryPair(tx *world.Tx, pos cube.Pos) (*inventory.Inventory, bool) {
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
func (c Chest) WithName(a ...any) world.Item {
	c.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return c
}

// SideClosed ...
func (Chest) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// AddViewer adds a viewer to the chest, so that it is updated whenever the inventory of the chest is changed.
func (c Chest) AddViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	if _, changed := c.tryPair(tx, pos); changed {
		c = tx.Block(pos).(Chest)
	}
	c.baseChest.addViewer(v, tx, pos)
}

// RemoveViewer removes a viewer from the chest, so that slot updates in the inventory are no longer sent to it.
func (c Chest) RemoveViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	if _, changed := c.tryPair(tx, pos); changed {
		c = tx.Block(pos).(Chest)
	}
	c.baseChest.removeViewer(v, tx, pos)
}

// Activate ...
func (c Chest) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
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
func (c Chest) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, c)
	if !used {
		return
	}
	c = NewChest()
	c.Facing = user.Rotation().Direction().Opposite()

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
func (c Chest) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, oneOf(c)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
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

// FuelInfo ...
func (Chest) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// FlammabilityInfo ...
func (c Chest) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, true)
}

// Paired returns whether the chest is paired with another chest.
func (c Chest) Paired() bool {
	return c.paired
}

// pair pairs this chest with the given chest position.
func (c Chest) pair(tx *world.Tx, pos, pairPos cube.Pos) (ch, pair Chest, ok bool) {
	pair, ok = tx.Block(pairPos).(Chest)
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
func (c Chest) unpair(tx *world.Tx, pos cube.Pos) (ch, pair Chest, ok bool) {
	if !c.paired {
		return c, Chest{}, false
	}

	pair, ok = tx.Block(c.pairPos(pos)).(Chest)
	if !ok || c.Facing != pair.Facing || pair.paired && (pair.pairX != pos[0] || pair.pairZ != pos[2]) {
		return c, pair, false
	}

	unpairChests(&c.baseChest, tx, pos)
	unpairChests(&pair.baseChest, tx, pos)
	return c, pair, true
}

// DecodeNBT ...
func (c Chest) DecodeNBT(data map[string]any) any {
	facing := c.Facing
	c = NewChest()
	c.Facing = facing
	c.CustomName = nbtconv.String(data, "CustomName")

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
func (c Chest) EncodeNBT() map[string]any {
	if c.inventory == nil {
		facing, customName := c.Facing, c.CustomName
		c = NewChest()
		c.Facing, c.CustomName = facing, customName
	}
	m := map[string]any{
		"Items": nbtconv.InvToNBT(c.inventory),
		"id":    "Chest",
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
func (Chest) EncodeItem() (name string, meta int16) {
	return "minecraft:chest", 0
}

// EncodeBlock ...
func (c Chest) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:chest", map[string]any{"minecraft:cardinal_direction": c.Facing.String()}
}

// allChests ...
func allChests() (chests []world.Block) {
	for _, direction := range cube.Directions() {
		chests = append(chests, Chest{Facing: direction})
	}
	return
}
