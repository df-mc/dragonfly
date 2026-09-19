package entity

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// cushionTestWorld returns a synchronous world to test cushions in.
func cushionTestWorld(t *testing.T) *world.World {
	t.Helper()
	w := world.Config{Synchronous: true, Entities: DefaultRegistry}.New()
	t.Cleanup(func() { _ = w.Close() })
	return w
}

func cushionDo(t *testing.T, w *world.World, f func(tx *world.Tx)) {
	t.Helper()
	if err := w.Do(f).Wait(context.Background()); err != nil {
		t.Fatalf("world task failed: %v", err)
	}
}

// TestCushionSurvivalMatchesVanilla checks cushion survival at heights 2/32 to 32/32 above each block, as
// recorded on a vanilla 1.26.50 dedicated server ('.' stayed, 'X' broke).
func TestCushionSurvivalMatchesVanilla(t *testing.T) {
	cases := []struct {
		name    string
		block   func() world.Block
		ceiling bool
		vanilla string
	}{
		{"poppy", func() world.Block { return block.Flower{Type: block.Poppy()} }, false, "......................XXXXXXXXX"},
		{"dandelion", func() world.Block { return block.Flower{Type: block.Dandelion()} }, false, "......................XXXXXXXXX"},
		{"wither rose", func() world.Block { return block.Flower{Type: block.WitherRose()} }, false, "......................XXXXXXXXX"},
		{"allium", func() world.Block { return block.Flower{Type: block.Allium()} }, false, "......................XXXXXXXXX"},
		{"cornflower", func() world.Block { return block.Flower{Type: block.Cornflower()} }, false, "......................XXXXXXXXX"},
		{"dead bush", func() world.Block { return block.DeadBush{} }, false, "..............................."},
		{"oak sapling", func() world.Block { return block.Sapling{Type: block.OakSapling()} }, false, ".............................XX"},
		{"bamboo sapling", func() world.Block { return block.BambooSapling{} }, false, ".............................XX"},
		{"nether sprouts", func() world.Block { return block.NetherSprouts{} }, false, ".............XXXXXXXXXXXXXXXXXX"},
		{"torch", func() world.Block { return block.Torch{Facing: cube.FaceDown, Type: block.NormalFire()} }, false, "......................XXXXXXXXX"},
		{"soul torch", func() world.Block { return block.Torch{Facing: cube.FaceDown, Type: block.SoulFire()} }, false, "......................XXXXXXXXX"},
		{"redstone torch", func() world.Block { return block.RedstoneTorch{Facing: cube.FaceDown} }, false, "......................XXXXXXXXX"},
		{"copper torch", func() world.Block { return block.CopperTorch{Facing: cube.FaceDown} }, false, "......................XXXXXXXXX"},
		{"wall torch", func() world.Block { return block.Torch{Facing: cube.FaceEast, Type: block.NormalFire()} }, false, "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"floor lever", func() world.Block { return block.Lever{Facing: cube.FaceUp, Direction: cube.North} }, false, "......................XXXXXXXXX"},
		{"ceiling lever", func() world.Block { return block.Lever{Facing: cube.FaceDown, Direction: cube.North} }, true, "XXXXXXXX......................X"},
		{"nether portal", func() world.Block { return block.Portal{Axis: cube.X} }, false, "..............................."},
		{"wall lever", func() world.Block { return block.Lever{Facing: cube.FaceEast} }, false, "XXX........................XXXX"},
		{"redstone wire", func() world.Block { return block.RedstoneWire{} }, false, ".....XXXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"tripwire", func() world.Block { return block.String{} }, false, "...................XXXXXXXXXXXX"},
		{"standing sign", func() world.Block { return block.Sign{Wood: block.OakWood(), Attach: block.StandingAttachment(0)} }, false, "..............................."},
		{"wall sign", func() world.Block { return block.Sign{Wood: block.OakWood(), Attach: block.WallAttachment(cube.West)} }, false, "XXXX........................XXX"},
		{"standing banner", func() world.Block { return block.Banner{Attach: block.StandingAttachment(0)} }, false, "..............................."},
		{"wall banner", func() world.Block { return block.Banner{Attach: block.WallAttachment(cube.West)} }, false, "............................XXX"},
		{"floor item frame", func() world.Block { return block.ItemFrame{Facing: cube.FaceDown} }, false, ".....XXXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"wall item frame", func() world.Block { return block.ItemFrame{Facing: cube.FaceEast} }, false, "..............................."},
		{"ceiling item frame", func() world.Block { return block.ItemFrame{Facing: cube.FaceUp} }, true, "XXXXXXXXXXXXXXXXXXXXXXXXX.....X"},
		{"sea pickle", func() world.Block { return block.SeaPickle{} }, false, "...................XXXXXXXXXXXX"},
		{"sea pickles", func() world.Block { return block.SeaPickle{AdditionalCount: 3} }, false, "...................XXXXXXXXXXXX"},
		{"cobweb", func() world.Block { return block.Cobweb{} }, false, "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"light", func() world.Block { return block.Light{Level: 15} }, false, "..............................."},
		{"pink petals", func() world.Block { return block.PinkPetals{AdditionalCount: 3} }, false, "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"red shrub", func() world.Block { return block.RedShrub{} }, false, "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"fire", func() world.Block { return block.Fire{Type: block.NormalFire()} }, false, "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"coral", func() world.Block { return block.Coral{Type: block.BrainCoral()} }, false, ".............................XX"},
		{"kelp", func() world.Block { return block.Kelp{} }, false, "..............................."},
		{"vines", func() world.Block { return block.Vines{EastDirection: true} }, false, "..............................."},
		{"sugar cane", func() world.Block { return block.SugarCane{} }, false, "..............................."},
		{"spore blossom", func() world.Block { return block.SporeBlossom{} }, true, "XXXXXXXXXXXXXXXXXXX...........X"},
		{"end portal", func() world.Block { return block.EndPortal{} }, false, "...........XXXXXXXXXXXXXXXXXXXX"},
		{"lily pad", func() world.Block { return block.LilyPad{} }, false, "......XXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"carpet", func() world.Block { return block.Carpet{Colour: item.ColourWhite()} }, false, ".....XXXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"open fence gate", func() world.Block { return block.WoodFenceGate{Wood: block.OakWood(), Open: true} }, false, "..............................."},
		{"closed fence gate", func() world.Block { return block.WoodFenceGate{Wood: block.OakWood()} }, false, "..............................."},
		{"enchanting table", func() world.Block { return block.EnchantingTable{} }, false, "...........................XXXX"},
		{"stonecutter", func() world.Block { return block.Stonecutter{} }, false, ".....................XXXXXXXXXX"},
		{"lantern", func() world.Block { return block.Lantern{Type: block.NormalFire()} }, false, "...................XXXXXXXXXXXX"},
		{"trapdoor", func() world.Block { return block.WoodTrapdoor{Wood: block.OakWood()} }, false, ".........XXXXXXXXXXXXXXXXXXXXXX"},
		{"cake", func() world.Block { return block.Cake{} }, false, "...................XXXXXXXXXXXX"},
		{"campfire", func() world.Block { return block.Campfire{Type: block.NormalFire()} }, false, ".................XXXXXXXXXXXXXX"},
		{"end portal frame", func() world.Block { return block.EndPortalFrame{} }, false, ".............................XX"},
		{"candle", func() world.Block { return block.Candle{} }, false, "...............XXXXXXXXXXXXXXXX"},
		{"slab", func() world.Block { return block.Slab{Block: block.Stone{Smooth: true}} }, false, "...................XXXXXXXXXXXX"},
		{"barrier", func() world.Block { return block.Barrier{} }, false, "XXXXXXXXXXXXXXXXXXXXXXX........"},
		{"farmland", func() world.Block { return block.Farmland{} }, false, "XXXXXXXXXXXXXXXXXXXXXXX........"},
		{"dirt path", func() world.Block { return block.DirtPath{} }, false, "XXXXXXXXXXXXXXXXXXXXXXX........"},
		{"soul sand", func() world.Block { return block.SoulSand{} }, false, "XXXXXXXXXXXXXXXXXXXXXXX........"},
		{"mud", func() world.Block { return block.Mud{} }, false, "XXXXXXXXXXXXXXXXXXXXXXX........"},
		{"wheat 0", func() world.Block { b := block.WheatSeeds{}; b.Growth = 0; return b }, false, "........XXXXXXXXXXXXXXXXXXXXXXX"},
		{"wheat 1", func() world.Block { b := block.WheatSeeds{}; b.Growth = 1; return b }, false, "............XXXXXXXXXXXXXXXXXXX"},
		{"wheat 2", func() world.Block { b := block.WheatSeeds{}; b.Growth = 2; return b }, false, ".................XXXXXXXXXXXXXX"},
		{"wheat 3", func() world.Block { b := block.WheatSeeds{}; b.Growth = 3; return b }, false, ".....................XXXXXXXXXX"},
		{"wheat 4", func() world.Block { b := block.WheatSeeds{}; b.Growth = 4; return b }, false, "..........................XXXXX"},
		{"wheat 5", func() world.Block { b := block.WheatSeeds{}; b.Growth = 5; return b }, false, "..............................X"},
		{"wheat 6", func() world.Block { b := block.WheatSeeds{}; b.Growth = 6; return b }, false, "..............................."},
		{"wheat 7", func() world.Block { b := block.WheatSeeds{}; b.Growth = 7; return b }, false, "..............................."},
		{"carrots 0", func() world.Block { b := block.Carrot{}; b.Growth = 0; return b }, false, "......XXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"carrots 1", func() world.Block { b := block.Carrot{}; b.Growth = 1; return b }, false, ".........XXXXXXXXXXXXXXXXXXXXXX"},
		{"carrots 2", func() world.Block { b := block.Carrot{}; b.Growth = 2; return b }, false, ".............XXXXXXXXXXXXXXXXXX"},
		{"carrots 3", func() world.Block { b := block.Carrot{}; b.Growth = 3; return b }, false, "................XXXXXXXXXXXXXXX"},
		{"carrots 4", func() world.Block { b := block.Carrot{}; b.Growth = 4; return b }, false, "...................XXXXXXXXXXXX"},
		{"carrots 5", func() world.Block { b := block.Carrot{}; b.Growth = 5; return b }, false, "......................XXXXXXXXX"},
		{"carrots 6", func() world.Block { b := block.Carrot{}; b.Growth = 6; return b }, false, ".........................XXXXXX"},
		{"carrots 7", func() world.Block { b := block.Carrot{}; b.Growth = 7; return b }, false, ".............................XX"},
		{"potatoes 0", func() world.Block { b := block.Potato{}; b.Growth = 0; return b }, false, "......XXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"potatoes 1", func() world.Block { b := block.Potato{}; b.Growth = 1; return b }, false, "........XXXXXXXXXXXXXXXXXXXXXXX"},
		{"potatoes 2", func() world.Block { b := block.Potato{}; b.Growth = 2; return b }, false, "...........XXXXXXXXXXXXXXXXXXXX"},
		{"potatoes 3", func() world.Block { b := block.Potato{}; b.Growth = 3; return b }, false, "..............XXXXXXXXXXXXXXXXX"},
		{"potatoes 4", func() world.Block { b := block.Potato{}; b.Growth = 4; return b }, false, ".................XXXXXXXXXXXXXX"},
		{"potatoes 5", func() world.Block { b := block.Potato{}; b.Growth = 5; return b }, false, "...................XXXXXXXXXXXX"},
		{"potatoes 6", func() world.Block { b := block.Potato{}; b.Growth = 6; return b }, false, "......................XXXXXXXXX"},
		{"pumpkin stem 0", func() world.Block { b := block.PumpkinSeeds{}; b.Growth = 0; return b }, false, ".......XXXXXXXXXXXXXXXXXXXXXXXX"},
		{"pumpkin stem 1", func() world.Block { b := block.PumpkinSeeds{}; b.Growth = 1; return b }, false, "...........XXXXXXXXXXXXXXXXXXXX"},
		{"pumpkin stem 2", func() world.Block { b := block.PumpkinSeeds{}; b.Growth = 2; return b }, false, "...............XXXXXXXXXXXXXXXX"},
		{"pumpkin stem 3", func() world.Block { b := block.PumpkinSeeds{}; b.Growth = 3; return b }, false, "...................XXXXXXXXXXXX"},
		{"pumpkin stem 4", func() world.Block { b := block.PumpkinSeeds{}; b.Growth = 4; return b }, false, ".......................XXXXXXXX"},
		{"pumpkin stem 5", func() world.Block { b := block.PumpkinSeeds{}; b.Growth = 5; return b }, false, "...........................XXXX"},
		{"pumpkin stem 6", func() world.Block { b := block.PumpkinSeeds{}; b.Growth = 6; return b }, false, "..............................."},
		{"pumpkin stem 7", func() world.Block { b := block.PumpkinSeeds{}; b.Growth = 7; return b }, false, "..............................."},
		{"melon stem 0", func() world.Block { b := block.MelonSeeds{}; b.Growth = 0; return b }, false, ".......XXXXXXXXXXXXXXXXXXXXXXXX"},
		{"nether wart 0", func() world.Block { return block.NetherWart{Age: 0} }, false, "...........XXXXXXXXXXXXXXXXXXXX"},
		{"nether wart 1", func() world.Block { return block.NetherWart{Age: 1} }, false, "...................XXXXXXXXXXXX"},
		{"nether wart 2", func() world.Block { return block.NetherWart{Age: 2} }, false, "...........................XXXX"},
		{"nether wart 3", func() world.Block { return block.NetherWart{Age: 3} }, false, "..............................."},
	}
	w := cushionTestWorld(t)
	cushionDo(t, w, func(tx *world.Tx) {
		for i, c := range cases {
			pos := cube.Pos{i * 2, 64, 0}
			tx.SetBlock(pos, c.block(), nil)
			if c.ceiling {
				tx.SetBlock(pos.Side(cube.FaceUp), block.Stone{}, nil)
			}
			got := make([]byte, 0, len(c.vanilla))
			for k := 2; k <= 32; k++ {
				p := pos.Vec3Middle()
				p[1] = float64(pos[1]) + float64(k)/32
				if cushionCanSurvive(p, tx) {
					got = append(got, '.')
				} else {
					got = append(got, 'X')
				}
			}
			if string(got) != c.vanilla {
				t.Errorf("%v:\n got     %v\n vanilla %v", c.name, string(got), c.vanilla)
			}
			tx.SetBlock(pos, nil, nil)
			tx.SetBlock(pos.Side(cube.FaceUp), nil, nil)
		}
	})
}

// TestCushionPlantOffsetMatchesVanilla checks the position based height of short grass and ferns, as measured
// on a vanilla 1.26.50 dedicated server.
func TestCushionPlantOffsetMatchesVanilla(t *testing.T) {
	cases := []struct {
		block   world.Block
		pos     cube.Pos
		highest float64
	}{
		{block.ShortGrass{}, cube.Pos{6, 64, -16}, 0.8046875},
		{block.ShortGrass{}, cube.Pos{6, 64, -15}, 0.8671875},
		{block.ShortGrass{}, cube.Pos{6, 64, -14}, 0.8671875},
		{block.ShortGrass{}, cube.Pos{6, 64, -13}, 0.921875},
		{block.ShortGrass{}, cube.Pos{6, 64, -12}, 0.8046875},
		{block.ShortGrass{}, cube.Pos{6, 64, -11}, 0.84375},
		{block.ShortGrass{}, cube.Pos{6, 64, -10}, 0.921875},
		{block.ShortGrass{}, cube.Pos{6, 64, -9}, 0.7734375},
		{block.ShortGrass{}, cube.Pos{6, 64, -8}, 0.90625},
		{block.ShortGrass{}, cube.Pos{6, 64, -7}, 0.859375},
		{block.ShortGrass{}, cube.Pos{6, 64, -6}, 0.734375},
		{block.ShortGrass{}, cube.Pos{6, 64, -5}, 0.765625},
		{block.ShortGrass{}, cube.Pos{6, 64, -4}, 0.8046875},
		{block.ShortGrass{}, cube.Pos{6, 64, -3}, 0.921875},
		{block.ShortGrass{}, cube.Pos{6, 64, -2}, 0.84375},
		{block.ShortGrass{}, cube.Pos{6, 64, -1}, 0.90625},
		{block.ShortGrass{}, cube.Pos{6, 64, 0}, 0.90625},
		{block.ShortGrass{}, cube.Pos{6, 64, 1}, 0.8046875},
		{block.ShortGrass{}, cube.Pos{6, 64, 2}, 0.8671875},
		{block.ShortGrass{}, cube.Pos{6, 64, 3}, 0.75},
		{block.ShortGrass{}, cube.Pos{6, 64, 4}, 0.7890625},
		{block.ShortGrass{}, cube.Pos{6, 64, 5}, 0.8046875},
		{block.ShortGrass{}, cube.Pos{6, 64, 6}, 0.734375},
		{block.ShortGrass{}, cube.Pos{6, 64, 7}, 0.734375},
		{block.ShortGrass{}, cube.Pos{6, 64, 8}, 0.828125},
		{block.ShortGrass{}, cube.Pos{6, 64, 9}, 0.8046875},
		{block.ShortGrass{}, cube.Pos{6, 64, 10}, 0.8671875},
		{block.ShortGrass{}, cube.Pos{6, 64, 11}, 0.84375},
		{block.ShortGrass{}, cube.Pos{6, 64, 12}, 0.921875},
		{block.ShortGrass{}, cube.Pos{6, 64, 13}, 0.8671875},
		{block.ShortGrass{}, cube.Pos{6, 64, 14}, 0.90625},
		{block.Fern{}, cube.Pos{7, 64, -16}, 0.734375},
		{block.Fern{}, cube.Pos{7, 64, -15}, 0.7890625},
		{block.Fern{}, cube.Pos{7, 64, -14}, 0.859375},
		{block.Fern{}, cube.Pos{7, 64, -13}, 0.8125},
		{block.Fern{}, cube.Pos{7, 64, -12}, 0.9375},
		{block.Fern{}, cube.Pos{7, 64, -11}, 0.7734375},
		{block.Fern{}, cube.Pos{7, 64, -10}, 0.8671875},
		{block.Fern{}, cube.Pos{7, 64, -9}, 0.765625},
		{block.Fern{}, cube.Pos{7, 64, -8}, 0.828125},
		{block.Fern{}, cube.Pos{7, 64, -7}, 0.75},
		{block.Fern{}, cube.Pos{7, 64, -6}, 0.7734375},
		{block.Fern{}, cube.Pos{7, 64, -5}, 0.8671875},
		{block.Fern{}, cube.Pos{7, 64, -4}, 0.7734375},
		{block.Fern{}, cube.Pos{7, 64, -3}, 0.7734375},
		{block.Fern{}, cube.Pos{7, 64, -2}, 0.7890625},
		{block.Fern{}, cube.Pos{7, 64, -1}, 0.765625},
		{block.Fern{}, cube.Pos{7, 64, 0}, 0.75},
		{block.Fern{}, cube.Pos{7, 64, 1}, 0.734375},
		{block.Fern{}, cube.Pos{7, 64, 2}, 0.9375},
		{block.Fern{}, cube.Pos{7, 64, 3}, 0.828125},
		{block.Fern{}, cube.Pos{7, 64, 4}, 0.7890625},
		{block.Fern{}, cube.Pos{7, 64, 5}, 0.765625},
		{block.Fern{}, cube.Pos{7, 64, 6}, 0.9375},
		{block.Fern{}, cube.Pos{7, 64, 7}, 0.84375},
		{block.Fern{}, cube.Pos{7, 64, 8}, 0.9375},
		{block.Fern{}, cube.Pos{7, 64, 9}, 0.8046875},
		{block.Fern{}, cube.Pos{7, 64, 10}, 0.8125},
		{block.Fern{}, cube.Pos{7, 64, 11}, 0.9375},
		{block.Fern{}, cube.Pos{7, 64, 12}, 0.90625},
		{block.Fern{}, cube.Pos{7, 64, 13}, 0.8125},
		{block.Fern{}, cube.Pos{7, 64, 14}, 0.7734375},
	}
	w := cushionTestWorld(t)
	cushionDo(t, w, func(tx *world.Tx) {
		for _, c := range cases {
			tx.SetBlock(c.pos, c.block, nil)
			at := func(h float64) mgl64.Vec3 {
				p := c.pos.Vec3Middle()
				p[1] = float64(c.pos[1]) + h
				return p
			}
			if !cushionSupported(at(c.highest), tx) || cushionSupported(at(c.highest+1.0/128), tx) {
				t.Errorf("%T at %v: vanilla supports cushions up to %v above it", c.block, c.pos, c.highest)
			}
			tx.SetBlock(c.pos, nil, nil)
		}
	})
}

// TestCushionCoveringBlocks checks which blocks break a cushion inside of them, as recorded on vanilla.
func TestCushionCoveringBlocks(t *testing.T) {
	covering := []world.Block{
		block.Stone{}, block.PackedIce{}, block.BlueIce{}, block.Slime{}, block.Barrier{}, block.SoulSand{},
		block.Farmland{}, block.DirtPath{}, block.Mud{}, block.Bedrock{}, block.Log{Wood: block.OakWood()},
		block.Cactus{}, block.Sponge{}, block.Snow{}, block.BambooBlock{}, block.Dirt{}, block.HayBale{},
		block.CraftingTable{}, block.Jukebox{}, block.Wool{Colour: item.ColourWhite()},
		block.Concrete{Colour: item.ColourWhite()}, block.ConcretePowder{Colour: item.ColourWhite()},
		block.DriedKelp{}, block.CopperGrate{}, block.StoneBricks{}, block.Note{},
	}
	notCovering := []world.Block{
		block.Glass{}, block.StainedGlass{Colour: item.ColourWhite()}, block.TintedGlass{},
		block.Leaves{Type: block.OakLeaves()}, block.Ice{}, block.Glowstone{}, block.SeaLantern{},
		block.Light{Level: 15}, block.NewChest(), block.GlassPane{}, block.IronBars{},
		block.WoodFence{Wood: block.OakWood()}, block.Beacon{}, block.TNT{}, block.Anvil{}, block.EndPortalFrame{},
	}
	for _, b := range covering {
		if !cushionCovers(b) {
			t.Errorf("%T should cover a cushion", b)
		}
	}
	for _, b := range notCovering {
		if cushionCovers(b) {
			t.Errorf("%T should not cover a cushion", b)
		}
	}
}

// TestCushionVariant checks the colours of the cushion variants, which count down from white to black.
func TestCushionVariant(t *testing.T) {
	for _, c := range item.Colours() {
		b := &cushionBehaviour{colour: c}
		if cushionColour(b.Variant()) != c {
			t.Errorf("colour %v does not survive a round trip through variant %v", c, b.Variant())
		}
	}
	if (&cushionBehaviour{colour: item.ColourWhite()}).Variant() != 15 || (&cushionBehaviour{colour: item.ColourBlack()}).Variant() != 0 {
		t.Error("white cushions must have variant 15 and black cushions variant 0")
	}
}
