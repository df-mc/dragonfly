package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/colour"
	"github.com/df-mc/dragonfly/dragonfly/block/fire"
	"github.com/df-mc/dragonfly/dragonfly/block/grass"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/world"
	_ "unsafe" // Imported for compiler directives.
)

// init registers all blocks implemented by Dragonfly.
func init() {
	world.RegisterBlock(Air{})
	world.RegisterBlock(Stone{})
	world.RegisterBlock(Stone{Smooth: true})
	world.RegisterBlock(Granite{})
	world.RegisterBlock(Granite{Polished: true})
	world.RegisterBlock(Diorite{})
	world.RegisterBlock(Diorite{Polished: true})
	world.RegisterBlock(Andesite{})
	world.RegisterBlock(Andesite{Polished: true})
	world.RegisterBlock(Grass{})
	world.RegisterBlock(GrassPath{})
	world.RegisterBlock(Dirt{})
	world.RegisterBlock(Dirt{Coarse: true})
	world.RegisterBlock(Cobblestone{})
	world.RegisterBlock(Cobblestone{Mossy: true})
	world.RegisterBlock(Bedrock{})
	world.RegisterBlock(Bedrock{InfiniteBurning: true})
	world.RegisterBlock(Obsidian{})
	world.RegisterBlock(DiamondBlock{})
	world.RegisterBlock(Glass{})
	world.RegisterBlock(Glowstone{})
	world.RegisterBlock(EmeraldBlock{})
	world.RegisterBlock(EndBricks{})
	world.RegisterBlock(GoldBlock{})
	world.RegisterBlock(NetheriteBlock{})
	world.RegisterBlock(IronBlock{})
	world.RegisterBlock(CoalBlock{})
	world.RegisterBlock(Beacon{})
	world.RegisterBlock(Sponge{})
	world.RegisterBlock(Sponge{Wet: true})
	world.RegisterBlock(LapisBlock{})
	world.RegisterBlock(Terracotta{})
	world.RegisterBlock(GlassPane{})
	world.RegisterBlock(IronBars{})
	world.RegisterBlock(NetherBrickFence{})
	world.RegisterBlock(EndStone{})
	world.RegisterBlock(Netherrack{})
	world.RegisterBlock(QuartzBricks{})
	world.RegisterBlock(Clay{})
	world.RegisterBlock(AncientDebris{})
	world.RegisterBlock(EmeraldOre{})
	world.RegisterBlock(DiamondOre{})
	world.RegisterBlock(LapisOre{})
	world.RegisterBlock(NetherGoldOre{})
	world.RegisterBlock(GoldOre{})
	world.RegisterBlock(IronOre{})
	world.RegisterBlock(CoalOre{})
	world.RegisterBlock(NetherQuartzOre{})
	world.RegisterBlock(Melon{})
	world.RegisterBlock(Sand{})
	world.RegisterBlock(Sand{Red: true})
	world.RegisterBlock(Gravel{})
	world.RegisterBlock(Bricks{})
	world.RegisterBlock(SoulSand{})
	world.RegisterBlock(Barrier{})
	world.RegisterBlock(CryingObsidian{})
	world.RegisterBlock(SeaLantern{})
	world.RegisterBlock(SoulSoil{})
	world.RegisterBlock(BlueIce{})
	world.RegisterBlock(GildedBlackstone{})
	world.RegisterBlock(Shroomlight{})
	world.RegisterBlock(InvisibleBedrock{})
	world.RegisterBlock(DragonEgg{})

	registerAll(allBasalt())
	registerAll(allBeetroot())
	registerAll(allBoneBlock())
	registerAll(allCake())
	registerAll(allCarpet())
	registerAll(allCarrots())
	registerAll(allChests())
	registerAll(allConcrete())
	registerAll(allConcretePowder())
	registerAll(allCocoaBeans())
	registerAll(allCoral())
	registerAll(allCoralBlocks())
	registerAll(allEndBrickStairs())
	registerAll(allNoteBlocks())
	registerAll(allWool())
	registerAll(allStainedTerracotta())
	registerAll(allGlazedTerracotta())
	registerAll(allStainedGlass())
	registerAll(allStainedGlassPane())
	registerAll(allLanterns())
	registerAll(allFire())
	registerAll(allPlanks())
	registerAll(allFence())
	registerAll(allFenceGates())
	registerAll(allWoodStairs())
	registerAll(allDoors())
	registerAll(allTrapdoors())
	registerAll(allWoodSlabs())
	registerAll(allLogs())
	registerAll(allLeaves())
	registerAll(allTorch())
	registerAll(allPumpkinStems())
	registerAll(allPumpkins())
	registerAll(allLitPumpkins())
	registerAll(allMelonStems())
	registerAll(allFarmland())
	registerAll(allLava())
	registerAll(allWater())
	registerAll(allKelp())
	registerAll(allPotato())
	registerAll(allWheat())
	registerAll(allQuartz())
	registerAll(allNetherWart())
	registerAll(allGrassPlants())
}

