package player

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type shieldTestEntity struct {
	pos mgl64.Vec3
}

func (e shieldTestEntity) Close() error           { return nil }
func (e shieldTestEntity) H() *world.EntityHandle { return nil }
func (e shieldTestEntity) Position() mgl64.Vec3   { return e.pos }
func (e shieldTestEntity) Rotation() cube.Rotation {
	return cube.Rotation{}
}
func (e shieldTestEntity) HeldItems() (item.Stack, item.Stack) {
	return item.Stack{}, item.Stack{}
}

type shieldAxeAttacker struct {
	shieldTestEntity
	mainHand item.Stack
}

func (a shieldAxeAttacker) HeldItems() (item.Stack, item.Stack) {
	return a.mainHand, item.Stack{}
}

type shieldKnockBackAttacker struct {
	shieldTestEntity
	src           mgl64.Vec3
	force, height float64
}

func (a *shieldKnockBackAttacker) KnockBack(src mgl64.Vec3, force, height float64) {
	a.src, a.force, a.height = src, force, height
}

func newShieldTestPlayer(rot cube.Rotation, mainHand, offHand item.Stack) *Player {
	heldSlot := uint32(0)
	inv := inventory.New(36, nil)
	_ = inv.SetItem(0, mainHand)

	off := inventory.New(1, nil)
	_ = off.SetItem(0, offHand)

	return &Player{
		data: &world.EntityData{Rot: rot},
		playerData: &playerData{
			gameMode:  world.GameModeSurvival,
			h:         NopHandler{},
			inv:       inv,
			offHand:   off,
			heldSlot:  &heldSlot,
			sneaking:  true,
			cooldowns: map[string]time.Time{},
			health:    entity.NewHealthManager(20, 20),
			effects:   entity.NewEffectManager(),
			armour:    inventory.NewArmour(nil),
			hunger:    newHungerManager(),
		},
	}
}

func TestShieldBlockingRequiresSneakingReadyShieldAndStartupDelay(t *testing.T) {
	now := time.Unix(10, 0)
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))

	if changed := p.updateShieldBlockingState(now); !changed {
		t.Fatal("expected shield blocking state to start when sneaking with a shield")
	}
	if p.shieldBlockingAt(now.Add(shieldBlockDelay - time.Nanosecond)) {
		t.Fatal("expected shield to wait for the vanilla startup delay before blocking")
	}
	if !p.shieldBlockingAt(now.Add(shieldBlockDelay)) {
		t.Fatal("expected shield to block after the vanilla startup delay")
	}

	p.sneaking = false
	if p.shieldBlockingAt(now.Add(shieldBlockDelay)) {
		t.Fatal("expected shield not to block while not sneaking")
	}

	p.shieldBlockingInput = true
	if !p.shieldBlockingAt(now.Add(shieldBlockDelay)) {
		t.Fatal("expected shield to block while the shield input is held")
	}
	p.shieldBlockingInput = false

	p.sneaking = true
	p.cooldowns[shieldItemName] = now.Add(shieldDisableCooldown)
	if p.shieldBlockingAt(now.Add(shieldBlockDelay)) {
		t.Fatal("expected shield not to block while its item cooldown is active")
	}
}

func TestUseItemStartsShieldBlockingInput(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.sneaking = false

	p.UseItem()
	if !p.shieldBlockingInput {
		t.Fatal("expected using a held shield to start shield blocking input")
	}
	if !p.shieldBlockingAt(p.shieldBlockingSince.Add(shieldBlockDelay)) {
		t.Fatal("expected shield to block after use input and startup delay")
	}
}

func TestReleaseItemStopsShieldBlockingInput(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	p.shieldBlockingInput = true
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)

	p.ReleaseItem()
	if p.shieldBlockingInput {
		t.Fatal("expected releasing item to stop shield blocking input")
	}
	if p.Blocking() {
		t.Fatal("expected shield not to block after use input is released")
	}
}

func TestStartShieldBlockingInputHonoursUsePriority(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.NewStack(item.Bow{}, 1), item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	handler := &countingItemUseHandler{}
	p.h = handler

	if p.StartShieldBlockingInput() {
		t.Fatal("expected main-hand bow use to take priority over offhand shield blocking")
	}
	if p.shieldBlockingInput {
		t.Fatal("expected shield input to stay inactive while main-hand bow use has priority")
	}
	if handler.count != 0 {
		t.Fatalf("expected item-use handler not to run for non-shield-priority use, got %v calls", handler.count)
	}
}

