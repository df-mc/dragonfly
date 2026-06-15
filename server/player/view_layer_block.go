package player

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

type blockBreakTarget struct {
	pos     cube.Pos
	block   world.Block
	private bool
}

// visibleBlock returns the block currently shown to the player at pos.
func (p *Player) visibleBlock(pos cube.Pos) (world.Block, bool) {
	if b, ok := p.privateBlock(pos); ok {
		return b, true
	}
	return p.tx.Block(pos), false
}

type blockAudience interface {
	PlaySound(pos mgl64.Vec3, s world.Sound)
	AddParticle(pos mgl64.Vec3, particle world.Particle)
	ViewBlockAction(pos cube.Pos, action world.BlockAction)
	Resend(pos cube.Pos)
	ClearOverride(pos cube.Pos)
}

func (p *Player) blockAudience(private bool) blockAudience {
	if private {
		return privateBlockAudience{p: p}
	}
	return publicBlockAudience{p: p}
}

type privateBlockAudience struct {
	p *Player
}

func (a privateBlockAudience) PlaySound(pos mgl64.Vec3, s world.Sound) {
	a.p.session().ViewSound(pos, s)
}

func (a privateBlockAudience) AddParticle(pos mgl64.Vec3, particle world.Particle) {
	a.p.ShowParticle(pos, particle)
}

func (a privateBlockAudience) ViewBlockAction(pos cube.Pos, action world.BlockAction) {
	a.p.session().ViewPrivateBlockAction(pos, action)
}

func (a privateBlockAudience) Resend(pos cube.Pos) {
	a.p.session().ViewLayerBlockChanged(pos)
}

func (a privateBlockAudience) ClearOverride(pos cube.Pos) {
	a.p.ViewPublicBlock(pos)
}

type publicBlockAudience struct {
	p *Player
}

func (a publicBlockAudience) PlaySound(pos mgl64.Vec3, s world.Sound) {
	a.p.tx.PlaySound(pos, s)
}

func (a publicBlockAudience) AddParticle(pos mgl64.Vec3, particle world.Particle) {
	a.p.tx.AddParticle(pos, particle)
}

func (a publicBlockAudience) ViewBlockAction(pos cube.Pos, action world.BlockAction) {
	for _, viewer := range a.p.viewers() {
		viewer.ViewBlockAction(pos, action)
	}
}

func (a publicBlockAudience) Resend(pos cube.Pos) {
	a.p.resendNearbyBlocks(pos)
}

func (a publicBlockAudience) ClearOverride(cube.Pos) {}

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

func (p *Player) breakTarget(target blockBreakTarget) {
	if target.private {
		if _, ok := p.privateBlock(target.pos); !ok {
			p.blockAudience(false).Resend(target.pos)
			return
		}
	}
	p.breakBlock(target.pos, target.block, target.private)
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

	ctx := event.C(p)
	if p.Handler().HandleBlockBreak(ctx, pos, &drops, &xp); ctx.Cancelled() {
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
	if block.BreaksInstantly(b, held) {
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
	return p.ViewLayer().Block(pos)
}

// resendBreakingBlock resends the block being broken without overwriting private view-layer overrides.
func (p *Player) resendBreakingBlock(pos cube.Pos, private bool) {
	p.blockAudience(private).Resend(pos)
}
