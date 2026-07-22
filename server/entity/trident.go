package entity

import (
	"math"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// NewTrident creates a thrown trident entity using the item.Stack passed.
func NewTrident(opts world.EntitySpawnOpts, owner world.Entity, it item.Stack) *world.EntityHandle {
	conf := TridentBehaviourConfig{Item: it}
	if owner != nil {
		conf.Owner = owner.H()
	}
	return opts.New(TridentType, conf)
}

// TridentBehaviourConfig allows the configuration of thrown tridents.
type TridentBehaviourConfig struct {
	// Owner is the entity that threw the trident.
	Owner *world.EntityHandle
	// Damage is the base damage dealt by the trident. Defaults to 8 if left
	// as 0.
	Damage float64
	// Item is the trident item.Stack the projectile was thrown with. The
	// enchantments on the stack, such as loyalty, channeling and impaling,
	// influence the behaviour of the trident.
	Item item.Stack
	// DisablePickup specifies if picking up the trident should be disabled.
	// This is the case for tridents thrown in creative mode.
	DisablePickup bool
	// CollisionPosition specifies the position of the block the trident is
	// stuck in. If non-empty, the trident will not move.
	CollisionPosition cube.Pos
}

func (conf TridentBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
}

// New creates a TridentBehaviour using conf.
func (conf TridentBehaviourConfig) New() *TridentBehaviour {
	if conf.Damage == 0 {
		conf.Damage = 8
	}
	proj := ProjectileBehaviourConfig{
		Owner:                 conf.Owner,
		Gravity:               0.05,
		Drag:                  0.01,
		Damage:                -1,
		SurviveBlockCollision: true,
		DisablePickup:         conf.DisablePickup,
		CollisionPosition:     conf.CollisionPosition,
	}
	if !conf.DisablePickup {
		proj.PickupItem = conf.Item
	}
	return &TridentBehaviour{ProjectileBehaviour: proj.New(), conf: conf}
}

// TridentBehaviour implements the behaviour of thrown tridents. Unlike most
// projectiles, tridents survive hitting an entity and may return to their
// owner if enchanted with loyalty.
type TridentBehaviour struct {
	*ProjectileBehaviour
	conf TridentBehaviourConfig

	dealtDamage bool
	returning   bool
	returnAge   int
}

// Returning returns true if the trident is currently returning to its owner
// as a result of the loyalty enchantment.
func (b *TridentBehaviour) Returning() bool {
	return b.returning
}

// Glint returns true if the trident stack carried by the entity is enchanted.
func (b *TridentBehaviour) Glint() bool {
	return len(b.conf.Item.Enchantments()) > 0
}

// Tick runs the tick-based behaviour of a TridentBehaviour and returns the
// Movement within the tick.
func (b *TridentBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	if b.close {
		_ = e.Close()
		return nil
	}
	if b.returning {
		return b.tickReturning(e, tx)
	}
	if b.collided && b.tickAttached(e, tx) {
		if b.loyaltyLevel() > 0 && b.ageCollided > 4 {
			b.startReturning(e, tx)
			return nil
		}
		if b.ageCollided > 1200 {
			b.close = true
		}
		return nil
	}
	vel := e.Velocity()
	m, result := b.tickMovement(e, tx)
	e.data.Pos, e.data.Vel, e.data.Rot = m.pos, m.vel, m.rot

	b.collisionPos, b.collided, b.ageCollided = cube.Pos{}, false, 0
	if result == nil {
		return m
	}

	switch r := result.(type) {
	case trace.EntityResult:
		if l, ok := r.Entity().(Living); ok {
			if !b.dealtDamage {
				b.hitEntity(l, e, tx, vel)
			}
			b.collidedEntities = append(b.collidedEntities, l.H())
		}
		// The trident deflects off the entity hit and drops to the ground.
		e.data.Vel = mgl64.Vec3{vel[0] * -0.01, vel[1] * -0.1, vel[2] * -0.01}
		if b.loyaltyLevel() > 0 {
			b.startReturning(e, tx)
		}
	case trace.BlockResult:
		bpos := r.BlockPosition()
		if h, ok := tx.Block(bpos).(block.ProjectileHitter); ok {
			h.ProjectileHit(bpos, tx, e, r.Face())
		}
		tx.PlaySound(result.Position(), sound.TridentHitGround{})
		b.hitBlockSurviving(e, r, m, tx)
	}
	return m
}

// hitEntity is called when the trident hits a Living entity. It deals damage
// to the entity, knocks it back and summons a lightning bolt if the trident
// is enchanted with channeling during a thunderstorm.
func (b *TridentBehaviour) hitEntity(l Living, e *Ent, tx *world.Tx, vel mgl64.Vec3) {
	b.dealtDamage = true
	owner, _ := b.conf.Owner.Entity(e.tx)

	dmg := b.conf.Damage
	if ench, ok := b.conf.Item.Enchantment(enchantment.Impaling); ok && Wet(l, tx) {
		dmg += enchantment.Impaling.Addend(ench.Level())
	}
	if _, vulnerable := l.Hurt(dmg, ProjectileDamageSource{Projectile: e, Owner: owner}); vulnerable {
		l.KnockBack(l.Position().Sub(vel), 0.45, 0.3608)
	}
	tx.PlaySound(e.Position(), sound.TridentHit{})

	if _, ok := b.conf.Item.Enchantment(enchantment.Channeling); ok && tx.ThunderingAt(cube.PosFromVec3(l.Position())) {
		tx.AddEntity(NewLightning(world.EntitySpawnOpts{Position: l.Position()}))
		tx.PlaySound(l.Position(), sound.TridentThunder{})
	}
}

// loyaltyLevel returns the level of the loyalty enchantment on the trident
// stack held by the entity, or 0 if it is not enchanted with loyalty.
func (b *TridentBehaviour) loyaltyLevel() int {
	if ench, ok := b.conf.Item.Enchantment(enchantment.Loyalty); ok {
		return ench.Level()
	}
	return 0
}

// startReturning makes the trident start returning to its owner. If the owner
// is not found in the world, the trident is dropped as an item instead.
func (b *TridentBehaviour) startReturning(e *Ent, tx *world.Tx) {
	if _, ok := b.conf.Owner.Entity(tx); !ok {
		b.drop(e, tx)
		return
	}
	b.returning = true
	b.collisionPos, b.collided, b.ageCollided = cube.Pos{}, false, 0

	tx.PlaySound(e.Position(), sound.TridentReturn{})
	for _, v := range tx.Viewers(e.Position()) {
		v.ViewEntityState(e)
	}
}

// tickReturning ticks the trident as it returns to its owner, accelerating
// towards the owner's eye position. The trident is dropped as an item if the
// owner is no longer available.
func (b *TridentBehaviour) tickReturning(e *Ent, tx *world.Tx) *Movement {
	owner, ok := b.conf.Owner.Entity(tx)
	living, alive := owner.(Living)
	if !ok || (alive && living.Dead()) || b.returnAge > 1200 {
		b.drop(e, tx)
		return nil
	}
	b.returnAge++

	pos, level := e.Position(), float64(b.loyaltyLevel())
	diff := EyePosition(owner).Sub(pos)
	if diff.Len() < 1 {
		b.pickUpReturned(e, tx, owner)
		return nil
	}
	pos[1] += diff[1] * 0.015 * level

	vel := e.Velocity().Mul(0.95).Add(diff.Normalize().Mul(0.05 * level))
	end := pos.Add(vel)
	rot := cube.Rotation{
		mgl64.RadToDeg(math.Atan2(vel[0], vel[2])),
		mgl64.RadToDeg(math.Atan2(vel[1], math.Hypot(vel[0], vel[2]))),
	}
	m := &Movement{v: tx.Viewers(e.data.Pos), e: e, pos: end, vel: vel, dpos: end.Sub(e.data.Pos), dvel: vel.Sub(e.data.Vel), rot: rot}
	e.data.Pos, e.data.Vel, e.data.Rot = end, vel, rot
	return m
}

// pickUpReturned makes the owner of the trident pick it up after it returned.
// If the owner has no space in its inventory, the trident is dropped as an
// item instead.
func (b *TridentBehaviour) pickUpReturned(e *Ent, tx *world.Tx, owner world.Entity) {
	if b.conf.DisablePickup || b.conf.Item.Empty() {
		_ = e.Close()
		return
	}
	collector, ok := owner.(Collector)
	if !ok {
		b.drop(e, tx)
		return
	}
	if n, _ := collector.Collect(b.conf.Item); n == 0 {
		b.drop(e, tx)
		return
	}
	for _, viewer := range tx.Viewers(e.Position()) {
		viewer.ViewEntityAction(e, PickedUpAction{Collector: collector})
	}
	_ = e.Close()
}

// drop drops the trident stack held by the entity as an item and closes the
// entity.
func (b *TridentBehaviour) drop(e *Ent, tx *world.Tx) {
	if !b.conf.DisablePickup && !b.conf.Item.Empty() {
		create := tx.World().EntityRegistry().Config().Item
		tx.AddEntity(create(world.EntitySpawnOpts{Position: e.Position()}, b.conf.Item))
	}
	_ = e.Close()
}

// Wet checks if the world.Entity passed is currently standing in water or
// exposed to rain.
func Wet(e world.Entity, tx *world.Tx) bool {
	pos := cube.PosFromVec3(e.Position())
	if tx.RainingAt(pos) {
		return true
	}
	if l, ok := tx.Liquid(pos); ok {
		_, isWater := l.(block.Water)
		return isWater
	}
	return false
}

// TridentType is a world.EntityType implementation for thrown tridents.
var TridentType tridentType

type tridentType struct{}

func (t tridentType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (tridentType) EncodeEntity() string { return "minecraft:thrown_trident" }
func (tridentType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.35, 0.125)
}

func (tridentType) DecodeNBT(m map[string]any, data *world.EntityData) {
	conf := TridentBehaviourConfig{
		Damage:            float64(nbtconv.Float32(m, "Damage")),
		Item:              nbtconv.MapItem(m, "Trident"),
		DisablePickup:     !nbtconv.Bool(m, "player"),
		CollisionPosition: nbtconv.Pos(m, "StuckToBlockPos"),
	}
	if conf.Item.Empty() {
		conf.Item = item.NewStack(item.Trident{}, 1)
	}
	data.Data = conf.New()
}

func (tridentType) EncodeNBT(data *world.EntityData) map[string]any {
	b := data.Data.(*TridentBehaviour)
	m := map[string]any{
		"Damage": float32(b.conf.Damage),
		"player": boolByte(!b.conf.DisablePickup),
	}
	if !b.conf.Item.Empty() {
		m["Trident"] = nbtconv.WriteItem(b.conf.Item, true)
	}
	if b.collided {
		m["StuckToBlockPos"] = nbtconv.PosToInt32Slice(b.collisionPos)
	}
	return m
}
