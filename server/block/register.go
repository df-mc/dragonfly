package block

import (
	"github.com/df-mc/dragonfly/server/block/fire"
	"github.com/df-mc/dragonfly/server/block/grass"
	"github.com/df-mc/dragonfly/server/block/wood"
	"github.com/df-mc/dragonfly/server/internal/item_internal"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
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
	world.RegisterBlock(DirtPath{})
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
	world.RegisterBlock(NoteBlock{})

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
	registerAll(allTorches())
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
	world.RegisterItem(Air{})
	world.RegisterItem(Stone{})
	world.RegisterItem(Stone{Smooth: true})
	world.RegisterItem(Granite{})
	world.RegisterItem(Granite{Polished: true})
	world.RegisterItem(Diorite{})
	world.RegisterItem(Diorite{Polished: true})
	world.RegisterItem(Andesite{})
	world.RegisterItem(Andesite{Polished: true})
	world.RegisterItem(Grass{})
	world.RegisterItem(DirtPath{})
	world.RegisterItem(Dirt{})
	world.RegisterItem(Dirt{Coarse: true})
	world.RegisterItem(Cobblestone{})
	world.RegisterItem(Bedrock{})
	world.RegisterItem(Kelp{})
	world.RegisterItem(Log{Wood: wood.Oak()})
	world.RegisterItem(Log{Wood: wood.Spruce()})
	world.RegisterItem(Log{Wood: wood.Birch()})
	world.RegisterItem(Log{Wood: wood.Jungle()})
	world.RegisterItem(Leaves{Wood: wood.Oak(), Persistent: true})
	world.RegisterItem(Leaves{Wood: wood.Spruce(), Persistent: true})
	world.RegisterItem(Leaves{Wood: wood.Birch(), Persistent: true})
	world.RegisterItem(Leaves{Wood: wood.Jungle(), Persistent: true})
	world.RegisterItem(Chest{})
	world.RegisterItem(Cobblestone{Mossy: true})
	world.RegisterItem(Leaves{Wood: wood.Acacia(), Persistent: true})
	world.RegisterItem(Leaves{Wood: wood.DarkOak(), Persistent: true})
	world.RegisterItem(Log{Wood: wood.Acacia()})
	world.RegisterItem(Log{Wood: wood.DarkOak()})
	world.RegisterItem(Log{Wood: wood.Spruce(), Stripped: true})
	world.RegisterItem(Log{Wood: wood.Birch(), Stripped: true})
	world.RegisterItem(Log{Wood: wood.Jungle(), Stripped: true})
	world.RegisterItem(Log{Wood: wood.Acacia(), Stripped: true})
	world.RegisterItem(Log{Wood: wood.DarkOak(), Stripped: true})
	world.RegisterItem(Log{Wood: wood.Oak(), Stripped: true})
	for _, c := range Colours() {
		world.RegisterItem(Concrete{Colour: c})
		world.RegisterItem(ConcretePowder{Colour: c})
		world.RegisterItem(StainedTerracotta{Colour: c})
		world.RegisterItem(Carpet{Colour: c})
		world.RegisterItem(Wool{Colour: c})
		world.RegisterItem(StainedGlass{Colour: c})
		world.RegisterItem(StainedGlassPane{Colour: c})
		world.RegisterItem(GlazedTerracotta{Colour: c})
	}
	for _, b := range allLight() {
		world.RegisterItem(b.(world.Item))
	}
	for _, w := range wood.All() {
		world.RegisterItem(Planks{Wood: w})
	}
	world.RegisterItem(WoodStairs{Wood: wood.Oak()})
	world.RegisterItem(WoodStairs{Wood: wood.Spruce()})
	world.RegisterItem(WoodStairs{Wood: wood.Birch()})
	world.RegisterItem(WoodStairs{Wood: wood.Jungle()})
	world.RegisterItem(WoodStairs{Wood: wood.Acacia()})
	world.RegisterItem(WoodStairs{Wood: wood.DarkOak()})
	world.RegisterItem(WoodStairs{Wood: wood.Crimson()})
	world.RegisterItem(WoodStairs{Wood: wood.Warped()})
	world.RegisterItem(WoodSlab{Wood: wood.Oak()})
	world.RegisterItem(WoodSlab{Wood: wood.Spruce()})
	world.RegisterItem(WoodSlab{Wood: wood.Birch()})
	world.RegisterItem(WoodSlab{Wood: wood.Jungle()})
	world.RegisterItem(WoodSlab{Wood: wood.Acacia()})
	world.RegisterItem(WoodSlab{Wood: wood.DarkOak()})
	world.RegisterItem(WoodSlab{Wood: wood.Crimson()})
	world.RegisterItem(WoodSlab{Wood: wood.Warped()})
	world.RegisterItem(WoodSlab{Wood: wood.Oak(), Double: true})
	world.RegisterItem(WoodSlab{Wood: wood.Spruce(), Double: true})
	world.RegisterItem(WoodSlab{Wood: wood.Birch(), Double: true})
	world.RegisterItem(WoodSlab{Wood: wood.Jungle(), Double: true})
	world.RegisterItem(WoodSlab{Wood: wood.Acacia(), Double: true})
	world.RegisterItem(WoodSlab{Wood: wood.DarkOak(), Double: true})
	world.RegisterItem(WoodSlab{Wood: wood.Crimson(), Double: true})
	world.RegisterItem(WoodSlab{Wood: wood.Warped(), Double: true})
	world.RegisterItem(Obsidian{})
	world.RegisterItem(DiamondBlock{})
	world.RegisterItem(Glass{})
	world.RegisterItem(Glowstone{})
	world.RegisterItem(EmeraldBlock{})
	world.RegisterItem(EndBricks{})
	world.RegisterItem(EndBrickStairs{})
	world.RegisterItem(NetheriteBlock{})
	world.RegisterItem(GoldBlock{})
	world.RegisterItem(IronBlock{})
	world.RegisterItem(CoalBlock{})
	world.RegisterItem(Beacon{})
	world.RegisterItem(Sponge{})
	world.RegisterItem(Sponge{Wet: true})
	world.RegisterItem(LapisBlock{})
	world.RegisterItem(Terracotta{})
	world.RegisterItem(Quartz{})
	world.RegisterItem(Quartz{Smooth: true})
	world.RegisterItem(ChiseledQuartz{})
	world.RegisterItem(QuartzPillar{})
	world.RegisterItem(QuartzBricks{})
	world.RegisterItem(GlassPane{})
	world.RegisterItem(IronBars{})
	world.RegisterItem(WoodFence{Wood: wood.Oak()})
	world.RegisterItem(WoodFence{Wood: wood.Spruce()})
	world.RegisterItem(WoodFence{Wood: wood.Birch()})
	world.RegisterItem(WoodFence{Wood: wood.Jungle()})
	world.RegisterItem(WoodFence{Wood: wood.Acacia()})
	world.RegisterItem(WoodFence{Wood: wood.DarkOak()})
	world.RegisterItem(WoodFence{Wood: wood.Crimson()})
	world.RegisterItem(WoodFence{Wood: wood.Warped()})
	world.RegisterItem(NetherBrickFence{})
	world.RegisterItem(WoodFenceGate{Wood: wood.Oak()})
	world.RegisterItem(WoodFenceGate{Wood: wood.Spruce()})
	world.RegisterItem(WoodFenceGate{Wood: wood.Birch()})
	world.RegisterItem(WoodFenceGate{Wood: wood.Jungle()})
	world.RegisterItem(WoodFenceGate{Wood: wood.Acacia()})
	world.RegisterItem(WoodFenceGate{Wood: wood.DarkOak()})
	world.RegisterItem(WoodFenceGate{Wood: wood.Crimson()})
	world.RegisterItem(WoodFenceGate{Wood: wood.Warped()})
	world.RegisterItem(WoodTrapdoor{Wood: wood.Oak()})
	world.RegisterItem(WoodTrapdoor{Wood: wood.Spruce()})
	world.RegisterItem(WoodTrapdoor{Wood: wood.Birch()})
	world.RegisterItem(WoodTrapdoor{Wood: wood.Jungle()})
	world.RegisterItem(WoodTrapdoor{Wood: wood.Acacia()})
	world.RegisterItem(WoodTrapdoor{Wood: wood.DarkOak()})
	world.RegisterItem(WoodTrapdoor{Wood: wood.Crimson()})
	world.RegisterItem(WoodTrapdoor{Wood: wood.Warped()})
	world.RegisterItem(WoodDoor{Wood: wood.Oak()})
	world.RegisterItem(WoodDoor{Wood: wood.Spruce()})
	world.RegisterItem(WoodDoor{Wood: wood.Birch()})
	world.RegisterItem(WoodDoor{Wood: wood.Jungle()})
	world.RegisterItem(WoodDoor{Wood: wood.Acacia()})
	world.RegisterItem(WoodDoor{Wood: wood.DarkOak()})
	world.RegisterItem(WoodDoor{Wood: wood.Crimson()})
	world.RegisterItem(WoodDoor{Wood: wood.Warped()})
	for _, c := range allCoral() {
		world.RegisterItem(c.(world.Item))
	}
	for _, c := range allCoralBlocks() {
		world.RegisterItem(c.(world.Item))
	}
	world.RegisterItem(Pumpkin{})
	world.RegisterItem(LitPumpkin{})
	world.RegisterItem(Pumpkin{Carved: true})
	world.RegisterItem(EndStone{})
	world.RegisterItem(Netherrack{})
	world.RegisterItem(Clay{})
	world.RegisterItem(BoneBlock{})
	world.RegisterItem(Lantern{Type: fire.Normal()})
	world.RegisterItem(Lantern{Type: fire.Soul()})
	world.RegisterItem(AncientDebris{})
	world.RegisterItem(EmeraldOre{})
	world.RegisterItem(DiamondOre{})
	world.RegisterItem(LapisOre{})
	world.RegisterItem(NetherGoldOre{})
	world.RegisterItem(GoldOre{})
	world.RegisterItem(IronOre{})
	world.RegisterItem(CoalOre{})
	world.RegisterItem(NetherQuartzOre{})
	world.RegisterItem(CocoaBean{})
	world.RegisterItem(WheatSeeds{})
	world.RegisterItem(BeetrootSeeds{})
	world.RegisterItem(Potato{})
	world.RegisterItem(Carrot{})
	world.RegisterItem(PumpkinSeeds{})
	world.RegisterItem(MelonSeeds{})
	world.RegisterItem(Melon{})
	world.RegisterItem(Sand{})
	world.RegisterItem(Sand{Red: true})
	world.RegisterItem(Gravel{})
	world.RegisterItem(Bricks{})
	world.RegisterItem(SoulSand{})
	world.RegisterItem(Barrier{})
	world.RegisterItem(Basalt{})
	world.RegisterItem(Basalt{Polished: true})
	world.RegisterItem(CryingObsidian{})
	world.RegisterItem(SeaLantern{})
	world.RegisterItem(SoulSoil{})
	world.RegisterItem(BlueIce{})
	world.RegisterItem(GildedBlackstone{})
	world.RegisterItem(Shroomlight{})
	world.RegisterItem(Torch{Type: fire.Normal()})
	world.RegisterItem(Torch{Type: fire.Soul()})
	world.RegisterItem(Cake{})
	world.RegisterItem(NetherWart{})
	world.RegisterItem(InvisibleBedrock{})
	world.RegisterItem(NoteBlock{Pitch: 24})
	world.RegisterItem(DragonEgg{})
	world.RegisterItem(GrassPlant{})
	world.RegisterItem(GrassPlant{Type: grass.NetherSprouts()})
	world.RegisterItem(GrassPlant{Type: grass.Fern()})
	world.RegisterItem(GrassPlant{Type: grass.TallGrass()})
	world.RegisterItem(GrassPlant{Type: grass.LargeFern()})
	world.RegisterItem(Farmland{})

	world.RegisterItem(item.Bucket{Content: Water{}})
	world.RegisterItem(item.Bucket{Content: Lava{}})
}

func init() {
	item_internal.Air = Air{}
	item_internal.IsCarvedPumpkin = func(b world.Item) bool {
		p, ok := b.(Pumpkin)
		return ok && p.Carved
	}
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
