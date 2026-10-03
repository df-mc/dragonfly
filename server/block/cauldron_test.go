package block_test

import (
	"context"
	"fmt"
	"image/color"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/biome"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func TestCauldronBucketInteractions(t *testing.T) {
	water := item.Bucket{Content: item.LiquidBucketContent(block.Water{Depth: 8})}
	lava := item.Bucket{Content: item.LiquidBucketContent(block.Lava{Depth: 8})}
	snow := item.Bucket{Content: item.PowderSnowBucketContent()}
	for _, test := range []struct {
		name     string
		before   block.Cauldron
		bucket   item.Bucket
		after    block.Cauldron
		returned item.Bucket
		used     bool
	}{
		{name: "fill water", bucket: water, after: block.Cauldron{Level: 6}, used: true},
		{name: "fill lava", bucket: lava, after: block.Cauldron{Level: 6, Content: block.LavaCauldronContent()}, used: true},
		{name: "fill powder snow", bucket: snow, after: block.Cauldron{Level: 6, Content: block.PowderSnowCauldronContent()}, used: true},
		{name: "collect water", before: block.Cauldron{Level: 6}, returned: water, used: true},
		{name: "collect lava", before: block.Cauldron{Level: 6, Content: block.LavaCauldronContent()}, returned: lava, used: true},
		{name: "collect powder snow", before: block.Cauldron{Level: 6, Content: block.PowderSnowCauldronContent()}, returned: snow, used: true},
		{name: "partial water cannot fill bucket", before: block.Cauldron{Level: 5}, after: block.Cauldron{Level: 5}},
		{name: "empty cannot fill bucket"},
		{name: "already full water", before: block.Cauldron{Level: 6}, bucket: water, after: block.Cauldron{Level: 6}, used: true},
		{name: "clear dye with water bucket", before: block.Cauldron{Level: 6, Colour: color.RGBA{R: 255, A: 255}}, bucket: water, after: block.Cauldron{Level: 6}, used: true},
		{name: "water and lava mix", before: block.Cauldron{Level: 6}, bucket: lava, used: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, ctx, used := activateCauldron(t, test.before, item.NewStack(test.bucket, 1))
			if used != test.used || !reflect.DeepEqual(got, test.after) {
				t.Fatalf("interaction = (%#v, %v), want (%#v, %v)", got, used, test.after, test.used)
			}
			if !used {
				if ctx.CountSub != 0 || !ctx.NewItem.Empty() {
					t.Fatalf("unsuccessful interaction changed items: %#v", ctx)
				}
				return
			}
			assertCauldronReturnedItem(t, ctx, test.returned, 1)
			if ctx.CountSub != 1 || !ctx.NewItemSurvivalOnly {
				t.Fatalf("bucket exchange must consume one bucket only in survival: %#v", ctx)
			}
		})
	}
}

func TestCauldronBottleLevels(t *testing.T) {
	for _, level := range []int{0, 2, 4, 6} {
		t.Run(fmt.Sprintf("add at %d", level), func(t *testing.T) {
			got, ctx, used := activateCauldron(t, block.Cauldron{Level: level}, item.NewStack(item.Potion{Type: potion.Water()}, 1))
			if !used || got.Level != min(level+2, 6) || got.Potion != nil {
				t.Fatalf("add water bottle = (%#v, %v), want plain water level %d", got, used, min(level+2, 6))
			}
			assertCauldronReturnedItem(t, ctx, item.GlassBottle{}, 1)
			if ctx.CountSub != 1 {
				t.Fatalf("water bottle consumption = %d, want 1", ctx.CountSub)
			}
		})
	}
	for _, level := range []int{2, 4, 6} {
		t.Run(fmt.Sprintf("take at %d", level), func(t *testing.T) {
			got, ctx, used := activateCauldron(t, block.Cauldron{Level: level}, item.NewStack(item.GlassBottle{}, 1))
			if !used || got.Level != level-2 {
				t.Fatalf("take water bottle = (%#v, %v), want level %d", got, used, level-2)
			}
			assertCauldronReturnedItem(t, ctx, item.Potion{Type: potion.Water()}, 1)
			if ctx.CountSub != 1 {
				t.Fatalf("glass bottle consumption = %d, want 1", ctx.CountSub)
			}
		})
	}
	for _, test := range []struct {
		name string
		c    block.Cauldron
		i    world.Item
	}{
		{name: "empty", i: item.GlassBottle{}},
		{name: "insufficient water", c: block.Cauldron{Level: 1}, i: item.GlassBottle{}},
		{name: "lava", c: block.Cauldron{Level: 6, Content: block.LavaCauldronContent()}, i: item.GlassBottle{}},
		{name: "powder snow", c: block.Cauldron{Level: 6, Content: block.PowderSnowCauldronContent()}, i: item.GlassBottle{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, ctx, used := activateCauldron(t, test.c, item.NewStack(test.i, 1))
			if used || !reflect.DeepEqual(got, test.c) || ctx.CountSub != 0 || !ctx.NewItem.Empty() {
				t.Fatalf("invalid bottle interaction changed block/items: block=%#v used=%v context=%#v", got, used, ctx)
			}
		})
	}
}

