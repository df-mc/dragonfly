package player

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// blockBreakTarget holds the position and break mode that a player started breaking.
type blockBreakTarget struct {
	pos     cube.Pos
	private bool
}

// visibleBlock returns the block currently shown to the player at pos.
func (p *Player) visibleBlock(pos cube.Pos) (world.Block, bool) {
	if b, ok := p.privateBlock(pos); ok {
		return b, true
	}
	return p.tx.Block(pos), false
}

// blockAudience handles block breaking side effects for either public world blocks or private view-layer blocks.
type blockAudience interface {
	PlaySound(pos mgl64.Vec3, s world.Sound)
	AddParticle(pos mgl64.Vec3, particle world.Particle)
	ViewBlockAction(pos cube.Pos, action world.BlockAction)
	Resend(pos cube.Pos)
	ClearOverride(pos cube.Pos)
}

// blockAudience returns the audience to use for the break mode passed.
func (p *Player) blockAudience(private bool) blockAudience {
	if private {
		return privateBlockAudience{p: p}
	}
	return publicBlockAudience{p: p}
}

// privateBlockAudience handles block breaking side effects for private view-layer blocks.
type privateBlockAudience struct {
	p *Player
}

// PlaySound plays the sound only to the player breaking the private block.
func (a privateBlockAudience) PlaySound(pos mgl64.Vec3, s world.Sound) {
	a.p.session().ViewSound(pos, s)
}

// AddParticle shows the particle only to the player breaking the private block.
func (a privateBlockAudience) AddParticle(pos mgl64.Vec3, particle world.Particle) {
	a.p.ShowParticle(pos, particle)
}

// ViewBlockAction shows the block action only to the player breaking the private block.
func (a privateBlockAudience) ViewBlockAction(pos cube.Pos, action world.BlockAction) {
	a.p.session().ViewPrivateBlockAction(pos, action)
}

// Resend resends the private block override to the player.
func (a privateBlockAudience) Resend(pos cube.Pos) {
	a.p.session().ViewLayerBlockChanged(a.p.tx.World(), pos)
}

// ClearOverride removes the private block override for the player.
func (a privateBlockAudience) ClearOverride(pos cube.Pos) {
	a.p.ViewPublicBlock(pos)
}

// publicBlockAudience handles block breaking side effects for public world blocks.
type publicBlockAudience struct {
	p *Player
}

// PlaySound plays the sound to all players viewing the public block.
func (a publicBlockAudience) PlaySound(pos mgl64.Vec3, s world.Sound) {
	s.Play(a.p.tx.World(), pos)
	for _, viewer := range a.viewers(cube.PosFromVec3(pos)) {
		viewer.ViewSound(pos, s)
	}
}

// AddParticle adds the particle to all players viewing the public block.
func (a publicBlockAudience) AddParticle(pos mgl64.Vec3, particle world.Particle) {
	particle.Spawn(a.p.tx.World(), pos)
	for _, viewer := range a.viewers(cube.PosFromVec3(pos)) {
		viewer.ViewParticle(pos, particle)
	}
}

// ViewBlockAction shows the block action to all players viewing the public block.
func (a publicBlockAudience) ViewBlockAction(pos cube.Pos, action world.BlockAction) {
	for _, viewer := range a.viewers(pos) {
		viewer.ViewBlockAction(pos, action)
	}
}

// Resend resends the public block to the player.
func (a publicBlockAudience) Resend(pos cube.Pos) {
	a.p.resendNearbyBlocks(pos)
}

// ClearOverride does nothing for public blocks.
func (a publicBlockAudience) ClearOverride(cube.Pos) {}

// viewers returns all viewers that do not have a private block override at pos.
func (a publicBlockAudience) viewers(pos cube.Pos) []world.Viewer {
	viewers := a.p.viewers()
	filtered := viewers[:0]
	for _, viewer := range viewers {
		if viewerViewsPublicBlock(viewer, a.p.tx.World(), pos) {
			filtered = append(filtered, viewer)
		}
	}
	return filtered
}

type blockOverrideViewer interface {
	ViewLayer() *world.ViewLayer
}

// viewerViewsPublicBlock returns whether viewer sees the public block at pos.
func viewerViewsPublicBlock(viewer world.Viewer, w *world.World, pos cube.Pos) bool {
	v, ok := viewer.(blockOverrideViewer)
	if !ok || v.ViewLayer() == nil {
		return true
	}
	_, ok = v.ViewLayer().Block(w, pos)
	return !ok
}

