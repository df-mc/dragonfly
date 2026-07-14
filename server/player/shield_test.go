package player

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type shieldTestEntity struct {
	h   *world.EntityHandle
	pos mgl64.Vec3
}

func (e shieldTestEntity) Close() error                        { return nil }
func (e shieldTestEntity) H() *world.EntityHandle              { return e.h }
func (e shieldTestEntity) Position() mgl64.Vec3                { return e.pos }
func (e shieldTestEntity) Rotation() cube.Rotation             { return cube.Rotation{} }
func (e shieldTestEntity) HeldItems() (item.Stack, item.Stack) { return item.Stack{}, item.Stack{} }

type shieldTestAttacker struct {
	shieldTestEntity
	main          item.Stack
	src           mgl64.Vec3
	force, height float64
}

func (a *shieldTestAttacker) HeldItems() (item.Stack, item.Stack) { return a.main, item.Stack{} }
func (a *shieldTestAttacker) KnockBack(src mgl64.Vec3, force, height float64) {
	a.src, a.force, a.height = src, force, height
}

type shieldCustomAxe struct{}

func (shieldCustomAxe) EncodeItem() (string, int16)              { return "dragonfly:test_axe", 0 }
func (shieldCustomAxe) ToolType() item.ToolType                  { return item.TypeAxe }
func (shieldCustomAxe) HarvestLevel() int                        { return 0 }
func (shieldCustomAxe) BaseMiningEfficiency(world.Block) float64 { return 1 }

type cancellingHurtHandler struct{ NopHandler }

func (cancellingHurtHandler) HandleHurt(ctx *Context, _ *float64, _ bool, _ *time.Duration, _ world.DamageSource) {
	ctx.Cancel()
}

type countingHurtHandler struct {
	NopHandler
	called bool
}

func (h *countingHurtHandler) HandleHurt(*Context, *float64, bool, *time.Duration, world.DamageSource) {
	h.called = true
}

type cancellingItemUseHandler struct{ NopHandler }

func (cancellingItemUseHandler) HandleItemUse(ctx *Context) { ctx.Cancel() }

type countingItemUseHandler struct {
	NopHandler
	count int
}

func (h *countingItemUseHandler) HandleItemUse(*Context) { h.count++ }

type changingItemUseHandler struct {
	NopHandler
	p             *Player
	main, offHand item.Stack
}

func (h changingItemUseHandler) HandleItemUse(*Context) { h.p.SetHeldItems(h.main, h.offHand) }

type nestedShieldBlockHandler struct {
	NopHandler
	p    *Player
	done bool
}

func (h *nestedShieldBlockHandler) HandleHurt(*Context, *float64, bool, *time.Duration, world.DamageSource) {
	if !h.done {
		h.done = true
		h.p.Hurt(4, entity.AttackDamageSource{Attacker: shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}})
	}
}

