package player

import (
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	shieldBlockDelay                   = time.Second / 4
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
	if p.shieldBlockingSince.IsZero() || !p.canBlockWithShieldAt(now, false) {
		return false
	}
	return !now.Before(p.shieldBlockingSince.Add(shieldBlockDelay))
}

// updateShieldBlockingState refreshes shield state, reporting whether visible state changed.
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
	p.shieldBlockingUseHandled = false
	return wasPrepared || wasBlocking
}

// canBlockWithShieldAt reports whether the player may raise a shield at now.
func (p *Player) canBlockWithShieldAt(now time.Time, cleanExpiredCooldown bool) bool {
	if (!p.Sneaking() && !p.shieldBlockingInput) || p.hasCooldownAt(item.Shield{}, now, cleanExpiredCooldown) {
		return false
	}
	_, _, ok := p.heldShield()
	return ok
}

// heldShield returns the preferred held shield and its hand.
func (p *Player) heldShield() (item.Stack, shieldHand, bool) {
	mainHand, offHand := p.HeldItems()
	if _, ok := mainHand.Item().(item.Shield); ok {
		return mainHand, shieldHandMain, true
	}
	if _, ok := offHand.Item().(item.Shield); ok {
		return offHand, shieldHandOff, true
	}
	return item.Stack{}, 0, false
}

// setHeldShield writes a shield stack back to hand.
func (p *Player) setHeldShield(hand shieldHand, shield item.Stack) {
	if hand == shieldHandMain {
		_ = p.inv.SetItem(int(*p.heldSlot), shield)
		return
	}
	_ = p.offHand.SetItem(0, shield)
}

// shieldBlocksDamageAt reports whether the raised shield blocks src at now.
func (p *Player) shieldBlocksDamageAt(src world.DamageSource, now time.Time) bool {
	if !p.shieldBlockingAt(now) {
		return false
	}
	source, ok := shieldDamageSourcePosition(src)
	if !ok {
		return false
	}
	return p.facingShieldDamageSource(source)
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

// shieldDamageSourcePosition returns the origin of shield-blockable damage.
func shieldDamageSourcePosition(src world.DamageSource) (mgl64.Vec3, bool) {
	switch s := src.(type) {
	case entity.AttackDamageSource:
		if s.Attacker == nil {
			return mgl64.Vec3{}, false
		}
		return s.Attacker.Position(), true
	case entity.ProjectileDamageSource:
		if s.Projectile != nil {
			return s.Projectile.Position(), true
		}
		if s.Owner != nil {
			return s.Owner.Position(), true
		}
	case entity.ExplosionDamageSource:
		if !s.HasOrigin || !s.BlockableByShield {
			return mgl64.Vec3{}, false
		}
		return s.Origin, true
	case enchantment.ThornsDamageSource:
		if s.Owner == nil {
			return mgl64.Vec3{}, false
		}
		return s.Owner.Position(), true
	}
	return mgl64.Vec3{}, false
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

// shouldAttemptShieldBlockBeforeHurtHandler reports whether src should be checked before the hurt handler.
func shouldAttemptShieldBlockBeforeHurtHandler(_ float64, src world.DamageSource) bool {
	switch src.(type) {
	case entity.ProjectileDamageSource, entity.ExplosionDamageSource:
		return true
	default:
		return false
	}
}

// shouldAttemptShieldBlockAfterHurtHandler reports whether handled damage should still be blocked.
func shouldAttemptShieldBlockAfterHurtHandler(rawDamage, damageLeft, damageBeforeHandler float64, src world.DamageSource) bool {
	if damageLeft > 0 {
		return true
	}
	if damageBeforeHandler > 0 {
		return false
	}
	return isZeroDamageProjectile(rawDamage, src)
}

// isZeroDamageProjectile reports whether src is a harmless projectile that may still be deflected.
func isZeroDamageProjectile(rawDamage float64, src world.DamageSource) bool {
	_, ok := src.(entity.ProjectileDamageSource)
	return ok && rawDamage == 0
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
	p.shieldBlockingUseHandled = true
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

// consumeShieldBlockingUseHandled prevents handling one shield-use input twice.
func (p *Player) consumeShieldBlockingUseHandled(mainHand item.Stack) bool {
	if !p.shieldBlockingUseHandled {
		return false
	}
	p.shieldBlockingUseHandled = false
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
func (p *Player) blockDamageWithShield(dmg float64, src world.DamageSource) bool {
	now := time.Now()
	if s, ok := src.(entity.ExplosionDamageSource); ok && s.Source != nil && s.Source.H() != nil && p.H() != nil && s.Source.H().UUID() == p.H().UUID() {
		return false
	}
	if !p.shieldBlocksDamageAt(src, now) {
		return false
	}
	shield, hand, ok := p.heldShield()
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