// BreakBlock makes the player break the public world block at the position passed. Private view-layer
// overrides are ignored by this method: Call BreakVisibleBlock to break what the player currently sees instead.
// If the player is unable to reach the block passed, the method returns immediately.
func (p *Player) BreakBlock(pos cube.Pos) {
	p.breakBlock(pos, p.tx.Block(pos), false)
}

// BreakVisibleBlock makes the player break the block currently shown to them at the position passed. If the
// player has a private block override at that position, it is removed instead of breaking the public world block.
// If the player is unable to reach the block passed, the method returns immediately.
func (p *Player) BreakVisibleBlock(pos cube.Pos) {
	b, private := p.visibleBlock(pos)
	p.breakBlock(pos, b, private)
}

// breakTarget breaks the target using the same mode as when breaking started.
func (p *Player) breakTarget(target blockBreakTarget) {
	b, ok := p.breakTargetBlock(target)
	if !ok {
		p.blockAudience(false).Resend(target.pos)
		return
	}
	p.breakBlock(target.pos, b, target.private)
}

// breakTargetBlock returns the current block matching the break mode of target.
func (p *Player) breakTargetBlock(target blockBreakTarget) (world.Block, bool) {
	if target.private {
		return p.privateBlock(target.pos)
	}
	return p.tx.Block(target.pos), true
}

// breakBlock makes the player break the block passed at the position passed. Private blocks are removed
// from the player's view layer instead of the world.
func (p *Player) breakBlock(pos cube.Pos, b world.Block, private bool) {
	audience := p.blockAudience(private)
	if _, air := b.(block.Air); air {
		// Don't do anything if the position broken is already air.
		return
	}
	if !p.canReach(pos.Vec3Centre()) || !p.GameMode().AllowsEditing() {
		audience.Resend(pos)
		return
	}
	breakable, ok := b.(block.Breakable)
	if !ok && !p.GameMode().CreativeInventory() {
		audience.Resend(pos)
		return
	}
	held, _ := p.HeldItems()
	var drops []item.Stack

	xp := 0
	if !private {
		drops = p.drops(held, b)
		if ok && !p.GameMode().CreativeInventory() {
			if _, hasSilkTouch := held.Enchantment(enchantment.SilkTouch); !hasSilkTouch {
				xp = breakable.BreakInfo().XPDrops.RandomValue()
			}
		}
	}

	ctx := newContext(p)
	if p.Handler().HandleBlockBreak(ctx, pos, private, &drops, &xp); ctx.Cancelled() {
		audience.Resend(pos)
		return
	}
	held, left := p.HeldItems()

	p.SwingArm()
	if private {
		audience.ClearOverride(pos)
		audience.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: b})
		return
	}
	p.tx.SetBlock(pos, nil, nil)
	audience.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: b})
	if ok {
		info := breakable.BreakInfo()
		if info.BreakHandler != nil {
			info.BreakHandler(pos, p.tx, p)
		}
		for _, orb := range entity.NewExperienceOrbs(pos.Vec3Centre(), xp) {
			p.tx.AddEntity(orb)
		}
	}
	for _, drop := range drops {
		opts := world.EntitySpawnOpts{Position: pos.Vec3Centre(), Velocity: mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1}}
		p.tx.AddEntity(entity.NewItem(opts, drop))
	}

	p.Exhaust(0.005)
	if block.BreaksInstantly(b) {
		return
	}
	if durable, ok := held.Item().(item.Durable); ok {
		p.SetHeldItems(p.damageItem(held, durable.DurabilityInfo().BreakDurability), left)
	}
}

// drops returns the drops that the player can get from the block passed using the item held.
func (p *Player) drops(held item.Stack, b world.Block) []item.Stack {
	t, ok := held.Item().(item.Tool)
	if !ok {
		t = item.ToolNone{}
	}
	var drops []item.Stack
	if breakable, ok := b.(block.Breakable); ok && !p.GameMode().CreativeInventory() {
		if breakable.BreakInfo().Harvestable(t) {
			drops = breakable.BreakInfo().Drops(t, held.Enchantments())
		}
	} else if it, ok := b.(world.Item); ok && !p.GameMode().CreativeInventory() {
		drops = []item.Stack{item.NewStack(it, 1)}
	}
	return drops
}

// privateBlock returns this player's view-layer block override at pos, if present.
func (p *Player) privateBlock(pos cube.Pos) (world.Block, bool) {
	if p.session() == session.Nop {
		return nil, false
	}
	return p.ViewLayer().Block(p.tx.World(), pos)
}

// resendBreakingBlock resends the block being broken without overwriting private view-layer overrides.
func (p *Player) resendBreakingBlock(pos cube.Pos, private bool) {
	p.blockAudience(private).Resend(pos)
}