func newShieldTestPlayer(mainHand, offHand item.Stack) *Player {
	heldSlot := uint32(0)
	inv := inventory.New(36, nil)
	_ = inv.SetItem(0, mainHand)
	off := inventory.New(1, nil)
	_ = off.SetItem(0, offHand)
	return &Player{
		data: &world.EntityData{},
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

func readyShieldPlayer() *Player {
	p := newShieldTestPlayer(item.Stack{}, item.NewStack(item.Shield{}, 1))
	p.shieldBlockingSince = time.Now().Add(-shieldBlockDelay)
	return p
}

func TestShieldActivation(t *testing.T) {
	t.Run("delay", func(t *testing.T) {
		now := time.Unix(10, 0)
		p := newShieldTestPlayer(item.Stack{}, item.NewStack(item.Shield{}, 1))
		p.updateShieldBlockingState(now)
		if p.shieldBlockingAt(now.Add(shieldBlockDelay - time.Nanosecond)) {
			t.Fatal("shield blocked before startup delay")
		}
		if !p.shieldBlockingAt(now.Add(shieldBlockDelay)) {
			t.Fatal("shield did not block after startup delay")
		}
	})

	tests := []struct {
		name  string
		setup func(*Player)
		want  bool
	}{
		{name: "sneaking", want: true},
		{name: "held input", setup: func(p *Player) { p.sneaking, p.shieldBlockingInput = false, true }, want: true},
		{name: "no input", setup: func(p *Player) { p.sneaking = false }},
		{name: "cooldown", setup: func(p *Player) { p.cooldowns[shieldItemName] = time.Now().Add(time.Second) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := readyShieldPlayer()
			if tt.setup != nil {
				tt.setup(p)
			}
			if got := p.ShieldBlocking(); got != tt.want {
				t.Fatalf("ShieldBlocking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShieldItemUse(t *testing.T) {
	tests := []struct {
		name  string
		main  item.Stack
		setup func(*Player)
		want  bool
	}{
		{name: "off hand shield", want: true},
		{name: "main hand priority", main: item.NewStack(item.Bow{}, 1)},
		{name: "cooldown", setup: func(p *Player) { p.cooldowns[shieldItemName] = time.Now().Add(time.Second) }},
		{name: "cancelled", setup: func(p *Player) { p.h = cancellingItemUseHandler{} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newShieldTestPlayer(tt.main, item.NewStack(item.Shield{}, 1))
			p.sneaking = false
			if tt.setup != nil {
				tt.setup(p)
			}
			if got := p.StartShieldBlockingInput(); got != tt.want {
				t.Fatalf("StartShieldBlockingInput() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("handler runs once", func(t *testing.T) {
		p := newShieldTestPlayer(item.Stack{}, item.NewStack(item.Shield{}, 1))
		p.sneaking = false
		h := &countingItemUseHandler{}
		p.h = h
		if !p.StartShieldBlockingInput() {
			t.Fatal("expected input to start shield use")
		}
		p.UseItem()
		if h.count != 1 {
			t.Fatalf("item-use handler called %v times, want 1", h.count)
		}
	})

	t.Run("fallbacks", func(t *testing.T) {
		for _, tt := range []struct {
			name      string
			main      item.Stack
			cancel    bool
			cooldown  bool
			wantInput bool
		}{
			{name: "full hunger", main: item.NewStack(item.Apple{}, 1), wantInput: true},
			{name: "main item cooldown", main: item.NewStack(item.GoatHorn{}, 1), cooldown: true, wantInput: true},
			{name: "cancelled fallback", main: item.NewStack(item.GoatHorn{}, 1), cooldown: true, cancel: true},
		} {
			t.Run(tt.name, func(t *testing.T) {
				p := newShieldTestPlayer(tt.main, item.NewStack(item.Shield{}, 1))
				p.sneaking = false
				if tt.cooldown {
					p.SetCooldown(item.GoatHorn{}, time.Second)
				}
				if tt.cancel {
					p.h = cancellingItemUseHandler{}
				}
				p.UseItem()
				if p.shieldBlockingInput != tt.wantInput {
					t.Fatalf("shield input = %v, want %v", p.shieldBlockingInput, tt.wantInput)
				}
			})
		}
	})

	t.Run("handler item change is refreshed", func(t *testing.T) {
		p := newShieldTestPlayer(item.Stack{}, item.NewStack(item.Shield{}, 1))
		p.sneaking = false
		p.h = changingItemUseHandler{p: p, main: item.NewStack(item.Bow{}, 1), offHand: item.NewStack(item.Shield{}, 1)}
		if p.StartShieldBlockingInput() || p.shieldBlockingInput {
			t.Fatal("handler-swapped priority item did not prevent shield input")
		}
	})

	t.Run("held item change resets latch", func(t *testing.T) {
		p := newShieldTestPlayer(item.Stack{}, item.NewStack(item.Shield{}, 1))
		p.sneaking = false
		h := &countingItemUseHandler{}
		p.h = h
		if !p.StartShieldBlockingInput() {
			t.Fatal("initial shield input did not start")
		}
		p.SetHeldItems(item.NewStack(item.Bow{}, 1), item.NewStack(item.Shield{}, 1))
		p.SetHeldItems(item.Stack{}, item.NewStack(item.Shield{}, 1))
		p.UseItem()
		if h.count != 2 {
			t.Fatalf("item-use handler called %v times, want 2", h.count)
		}
	})

	t.Run("release clears held input", func(t *testing.T) {
		p := readyShieldPlayer()
		p.sneaking, p.shieldBlockingInput = false, true
		p.ReleaseItem()
		if p.shieldBlockingInput || p.ShieldBlocking() {
			t.Fatal("release did not clear held shield input")
		}
	})

	t.Run("stop sneaking preserves held input", func(t *testing.T) {
		p := readyShieldPlayer()
		p.shieldBlockingInput = true
		w := world.Config{Synchronous: true}.New()
		defer w.Close()
		w.Do(func(tx *world.Tx) {
			p.tx = tx
			p.StopSneaking()
		})
		if !p.shieldBlockingInput || p.shieldBlockingSince.IsZero() {
			t.Fatal("stopping sneak discarded held shield input or warmup")
		}
	})
}

func TestShieldHeldItemChanges(t *testing.T) {
	t.Run("item replacement", func(t *testing.T) {
		p := readyShieldPlayer()
		p.sneaking, p.shieldBlockingInput = false, true
		p.SetHeldItems(item.NewStack(item.Bow{}, 1), item.NewStack(item.Shield{}, 1))
		if p.shieldBlockingInput || p.ShieldBlocking() {
			t.Fatal("priority main-hand item did not clear shield input")
		}
	})

	t.Run("slot change", func(t *testing.T) {
		p := readyShieldPlayer()
		p.sneaking, p.shieldBlockingInput = false, true
		_ = p.inv.SetItem(1, item.NewStack(item.Bow{}, 1))
		w := world.Config{Synchronous: true}.New()
		defer w.Close()
		w.Do(func(tx *world.Tx) {
			p.tx = tx
			if err := p.SetHeldSlot(1); err != nil {
				t.Fatal(err)
			}
		})
		if p.shieldBlockingInput || p.ShieldBlocking() {
			t.Fatal("priority slot did not clear shield input")
		}
	})
}

func TestShieldDamageOutcomes(t *testing.T) {
	front, back := mgl64.Vec3{0, 0, 4}, mgl64.Vec3{0, 0, -4}
	tests := []struct {
		name  string
		src   func(*Player) world.DamageSource
		setup func(*Player)
		want  world.HurtResult
	}{
		{name: "front melee", src: func(*Player) world.DamageSource {
			return entity.AttackDamageSource{Attacker: shieldTestEntity{pos: front}}
		}, want: world.HurtBlocked},
		{name: "back melee", src: func(*Player) world.DamageSource {
			return entity.AttackDamageSource{Attacker: shieldTestEntity{pos: back}}
		}, want: world.HurtDamaged},
		{name: "projectile", src: func(*Player) world.DamageSource {
			return entity.ProjectileDamageSource{Projectile: shieldTestEntity{pos: front}}
		}, want: world.HurtBlocked},
		{name: "zero-damage projectile", src: func(*Player) world.DamageSource {
			return entity.ProjectileDamageSource{Projectile: shieldTestEntity{pos: front}}
		}, setup: func(p *Player) { p.lastDamage = 0 }, want: world.HurtBlocked},
		{name: "immune projectile", src: func(*Player) world.DamageSource {
			return entity.ProjectileDamageSource{Projectile: shieldTestEntity{pos: front}}
		}, setup: func(p *Player) { p.immuneUntil, p.lastDamage = time.Now().Add(time.Second), 10 }, want: world.HurtBlocked},
		{name: "blockable explosion", src: func(*Player) world.DamageSource {
			return entity.ExplosionDamageSource{Origin: front, HasOrigin: true, BlockableByShield: true}
		}, want: world.HurtBlocked},
		{name: "unblockable explosion", src: func(*Player) world.DamageSource { return entity.ExplosionDamageSource{Origin: front, HasOrigin: true} }, want: world.HurtDamaged},
		{name: "self-sourced explosion", src: func(p *Player) world.DamageSource {
			return entity.ExplosionDamageSource{Origin: front, HasOrigin: true, BlockableByShield: true, Source: p}
		}, setup: func(p *Player) { p.handle = entity.NewText("", mgl64.Vec3{}) }, want: world.HurtDamaged},
		{name: "cancelled", src: func(*Player) world.DamageSource {
			return entity.AttackDamageSource{Attacker: shieldTestEntity{pos: front}}
		}, setup: func(p *Player) { p.h = cancellingHurtHandler{} }, want: world.HurtCancelled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := readyShieldPlayer()
			if tt.setup != nil {
				tt.setup(p)
			}
			w := world.Config{Synchronous: true}.New()
			defer w.Close()
			var result world.HurtResult
			w.Do(func(tx *world.Tx) {
				p.tx = tx
				damage := 4.0
				if tt.name == "zero-damage projectile" {
					damage = 0
				}
				_, result = p.Hurt(damage, tt.src(p))
			})
			if result != tt.want {
				t.Fatalf("Hurt() result = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestShieldHitEffects(t *testing.T) {
	p := readyShieldPlayer()
	attacker := &shieldTestAttacker{
		shieldTestEntity: shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}},
		main:             item.NewStack(shieldCustomAxe{}, 1),
	}
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	var result world.HurtResult
	w.Do(func(tx *world.Tx) {
		p.tx = tx
		_, result = p.Hurt(4, entity.AttackDamageSource{Attacker: attacker})
	})
	if !result.Blocked() {
		t.Fatalf("Hurt() result = %v, want blocked", result)
	}
	_, shield := p.HeldItems()
	if got, want := shield.Durability(), shield.MaxDurability()-5; got != want {
		t.Fatalf("shield durability = %v, want %v", got, want)
	}
	if !p.HasCooldown(item.Shield{}) {
		t.Fatal("axe hit did not disable shield")
	}
	if attacker.force != shieldAttackerKnockBackForce || attacker.height != shieldAttackerKnockBackHeight {
		t.Fatalf("attacker knockback = %v/%v", attacker.force, attacker.height)
	}

	for _, tt := range []struct {
		damage float64
		want   int
	}{{2.9, 0}, {3, 4}, {7.2, 8}} {
		if got := shieldDurabilityDamage(tt.damage); got != tt.want {
			t.Errorf("shieldDurabilityDamage(%v) = %v, want %v", tt.damage, got, tt.want)
		}
	}

	t.Run("ignored immune melee has no shield effects", func(t *testing.T) {
		p := readyShieldPlayer()
		p.immuneUntil, p.lastDamage = time.Now().Add(time.Second), 10
		h := &countingHurtHandler{}
		p.h = h
		attacker := &shieldTestAttacker{shieldTestEntity: shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}}
		if damage, result := p.Hurt(1, entity.AttackDamageSource{Attacker: attacker}); damage != 0 || result != world.HurtImmune {
			t.Fatalf("Hurt() = (%v, %v), want (0, immune)", damage, result)
		}
		_, shield := p.HeldItems()
		if h.called || attacker.force != 0 || shield.Durability() != shield.MaxDurability() {
			t.Fatalf("immune hit produced effects: handler=%v knockback=%v durability=%v", h.called, attacker.force, shield.Durability())
		}
	})

	t.Run("durability follows effective pre-shield damage", func(t *testing.T) {
		for _, tt := range []struct {
			name  string
			setup func(*Player)
			want  int
		}{
			{name: "armour uses raw damage", setup: func(p *Player) {
				p.armour.SetChestplate(item.NewStack(item.Chestplate{Tier: item.ArmourTierDiamond{}}, 1))
			}, want: 5},
			{name: "fully reduced", setup: func(p *Player) {
				p.effects.Add(effect.New(effect.Resistance, 5, time.Second), p)
			}},
		} {
			t.Run(tt.name, func(t *testing.T) {
				p := readyShieldPlayer()
				tt.setup(p)
				w := world.Config{Synchronous: true}.New()
				defer w.Close()
				w.Do(func(tx *world.Tx) {
					p.tx = tx
					p.Hurt(4, entity.AttackDamageSource{Attacker: shieldTestEntity{pos: mgl64.Vec3{0, 0, 4}}})
				})
				_, shield := p.HeldItems()
				if lost := shield.MaxDurability() - shield.Durability(); lost != tt.want {
					t.Fatalf("shield durability loss = %v, want %v", lost, tt.want)
				}
			})
		}
	})
}

func TestShieldExplosionKnockBack(t *testing.T) {
	for _, immune := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal", true: "immune"}[immune], func(t *testing.T) {
			blocked := readyShieldPlayer()
			unblocked := newShieldTestPlayer(item.Stack{}, item.Stack{})
			blocked.data.Pos, unblocked.data.Pos = mgl64.Vec3{0, 0, 1}, mgl64.Vec3{0, 0, 1}
			if immune {
				blocked.immuneUntil, unblocked.immuneUntil = time.Now().Add(time.Second), time.Now().Add(time.Second)
				blocked.lastDamage, unblocked.lastDamage = 10, 10
			}
			w := world.Config{Synchronous: true}.New()
			defer w.Close()
			w.Do(func(tx *world.Tx) {
				blocked.tx, unblocked.tx = tx, tx
				conf := block.ExplosionConfig{Size: 1}
				blocked.Explode(mgl64.Vec3{0, 0, 4}, 0.2, conf)
				unblocked.Explode(mgl64.Vec3{0, 0, 4}, 0.2, conf)
			})
			if blocked.Velocity().Len() == 0 || blocked.Velocity().Len() >= unblocked.Velocity().Len() {
				t.Fatalf("blocked knockback %v, unblocked %v", blocked.Velocity(), unblocked.Velocity())
			}
		})
	}

	t.Run("nested block does not suppress unrelated explosion", func(t *testing.T) {
		p := readyShieldPlayer()
		p.data.Pos = mgl64.Vec3{0, 0, 1}
		p.h = &nestedShieldBlockHandler{p: p}
		w := world.Config{Synchronous: true}.New()
		defer w.Close()
		w.Do(func(tx *world.Tx) {
			p.tx = tx
			p.Explode(mgl64.Vec3{}, 0.2, block.ExplosionConfig{Size: 1, UnblockableByShield: true})
		})
		if p.Velocity().Len() == 0 {
			t.Fatal("nested shield block suppressed unrelated explosion knockback")
		}
	})
}