func TestCauldronBottleDyedWater(t *testing.T) {
	colour := color.RGBA{R: 200, G: 50, B: 80, A: 255}
	got, ctx, used := activateCauldron(t, block.Cauldron{Level: 3, Colour: colour}, item.NewStack(item.GlassBottle{}, 1))
	if !used || got.Level != 1 || got.Colour != colour {
		t.Fatalf("bottle dyed water = (%#v, %v), want level 1 retaining dye", got, used)
	}
	assertCauldronReturnedItem(t, ctx, item.Potion{Type: potion.Water()}, 1)
	got, _, used = activateCauldron(t, block.Cauldron{Level: 2, Colour: colour}, item.NewStack(item.GlassBottle{}, 1))
	if !used || !reflect.DeepEqual(got, block.Cauldron{}) {
		t.Fatalf("bottle last dyed water = (%#v, %v), want empty plain cauldron", got, used)
	}
	got, ctx, used = activateCauldron(t, block.Cauldron{Level: 3, Colour: colour}, item.NewStack(item.Potion{Type: potion.Water()}, 1))
	if !used || got.Level != 5 || got.Colour != (color.RGBA{}) || ctx.CountSub != 1 {
		t.Fatalf("add water to dyed water = (%#v, %v, %#v), want plain water level 5", got, used, ctx)
	}
}

