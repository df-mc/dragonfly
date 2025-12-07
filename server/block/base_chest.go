package block

import (
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// chest contains common Chest functionality.
type chest struct {
	paired       bool
	pairX, pairZ int
	pairInv      *inventory.Inventory
	inventory    *inventory.Inventory
	viewerMu     *sync.RWMutex
	viewers      map[ContainerViewer]struct{}
}

// newBaseChest creates a new initialized base chest with inventory
func newBaseChest() chest {
	b := chest{
		viewerMu: new(sync.RWMutex),
		viewers:  make(map[ContainerViewer]struct{}, 1),
	}
	b.inventory = inventory.New(27, func(slot int, _, after item.Stack) {
		b.viewerMu.RLock()
		defer b.viewerMu.RUnlock()
		for viewer := range b.viewers {
			viewer.ViewSlotChange(slot, after)
		}
	})
	return b
}

// Model ...
func (chest) Model() world.BlockModel {
	return model.Chest{}
}

// pairPos returns the position of the paired chest.
func (b chest) pairPos(pos cube.Pos) cube.Pos {
	return cube.Pos{b.pairX, pos[1], b.pairZ}
}

// open opens the chest, displaying the animation and playing a sound.
func (b chest) open(tx *world.Tx, pos cube.Pos) {
	for _, v := range tx.Viewers(pos.Vec3()) {
		if b.paired {
			v.ViewBlockAction(b.pairPos(pos), OpenAction{})
		}
		v.ViewBlockAction(pos, OpenAction{})
	}
	tx.PlaySound(pos.Vec3Centre(), sound.ChestOpen{})
}

// close closes the chest, displaying the animation and playing a sound.
func (b chest) close(tx *world.Tx, pos cube.Pos) {
	for _, v := range tx.Viewers(pos.Vec3()) {
		if b.paired {
			v.ViewBlockAction(b.pairPos(pos), CloseAction{})
		}
		v.ViewBlockAction(pos, CloseAction{})
	}
	tx.PlaySound(pos.Vec3Centre(), sound.ChestClose{})
}

// addViewer adds a viewer to the chest, so that it is updated whenever the inventory of the chest is changed.
func (b chest) addViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	b.viewerMu.Lock()
	defer b.viewerMu.Unlock()
	if len(b.viewers) == 0 {
		b.open(tx, pos)
	}
	b.viewers[v] = struct{}{}
}

// removeViewer removes a viewer from the chest, so that slot updates in the inventory are no longer sent to it.
func (b chest) removeViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	b.viewerMu.Lock()
	defer b.viewerMu.Unlock()
	if len(b.viewers) == 0 {
		return
	}
	delete(b.viewers, v)
	if len(b.viewers) == 0 {
		b.close(tx, pos)
	}
}

// mergeInventories merges the inventories of the two chests and returns the merged inventories.
func mergeInventories(c1Inv, c2Inv *inventory.Inventory, pos, pairPos cube.Pos, facing cube.Direction) (
	left, right, double *inventory.Inventory,
	mu *sync.RWMutex,
	viewers map[ContainerViewer]struct{},
) {
	mu = new(sync.RWMutex)
	viewers = make(map[ContainerViewer]struct{})
	left, right = c1Inv.Clone(nil), c2Inv.Clone(nil)

	if pos.Side(facing.RotateRight().Face()) == pairPos {
		left, right = right, left
	}

	double = left.Merge(right, func(slot int, _, item item.Stack) {
		if slot < 27 {
			_ = left.SetItem(slot, item)
		} else {
			_ = right.SetItem(slot-27, item)
		}
		mu.RLock()
		defer mu.RUnlock()
		for viewer := range viewers {
			viewer.ViewSlotChange(slot, item)
		}
	})

	return left, right, double, mu, viewers
}

// unpairChests unpairs the chests from each other.
func unpairChests(b *chest, tx *world.Tx, pos cube.Pos) {
	if len(b.viewers) != 0 {
		b.close(tx, pos)
	}

	b.inventory = b.inventory.Clone(func(slot int, _, after item.Stack) {
		b.viewerMu.RLock()
		defer b.viewerMu.RUnlock()
		for viewer := range b.viewers {
			viewer.ViewSlotChange(slot, after)
		}
	})
	b.paired = false
	b.viewerMu = new(sync.RWMutex)
	b.viewers = make(map[ContainerViewer]struct{}, 1)
	b.pairInv = nil
}
