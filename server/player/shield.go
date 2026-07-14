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
	// shieldBlockDelay is the time a shield must be raised for before it starts blocking damage.
	shieldBlockDelay = time.Second / 4
	// shieldDisableCooldown is how long a shield is unusable for after being disabled by an axe.
	shieldDisableCooldown = 5 * time.Second
	// shieldExplosionKnockBackMultiplier is the fraction of explosion knock back still applied to a player
	// that blocked the blast with a shield.
	shieldExplosionKnockBackMultiplier = 0.2
	// shieldDamageThreshold is the minimum damage a blocked hit must deal for the shield to lose durability.
	shieldDamageThreshold = 3
	// shieldItemName is the encoded item name of item.Shield, used to recognise shield cooldowns by name.
	shieldItemName = "minecraft:shield"

	shieldAttackerKnockBackForce  = 0.4
	shieldAttackerKnockBackHeight = 0.4
)

// shieldHand identifies the hand a shield is held in. A shield in the main hand takes precedence over one in
// the off hand, both for blocking and for taking durability damage.
type shieldHand int

const (
	shieldHandMain shieldHand = iota
	shieldHandOff
)

// shieldDisabler is implemented by attackers whose held item may disable a shield, such as a player wielding
// an axe.
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

// shieldBlockingAt reports whether the shield is actually up at time now, meaning it was raised at least
// shieldBlockDelay ago and the player may still block. It reads state without mutating it, so that queries
// such as ShieldBlocking do not expire cooldowns as a side effect.
func (p *Player) shieldBlockingAt(now time.Time) bool {
	if p.shieldBlockingSince.IsZero() || !p.canBlockWithShieldAt(now, false) {
		return false
	}
	return !now.Before(p.shieldBlockingSince.Add(shieldBlockDelay))
}

// updateShieldBlockingState recomputes the player's shield state at time now, starting the raise timer when
// the shield first goes up and clearing it once the player can no longer block. It returns true if the state
// visible to other players changed, in which case the caller must send an entity state update.
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

// resetShieldBlocking lowers the shield immediately, forcing the raise delay to be served again before the
// player can block once more. It is used when the shield is disabled or put on cooldown. It returns true if
// the state visible to other players changed.
func (p *Player) resetShieldBlocking() bool {
	wasPrepared, wasBlocking := !p.shieldBlockingSince.IsZero(), p.shieldBlockingCached
	p.shieldBlockingSince = time.Time{}
	p.shieldBlockingCached = false
	p.shieldBlockingUseHandled = false
	return wasPrepared || wasBlocking
}

// canBlockWithShieldAt reports whether the player holds a shield, is raising it (either by sneaking or by
// holding the use control) and is not on shield cooldown at time now. cleanExpiredCooldown must only be set
// by callers that are allowed to mutate the player, as it deletes an expired shield cooldown entry.
func (p *Player) canBlockWithShieldAt(now time.Time, cleanExpiredCooldown bool) bool {
	if (!p.Sneaking() && !p.shieldBlockingInput) || p.hasCooldownAt(item.Shield{}, now, cleanExpiredCooldown) {
		return false
	}
	_, _, ok := p.heldShield()
	return ok
}

// heldShield returns the shield the player would block with and the hand it is held in, preferring the main
// hand over the off hand. The final return value is false if the player holds no shield.
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

// setHeldShield writes a shield stack back to the hand it was taken from with heldShield, used to persist
// durability damage taken while blocking.
func (p *Player) setHeldShield(hand shieldHand, shield item.Stack) {
	if hand == shieldHandMain {
		_ = p.inv.SetItem(int(*p.heldSlot), shield)
		return
	}
	_ = p.offHand.SetItem(0, shield)
}

// shieldBlocksDamageAt reports whether a shield that is up at time now blocks damage from src. Only damage
// that has a position in the world can be blocked, and only if it comes from in front of the player.
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

// facingShieldDamageSource reports whether source lies in the half of the horizontal plane the player is
// looking at. Damage from directly above, below or exactly beside the player is not blocked.
func (p *Player) facingShieldDamageSource(source mgl64.Vec3) bool {
	direction := source.Sub(p.Position())
	direction[1] = 0
	if direction.Len() == 0 {
		return false
	}
	look := cube.Rotation{p.Rotation().Yaw(), 0}.Vec3()
	return direction.Normalize().Dot(look) > 0
}

// shieldDamageSourcePosition returns the position a shield-blockable damage source came from. The final
// return value is false for damage that a shield can never block, either because the source has no position
// (fall damage, drowning, ...) or because it explicitly opted out, such as an explosion configured as
// unblockable or one with no known origin.
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

// shieldDisableCooldownFrom returns the cooldown a blocked hit from src puts the shield on. Only melee hits
// from an attacker wielding an axe disable a shield; every other blocked hit leaves it usable, in which case
// the final return value is false.
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