func TestCauldronPotionForms(t *testing.T) {
	for _, p := range []world.Item{
		item.Potion{Type: potion.Healing()},
		item.SplashPotion{Type: potion.Healing()},
		item.LingeringPotion{Type: potion.Healing()},
	} {
		name, _ := p.EncodeItem()
		t.Run(name, func(t *testing.T) {
			filled, ctx, used := activateCauldron(t, block.Cauldron{}, item.NewStack(p, 1))
			if !used || filled.Level != 2 || !reflect.DeepEqual(filled.Potion, p) {
				t.Fatalf("fill potion = (%#v, %v), want level 2 with %#v", filled, used, p)
			}
			assertCauldronReturnedItem(t, ctx, item.GlassBottle{}, 1)
			empty, ctx, used := activateCauldron(t, filled, item.NewStack(item.GlassBottle{}, 1))
			if !used || !reflect.DeepEqual(empty, block.Cauldron{}) {
				t.Fatalf("take potion = (%#v, %v), want empty cauldron", empty, used)
			}
			assertCauldronReturnedItem(t, ctx, p, 1)
		})
	}
	for _, incoming := range []world.Item{
		item.Potion{Type: potion.Poison()},
		item.Potion{Type: potion.Water()},
	} {
		name, meta := incoming.EncodeItem()
		t.Run(fmt.Sprintf("mix %s %d", name, meta), func(t *testing.T) {
			got, ctx, used := activateCauldron(t, block.Cauldron{Level: 2, Potion: item.Potion{Type: potion.Healing()}}, item.NewStack(incoming, 1))
			if !used || !reflect.DeepEqual(got, block.Cauldron{}) {
				t.Fatalf("incompatible potion mix = (%#v, %v), want empty cauldron", got, used)
			}
			assertCauldronReturnedItem(t, ctx, item.GlassBottle{}, 1)
			if ctx.CountSub != 1 {
				t.Fatalf("incompatible potion consumption = %d, want 1", ctx.CountSub)
			}
		})
	}
	t.Run("same potion changes delivery form", func(t *testing.T) {
		incoming := item.SplashPotion{Type: potion.Healing()}
		got, ctx, used := activateCauldron(t, block.Cauldron{Level: 2, Potion: item.Potion{Type: potion.Healing()}}, item.NewStack(incoming, 1))
		if !used || got.Level != 4 || !reflect.DeepEqual(got.Potion, incoming) || ctx.CountSub != 1 {
			t.Fatalf("same potion different form = (%#v, %v, %#v), want splash potion level 4", got, used, ctx)
		}
	})
	t.Run("full potion handles use without consumption", func(t *testing.T) {
		before := block.Cauldron{Level: 6, Potion: item.Potion{Type: potion.Healing()}}
		got, ctx, used := activateCauldron(t, before, item.NewStack(item.SplashPotion{Type: potion.Healing()}, 1))
		if !used || !reflect.DeepEqual(got, before) || ctx.CountSub != 0 || !ctx.NewItem.Empty() {
			t.Fatalf("full potion interaction = (%#v, %v, %#v), want unchanged handled interaction", got, used, ctx)
		}
	})
	t.Run("water forms store plain water", func(t *testing.T) {
		for _, water := range []world.Item{item.Potion{Type: potion.Water()}, item.SplashPotion{Type: potion.Water()}, item.LingeringPotion{Type: potion.Water()}} {
			got, _, used := activateCauldron(t, block.Cauldron{}, item.NewStack(water, 1))
			if !used || got.Level != 2 || got.Potion != nil {
				t.Fatalf("fill with %#v = (%#v, %v), want plain water level 2", water, got, used)
			}
		}
	})
}

func TestCauldronArrowQuantities(t *testing.T) {
	for _, test := range []struct {
		level, count, tipped, remaining int
	}{
		{1, 64, 16, 0},
		{2, 64, 16, 0},
		{3, 64, 16, 0},
		{4, 64, 32, 0},
		{5, 64, 48, 0},
		{6, 64, 64, 0},
		{6, 1, 1, 5},
		{6, 17, 17, 4},
	} {
		t.Run(fmt.Sprintf("level %d count %d", test.level, test.count), func(t *testing.T) {
			got, ctx, used := activateCauldron(t, block.Cauldron{Level: test.level, Potion: item.Potion{Type: potion.Poison()}}, item.NewStack(item.Arrow{}, test.count))
			if !used || got.Level != test.remaining {
				t.Fatalf("tip arrows = (%#v, %v), want level %d", got, used, test.remaining)
			}
			if got.Level == 0 && got.Potion != nil {
				t.Fatalf("emptied cauldron retained potion: %#v", got)
			}
			assertCauldronReturnedItem(t, ctx, item.Arrow{Tip: potion.Poison()}, test.tipped)
			if ctx.CountSub != test.tipped {
				t.Fatalf("consumed %d arrows, want %d", ctx.CountSub, test.tipped)
			}
		})
	}
}

