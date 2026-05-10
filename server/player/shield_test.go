package player

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type shieldTestEntity struct {
	h   *world.EntityHandle
	pos mgl64.Vec3
}

func (e shieldTestEntity) Close() error           { return nil }
func (e shieldTestEntity) H() *world.EntityHandle { return e.h }
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

type shieldMarkedProjectile struct {
	shieldTestEntity
	marked bool
}

func (p *shieldMarkedProjectile) MarkShieldBlocked() {
	p.marked = true
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

func TestOffHandShieldBlockingDoesNotReportMainHandItemUse(t *testing.T) {
	now := time.Now()
	p := newShieldTestPlayer(cube.Rotation{}, item.NewStack(item.Bow{}, 1), item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = now.Add(-shieldBlockDelay)

	if !p.shieldBlockingAt(now) {
		t.Fatal("expected off-hand shield to block while sneaking")
	}
	if p.UsingItem() {
		t.Fatal("expected off-hand shield blocking not to report generic main-hand item use")
	}
}

func TestMainHandShieldBlockingReportsItemUse(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.NewStack(item.Shield{}, 1), item.Stack{})
	p.shieldBlockingSince = time.Now()

	if !p.UsingItem() {
		t.Fatal("expected main-hand shield blocking to report item use")
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
	if p.ShieldBlocking() {
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

func TestUseItemWithPriorityMainHandClearsHeldShieldInput(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.NewStack(item.Bow{}, 1), item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	p.shieldBlockingInput = true
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)
	_ = p.inv.SetItem(1, item.NewStack(item.Arrow{}, 1))

	w := world.New()
	defer func() {
		_ = w.Close()
	}()
	<-w.Exec(func(tx *world.Tx) {
		p.tx = tx
		p.UseItem()
	})
	if p.shieldBlockingInput {
		t.Fatal("expected priority main-hand use to clear held shield input")
	}
	if p.ShieldBlocking() {
		t.Fatal("expected priority main-hand use to stop shield blocking")
	}
}

func TestUseItemFallsBackToOffHandShieldWhenMainHandFoodCannotBeConsumed(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.NewStack(item.Apple{}, 1), item.NewStack(item.Shield{}, 1))
	p.sneaking = false

	p.UseItem()
	if !p.shieldBlockingInput {
		t.Fatal("expected off-hand shield input to start when full hunger prevents main-hand food use")
	}
}

func TestUseItemFallsBackToOffHandShieldWhenMainHandItemIsOnCooldown(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.NewStack(item.GoatHorn{}, 1), item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	p.SetCooldown(item.GoatHorn{}, time.Second)

	p.UseItem()
	if !p.shieldBlockingInput {
		t.Fatal("expected off-hand shield input to start when main-hand item is on cooldown")
	}
}

func TestUseItemWithCooldownOffHandShieldHonoursCancelledItemUse(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.NewStack(item.GoatHorn{}, 1), item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	p.SetCooldown(item.GoatHorn{}, time.Second)
	p.h = cancellingItemUseHandler{}

	p.UseItem()

	if p.shieldBlockingInput {
		t.Fatal("expected cancelled item use to prevent cooldown off-hand shield fallback")
	}
}

func TestUseItemDoesNotHandleAlreadyStartedShieldUseTwice(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	handler := &countingItemUseHandler{}
	p.h = handler

	if !p.StartShieldBlockingInput() {
		t.Fatal("expected auth input to start shield blocking")
	}
	p.UseItem()

	if handler.count != 1 {
		t.Fatalf("expected shield use handler to run once, got %v calls", handler.count)
	}
}

func TestSetHeldSlotWithPriorityMainHandClearsHeldShieldInput(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	p.shieldBlockingInput = true
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)
	_ = p.inv.SetItem(1, item.NewStack(item.Bow{}, 1))

	w := world.New()
	defer func() {
		_ = w.Close()
	}()
	<-w.Exec(func(tx *world.Tx) {
		p.tx = tx
		if err := p.SetHeldSlot(1); err != nil {
			t.Fatalf("expected held slot change to succeed: %v", err)
		}
	})
	if p.shieldBlockingInput {
		t.Fatal("expected priority main-hand slot change to clear held shield input")
	}
	if p.ShieldBlocking() {
		t.Fatal("expected priority main-hand slot change to stop shield blocking")
	}
}

func TestSetHeldItemsWithPriorityMainHandClearsHeldShieldInput(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	p.shieldBlockingInput = true
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)

	p.SetHeldItems(item.NewStack(item.Bow{}, 1), item.NewStack(item.Shield{}, 1))
	if p.shieldBlockingInput {
		t.Fatal("expected priority main-hand item update to clear held shield input")
	}
	if p.ShieldBlocking() {
		t.Fatal("expected priority main-hand item update to stop shield blocking")
	}
}