func TestStartShieldBlockingInputHonoursShieldCooldown(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	p.cooldowns[shieldItemName] = time.Now().Add(shieldDisableCooldown)
	handler := &countingItemUseHandler{}
	p.h = handler

	if p.StartShieldBlockingInput() {
		t.Fatal("expected shield use not to start while shield is on cooldown")
	}
	if p.shieldBlockingInput {
		t.Fatal("expected shield input to stay inactive while shield is on cooldown")
	}
	if handler.count != 0 {
		t.Fatalf("expected item-use handler not to run while shield is on cooldown, got %v calls", handler.count)
	}
}

func TestStartShieldBlockingInputHonoursCancelledItemUse(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	p.h = cancellingItemUseHandler{}

	if p.StartShieldBlockingInput() {
		t.Fatal("expected cancelled item use not to start shield blocking input")
	}
	if p.shieldBlockingInput {
		t.Fatal("expected shield input to stay inactive after cancelled item use")
	}
}

func TestStopSneakingPreservesHeldShieldInput(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingInput = true
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)

	w := world.New()
	defer func() {
		_ = w.Close()
	}()
	<-w.Exec(func(tx *world.Tx) {
		p.tx = tx
		p.StopSneaking()
	})
	if !p.shieldBlockingInput {
		t.Fatal("expected stop sneaking not to clear held raw shield input")
	}
	if p.shieldBlockingSince.IsZero() {
		t.Fatal("expected shield warmup to be preserved while raw shield input is still held")
	}
}

func TestShieldBlocksOnlyFrontBlockableDamage(t *testing.T) {
	now := time.Unix(10, 0)
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = now.Add(-shieldBlockDelay)

	front := shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}
	behind := shieldTestEntity{pos: mgl64.Vec3{0, 0, -4}}

	if !p.shieldBlocksDamageAt(entity.AttackDamageSource{Attacker: front}, now) {
		t.Fatal("expected shield to block a front melee attack")
	}
	if p.shieldBlocksDamageAt(entity.AttackDamageSource{Attacker: behind}, now) {
		t.Fatal("expected shield not to block an attack from behind")
	}
	if p.shieldBlocksDamageAt(entity.FallDamageSource{}, now) {
		t.Fatal("expected shield not to block fall damage")
	}
	if !p.shieldBlocksDamageAt(enchantment.ThornsDamageSource{Owner: front}, now) {
		t.Fatal("expected shield to block front thorns damage")
	}
}

func TestShieldBlocksProjectileAndExplosionFromFront(t *testing.T) {
	now := time.Unix(10, 0)
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = now.Add(-shieldBlockDelay)

	projectile := shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}
	if !p.shieldBlocksDamageAt(entity.ProjectileDamageSource{Projectile: projectile}, now) {
		t.Fatal("expected shield to block a front projectile")
	}
	if !p.shieldBlocksDamageAt(entity.ExplosionDamageSource{Origin: mgl64.Vec3{0, 0, 4}, HasOrigin: true, BlockableByShield: true}, now) {
		t.Fatal("expected shield to block a front explosion")
	}
	if p.shieldBlocksDamageAt(entity.ExplosionDamageSource{}, now) {
		t.Fatal("expected shield not to block an explosion with no origin")
	}
	if p.shieldBlocksDamageAt(entity.ExplosionDamageSource{Origin: mgl64.Vec3{0, 0, 4}, HasOrigin: true}, now) {
		t.Fatal("expected shield not to block an explosion marked unblockable")
	}
}

func TestShieldDoesNotBlockSelfSourcedExplosion(t *testing.T) {
	now := time.Unix(10, 0)
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.handle = entity.NewText("", mgl64.Vec3{})
	p.shieldBlockingSince = now.Add(-shieldBlockDelay)

	src := entity.ExplosionDamageSource{
		Origin:            mgl64.Vec3{0, 0, 4},
		HasOrigin:         true,
		BlockableByShield: true,
		Source:            p,
	}
	if p.blockDamageWithShield(4, src) {
		t.Fatal("expected shield not to block a self-sourced explosion")
	}
}