func TestCauldronLeatherArmourPreservesMetadata(t *testing.T) {
	colour := color.RGBA{R: 40, G: 120, B: 200, A: 255}
	trim := item.ArmourTrim{Template: item.TemplateSentry(), Material: item.GoldIngot{}}
	for _, wash := range []bool{false, true} {
		initial, target := color.RGBA{}, colour
		if wash {
			initial, target = colour, color.RGBA{}
		}
		for _, armour := range []world.Item{
			item.Helmet{Tier: item.ArmourTierLeather{Colour: initial}, Trim: trim},
			item.Chestplate{Tier: item.ArmourTierLeather{Colour: initial}, Trim: trim},
			item.Leggings{Tier: item.ArmourTierLeather{Colour: initial}, Trim: trim},
			item.Boots{Tier: item.ArmourTierLeather{Colour: initial}, Trim: trim},
		} {
			name, _ := armour.EncodeItem()
			t.Run(fmt.Sprintf("%s wash=%v", name, wash), func(t *testing.T) {
				c := block.Cauldron{Level: 4}
				if !wash {
					c.Colour = colour
				}
				held := item.NewStack(armour, 1).
					WithDurability(21).
					WithCustomName("A traveller's armour").
					WithLore("Keep this inscription").
					WithValue("owner", "traveller").
					WithEnchantments(item.NewEnchantment(enchantment.Protection, 2)).
					WithAnvilCost(3).
					AsUnbreakable()
				got, ctx, used := activateCauldron(t, c, held)
				if !used || got.Level != 3 || got.Colour != c.Colour {
					t.Fatalf("armour interaction = (%#v, %v), want one water level consumed", got, used)
				}
				if !ctx.NewItemReplaceHeld || ctx.NewItemSurvivalOnly || ctx.CountSub != 0 {
					t.Fatalf("armour must replace held item in all game modes: %#v", ctx)
				}
				var gotTier item.ArmourTier
				var gotTrim item.ArmourTrim
				switch a := ctx.NewItem.Item().(type) {
				case item.Helmet:
					gotTier, gotTrim = a.Tier, a.Trim
				case item.Chestplate:
					gotTier, gotTrim = a.Tier, a.Trim
				case item.Leggings:
					gotTier, gotTrim = a.Tier, a.Trim
				case item.Boots:
					gotTier, gotTrim = a.Tier, a.Trim
				default:
					t.Fatalf("returned armour has type %T", ctx.NewItem.Item())
				}
				if !reflect.DeepEqual(gotTier, item.ArmourTierLeather{Colour: target}) || !reflect.DeepEqual(gotTrim, trim) {
					t.Fatalf("armour tier/trim = (%#v, %#v), want colour %#v and original trim", gotTier, gotTrim, target)
				}
				assertCauldronStackMetadata(t, ctx.NewItem, held)
				if ctx.NewItem.Durability() != held.Durability() || !ctx.NewItem.Unbreakable() || ctx.NewItem.AnvilCost() != held.AnvilCost() || !reflect.DeepEqual(ctx.NewItem.Enchantments(), held.Enchantments()) {
					t.Fatalf("armour lost durability, enchantments or repair metadata: %#v", ctx.NewItem)
				}
			})
		}
	}
}

func TestCauldronBannerRemovesLastPattern(t *testing.T) {
	patterns := []block.BannerPatternLayer{
		{Type: block.BorderBannerPattern(), Colour: item.ColourRed()},
		{Type: block.CrossBannerPattern(), Colour: item.ColourBlue()},
	}
	held := item.NewStack(block.Banner{Colour: item.ColourWhite(), Patterns: patterns}, 2).
		WithCustomName("Company standard").WithLore("Two woven patterns").WithValue("owner", "company")
	got, ctx, used := activateCauldron(t, block.Cauldron{Level: 2}, held)
	if !used || got.Level != 1 || ctx.CountSub != 1 || ctx.NewItem.Count() != 1 {
		t.Fatalf("wash banner = (%#v, %v, %#v), want one banner and one water level consumed", got, used, ctx)
	}
	banner, ok := ctx.NewItem.Item().(block.Banner)
	if !ok || banner.Colour != item.ColourWhite() || !reflect.DeepEqual(banner.Patterns, patterns[:1]) {
		t.Fatalf("washed banner = %#v, want only original first pattern", banner)
	}
	if original := held.Item().(block.Banner); len(original.Patterns) != 2 || original.Patterns[1] != patterns[1] {
		t.Fatalf("washing mutated remaining banner stack: %#v", original)
	}
	assertCauldronStackMetadata(t, ctx.NewItem, held)
}