func TestSetHeldItemsClearsHandledShieldUseLatch(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	handler := &countingItemUseHandler{}
	p.h = handler

	if !p.StartShieldBlockingInput() {
		t.Fatal("expected auth input to start shield blocking")
	}
	p.SetHeldItems(item.NewStack(item.Bow{}, 1), item.NewStack(item.Shield{}, 1))
	p.SetHeldItems(item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.UseItem()

	if handler.count != 2 {
		t.Fatalf("expected item-use handler to run for next shield raise after held-item cancellation, got %v calls", handler.count)
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

func TestStartShieldBlockingInputRefreshesHeldItemsAfterHandler(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.sneaking = false
	p.h = &changingItemUseHandler{
		player:  p,
		main:    item.NewStack(item.Bow{}, 1),
		offHand: item.NewStack(item.Shield{}, 1),
	}

	if p.StartShieldBlockingInput() {
		t.Fatal("expected handler-swapped priority main-hand item to prevent shield blocking input")
	}
	if p.shieldBlockingInput {
		t.Fatal("expected shield input to stay inactive after handler gives main-hand use priority")
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
	h := world.EntitySpawnOpts{Position: mgl64.Vec3{0, 0, 4}}.New(entity.SnowballType, entity.ProjectileBehaviourConfig{})
	w := world.New()
	defer func() {
		_ = w.Close()
	}()
	var (
		dmg           float64
		vulnerable    bool
		shieldBlocked bool
	)
	<-w.Exec(func(tx *world.Tx) {
		p.tx = tx
		projectile := tx.AddEntity(h).(*entity.Ent)
		dmg, vulnerable = p.Hurt(0, entity.ProjectileDamageSource{Projectile: projectile})
		shieldBlocked = projectile.Behaviour().(*entity.ProjectileBehaviour).ShieldBlocked()
	})
	if dmg != 0 || vulnerable {
		t.Fatalf("expected shield-blocked zero damage projectile to deal no vulnerable damage, got damage %v vulnerable %v", dmg, vulnerable)
	}
	if !shieldBlocked {
		t.Fatal("expected zero damage projectile shield block marker to be set")
	}
}

func TestShieldBlocksProjectileDuringDamageImmunity(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)
	p.immuneUntil = time.Now().Add(time.Second)
	p.lastDamage = 10
	h := world.EntitySpawnOpts{Position: mgl64.Vec3{0, 0, 4}}.New(entity.SnowballType, entity.ProjectileBehaviourConfig{})
	w := world.New()
	defer func() {
		_ = w.Close()
	}()
	var (
		dmg           float64
		vulnerable    bool
		shieldBlocked bool
	)
	<-w.Exec(func(tx *world.Tx) {
		p.tx = tx
		projectile := tx.AddEntity(h).(*entity.Ent)
		dmg, vulnerable = p.Hurt(1, entity.ProjectileDamageSource{Projectile: projectile})
		shieldBlocked = projectile.Behaviour().(*entity.ProjectileBehaviour).ShieldBlocked()
	})
	if dmg != 0 || vulnerable {
		t.Fatalf("expected immune shield-blocked projectile to deal no vulnerable damage, got damage %v vulnerable %v", dmg, vulnerable)
	}
	if !shieldBlocked {
		t.Fatal("expected shield-blocked projectile to be marked even during damage immunity")
	}
}

func TestShieldBlocksCustomMarkedProjectile(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)
	projectile := &shieldMarkedProjectile{shieldTestEntity: shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}}

	if dmg, vulnerable := p.Hurt(1, entity.ProjectileDamageSource{Projectile: projectile}); dmg != 0 || vulnerable {
		t.Fatalf("expected shield-blocked custom projectile to deal no vulnerable damage, got damage %v vulnerable %v", dmg, vulnerable)
	}
	if !projectile.marked {
		t.Fatal("expected custom projectile shield block marker to be set")
	}
}

func TestShieldBlocksMeleeDuringDamageImmunity(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)
	p.immuneUntil = time.Now().Add(time.Second)
	p.lastDamage = 10
	attacker := &shieldKnockBackAttacker{shieldTestEntity: shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}}

	if dmg, vulnerable := p.Hurt(1, entity.AttackDamageSource{Attacker: attacker}); dmg != 0 || vulnerable {
		t.Fatalf("expected immune shield-blocked melee hit to deal no vulnerable damage, got damage %v vulnerable %v", dmg, vulnerable)
	}
	if attacker.force != shieldAttackerKnockBackForce {
		t.Fatal("expected shield-blocked melee attacker to be knocked back during damage immunity")
	}
}

func TestIgnoredImmuneHitDoesNotNotifyHurtHandler(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.Stack{})
	p.immuneUntil = time.Now().Add(time.Second)
	p.lastDamage = 10
	handler := &minimumDamageHurtHandler{}
	p.h = handler

	if dmg, vulnerable := p.Hurt(1, entity.AttackDamageSource{Attacker: shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}}); dmg != 0 || vulnerable {
		t.Fatalf("expected immune ignored hit to deal no damage, got damage %v vulnerable %v", dmg, vulnerable)
	}
	if handler.called {
		t.Fatal("expected fully ignored immune hit not to notify hurt handler")
	}
}

