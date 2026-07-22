package player

import (
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	shieldBlockDelay                   = time.Second / 4
	shieldAttackCooldown               = time.Second / 2
	shieldDisableCooldown              = 5 * time.Second
	shieldExplosionKnockBackMultiplier = 0.2
	shieldDamageThreshold              = 3
	shieldItemName                     = "minecraft:shield"

	shieldAttackerKnockBackForce  = 0.4
	shieldAttackerKnockBackHeight = 0.4
)

// shieldHand identifies the hand holding a shield.
type shieldHand int

const (
	shieldHandMain shieldHand = iota
	shieldHandOff
)

// shieldDisabler exposes the item used to attack a shield.
type shieldDisabler interface {
	HeldItems() (item.Stack, item.Stack)
}

// shieldKnockBacker is implemented by attackers that can be knocked back when their melee hit is blocked.
type shieldKnockBacker interface {
	KnockBack(src mgl64.Vec3, force, height float64)
}

// ShieldBlocking returns true if the player is currently blocking with a shield.
func (p *Player) ShieldBlocking() bool {
	return p.shieldBlockingAt(time.Now())
}

// shieldBlockingAt reports whether the shield is ready to block at now.
func (p *Player) shieldBlockingAt(now time.Time) bool {
	_, _, ok := p.blockingShieldAt(now)
	return ok
}

// updateShieldBlockingState refreshes cached visible shield state, reporting whether it changed.
func (p *Player) updateShieldBlockingState(now time.Time) bool {
	wasPrepared, wasBlocking := !p.shieldBlockingSince.IsZero(), p.shieldBlockingCached
	if !p.canBlockWithShieldAt(now, true) {
		p.shieldBlockingSince = time.Time{}
		p.shieldBlockingCached = false
		return wasPrepared || wasBlocking
	}
	if !wasPrepared {
		p.shieldBlockingSince = now
	}
	p.shieldBlockingCached = !now.Before(p.shieldBlockingSince.Add(shieldBlockDelay))
	return !wasPrepared || wasBlocking != p.shieldBlockingCached
}

// resetShieldBlocking lowers the shield and reports whether visible state changed.
func (p *Player) resetShieldBlocking() bool {
	wasPrepared, wasBlocking := !p.shieldBlockingSince.IsZero(), p.shieldBlockingCached
	p.shieldBlockingSince = time.Time{}
	p.shieldBlockingCached = false
	p.shieldUsePending = false
	return wasPrepared || wasBlocking
}

// canBlockWithShieldAt reports whether the player may raise a shield at now.
func (p *Player) canBlockWithShieldAt(now time.Time, cleanExpiredCooldown bool) bool {
	if !p.shieldInputActiveAt(now, cleanExpiredCooldown) {
		return false
	}
	_, _, ok := p.heldShield()
	return ok
}

// shieldInputActiveAt reports whether shield input is active without a cooldown at now.
func (p *Player) shieldInputActiveAt(now time.Time, cleanExpiredCooldown bool) bool {
	return (p.Sneaking() || p.shieldBlockingInput) && !p.hasCooldownAt(item.Shield{}, now, cleanExpiredCooldown)
}

// heldShield returns the active shield, preferring the off hand as the client does.
func (p *Player) heldShield() (item.Stack, shieldHand, bool) {
	mainHand, offHand := p.HeldItems()
	if _, ok := offHand.Item().(item.Shield); ok {
		return offHand, shieldHandOff, true
	}
	if _, ok := mainHand.Item().(item.Shield); ok {
		return mainHand, shieldHandMain, true
	}
	return item.Stack{}, 0, false
}

// blockingShieldAt returns the shield ready to block at now.
func (p *Player) blockingShieldAt(now time.Time) (item.Stack, shieldHand, bool) {
	if p.shieldBlockingSince.IsZero() || now.Before(p.shieldBlockingSince.Add(shieldBlockDelay)) ||
		!p.shieldInputActiveAt(now, false) {
		return item.Stack{}, 0, false
	}
	return p.heldShield()
}