func TestCauldronDyeMixesWater(t *testing.T) {
	red, blue := item.ColourRed().RGBA(), item.ColourBlue().RGBA()
	c, ctx, used := activateCauldron(t, block.Cauldron{Level: 5}, item.NewStack(item.Dye{Colour: item.ColourRed()}, 1))
	if !used || c.Level != 5 || c.Colour != red || ctx.CountSub != 1 {
		t.Fatalf("dye water = (%#v, %v, %#v), want red without consuming water", c, used, ctx)
	}
	got, ctx, used := activateCauldron(t, c, item.NewStack(item.Dye{Colour: item.ColourBlue()}, 1))
	want := color.RGBA{
		R: uint8((uint16(red.R) + uint16(blue.R)) / 2),
		G: uint8((uint16(red.G) + uint16(blue.G)) / 2),
		B: uint8((uint16(red.B) + uint16(blue.B)) / 2),
		A: 255,
	}
	if !used || got.Level != 5 || got.Colour != want || ctx.CountSub != 1 {
		t.Fatalf("mix dyes = (%#v, %v, %#v), want colour %#v", got, used, ctx, want)
	}
	for _, c := range []block.Cauldron{
		{},
		{Level: 6, Content: block.LavaCauldronContent()},
		{Level: 6, Content: block.PowderSnowCauldronContent()},
		{Level: 2, Potion: item.Potion{Type: potion.Healing()}},
	} {
		got, ctx, used := activateCauldron(t, c, item.NewStack(item.Dye{Colour: item.ColourRed()}, 1))
		if used || !reflect.DeepEqual(got, c) || ctx.CountSub != 0 {
			t.Fatalf("dye on %#v = (%#v, %v, %#v), want unchanged unhandled interaction", c, got, used, ctx)
		}
	}
}

func TestCauldronEntityInside(t *testing.T) {
	for _, test := range []struct {
		name     string
		before   block.Cauldron
		after    block.Cauldron
		position mgl64.Vec3
		fire     time.Duration
	}{
		{name: "empty", position: mgl64.Vec3{0.5, 1.5, 0.5}, fire: 2 * time.Second},
		{name: "water extinguishes at native level one", before: block.Cauldron{Level: 1}, position: mgl64.Vec3{0.5, 1.5, 0.5}},
		{name: "feet above liquid surface", before: block.Cauldron{Level: 4}, after: block.Cauldron{Level: 3}, position: mgl64.Vec3{0.5, 1.9, 0.5}},
		{name: "dyed water", before: block.Cauldron{Level: 4, Colour: color.RGBA{R: 255, A: 255}}, after: block.Cauldron{Level: 3, Colour: color.RGBA{R: 255, A: 255}}, position: mgl64.Vec3{0.5, 1.5, 0.5}},
		{name: "snow melts into water", before: block.Cauldron{Level: 4, Content: block.PowderSnowCauldronContent()}, after: block.Cauldron{Level: 3}, position: mgl64.Vec3{0.5, 1.5, 0.5}},
		{name: "potions do not extinguish", before: block.Cauldron{Level: 4, Potion: item.Potion{Type: potion.FireResistance()}}, after: block.Cauldron{Level: 4, Potion: item.Potion{Type: potion.FireResistance()}}, position: mgl64.Vec3{0.5, 1.5, 0.5}, fire: 2 * time.Second},
		{name: "lava ignites", before: block.Cauldron{Level: 6, Content: block.LavaCauldronContent()}, after: block.Cauldron{Level: 6, Content: block.LavaCauldronContent()}, position: mgl64.Vec3{0.5, 1.5, 0.5}, fire: 8 * time.Second},
		{name: "feet outside cauldron", before: block.Cauldron{Level: 4}, after: block.Cauldron{Level: 4}, position: mgl64.Vec3{1.01, 1.5, 0.5}, fire: 2 * time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
			t.Cleanup(func() { _ = w.Close() })
			h := entity.NewItem(world.EntitySpawnOpts{Position: test.position}, item.NewStack(item.Stick{}, 1))
			_, err := world.Call(context.Background(), w, func(tx *world.Tx) (struct{}, error) {
				pos := cube.Pos{0, 1, 0}
				tx.SetBlock(pos, test.before, nil)
				e := tx.AddEntity(h).(*entity.Ent)
				e.SetOnFire(2 * time.Second)
				test.before.EntityInside(pos, tx, e)
				if got := tx.Block(pos).(block.Cauldron); !reflect.DeepEqual(got, test.after) {
					t.Errorf("cauldron after contact = %#v, want %#v", got, test.after)
				}
				if got := e.OnFireDuration(); got != test.fire {
					t.Errorf("fire duration after contact = %v, want %v", got, test.fire)
				}
				return struct{}{}, nil
			})
			if err != nil {
				t.Fatalf("entity contact: %v", err)
			}
		})
	}
}