func TestImmuneHitZeroedByHandlerUpdatesAttackImmunity(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.Stack{})
	p.immuneUntil = time.Now().Add(time.Second)
	p.lastDamage = 1
	p.h = zeroDamageHurtHandler{}

	w := world.New()
	defer func() {
		_ = w.Close()
	}()
	<-w.Exec(func(tx *world.Tx) {
		p.tx = tx
		p.Hurt(4, entity.SuffocationDamageSource{})
	})

	if p.lastDamage != 4 {
		t.Fatalf("expected zeroed immune hit to refresh last damage to 4, got %v", p.lastDamage)
	}
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

func TestShieldDoesNotLoseDurabilityWhenDamageFullyReduced(t *testing.T) {
	p := newShieldTestPlayer(cube.Rotation{}, item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)
	p.effects.Add(effect.New(effect.Resistance, 5, time.Second), p)
	attacker := shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}

	w := world.New()
	defer func() {
		_ = w.Close()
	}()
	<-w.Exec(func(tx *world.Tx) {
		p.tx = tx
		p.Hurt(4, entity.AttackDamageSource{Attacker: attacker})
	})

	_, offHand := p.HeldItems()
	if got, want := offHand.Durability(), item.NewStack(item.Shield{}, 1).MaxDurability(); got != want {
		t.Fatalf("expected shield durability to remain %v after fully reduced damage, got %v", want, got)
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

type minimumDamageHurtHandler struct {
	NopHandler
	called bool
}

func (h *minimumDamageHurtHandler) HandleHurt(_ *Context, damage *float64, _ bool, _ *time.Duration, _ world.DamageSource) {
	h.called = true
	*damage = 1
}

type zeroDamageHurtHandler struct {
	NopHandler
}

func (zeroDamageHurtHandler) HandleHurt(_ *Context, damage *float64, _ bool, _ *time.Duration, _ world.DamageSource) {
	*damage = 0
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

type changingItemUseHandler struct {
	NopHandler
	player        *Player
	main, offHand item.Stack
}

func (h *changingItemUseHandler) HandleItemUse(*Context) {
	h.player.SetHeldItems(h.main, h.offHand)
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

func TestShieldDisableCooldownFromCustomAxeToolAttack(t *testing.T) {
	attacker := shieldAxeAttacker{mainHand: item.NewStack(shieldCustomAxeTool{}, 1)}
	cooldown, ok := shieldDisableCooldownFrom(entity.AttackDamageSource{Attacker: attacker})
	if !ok {
		t.Fatal("expected a custom axe tool attack to disable shields")
	}
	if cooldown != shieldDisableCooldown {
		t.Fatalf("expected shield disable cooldown %v, got %v", shieldDisableCooldown, cooldown)
	}
}

type shieldCustomAxeTool struct{}

func (shieldCustomAxeTool) EncodeItem() (string, int16) {
	return "dragonfly:shield_custom_axe", 0
}

func (shieldCustomAxeTool) ToolType() item.ToolType {
	return item.TypeAxe
}

func (shieldCustomAxeTool) HarvestLevel() int {
	return 0
}

func (shieldCustomAxeTool) BaseMiningEfficiency(world.Block) float64 {
	return 1
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

	if !p.ShieldBlocking() {
		t.Fatal("expected expired shield cooldown not to prevent blocking")
	}
	if _, ok := p.cooldowns[shieldItemName]; !ok {
		t.Fatal("expected shield blocking metadata read not to mutate cooldown state")
	}
}

func TestShouldAttemptShieldBlockBeforeHurtHandler(t *testing.T) {
	for _, test := range []struct {
		name      string
		rawDamage float64
		src       world.DamageSource
		want      bool
	}{
		{
			name:      "positive melee damage",
			rawDamage: 4,
			src:       entity.AttackDamageSource{Attacker: shieldTestEntity{}},
			want:      true,
		},
		{
			name: "zero damage projectile",
			src:  entity.ProjectileDamageSource{Projectile: shieldTestEntity{}},
			want: true,
		},
		{
			name: "zero damage melee",
			src:  entity.AttackDamageSource{Attacker: shieldTestEntity{}},
		},
		{
			name:      "negative damage",
			rawDamage: -1,
			src:       entity.AttackDamageSource{Attacker: shieldTestEntity{}},
		},
	} {
		if got := shouldAttemptShieldBlockBeforeHurtHandler(test.rawDamage, test.src); got != test.want {
			t.Fatalf("%v: expected shouldAttemptShieldBlockBeforeHurtHandler to return %v, got %v", test.name, test.want, got)
		}
	}
}

func TestShouldAttemptShieldBlockAfterHurtHandlerWithHandlerMutatedDamage(t *testing.T) {
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
			name:       "positive raw damage fully reduced",
			rawDamage:  4,
			damageLeft: 0,
			src:        entity.AttackDamageSource{Attacker: shieldTestEntity{}},
		},
		{
			name: "zero damage projectile",
			src:  entity.ProjectileDamageSource{Projectile: shieldTestEntity{}},
			want: true,
		},
	} {
		if got := shouldAttemptShieldBlockAfterHurtHandler(test.rawDamage, test.damageLeft, test.damageBeforeHandler, test.src); got != test.want {
			t.Fatalf("%v: expected shouldAttemptShieldBlockAfterHurtHandler to return %v, got %v", test.name, test.want, got)
		}
	}
}