// delayShieldAfterAttack lowers shields briefly after an attack.
func (p *Player) delayShieldAfterAttack() {
	until := time.Now().Add(shieldAttackCooldown)
	if current, ok := p.cooldowns[shieldItemName]; ok && current.After(until) {
		return
	}
	p.setCooldown(item.Shield{}, shieldAttackCooldown, true)
}

// setHeldShield writes a shield stack back to hand.
func (p *Player) setHeldShield(hand shieldHand, shield item.Stack) {
	if hand == shieldHandMain {
		_ = p.inv.SetItem(int(*p.heldSlot), shield)
		return
	}
	_ = p.offHand.SetItem(0, shield)
}

// shieldForDamageAt returns the raised shield if it blocks info at now.
func (p *Player) shieldForDamageAt(info world.ShieldBlockInfo, now time.Time) (item.Stack, shieldHand, bool) {
	shield, hand, ok := p.blockingShieldAt(now)
	if !ok || !p.facingShieldDamageSource(info.Origin) {
		return item.Stack{}, 0, false
	}
	return shield, hand, true
}

// facingShieldDamageSource reports whether source is in front of the player.
func (p *Player) facingShieldDamageSource(source mgl64.Vec3) bool {
	direction := source.Sub(p.Position())
	direction[1] = 0
	if direction.Len() == 0 {
		return false
	}
	look := cube.Rotation{p.Rotation().Yaw(), 0}.Vec3()
	return direction.Normalize().Dot(look) > 0
}

// shieldBlockInfo returns shield-blocking information exposed by src.
func shieldBlockInfo(src world.DamageSource) (world.ShieldBlockInfo, bool) {
	s, ok := src.(world.ShieldBlockSource)
	if !ok {
		return world.ShieldBlockInfo{}, false
	}
	return s.ShieldBlockInfo()
}

// shieldDisableCooldownFrom returns the shield cooldown caused by an axe attack.
func shieldDisableCooldownFrom(src world.DamageSource) (time.Duration, bool) {
	attack, ok := src.(entity.AttackDamageSource)
	if !ok {
		return 0, false
	}
	attacker, ok := attack.Attacker.(shieldDisabler)
	if !ok {
		return 0, false
	}
	mainHand, _ := attacker.HeldItems()
	tool, ok := mainHand.Item().(item.Tool)
	if !ok || tool.ToolType() != item.TypeAxe {
		return 0, false
	}
	return shieldDisableCooldown, true
}

// shieldDurabilityDamage returns durability lost from blocking dmg.
func shieldDurabilityDamage(dmg float64) int {
	if dmg < shieldDamageThreshold {
		return 0
	}
	return int(math.Floor(dmg)) + 1
}

// useItemStartsShieldBlocking reports whether item use should raise a held shield.
func (p *Player) useItemStartsShieldBlocking(mainHand item.Stack) bool {
	if _, ok := mainHand.Item().(item.Shield); ok {
		return true
	}
	switch mainHand.Item().(type) {
	case item.Releasable, item.Chargeable, item.Usable, item.Consumable:
		return false
	}
	_, offHand := p.HeldItems()
	_, ok := offHand.Item().(item.Shield)
	return ok
}

// StartShieldBlockingInput handles an item-use input that may raise a shield.
func (p *Player) StartShieldBlockingInput() bool {
	mainHand, _ := p.HeldItems()
	if !p.canStartShieldBlockingInput(mainHand) {
		return false
	}
	if !p.handleShieldItemUse() {
		return false
	}
	mainHand, _ = p.HeldItems()
	if !p.startShieldBlockingInput(mainHand) {
		return false
	}
	p.shieldUsePending = true
	return true
}

