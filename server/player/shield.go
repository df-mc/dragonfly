package player

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	shieldBlockDelayTicks              int64 = 5
	shieldAttackCooldown                     = time.Second / 2
	shieldTickDuration                       = time.Second / 20
	shieldExplosionKnockBackMultiplier       = 0.2
	shieldItemName                           = "minecraft:shield"

	shieldAttackerKnockBackForce  = 0.4
	shieldAttackerKnockBackHeight = 0.4
)

// shieldState holds the player's active shield input, timing and transient visual state.
type shieldState struct {
	tick, worldTick, readyAt, cooldownUntil int64
	input, prepared, blocking               bool
	blocked, damaged                        bool
	blockTick                               int64
	tickWorld, blockWorld                   *world.World
}

// shieldTick returns a monotonic player-local tick synchronised with the current world tick.
func (p *Player) shieldTick() int64 {
	if p.tx == nil {
		return p.shield.tick
	}
	w, current := p.tx.World(), p.tx.CurrentTick()
	if p.shield.tickWorld != w {
		p.shield.tickWorld, p.shield.worldTick = w, current
		return p.shield.tick
	}
	if current > p.shield.worldTick {
		p.shield.tick += current - p.shield.worldTick
		p.shield.worldTick = current
	}
	return p.shield.tick
}

// shieldCooldownActive reports whether the shield cooldown is still running.
func (p *Player) shieldCooldownActive() bool {
	return p.shieldTick() < p.shield.cooldownUntil
}

// shieldTicks returns the number of game ticks covered by duration.
func shieldTicks(duration time.Duration) int64 {
	if duration <= 0 {
		return 0
	}
	return int64((duration + shieldTickDuration - 1) / shieldTickDuration)
}

// setShieldCooldown sets the shield cooldown using the player's monotonic shield tick counter.
func (p *Player) setShieldCooldown(duration time.Duration) {
	p.shield.cooldownUntil = p.shieldTick() + shieldTicks(duration)
}

// shieldHand identifies the hand holding a shield.
type shieldHand int

const (
	shieldHandMain shieldHand = iota
	shieldHandOff
)

// shieldKnockBacker is implemented by attackers that can be knocked back when their melee hit is blocked.
type shieldKnockBacker interface {
	KnockBack(src mgl64.Vec3, force, height float64)
}

// ShieldBlocking returns true if the player is currently blocking with a shield.
func (p *Player) ShieldBlocking() bool {
	_, _, ok := p.blockingShield()
	return ok
}

// ShieldBlockState reports the transient visual state of the most recent shield block.
func (p *Player) ShieldBlockState() (blocked, damaged bool) {
	return p.shield.blocked, p.shield.damaged
}

// recordShieldBlock records the visual state for a shield block until the next tick.
func (p *Player) recordShieldBlock(durabilityDamage int, currentWorld *world.World, currentTick int64) {
	p.shield.blocked, p.shield.damaged = durabilityDamage == 0, durabilityDamage > 0
	p.shield.blockWorld, p.shield.blockTick = currentWorld, currentTick
}

// clearShieldBlockState clears transient shield block visuals and reports whether they changed.
func (p *Player) clearShieldBlockState(currentWorld *world.World, currentTick int64) bool {
	changed := p.shield.blocked || p.shield.damaged
	if !changed || currentWorld == p.shield.blockWorld && currentTick <= p.shield.blockTick {
		return false
	}
	p.shield.blocked, p.shield.damaged = false, false
	p.shield.blockWorld = nil
	return true
}

// updateShieldBlockingState refreshes cached visible shield state, reporting whether it changed.
func (p *Player) updateShieldBlockingState() bool {
	now := p.shieldTick()
	wasPrepared, wasBlocking := p.shield.prepared, p.shield.blocking
	if !p.canBlockWithShield() {
		p.shield.prepared = false
		p.shield.blocking = false
		return wasPrepared || wasBlocking
	}
	if !wasPrepared {
		p.shield.prepared = true
		p.shield.readyAt = now + shieldBlockDelayTicks
	}
	p.shield.blocking = now >= p.shield.readyAt
	return !wasPrepared || wasBlocking != p.shield.blocking
}

// resetShieldBlocking lowers the shield and reports whether visible state changed.
func (p *Player) resetShieldBlocking() bool {
	wasPrepared, wasBlocking := p.shield.prepared, p.shield.blocking
	p.shield.prepared = false
	p.shield.blocking = false
	return wasPrepared || wasBlocking
}