func TestShieldDoesNotBlockCancelledDamage(t *testing.T) {
	now := time.Unix(10, 0)
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = now.Add(-shieldBlockDelay)
	p.h = cancellingHurtHandler{}

	front := shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}
	if _, vulnerable := p.Hurt(4, entity.AttackDamageSource{Attacker: front}); vulnerable {
		t.Fatal("expected cancelled damage not to be vulnerable")
	}
	_, offHand := p.HeldItems()
	if offHand.Durability() != offHand.MaxDurability() {
		t.Fatal("expected shield not to lose durability for damage cancelled by the hurt handler")
	}
}

func TestShieldBlocksZeroDamageProjectile(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)
	projectile := shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}
	handler := &shieldBlockTestHandler{}

	if dmg, vulnerable := p.Hurt(0, entity.ProjectileDamageSource{Projectile: projectile, ShieldBlockMarker: &handler.ProjectileShieldBlockMarker}); dmg != 0 || vulnerable {
		t.Fatalf("expected shield-blocked zero damage projectile to deal no vulnerable damage, got damage %v vulnerable %v", dmg, vulnerable)
	}
	if !handler.ShieldBlocked() {
		t.Fatal("expected zero damage projectile shield block callback to run")
	}
}

type shieldBlockTestHandler struct {
	entity.ProjectileShieldBlockMarker
}

func TestShieldDurabilityUsesDamageBeforeArmourReduction(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)
	p.armour.SetChestplate(item.NewStack(item.Chestplate{Tier: item.ArmourTierDiamond{}}, 1))
	attacker := shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}

	p.Hurt(4, entity.AttackDamageSource{Attacker: attacker})

	_, offHand := p.HeldItems()
	if got, want := offHand.Durability(), item.NewStack(item.Shield{}, 1).MaxDurability()-5; got != want {
		t.Fatalf("expected shield durability %v after blocking 4 raw damage, got %v", want, got)
	}
}

func TestExplosionKnockBackNotSuppressedByNestedShieldBlock(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.data.Pos = mgl64.Vec3{0, 0, 1}
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)
	p.h = &nestedShieldBlockHandler{
		player: p,
		src:    entity.AttackDamageSource{Attacker: shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}},
	}

	w := world.New()
	defer func() {
		_ = w.Close()
	}()
	<-w.Exec(func(tx *world.Tx) {
		p.tx = tx
		p.Explode(mgl64.Vec3{}, 0.2, block.ExplosionConfig{Size: 1, UnblockableByShield: true})
	})
	if p.Velocity().Len() == 0 {
		t.Fatal("expected unblockable explosion to knock back even if a nested hurt was shield-blocked")
	}
}

func TestShieldBlockedExplosionAppliesReducedKnockBack(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.data.Pos = mgl64.Vec3{0, 0, 1}
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)

	unblocked := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.Stack{})
	unblocked.data.Pos = p.data.Pos
	w := world.New()
	defer func() {
		_ = w.Close()
	}()
	explosionPos := mgl64.Vec3{0, 0, 4}
	<-w.Exec(func(tx *world.Tx) {
		p.tx = tx
		unblocked.tx = tx

		conf := block.ExplosionConfig{Size: 1}
		p.Explode(explosionPos, 0.2, conf)
		unblocked.Explode(explosionPos, 0.2, conf)
	})
	if p.Velocity().Len() == 0 {
		t.Fatal("expected shield-blocked explosion to still apply reduced knockback")
	}
	if p.Velocity().Len() >= unblocked.Velocity().Len() {
		t.Fatalf("expected shield-blocked explosion knockback %v to be less than unblocked knockback %v", p.Velocity(), unblocked.Velocity())
	}
}

type cancellingHurtHandler struct {
	NopHandler
}

func (cancellingHurtHandler) HandleHurt(ctx *Context, _ *float64, _ bool, _ *time.Duration, _ world.DamageSource) {
	ctx.Cancel()
}

type cancellingItemUseHandler struct {
	NopHandler
}

func (cancellingItemUseHandler) HandleItemUse(ctx *Context) {
	ctx.Cancel()
}

type countingItemUseHandler struct {
	NopHandler
	count int
}

func (h *countingItemUseHandler) HandleItemUse(*Context) {
	h.count++
}

type nestedShieldBlockHandler struct {
	NopHandler
	player *Player
	src    world.DamageSource
	done   bool
}