// canStartShieldBlockingInput reports whether mainHand use may raise a shield.
func (p *Player) canStartShieldBlockingInput(mainHand item.Stack) bool {
	if !p.useItemStartsShieldBlocking(mainHand) {
		return false
	}
	return !p.HasCooldown(item.Shield{})
}

// startShieldBlockingInput raises the shield if using mainHand allows it, returning true if it did.
func (p *Player) startShieldBlockingInput(mainHand item.Stack) bool {
	if !p.canStartShieldBlockingInput(mainHand) {
		return false
	}
	p.SetShieldBlockingInput(true)
	return true
}

// startOffHandShieldBlockingInput falls back to raising an off-hand shield.
func (p *Player) startOffHandShieldBlockingInput() bool {
	mainHand, offHand := p.HeldItems()
	if _, ok := mainHand.Item().(item.Shield); ok {
		return p.startShieldBlockingInput(mainHand)
	}
	if _, ok := offHand.Item().(item.Shield); !ok {
		return false
	}
	return p.startShieldBlockingInput(item.Stack{})
}

// startOffHandShieldBlockingInputAfterItemUse runs the item-use handler before falling back to the off hand.
func (p *Player) startOffHandShieldBlockingInputAfterItemUse() bool {
	if !p.canStartOffHandShieldBlockingInput() {
		return false
	}
	if !p.handleShieldItemUse() {
		return false
	}
	return p.startOffHandShieldBlockingInput()
}

// handleShieldItemUse reports whether the item use handler permits raising the shield.
func (p *Player) handleShieldItemUse() bool {
	ctx := newContext(p)
	p.Handler().HandleItemUse(ctx)
	return !ctx.Cancelled()
}

// canStartOffHandShieldBlockingInput reports whether an off-hand shield may be raised.
func (p *Player) canStartOffHandShieldBlockingInput() bool {
	mainHand, offHand := p.HeldItems()
	if _, ok := mainHand.Item().(item.Shield); ok {
		return p.canStartShieldBlockingInput(mainHand)
	}
	if _, ok := offHand.Item().(item.Shield); !ok {
		return false
	}
	return p.canStartShieldBlockingInput(item.Stack{})
}

// consumePendingShieldUse consumes a shield use already handled through auth input.
func (p *Player) consumePendingShieldUse(mainHand item.Stack) bool {
	if !p.shieldUsePending {
		return false
	}
	p.shieldUsePending = false
	return p.canStartShieldBlockingInput(mainHand)
}

// knockBackShieldAttacker knocks back a blocked melee attacker.
func (p *Player) knockBackShieldAttacker(src world.DamageSource) bool {
	attack, ok := src.(entity.AttackDamageSource)
	if !ok {
		return false
	}
	attacker, ok := attack.Attacker.(shieldKnockBacker)
	if !ok {
		return false
	}
	attacker.KnockBack(p.Position(), shieldAttackerKnockBackForce, shieldAttackerKnockBackHeight)
	return true
}

// blockDamageWithShield applies shield effects and reports whether dmg was blocked.
func (p *Player) blockDamageWithShield(dmg float64, src world.DamageSource, info world.ShieldBlockInfo) bool {
	now := time.Now()
	if info.Source != nil && info.Source.H() != nil && p.H() != nil && info.Source.H().UUID() == p.H().UUID() {
		return false
	}
	shield, hand, ok := p.shieldForDamageAt(info, now)
	if !ok {
		return false
	}
	if damage := shieldDurabilityDamage(dmg); damage > 0 {
		p.setHeldShield(hand, p.damageItem(shield, damage))
	}
	if p.tx != nil {
		p.tx.PlaySound(p.Position(), sound.ShieldBlock{})
	}
	p.knockBackShieldAttacker(src)
	if cooldown, ok := shieldDisableCooldownFrom(src); ok {
		p.setCooldown(item.Shield{}, cooldown, false)
		p.resetShieldBlocking()
	} else {
		p.updateShieldBlockingState(now)
	}
	if p.tx != nil {
		p.updateState()
	}
	return true
}