func TestCauldronPrecipitation(t *testing.T) {
	for _, test := range []struct {
		name   string
		before block.Cauldron
		after  block.Cauldron
		biome  world.Biome
		roof   bool
	}{
		{name: "exposed rain", after: block.Cauldron{Level: 6}, biome: biome.Plains{}},
		{name: "exposed snow", after: block.Cauldron{Level: 6, Content: block.PowderSnowCauldronContent()}, biome: biome.SnowyPlains{}},
		{name: "covered rain", biome: biome.Plains{}, roof: true},
		{name: "rain does not dilute potion", before: block.Cauldron{Level: 2, Potion: item.Potion{Type: potion.Healing()}}, after: block.Cauldron{Level: 2, Potion: item.Potion{Type: potion.Healing()}}, biome: biome.Plains{}},
		{name: "rain does not dilute dyed water", before: block.Cauldron{Level: 2, Colour: color.RGBA{G: 255, A: 255}}, after: block.Cauldron{Level: 2, Colour: color.RGBA{G: 255, A: 255}}, biome: biome.Plains{}},
		{name: "rain cannot fill lava", before: block.Cauldron{Level: 2, Content: block.LavaCauldronContent()}, after: block.Cauldron{Level: 2, Content: block.LavaCauldronContent()}, biome: biome.Plains{}},
		{name: "snow cannot fill water", before: block.Cauldron{Level: 2}, after: block.Cauldron{Level: 2}, biome: biome.SnowyPlains{}},
		{name: "rain cannot fill powder snow", before: block.Cauldron{Level: 2, Content: block.PowderSnowCauldronContent()}, after: block.Cauldron{Level: 2, Content: block.PowderSnowCauldronContent()}, biome: biome.Plains{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
			t.Cleanup(func() { _ = w.Close() })
			w.StartRaining(time.Minute)
			got, err := world.Call(context.Background(), w, func(tx *world.Tx) (block.Cauldron, error) {
				pos := cube.Pos{0, 1, 0}
				tx.SetBlock(pos, test.before, nil)
				tx.SetBiome(pos.Side(cube.FaceUp), test.biome)
				if test.roof {
					tx.SetBlock(cube.Pos{0, 4, 0}, block.Stone{}, nil)
				}
				r := rand.New(rand.NewPCG(1, 2))
				for range 1000 {
					tx.Block(pos).(block.Cauldron).RandomTick(pos, tx, r)
				}
				return tx.Block(pos).(block.Cauldron), nil
			})
			if err != nil {
				t.Fatalf("cauldron precipitation: %v", err)
			}
			if !reflect.DeepEqual(got, test.after) {
				t.Fatalf("cauldron after precipitation = %#v, want %#v", got, test.after)
			}
		})
	}
}