func init() {
	world.RegisterItem("minecraft:air", Air{})
	world.RegisterItem("minecraft:stone", Stone{})
	world.RegisterItem("minecraft:smooth_stone", Stone{Smooth: true})
	world.RegisterItem("minecraft:stone", Granite{})
	world.RegisterItem("minecraft:stone", Granite{Polished: true})
	world.RegisterItem("minecraft:stone", Diorite{})
	world.RegisterItem("minecraft:stone", Diorite{Polished: true})
	world.RegisterItem("minecraft:stone", Andesite{})
	world.RegisterItem("minecraft:stone", Andesite{Polished: true})
	world.RegisterItem("minecraft:grass", Grass{})
	world.RegisterItem("minecraft:grass_path", GrassPath{})
	world.RegisterItem("minecraft:dirt", Dirt{})
	world.RegisterItem("minecraft:dirt", Dirt{Coarse: true})
	world.RegisterItem("minecraft:cobblestone", Cobblestone{})
	world.RegisterItem("minecraft:bedrock", Bedrock{})
	world.RegisterItem("minecraft:kelp", Kelp{})
	world.RegisterItem("minecraft:log", Log{Wood: wood.Oak()})
	world.RegisterItem("minecraft:log", Log{Wood: wood.Spruce()})
	world.RegisterItem("minecraft:log", Log{Wood: wood.Birch()})
	world.RegisterItem("minecraft:log", Log{Wood: wood.Jungle()})
	world.RegisterItem("minecraft:leaves", Leaves{Wood: wood.Oak(), Persistent: true})
	world.RegisterItem("minecraft:leaves", Leaves{Wood: wood.Spruce(), Persistent: true})
	world.RegisterItem("minecraft:leaves", Leaves{Wood: wood.Birch(), Persistent: true})
	world.RegisterItem("minecraft:leaves", Leaves{Wood: wood.Jungle(), Persistent: true})
	world.RegisterItem("minecraft:chest", Chest{})
	world.RegisterItem("minecraft:mossy_cobblestone", Cobblestone{Mossy: true})
	world.RegisterItem("minecraft:leaves2", Leaves{Wood: wood.Acacia(), Persistent: true})
	world.RegisterItem("minecraft:leaves2", Leaves{Wood: wood.DarkOak(), Persistent: true})
	world.RegisterItem("minecraft:log2", Log{Wood: wood.Acacia()})
	world.RegisterItem("minecraft:log2", Log{Wood: wood.DarkOak()})
	world.RegisterItem("minecraft:stripped_spruce_log", Log{Wood: wood.Spruce(), Stripped: true})
	world.RegisterItem("minecraft:stripped_birch_log", Log{Wood: wood.Birch(), Stripped: true})
	world.RegisterItem("minecraft:stripped_jungle_log", Log{Wood: wood.Jungle(), Stripped: true})
	world.RegisterItem("minecraft:stripped_acacia_log", Log{Wood: wood.Acacia(), Stripped: true})
	world.RegisterItem("minecraft:stripped_dark_oak_log", Log{Wood: wood.DarkOak(), Stripped: true})
	world.RegisterItem("minecraft:stripped_oak_log", Log{Wood: wood.Oak(), Stripped: true})
	for _, c := range colour.All() {
		world.RegisterItem("minecraft:concrete", Concrete{Colour: c})
		world.RegisterItem("minecraft:concrete_powder", ConcretePowder{Colour: c})
		world.RegisterItem("minecraft:stained_hardened_clay", StainedTerracotta{Colour: c})
		world.RegisterItem("minecraft:carpet", Carpet{Colour: c})
		world.RegisterItem("minecraft:wool", Wool{Colour: c})
		world.RegisterItem("minecraft:stained_glass", StainedGlass{Colour: c})
		world.RegisterItem("minecraft:stained_glass_pane", StainedGlassPane{Colour: c})

		colourName := c.String()
		if c == colour.LightGrey() {
			colourName = "silver"
		}

		world.RegisterItem("minecraft:"+colourName+"_glazed_terracotta", GlazedTerracotta{Colour: c})
	}
	for _, b := range allLight() {
		world.RegisterItem("minecraft:light_block", b.(world.Item))
	}
	for _, w := range wood.All() {
		if w == wood.Crimson() || w == wood.Warped() {
			world.RegisterItem("minecraft:"+w.String()+"_planks", Planks{Wood: w})
		} else {
			world.RegisterItem("minecraft:planks", Planks{Wood: w})
		}
	}
	world.RegisterItem("minecraft:oak_stairs", WoodStairs{Wood: wood.Oak()})
	world.RegisterItem("minecraft:spruce_stairs", WoodStairs{Wood: wood.Spruce()})
	world.RegisterItem("minecraft:birch_stairs", WoodStairs{Wood: wood.Birch()})
	world.RegisterItem("minecraft:jungle_stairs", WoodStairs{Wood: wood.Jungle()})
	world.RegisterItem("minecraft:acacia_stairs", WoodStairs{Wood: wood.Acacia()})
	world.RegisterItem("minecraft:dark_oak_stairs", WoodStairs{Wood: wood.DarkOak()})
	world.RegisterItem("minecraft:crimson_stairs", WoodStairs{Wood: wood.Crimson()})
	world.RegisterItem("minecraft:warped_stairs", WoodStairs{Wood: wood.Warped()})
	world.RegisterItem("minecraft:wooden_slab", WoodSlab{Wood: wood.Oak()})
	world.RegisterItem("minecraft:wooden_slab", WoodSlab{Wood: wood.Spruce()})
	world.RegisterItem("minecraft:wooden_slab", WoodSlab{Wood: wood.Birch()})
	world.RegisterItem("minecraft:wooden_slab", WoodSlab{Wood: wood.Jungle()})
	world.RegisterItem("minecraft:wooden_slab", WoodSlab{Wood: wood.Acacia()})
	world.RegisterItem("minecraft:wooden_slab", WoodSlab{Wood: wood.DarkOak()})
	world.RegisterItem("minecraft:crimson_slab", WoodSlab{Wood: wood.Crimson()})
	world.RegisterItem("minecraft:warped_slab", WoodSlab{Wood: wood.Warped()})
	world.RegisterItem("minecraft:double_wooden_slab", WoodSlab{Wood: wood.Oak(), Double: true})
	world.RegisterItem("minecraft:double_wooden_slab", WoodSlab{Wood: wood.Spruce(), Double: true})
	world.RegisterItem("minecraft:double_wooden_slab", WoodSlab{Wood: wood.Birch(), Double: true})
	world.RegisterItem("minecraft:double_wooden_slab", WoodSlab{Wood: wood.Jungle(), Double: true})
	world.RegisterItem("minecraft:double_wooden_slab", WoodSlab{Wood: wood.Acacia(), Double: true})
	world.RegisterItem("minecraft:double_wooden_slab", WoodSlab{Wood: wood.DarkOak(), Double: true})
	world.RegisterItem("minecraft:crimson_double_slab", WoodSlab{Wood: wood.Crimson(), Double: true})
	world.RegisterItem("minecraft:warped_double_slab", WoodSlab{Wood: wood.Warped(), Double: true})
	world.RegisterItem("minecraft:obsidian", Obsidian{})
	world.RegisterItem("minecraft:diamond_block", DiamondBlock{})
	world.RegisterItem("minecraft:glass", Glass{})
	world.RegisterItem("minecraft:glowstone", Glowstone{})
	world.RegisterItem("minecraft:emerald_block", EmeraldBlock{})
	world.RegisterItem("minecraft:end_bricks", EndBricks{})
	world.RegisterItem("minecraft:end_brick_stairs", EndBrickStairs{})
	world.RegisterItem("minecraft:netherite_block", NetheriteBlock{})
	world.RegisterItem("minecraft:gold_block", GoldBlock{})
	world.RegisterItem("minecraft:iron_block", IronBlock{})
	world.RegisterItem("minecraft:coal_block", CoalBlock{})
	world.RegisterItem("minecraft:beacon", Beacon{})
	world.RegisterItem("minecraft:sponge", Sponge{})
	world.RegisterItem("minecraft:wet_sponge", Sponge{Wet: true})
	world.RegisterItem("minecraft:lapis_block", LapisBlock{})
	world.RegisterItem("minecraft:hardened_clay", Terracotta{})
	world.RegisterItem("minecraft:quartz_block", Quartz{})
	world.RegisterItem("minecraft:quartz_block", Quartz{Smooth: true})
	world.RegisterItem("minecraft:quartz_block", ChiseledQuartz{})
	world.RegisterItem("minecraft:quartz_block", QuartzPillar{})
	world.RegisterItem("minecraft:quartz_bricks", QuartzBricks{})
	world.RegisterItem("minecraft:glass_pane", GlassPane{})
	world.RegisterItem("minecraft:iron_bars", IronBars{})
	world.RegisterItem("minecraft:fence", WoodFence{Wood: wood.Oak()})
	world.RegisterItem("minecraft:fence", WoodFence{Wood: wood.Spruce()})
	world.RegisterItem("minecraft:fence", WoodFence{Wood: wood.Birch()})
	world.RegisterItem("minecraft:fence", WoodFence{Wood: wood.Jungle()})
	world.RegisterItem("minecraft:fence", WoodFence{Wood: wood.Acacia()})
	world.RegisterItem("minecraft:fence", WoodFence{Wood: wood.DarkOak()})
	world.RegisterItem("minecraft:crimson_fence", WoodFence{Wood: wood.Crimson()})
	world.RegisterItem("minecraft:warped_fence", WoodFence{Wood: wood.Warped()})
	world.RegisterItem("minecraft:nether_brick_fence", NetherBrickFence{})
	world.RegisterItem("minecraft:fence_gate", WoodFenceGate{Wood: wood.Oak()})
	world.RegisterItem("minecraft:spruce_fence_gate", WoodFenceGate{Wood: wood.Spruce()})
	world.RegisterItem("minecraft:birch_fence_gate", WoodFenceGate{Wood: wood.Birch()})
	world.RegisterItem("minecraft:jungle_fence_gate", WoodFenceGate{Wood: wood.Jungle()})
	world.RegisterItem("minecraft:acacia_fence_gate", WoodFenceGate{Wood: wood.Acacia()})
	world.RegisterItem("minecraft:dark_oak_fence_gate", WoodFenceGate{Wood: wood.DarkOak()})
	world.RegisterItem("minecraft:crimson_fence_gate", WoodFenceGate{Wood: wood.Crimson()})
	world.RegisterItem("minecraft:warped_fence_gate", WoodFenceGate{Wood: wood.Warped()})
	world.RegisterItem("minecraft:trapdoor", WoodTrapdoor{Wood: wood.Oak()})
	world.RegisterItem("minecraft:spruce_trapdoor", WoodTrapdoor{Wood: wood.Spruce()})
	world.RegisterItem("minecraft:birch_trapdoor", WoodTrapdoor{Wood: wood.Birch()})
	world.RegisterItem("minecraft:jungle_trapdoor", WoodTrapdoor{Wood: wood.Jungle()})
	world.RegisterItem("minecraft:acacia_trapdoor", WoodTrapdoor{Wood: wood.Acacia()})
	world.RegisterItem("minecraft:dark_oak_trapdoor", WoodTrapdoor{Wood: wood.DarkOak()})
	world.RegisterItem("minecraft:crimson_trapdoor", WoodTrapdoor{Wood: wood.Crimson()})
	world.RegisterItem("minecraft:warped_trapdoor", WoodTrapdoor{Wood: wood.Warped()})
	world.RegisterItem("minecraft:wooden_door", WoodDoor{Wood: wood.Oak()})
	world.RegisterItem("minecraft:spruce_door", WoodDoor{Wood: wood.Spruce()})
	world.RegisterItem("minecraft:birch_door", WoodDoor{Wood: wood.Birch()})
	world.RegisterItem("minecraft:jungle_door", WoodDoor{Wood: wood.Jungle()})
	world.RegisterItem("minecraft:acacia_door", WoodDoor{Wood: wood.Acacia()})
	world.RegisterItem("minecraft:dark_oak_door", WoodDoor{Wood: wood.DarkOak()})
	world.RegisterItem("minecraft:crimson_door", WoodDoor{Wood: wood.Crimson()})
	world.RegisterItem("minecraft:warped_door", WoodDoor{Wood: wood.Warped()})
	for _, c := range allCoral() {
		world.RegisterItem("minecraft:coral", c.(world.Item))
	}
	for _, c := range allCoralBlocks() {
		world.RegisterItem("minecraft:coral_block", c.(world.Item))
	}
	world.RegisterItem("minecraft:pumpkin", Pumpkin{})
	world.RegisterItem("minecraft:lit_pumpkin", LitPumpkin{})
	world.RegisterItem("minecraft:carved_pumpkin", Pumpkin{Carved: true})
	world.RegisterItem("minecraft:end_stone", EndStone{})
	world.RegisterItem("minecraft:netherrack", Netherrack{})
	world.RegisterItem("minecraft:clay", Clay{})
	world.RegisterItem("minecraft:bone_block", BoneBlock{})
	world.RegisterItem("minecraft:lantern", Lantern{Type: fire.Normal()})
	world.RegisterItem("minecraft:soul_lantern", Lantern{Type: fire.Soul()})
	world.RegisterItem("minecraft:ancient_debris", AncientDebris{})
	world.RegisterItem("minecraft:emerald_ore", EmeraldOre{})
	world.RegisterItem("minecraft:diamond_ore", DiamondOre{})
	world.RegisterItem("minecraft:lapis_ore", LapisOre{})
	world.RegisterItem("minecraft:nether_gold_ore", NetherGoldOre{})
	world.RegisterItem("minecraft:gold_ore", GoldOre{})
	world.RegisterItem("minecraft:iron_ore", IronOre{})
	world.RegisterItem("minecraft:coal_ore", CoalOre{})
	world.RegisterItem("minecraft:quartz_ore", NetherQuartzOre{})
	world.RegisterItem("minecraft:dye", CocoaBean{})
	world.RegisterItem("minecraft:wheat_seeds", WheatSeeds{})
	world.RegisterItem("minecraft:beetroot_seeds", BeetrootSeeds{})
	world.RegisterItem("minecraft:potato", Potato{})
	world.RegisterItem("minecraft:carrot", Carrot{})
	world.RegisterItem("minecraft:pumpkin_seeds", PumpkinSeeds{})
	world.RegisterItem("minecraft:melon_seeds", MelonSeeds{})
	world.RegisterItem("minecraft:melon_block", Melon{})
	world.RegisterItem("minecraft:sand", Sand{})
	world.RegisterItem("minecraft:sand", Sand{Red: true})
	world.RegisterItem("minecraft:gravel", Gravel{})
	world.RegisterItem("minecraft:brick_block", Bricks{})
	world.RegisterItem("minecraft:soul_sand", SoulSand{})
	world.RegisterItem("minecraft:barrier", Barrier{})
	world.RegisterItem("minecraft:basalt", Basalt{})
	world.RegisterItem("minecraft:polished_basalt", Basalt{Polished: true})
	world.RegisterItem("minecraft:crying_obsidian", CryingObsidian{})
	world.RegisterItem("minecraft:seaLantern", SeaLantern{})
	world.RegisterItem("minecraft:soul_soil", SoulSoil{})
	world.RegisterItem("minecraft:blue_ice", BlueIce{})
	world.RegisterItem("minecraft:gilded_blackstone", GildedBlackstone{})
	world.RegisterItem("minecraft:shroomlight", Shroomlight{})
	world.RegisterItem("minecraft:torch", Torch{Type: fire.Normal()})
	world.RegisterItem("minecraft:soul_torch", Torch{Type: fire.Soul()})
	world.RegisterItem("minecraft:cake", Cake{})
	world.RegisterItem("minecraft:nether_wart", NetherWart{})
	world.RegisterItem("minecraft:invisibleBedrock", InvisibleBedrock{})
	world.RegisterItem("minecraft:noteblock", NoteBlock{})
	world.RegisterItem("minecraft:dragon_egg", DragonEgg{})
	world.RegisterItem("minecraft:tallgrass", GrassPlant{})
	world.RegisterItem("minecraft:nether_sprouts", GrassPlant{Type: grass.NetherSprouts()})
	world.RegisterItem("minecraft:tallgrass", GrassPlant{Type: grass.Fern()})
	world.RegisterItem("minecraft:double_plant", GrassPlant{Type: grass.TallGrass()})
	world.RegisterItem("minecraft:double_plant", GrassPlant{Type: grass.LargeFern()})
}

func init() {
	item_internal.Air = Air{}
	item_internal.IsCarvedPumpkin = func(b world.Item) bool {
		p, ok := b.(Pumpkin)
		return ok && p.Carved
	}
	item_internal.Lava = Lava{Depth: 8, Still: true}
	item_internal.Water = Water{Depth: 8, Still: true}
	item_internal.IsWater = func(b world.Block) bool {
		_, ok := b.(Water)
		return ok
	}
	item_internal.Fire = Fire{Type: fire.Normal(), Age: 0}
}

// readSlice reads an interface slice from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readSlice(m map[string]interface{}, key string) []interface{} {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.([]interface{})
	return b
}

// readString reads a string from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readString(m map[string]interface{}, key string) string {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.(string)
	return b
}

// readInt32 reads an int32 from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readInt32(m map[string]interface{}, key string) int32 {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.(int32)
	return b
}

// readByte reads a byte from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readByte(m map[string]interface{}, key string) byte {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.(byte)
	return b
}

func registerAll(blocks []world.Block) {
	for _, b := range blocks {
		world.RegisterBlock(b)
	}
}