// shieldDurabilityDamage returns the durability a shield loses from blocking a hit of dmg damage. Weak hits
// below shieldDamageThreshold cost no durability at all.
func shieldDurabilityDamage(dmg float64) int {
	if dmg < shieldDamageThreshold {
		return 0
	}
	return int(math.Floor(dmg)) + 1
}

// shouldAttemptShieldBlockBeforeHurtHandler reports whether a shield block should be attempted for src before
// Hurt runs the damage handler. Projectiles and explosions must be able to block even while the player is
// still in their damage immunity window, so that arrows are deflected and blasts are shrugged off rather than
// silently ignored. Melee hits are left to the post-handler check, so a cancelled hit does not wear down the
// shield.
func shouldAttemptShieldBlockBeforeHurtHandler(_ float64, src world.DamageSource) bool {
	switch src.(type) {
	case entity.ProjectileDamageSource, entity.ExplosionDamageSource:
		return true
	default:
		return false
	}
}

// shouldAttemptShieldBlockAfterHurtHandler reports whether a shield block should be attempted once the damage
// handler has run. A hit that still deals damage is blockable. A hit that was reduced to nothing by the
// handler is not, as there is nothing left to block, unless it never carried damage to begin with: harmless
// projectiles such as eggs and snowballs must still be deflected by a shield.
func shouldAttemptShieldBlockAfterHurtHandler(rawDamage, damageLeft, damageBeforeHandler float64, src world.DamageSource) bool {
	if damageLeft > 0 {
		return true
	}
	if damageBeforeHandler > 0 {
		return false
	}
	return isZeroDamageProjectile(rawDamage, src)
}

// isZeroDamageProjectile reports whether src is a projectile that deals no damage, such as an egg or a
// snowball. These are still blocked (and thus deflected) by a shield.
func isZeroDamageProjectile(rawDamage float64, src world.DamageSource) bool {
	_, ok := src.(entity.ProjectileDamageSource)
	return ok && rawDamage == 0
}

// useItemStartsShieldBlocking reports whether using the item in the main hand raises a shield. A shield in the
// main hand always does. Otherwise the item in the off hand is used only if the main hand item has no use of
// its own that would take priority, as eating or drawing a bow must not raise the off-hand shield.
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

// StartShieldBlockingInput starts shield blocking from an item-use input if the held items allow it. It runs
// the item use handler, so that a cancelled use also prevents the shield from being raised, and returns true
// if the shield was raised. When it does, the use is marked as handled so that a UseItem call arriving for
// the same input does not run the handler a second time.
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

// canStartShieldBlockingInput reports whether using mainHand may raise a shield right now, i.e. whether the
// held items call for it and the shield is not on cooldown after being disabled.
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

// startOffHandShieldBlockingInput raises the shield after the main hand item turned out to do nothing on use,
// e.g. an empty hand or food eaten on a full hunger bar. Such a use falls through to the off-hand shield, so
// the check is made against an empty main hand stack, bypassing useItemStartsShieldBlocking's rule that a
// usable main hand item takes priority.
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

// startOffHandShieldBlockingInputAfterItemUse raises the shield on a UseItem that returns before reaching the
// item use handler, which happens when the main hand item is still on cooldown. The handler is run here so
// that the shield is only raised if the use is not cancelled.
func (p *Player) startOffHandShieldBlockingInputAfterItemUse() bool {
	if !p.canStartOffHandShieldBlockingInput() {
		return false
	}
	if !p.handleShieldItemUse() {
		return false
	}
	return p.startOffHandShieldBlockingInput()
}

// handleShieldItemUse calls the item use handler for a use that raises a shield, returning false if the
// handler cancelled it.
func (p *Player) handleShieldItemUse() bool {
	ctx := newContext(p)
	p.Handler().HandleItemUse(ctx)
	return !ctx.Cancelled()
}

// canStartOffHandShieldBlockingInput reports whether startOffHandShieldBlockingInput would raise the shield,
// checked up front so that the item use handler is not run for a use that could never raise it.
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

// consumeShieldBlockingUseHandled reports whether StartShieldBlockingInput already ran the item use handler
// for the input UseItem is now processing, clearing the flag so that it applies only to that one use. UseItem
// uses this to avoid calling the handler twice for a single client input.
func (p *Player) consumeShieldBlockingUseHandled(mainHand item.Stack) bool {
	if !p.shieldBlockingUseHandled {
		return false
	}
	p.shieldBlockingUseHandled = false
	return p.canStartShieldBlockingInput(mainHand)
}

// knockBackShieldAttacker pushes back the attacker of a blocked melee hit, away from the player that blocked
// it. It returns false if the damage did not come from an attacker that can be knocked back.
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

// blockDamageWithShield attempts to block dmg from src with the player's shield, returning true if it did, in
// which case the caller must not apply the damage. Blocking wears down the shield, plays the block sound,
// knocks back a melee attacker and, if the hit came from an axe, disables the shield for a while. A player
// never blocks an explosion they caused themselves.
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