func TestCauldronNBT(t *testing.T) {
	for _, test := range []struct {
		name string
		c    block.Cauldron
		id   int16
		form int16
	}{
		{name: "empty", id: -1},
		{name: "water", c: block.Cauldron{Level: 4}, id: -1},
		{name: "dyed water", c: block.Cauldron{Level: 3, Colour: color.RGBA{R: 21, G: 76, B: 189, A: 255}}, id: -1},
		{name: "potion", c: block.Cauldron{Level: 2, Potion: item.Potion{Type: potion.Healing()}}, id: int16(potion.Healing().Uint8()), form: 0},
		{name: "splash potion", c: block.Cauldron{Level: 4, Potion: item.SplashPotion{Type: potion.Poison()}}, id: int16(potion.Poison().Uint8()), form: 1},
		{name: "lingering potion", c: block.Cauldron{Level: 6, Potion: item.LingeringPotion{Type: potion.FireResistance()}}, id: int16(potion.FireResistance().Uint8()), form: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := test.c.EncodeNBT()
			if data["id"] != "Cauldron" || data["PotionId"] != test.id || data["PotionType"] != test.form {
				t.Fatalf("native potion tags must be shorts: %#v", data)
			}
			if test.c.Colour != (color.RGBA{}) {
				if _, ok := data["CustomColor"].(int32); !ok {
					t.Fatalf("dyed water missing native CustomColor int tag: %#v", data)
				}
			} else if _, exists := data["CustomColor"]; exists {
				t.Fatalf("uncoloured cauldron encoded CustomColor: %#v", data)
			}
			encoded, err := nbt.Marshal(data)
			if err != nil {
				t.Fatalf("encode cauldron NBT: %v", err)
			}
			var decoded map[string]any
			if err := nbt.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("decode cauldron NBT: %v", err)
			}
			base := block.Cauldron{Level: test.c.Level, Content: test.c.Content}
			if got := base.DecodeNBT(decoded).(block.Cauldron); !reflect.DeepEqual(got, test.c) {
				t.Fatalf("NBT round trip = %#v, want %#v", got, test.c)
			}
		})
	}
	t.Run("legacy splash", func(t *testing.T) {
		got := (block.Cauldron{Level: 2}).DecodeNBT(map[string]any{"PotionId": int16(potion.Healing().Uint8()), "IsSplash": uint8(1)}).(block.Cauldron)
		if !reflect.DeepEqual(got.Potion, item.SplashPotion{Type: potion.Healing()}) {
			t.Fatalf("legacy IsSplash decoded to %#v", got.Potion)
		}
	})
}

func TestCauldronRegisteredStates(t *testing.T) {
	for _, content := range []block.CauldronContent{block.WaterCauldronContent(), block.LavaCauldronContent(), block.PowderSnowCauldronContent()} {
		for level := 0; level <= 6; level++ {
			c := block.Cauldron{Level: level, Content: content}
			name, properties := c.EncodeBlock()
			got, ok := world.BlockByName(name, properties)
			if !ok {
				t.Fatalf("unregistered cauldron state %s %#v", name, properties)
			}
			if !reflect.DeepEqual(got, c) {
				t.Fatalf("registered state = %#v, want %#v", got, c)
			}
			world.BlockRuntimeID(c)
		}
	}
	if got, ok := world.ItemByName("minecraft:cauldron", 0); !ok || !reflect.DeepEqual(got, block.Cauldron{}) {
		t.Fatalf("registered cauldron item = %#v, %v", got, ok)
	}
}

type cauldronTestUser struct {
	item.User
	held item.Stack
}

func (u *cauldronTestUser) HeldItems() (item.Stack, item.Stack) { return u.held, item.Stack{} }
func (u *cauldronTestUser) SetHeldItems(main, _ item.Stack)     { u.held = main }

func activateCauldron(t *testing.T, c block.Cauldron, held item.Stack) (block.Cauldron, item.UseContext, bool) {
	t.Helper()
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	t.Cleanup(func() { _ = w.Close() })
	var ctx item.UseContext
	var used bool
	got, err := world.Call(context.Background(), w, func(tx *world.Tx) (block.Cauldron, error) {
		pos := cube.Pos{0, 1, 0}
		tx.SetBlock(pos, c, nil)
		used = c.Activate(pos, cube.FaceUp, tx, &cauldronTestUser{held: held}, &ctx)
		return tx.Block(pos).(block.Cauldron), nil
	})
	if err != nil {
		t.Fatalf("activate cauldron: %v", err)
	}
	return got, ctx, used
}

func assertCauldronReturnedItem(t *testing.T, ctx item.UseContext, want world.Item, count int) {
	t.Helper()
	if got := ctx.NewItem; got.Count() != count || !reflect.DeepEqual(got.Item(), want) {
		t.Fatalf("returned stack = %#v x%d, want %#v x%d", got.Item(), got.Count(), want, count)
	}
}

func assertCauldronStackMetadata(t *testing.T, got, original item.Stack) {
	t.Helper()
	if got.CustomName() != original.CustomName() || !reflect.DeepEqual(got.Lore(), original.Lore()) || !reflect.DeepEqual(got.Values(), original.Values()) {
		t.Fatalf("item lost name, lore or custom data: %#v", got)
	}
}