func (h *nestedShieldBlockHandler) HandleHurt(_ *Context, _ *float64, _ bool, _ *time.Duration, _ world.DamageSource) {
	if h.done {
		return
	}
	h.done = true
	h.player.Hurt(4, h.src)
}

func TestShieldDisableCooldownFromAxeAttack(t *testing.T) {
	attacker := shieldAxeAttacker{mainHand: item.NewStack(item.Axe{Tier: item.ToolTierWood}, 1)}
	cooldown, ok := shieldDisableCooldownFrom(entity.AttackDamageSource{Attacker: attacker})
	if !ok {
		t.Fatal("expected an axe attack to disable shields")
	}
	if cooldown != shieldDisableCooldown {
		t.Fatalf("expected shield disable cooldown %v, got %v", shieldDisableCooldown, cooldown)
	}

	if _, ok := shieldDisableCooldownFrom(entity.AttackDamageSource{Attacker: shieldTestEntity{}}); ok {
		t.Fatal("expected a non-axe attack not to disable shields")
	}
}

func TestShieldKnocksBackMeleeAttacker(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	attacker := &shieldKnockBackAttacker{shieldTestEntity: shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}}

	if !p.knockBackShieldAttacker(entity.AttackDamageSource{Attacker: attacker}) {
		t.Fatal("expected melee attacker to be knocked back")
	}
	if attacker.src != p.Position() {
		t.Fatalf("expected attacker to be knocked away from player position %v, got %v", p.Position(), attacker.src)
	}
	if attacker.force != shieldAttackerKnockBackForce || attacker.height != shieldAttackerKnockBackHeight {
		t.Fatalf("expected knockback %v/%v, got %v/%v", shieldAttackerKnockBackForce, shieldAttackerKnockBackHeight, attacker.force, attacker.height)
	}
	if p.knockBackShieldAttacker(entity.ProjectileDamageSource{Projectile: attacker}) {
		t.Fatal("expected projectile source not to knock back the attacker")
	}
}

func TestShieldDurabilityDamage(t *testing.T) {
	for _, test := range []struct {
		damage float64
		want   int
	}{
		{damage: 2.9, want: 0},
		{damage: 3, want: 4},
		{damage: 7.2, want: 8},
	} {
		if got := shieldDurabilityDamage(test.damage); got != test.want {
			t.Fatalf("shieldDurabilityDamage(%v) = %v, want %v", test.damage, got, test.want)
		}
	}
}

func TestShieldBlockingReadDoesNotClearExpiredCooldown(t *testing.T) {
	now := time.Now()
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = now.Add(-shieldBlockDelay)
	p.cooldowns[shieldItemName] = now.Add(-time.Second)

	if !p.Blocking() {
		t.Fatal("expected expired shield cooldown not to prevent blocking")
	}
	if _, ok := p.cooldowns[shieldItemName]; !ok {
		t.Fatal("expected shield blocking metadata read not to mutate cooldown state")
	}
}

func TestShouldAttemptShieldBlockWithHandlerMutatedDamage(t *testing.T) {
	for _, test := range []struct {
		name                string
		rawDamage           float64
		damageLeft          float64
		damageBeforeHandler float64
		src                 world.DamageSource
		want                bool
	}{
		{
			name:                "positive damage reduced to zero",
			rawDamage:           4,
			damageLeft:          0,
			damageBeforeHandler: 4,
			src:                 entity.AttackDamageSource{Attacker: shieldTestEntity{}},
		},
		{
			name:       "zero damage increased",
			damageLeft: 2,
			src:        entity.AttackDamageSource{Attacker: shieldTestEntity{}},
			want:       true,
		},
		{
			name:       "negative damage",
			rawDamage:  -1,
			damageLeft: -1,
			src:        entity.AttackDamageSource{Attacker: shieldTestEntity{}},
		},
		{
			name: "zero damage projectile",
			src:  entity.ProjectileDamageSource{Projectile: shieldTestEntity{}},
			want: true,
		},
	} {
		if got := shouldAttemptShieldBlock(test.rawDamage, test.damageLeft, test.damageBeforeHandler, test.src); got != test.want {
			t.Fatalf("%v: expected shouldAttemptShieldBlock to return %v, got %v", test.name, test.want, got)
		}
	}
}
