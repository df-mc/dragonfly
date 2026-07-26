package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestSweetBerriesSupportedSoil(t *testing.T) {
	tests := []struct {
		name string
		soil Soil
		want bool
	}{
		{name: "grass", soil: Grass{}, want: true},
		{name: "dirt", soil: Dirt{}, want: true},
		{name: "coarse dirt", soil: Dirt{Coarse: true}, want: true},
		{name: "podzol", soil: Podzol{}, want: true},
		{name: "mud", soil: Mud{}, want: true},
		{name: "muddy mangrove roots", soil: MuddyMangroveRoots{}, want: true},
		{name: "farmland", soil: Farmland{}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.soil.SoilFor(SweetBerries{}); got != test.want {
				t.Fatalf("SoilFor(SweetBerries{}) = %t, want %t", got, test.want)
			}
		})
	}
}

func TestSweetBerriesConsume(t *testing.T) {
	consumer := &sweetBerriesTestConsumer{}
	SweetBerries{}.Consume(nil, consumer)

	if consumer.food != 2 || consumer.saturation != 0.4 {
		t.Fatalf("Saturate(%d, %v), want Saturate(2, 0.4)", consumer.food, consumer.saturation)
	}
}

func TestSweetBerriesBoneMealPrecedesHarvest(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: redstoneBreakDropTestEntityRegistry()}.New()
	defer w.Close()

	pos := cube.Pos{0, 64, 0}
	user := &sweetBerriesTestConsumer{held: item.NewStack(item.BoneMeal{}, 1)}
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(pos, SweetBerries{Growth: 2}, nil)
		if activated := (SweetBerries{Growth: 2}).Activate(pos, cube.FaceUp, tx, user, &item.UseContext{}); activated {
			t.Fatal("expected bonemeal interaction to fall through to the held item")
		}
	})
}

func TestSweetBerriesHarvestPlaysPickSound(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: redstoneBreakDropTestEntityRegistry()}.New()
	defer w.Close()

	viewer := &sweetBerriesTestViewer{}
	loader := world.NewLoader(1, w, viewer)
	defer runWorld(w, func(tx *world.Tx) {
		loader.Close(tx)
	})

	pos := cube.Pos{0, 64, 0}
	runWorld(w, func(tx *world.Tx) {
		loader.Move(tx, pos.Vec3Centre())
		loader.Load(tx, 1)
		tx.SetBlock(pos, SweetBerries{Growth: 3}, nil)
		if ok := (SweetBerries{Growth: 3}).Activate(pos, cube.FaceUp, tx, nil, &item.UseContext{}); !ok {
			t.Fatal("expected mature sweet berry bush to be harvested")
		}
	})

	if len(viewer.sounds) != 1 {
		t.Fatalf("harvest played %d sounds, want 1", len(viewer.sounds))
	}
}

func TestSweetBerriesDamagePlaysHurtSound(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	viewer := &sweetBerriesTestViewer{}
	loader := world.NewLoader(1, w, viewer)
	defer runWorld(w, func(tx *world.Tx) {
		loader.Close(tx)
	})

	pos := cube.Pos{0, 64, 0}
	living := &sweetBerriesTestLiving{velocity: mgl64.Vec3{0.01, 0, 0}}
	runWorld(w, func(tx *world.Tx) {
		loader.Move(tx, pos.Vec3Centre())
		loader.Load(tx, 1)
		(SweetBerries{Growth: 1}).EntityInside(pos, tx, living)
	})

	if living.hurt != 1 {
		t.Fatalf("bush dealt damage %d times, want 1", living.hurt)
	}
	if living.fallResets != 1 {
		t.Fatalf("bush reset fall distance %d times, want 1", living.fallResets)
	}
	if len(viewer.sounds) != 1 {
		t.Fatalf("damage played %d sounds, want 1", len(viewer.sounds))
	}
}

func TestSweetBerriesSaplingDoesNotAffectEntities(t *testing.T) {
	velocity := mgl64.Vec3{0.01, -0.1, 0.01}
	living := &sweetBerriesTestLiving{velocity: velocity}

	SweetBerries{}.EntityInside(cube.Pos{}, nil, living)

	if living.velocity != velocity {
		t.Fatalf("sapling changed velocity from %v to %v", velocity, living.velocity)
	}
	if living.hurt != 0 || living.fallResets != 0 {
		t.Fatalf("sapling hurt entity %d times and reset fall distance %d times", living.hurt, living.fallResets)
	}
}

type sweetBerriesTestViewer struct {
	world.NopViewer
	sounds []world.Sound
}

func (v *sweetBerriesTestViewer) ViewSound(_ mgl64.Vec3, s world.Sound) {
	v.sounds = append(v.sounds, s)
}

type sweetBerriesTestLiving struct {
	redstoneTNTTestEntity
	velocity   mgl64.Vec3
	hurt       int
	fallResets int
}

func (e *sweetBerriesTestLiving) Velocity() mgl64.Vec3 {
	return e.velocity
}

func (e *sweetBerriesTestLiving) SetVelocity(velocity mgl64.Vec3) {
	e.velocity = velocity
}

func (e *sweetBerriesTestLiving) Hurt(float64, world.DamageSource) (float64, bool) {
	e.hurt++
	return 0.5, true
}

func (e *sweetBerriesTestLiving) ResetFallDistance() {
	e.fallResets++
}

func (*sweetBerriesTestLiving) FallDistance() float64 {
	return 0
}

type sweetBerriesTestConsumer struct {
	sweetBerriesTestLiving
	held       item.Stack
	food       int
	saturation float64
}

func (c *sweetBerriesTestConsumer) HeldItems() (item.Stack, item.Stack) {
	return c.held, item.Stack{}
}

func (c *sweetBerriesTestConsumer) SetHeldItems(mainHand, _ item.Stack) {
	c.held = mainHand
}

func (*sweetBerriesTestConsumer) UsingItem() bool          { return false }
func (*sweetBerriesTestConsumer) ReleaseItem()             {}
func (*sweetBerriesTestConsumer) UseItem()                 {}
func (*sweetBerriesTestConsumer) AddEffect(effect.Effect)  {}
func (*sweetBerriesTestConsumer) RemoveEffect(effect.Type) {}
func (*sweetBerriesTestConsumer) Effects() []effect.Effect { return nil }
func (*sweetBerriesTestConsumer) Absorption() float64      { return 0 }
func (*sweetBerriesTestConsumer) SetAbsorption(float64)    {}
func (c *sweetBerriesTestConsumer) Saturate(food int, saturation float64) {
	c.food, c.saturation = food, saturation
}
