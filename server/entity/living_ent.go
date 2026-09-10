package entity

import (
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// LivingBehaviour is a Behaviour for entities that are alive. The LivingState it returns holds the state
// that a LivingEnt exposes through the Living interface. Because an entity is reopened in every
// transaction, this state must be part of the Behaviour rather than the entity itself.
type LivingBehaviour interface {
	Behaviour
	// LivingState returns the state of the living entity. It must return the same LivingState for the
	// lifetime of the entity.
	LivingState() *LivingState
}

// LivingHurtHandler may be implemented by a LivingBehaviour to handle damage dealt to its entity. It is
// called before any other processing of the damage, so it also runs while the entity is immune to attacks.
type LivingHurtHandler interface {
	// HandleHurt handles damage about to be dealt to the entity. The damage may be changed through the
	// pointer passed. If true is returned, the damage is cancelled entirely.
	HandleHurt(e *LivingEnt, damage *float64, src world.DamageSource) bool
}

// LivingDeathHandler may be implemented by a LivingBehaviour to react to the death of its entity. It is
// called on the tick after the fatal damage, so a *world.Tx is available.
type LivingDeathHandler interface {
	// HandleDeath handles the death of the entity, before it is closed and removed from the world.
	HandleDeath(e *LivingEnt, tx *world.Tx, src world.DamageSource)
}

// LivingState holds the state of a living entity that persists across transactions. A LivingBehaviour
// creates one using NewLivingState and returns it from its LivingState method.
type LivingState struct {
	// Health manages the entity's current and maximum health.
	Health *HealthManager
	// Effects manages the effects active on the entity. They are ticked by LivingEnt.Tick.
	Effects *EffectManager
	// Immunity tracks the invulnerability window the entity has after being hurt.
	Immunity AttackImmunity
	// HurtImmunity is the duration of the immunity window armed when the entity is hurt.
	HurtImmunity time.Duration
	// Air tracks the air the entity has left while submerged.
	Air *AirSupply
	// Armour holds the armour worn by the entity. Worn items reduce incoming damage, resist knock back,
	// retaliate with thorns and lose durability exactly as they do for players, and are shown to viewers.
	Armour *inventory.Armour
	// MainHand and OffHand are the items held by the entity, shown to viewers. Change them through
	// LivingEnt.SetHeldItems so viewers see the change.
	MainHand, OffHand item.Stack
	// Speed is the movement speed of the entity.
	Speed float64
	// KnockBackResistance is the fraction, between 0 and 1, by which knock back on the entity is reduced.
	// Resistance from worn armour is added on top of it.
	KnockBackResistance float64
	// FireImmune makes the entity immune to fire damage and prevents it from being set on fire.
	FireImmune bool
	// Scale is the size modifier of the entity, with 1 being the regular size.
	Scale float64
	// Baby marks the entity as a baby to viewers.
	Baby bool
	// Variant and MarkVariant select the visual variant of the entity for viewers.
	Variant, MarkVariant int32
	// FallDistance is the distance the entity has been falling for. It is updated through
	// LivingEnt.UpdateFallState and converted to fall damage upon landing.
	FallDistance float64

	absorption float64
	dead       bool
	deathSrc   world.DamageSource
}

// NewLivingState returns a LivingState with the health passed and vanilla defaults for the remaining
// state: a speed of 0.1, half a second of hurt immunity, 15 seconds of air, a scale of 1 and empty armour
// and hands.
func NewLivingState(health float64) *LivingState {
	return &LivingState{
		Health:       NewHealthManager(health, health),
		Effects:      NewEffectManager(),
		Air:          NewAirSupply(0),
		Armour:       inventory.NewArmour(nil),
		Speed:        0.1,
		Scale:        1,
		HurtImmunity: time.Second / 2,
	}
}

// Dead checks if the living entity has taken fatal damage.
func (s *LivingState) Dead() bool {
	return s.dead
}

// SetAbsorption sets the absorption health of the living state. The value is clamped to zero or above.
func (s *LivingState) SetAbsorption(health float64) {
	s.absorption = max(health, 0)
}

// LivingEnt is an Ent that is alive. It implements Living on top of the LivingState of its Behaviour, which
// makes the entity part of everything gated on Living. Entity implementations based on a LivingBehaviour
// return a LivingEnt from the Open method of their world.EntityType by calling OpenLiving.
type LivingEnt struct {
	*Ent
}

var _ Living = (*LivingEnt)(nil)

// OpenLiving converts a world.EntityHandle to a LivingEnt in a world.Tx. It panics if the Behaviour of the
// entity does not implement LivingBehaviour.
func OpenLiving(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) *LivingEnt {
	e := &LivingEnt{Ent: Open(tx, handle, data)}
	_ = e.behaviour()
	return e
}

// behaviour returns the Behaviour of the entity as a LivingBehaviour.
func (e *LivingEnt) behaviour() LivingBehaviour {
	return e.Behaviour().(LivingBehaviour)
}

// state returns the LivingState of the entity's Behaviour.
func (e *LivingEnt) state() *LivingState {
	return e.behaviour().LivingState()
}

// Health returns the current health of the entity.
func (e *LivingEnt) Health() float64 {
	return e.state().Health.Health()
}

// MaxHealth returns the maximum health of the entity.
func (e *LivingEnt) MaxHealth() float64 {
	return e.state().Health.MaxHealth()
}

// SetMaxHealth changes the maximum health of the entity to the value passed.
func (e *LivingEnt) SetMaxHealth(v float64) {
	e.state().Health.SetMaxHealth(v)
}

// Dead checks if the entity has taken fatal damage.
func (e *LivingEnt) Dead() bool {
	return e.state().Dead()
}

// Hurt hurts the entity for a given amount of damage. The damage is first passed to the Behaviour's
// HandleHurt, if implemented, which may change or cancel it. Worn armour and an active resistance effect
// then reduce it, and within the immunity window of an earlier hit only the excess over the damage that
// armed the window is dealt. Absorption health is consumed before regular health. If the damage is fatal,
// the entity is killed on its next tick, so that drops spawned by the Behaviour's HandleDeath have a
// transaction to be added in.
func (e *LivingEnt) Hurt(damage float64, src world.DamageSource) (float64, bool) {
	s := e.state()
	if damage < 0 || s.dead {
		return 0, false
	}
	if src.Fire() {
		if _, ok := s.Effects.Effect(effect.FireResistance); ok || s.FireImmune {
			return 0, false
		}
	}
	ctx := e.tx.Event()
	if e.tx.World().Handler().HandleEntityHurt(ctx, e, &damage, src); ctx.Cancelled() {
		return 0, false
	}
	if h, ok := e.Behaviour().(LivingHurtHandler); ok && h.HandleHurt(e, &damage, src) {
		return 0, false
	}
	damage = FinalDamage(damage, src, s.Armour, s.Effects)
	damageLeft, immune := s.Immunity.Reduce(damage)
	if immune && damageLeft <= 0 {
		return 0, false
	}
	s.Immunity.Arm(s.HurtImmunity, damage)

	if a := s.absorption; a > 0 {
		s.SetAbsorption(a - damageLeft)
		damageLeft = max(0, damageLeft-a)
		if _, ok := s.Effects.Effect(effect.Absorption); ok && s.absorption <= 0 {
			e.RemoveEffect(effect.Absorption)
		}
	}

	s.Health.AddHealth(-damageLeft)
	e.PlayAction(HurtAction{})

	if src.ReducedByArmour() {
		s.Armour.Damage(damage, e.damageItem)

		var origin world.Entity
		if a, ok := src.(AttackDamageSource); ok {
			origin = a.Attacker
		} else if p, ok := src.(ProjectileDamageSource); ok {
			origin = p.Owner
		}
		if l, ok := origin.(Living); ok {
			if thorns := s.Armour.ThornsDamage(e.damageItem); thorns > 0 {
				l.Hurt(thorns, enchantment.ThornsDamageSource{Owner: e})
			}
		}
	}

	if s.Health.Health() <= 0 {
		s.dead, s.deathSrc = true, src
	}
	return damageLeft, true
}

// Heal heals the entity for a given amount of health, up to its maximum health, and returns the amount of
// health that was regenerated. A dead entity is not healed.
func (e *LivingEnt) Heal(health float64, _ world.HealingSource) float64 {
	s := e.state()
	if s.dead || health <= 0 {
		return 0
	}
	before := s.Health.Health()
	s.Health.AddHealth(health)
	return s.Health.Health() - before
}

// KnockBack knocks the entity back away from the source position passed. The knock back resistance of the
// entity's LivingState and its worn armour is applied. A dead entity is not knocked back.
func (e *LivingEnt) KnockBack(src mgl64.Vec3, force, height float64) {
	if e.state().dead {
		return
	}
	velocity := KnockBackVector(e.Position(), src, force, height)
	e.SetVelocity(velocity.Mul(1 - e.knockBackResistance()))
}

// knockBackResistance returns the total knock back resistance of the entity, from its LivingState and its
// worn armour, capped at 1.
func (e *LivingEnt) knockBackResistance() float64 {
	s := e.state()
	return min(1, s.KnockBackResistance+s.Armour.KnockBackResistance())
}

// Explode hurts the entity with vanilla explosion damage and knocks it back away from the explosion. It
// takes the place of the Behaviour's ExplodableBehaviour hook, which is not called for living entities.
func (e *LivingEnt) Explode(src world.ExplosionSource, impact float64) {
	diff := e.Position().Sub(src.Position())
	e.Hurt(ExplosionDamage(src.Size(), impact), ExplosionDamageSource{Source: src})

	height := 0.0
	if l := diff.Len(); l != 0 {
		height = diff[1] / l * impact
	}
	// The knock back is applied even to an entity killed by the explosion, so its body is still launched.
	velocity := KnockBackVector(e.Position(), src.Position(), impact, height)
	e.SetVelocity(velocity.Mul(1 - e.knockBackResistance()))
}

// SetOnFire sets the entity on fire for the duration passed. Fire protection on worn armour reduces the
// duration, and a fire-immune entity is never set on fire.
func (e *LivingEnt) SetOnFire(duration time.Duration) {
	s := e.state()
	if s.FireImmune {
		return
	}
	ticks := int64(duration.Seconds() * 20)
	if level := s.Armour.HighestEnchantmentLevel(enchantment.FireProtection); level > 0 {
		ticks -= int64(math.Floor(float64(ticks) * float64(level) * 0.15))
	}
	e.Ent.SetOnFire(time.Duration(ticks) * time.Second / 20)
}

// AddEffect adds an effect to the entity. If the effect is instant, it is applied immediately. If not, it is
// applied every tick until it expires.
func (e *LivingEnt) AddEffect(eff effect.Effect) {
	e.state().Effects.Add(eff, e)
}

// RemoveEffect removes any effect of the type passed that is active on the entity.
func (e *LivingEnt) RemoveEffect(t effect.Type) {
	e.state().Effects.Remove(t, e)
}

// Effects returns the effects currently active on the entity.
func (e *LivingEnt) Effects() []effect.Effect {
	return e.state().Effects.Effects()
}

// Speed returns the current speed of the entity.
func (e *LivingEnt) Speed() float64 {
	return e.state().Speed
}

// SetSpeed sets the speed of the entity to a new value.
func (e *LivingEnt) SetSpeed(v float64) {
	e.state().Speed = v
}

// Absorption returns the absorption health of the entity.
func (e *LivingEnt) Absorption() float64 {
	return e.state().absorption
}

// SetAbsorption sets the absorption health of the entity. The value is clamped to zero or above.
func (e *LivingEnt) SetAbsorption(health float64) {
	e.state().SetAbsorption(health)
}

// Armour returns the armour worn by the entity. After changing its contents, UpdateArmour shows the change
// to viewers.
func (e *LivingEnt) Armour() *inventory.Armour {
	return e.state().Armour
}

// UpdateArmour shows the entity's current armour to viewers.
func (e *LivingEnt) UpdateArmour() {
	for _, v := range e.tx.Viewers(e.data.Pos) {
		v.ViewEntityArmour(e)
	}
}

// HeldItems returns the items currently held by the entity in its main hand and off hand.
func (e *LivingEnt) HeldItems() (mainHand, offHand item.Stack) {
	s := e.state()
	return s.MainHand, s.OffHand
}

// SetHeldItems changes the items held by the entity and shows the change to viewers.
func (e *LivingEnt) SetHeldItems(mainHand, offHand item.Stack) {
	s := e.state()
	s.MainHand, s.OffHand = mainHand, offHand
	for _, v := range e.tx.Viewers(e.data.Pos) {
		v.ViewEntityItems(e)
	}
}

// Scale returns the size modifier of the entity.
func (e *LivingEnt) Scale() float64 {
	return e.state().Scale
}

// SetScale changes the size modifier of the entity and shows the change to viewers.
func (e *LivingEnt) SetScale(scale float64) {
	e.state().Scale = scale
	e.UpdateState()
}

// Baby checks if the entity is marked as a baby.
func (e *LivingEnt) Baby() bool {
	return e.state().Baby
}

// Variant returns the visual variant of the entity.
func (e *LivingEnt) Variant() int32 {
	return e.state().Variant
}

// MarkVariant returns the visual mark variant of the entity.
func (e *LivingEnt) MarkVariant() int32 {
	return e.state().MarkVariant
}

// Breathing checks if the entity is currently able to breathe.
func (e *LivingEnt) Breathing() bool {
	return e.state().Air.Breathing()
}

// AirSupply returns the air the entity has left.
func (e *LivingEnt) AirSupply() time.Duration {
	return e.state().Air.Supply()
}

// MaxAirSupply returns the maximum air supply of the entity.
func (e *LivingEnt) MaxAirSupply() time.Duration {
	return e.state().Air.Max()
}

// FallDistance returns the distance the entity has been falling for.
func (e *LivingEnt) FallDistance() float64 {
	return e.state().FallDistance
}

// ResetFallDistance resets the distance the entity has been falling for, cancelling any pending fall
// damage.
func (e *LivingEnt) ResetFallDistance() {
	e.state().FallDistance = 0
}

// UpdateFallState progresses the fall state of the entity with the vertical movement of one tick. A
// Behaviour running its own physics calls it every tick: while the entity falls, the fall distance grows,
// and upon landing, the block landed on may soften the fall before the remaining distance, reduced by any
// jump boost effect, is dealt as fall damage.
func (e *LivingEnt) UpdateFallState(deltaY float64, onGround bool) {
	s := e.state()
	switch {
	case onGround:
		s.FallDistance -= deltaY
		if s.FallDistance > 3 {
			Fall(e, e.tx, s.Effects, s.FallDistance)
		}
		s.FallDistance = 0
	case deltaY < 0 && deltaY < s.FallDistance:
		s.FallDistance -= deltaY
	default:
		s.FallDistance = 0
	}
}

// damageItem applies durability damage to an item worn by the entity, respecting its unbreaking
// enchantment and breaking it once its durability runs out.
func (e *LivingEnt) damageItem(s item.Stack, d int) item.Stack {
	if d == 0 || s.MaxDurability() == -1 {
		return s
	}
	if ench, ok := s.Enchantment(enchantment.Unbreaking); ok {
		d = enchantment.Unbreaking.Reduce(s.Item(), ench.Level(), d)
	}
	if s = s.Damage(d); s.Empty() {
		e.tx.PlaySound(e.data.Pos, sound.ItemBreak{})
	}
	return s
}

// Tick ticks the entity. A dead entity plays its death animation, has the Behaviour's HandleDeath called if
// implemented and is closed. A living entity is damaged by the void, fire, suffocation and drowning where
// applicable, dispatches the contact hooks of the blocks it is inside of, and has its effects ticked before
// it is ticked as an Ent as usual.
func (e *LivingEnt) Tick(tx *world.Tx, current int64) {
	s := e.state()
	if s.dead {
		// The death animation and removal go out in the same tick: clients play the animation on the
		// removed entity themselves.
		e.PlayAction(DeathAction{})
		if h, ok := e.Behaviour().(LivingDeathHandler); ok {
			h.HandleDeath(e, tx, s.deathSrc)
		}
		_ = e.Close()
		return
	}
	if e.data.Pos[1] < float64(tx.Range()[0]) && current%10 == 0 {
		e.Hurt(4, VoidDamageSource{})
	}
	TickOnFire(e, tx)
	if current%10 == 0 && InsideOfSolid(e, tx) {
		e.Hurt(1, SuffocationDamageSource{})
	}

	_, waterBreathing := s.Effects.Effect(effect.WaterBreathing)
	_, conduitPower := s.Effects.Effect(effect.ConduitPower)
	drowned, changed := s.Air.Tick(waterBreathing || conduitPower || !Submerged(e, tx), s.Armour.Helmet())
	if drowned {
		e.Hurt(2, DrowningDamageSource{})
	}
	if changed {
		e.updateState()
	}

	CheckEntityInsiders(tx, e, e.H().Type().BBox(e).Translate(e.data.Pos))

	s.Effects.Tick(e, tx)
	e.Ent.Tick(tx, current)
}