// canBlockWithShield reports whether the player may raise a shield.
func (p *Player) canBlockWithShield() bool {
	if !p.shieldInputActive() {
		return false
	}
	_, _, ok := p.heldShield()
	return ok
}

// shieldInputActive reports whether shield input is active without a cooldown.
func (p *Player) shieldInputActive() bool {
	return (p.Sneaking() || p.shield.input) && !p.shieldCooldownActive()
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

// blockingShield returns the shield ready to block.
func (p *Player) blockingShield() (item.Stack, shieldHand, bool) {
	if !p.shield.prepared || p.shieldTick() < p.shield.readyAt || !p.shieldInputActive() {
		return item.Stack{}, 0, false
	}
	return p.heldShield()
}

// delayShieldAfterAttack lowers shields briefly after an attack.
func (p *Player) delayShieldAfterAttack() {
	if p.shield.cooldownUntil > p.shieldTick()+shieldTicks(shieldAttackCooldown) {
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

// shieldForDamage returns the raised shield if it blocks info.
func (p *Player) shieldForDamage(info world.ShieldBlockInfo) (item.Stack, shieldHand, bool) {
	shield, hand, ok := p.blockingShield()
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
	s, ok := src.(world.ShieldBlockInfoSource)
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
	// TODO(2): Implement world.ShieldDisabler on the Warden when Warden support is added.
	if attacker, ok := attack.Attacker.(world.ShieldDisabler); ok {
		if cooldown := attacker.ShieldDisableDuration(); cooldown > 0 {
			return cooldown, true
		}
	}
	attacker, ok := attack.Attacker.(item.Carrier)
	if !ok {
		return 0, false
	}
	mainHand, _ := attacker.HeldItems()
	disabler, ok := mainHand.Item().(world.ShieldDisabler)
	if !ok {
		return 0, false
	}
	cooldown := disabler.ShieldDisableDuration()
	return cooldown, cooldown > 0
}

// shieldShouldBlockDamage reports whether an eligible hit should reach shield handling. Damage reduced to zero before
// the handler still reaches the shield, while a handler that deliberately reduces positive damage to zero suppresses it.
func shieldShouldBlockDamage(raw, beforeHandler, afterHandler float64, info world.ShieldBlockInfo) bool {
	return afterHandler > 0 || beforeHandler <= 0 && (raw > 0 || info.BlockZeroDamage)
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
	return p.startShieldBlockingInput(mainHand)
}

// canStartShieldBlockingInput reports whether mainHand use may raise a shield.
func (p *Player) canStartShieldBlockingInput(mainHand item.Stack) bool {
	if !p.useItemStartsShieldBlocking(mainHand) {
		return false
	}
	return !p.shieldCooldownActive()
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
	if !p.canStartOffHandShieldBlockingInput() {
		return false
	}
	p.SetShieldBlockingInput(true)
	return true
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
	ctx := NewEventContext(p.tx, p)
	p.Handler().HandleItemUse(ctx)
	return !ctx.Cancelled()
}

// canStartOffHandShieldBlockingInput reports whether an off-hand shield may be raised. A main-hand
// shield is checked as such; otherwise the main hand is treated as empty so only the off hand counts.
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
	if info.Source != nil {
		if h := info.Source.H(); h != nil && h == p.H() {
			return false
		}
	}
	shield, hand, ok := p.shieldForDamage(info)
	if !ok {
		return false
	}
	durabilityDamage := shield.Item().(item.Shield).BlockDurabilityDamage(dmg)
	if durabilityDamage > 0 {
		p.setHeldShield(hand, p.damageItem(shield, durabilityDamage))
	}
	if p.tx != nil {
		p.recordShieldBlock(durabilityDamage, p.tx.World(), p.tx.CurrentTick())
		p.tx.PlaySound(p.Position(), sound.ShieldBlock{})
	} else {
		p.recordShieldBlock(durabilityDamage, nil, 0)
	}
	p.knockBackShieldAttacker(src)
	if cooldown, ok := shieldDisableCooldownFrom(src); ok {
		p.setCooldown(item.Shield{}, cooldown, false)
		p.resetShieldBlocking()
	} else {
		p.updateShieldBlockingState()
	}
	if p.tx != nil {
		p.updateState()
	}
	return true
}
