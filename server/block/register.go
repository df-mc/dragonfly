package block

import (
	_ "github.com/df-mc/dragonfly/server/internal/block_internal"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	_ "unsafe" // Imported for compiler directives.
)

//go:generate go run ../../cmd/blockhash -o hash.go .

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
	world.RegisterBlock(Obsidian{Crying: true})
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
	world.RegisterBlock(NetherGoldOre{})
	world.RegisterBlock(NetherQuartzOre{})
	world.RegisterBlock(Melon{})
	world.RegisterBlock(Sand{})
	world.RegisterBlock(Sand{Red: true})
	world.RegisterBlock(Gravel{})
	world.RegisterBlock(Bricks{})
	world.RegisterBlock(SoulSand{})
	world.RegisterBlock(Barrier{})
	world.RegisterBlock(SeaLantern{})
	world.RegisterBlock(SoulSoil{})
	world.RegisterBlock(BlueIce{})
	world.RegisterBlock(GildedBlackstone{})
	world.RegisterBlock(Shroomlight{})
	world.RegisterBlock(InvisibleBedrock{})
	world.RegisterBlock(DragonEgg{})
	world.RegisterBlock(NoteBlock{})
	world.RegisterBlock(NetherSprouts{})
	world.RegisterBlock(Tuff{})
	world.RegisterBlock(Calcite{})
	for _, ore := range OreTypes() {
		world.RegisterBlock(CoalOre{Type: ore})
		world.RegisterBlock(IronOre{Type: ore})
		world.RegisterBlock(GoldOre{Type: ore})
		world.RegisterBlock(CopperOre{Type: ore})
		world.RegisterBlock(LapisOre{Type: ore})
		world.RegisterBlock(DiamondOre{Type: ore})
		world.RegisterBlock(EmeraldOre{Type: ore})
	}
	world.RegisterBlock(RawIronBlock{})
	world.RegisterBlock(RawGoldBlock{})
	world.RegisterBlock(RawCopperBlock{})
	world.RegisterBlock(MossCarpet{})
	world.RegisterBlock(SporeBlossom{})
	world.RegisterBlock(Dripstone{})
	world.RegisterBlock(DriedKelpBlock{})
	world.RegisterBlock(HoneycombBlock{})
	world.RegisterBlock(Podzol{})
	world.RegisterBlock(AmethystBlock{})
	world.RegisterBlock(PackedIce{})
	world.RegisterBlock(DeadBush{})
	world.RegisterBlock(Snow{})
	world.RegisterBlock(Bookshelf{})
	world.RegisterBlock(NetherWartBlock{})
	world.RegisterBlock(NetherWartBlock{Warped: true})

	registerAll(allBarrels())
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
	registerAll(allItemFrames())
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
	registerAll(allNetherBricks())
	registerAll(allTallGrass())
	registerAll(allDoubleTallGrass())
	registerAll(allSandstones())
	registerAll(allStoneBricks())
	registerAll(allDoubleFlowers())
	registerAll(allFlowers())
	registerAll(allPrismarine())
	registerAll(allSigns())
	registerAll(allLight())
	registerAll(allLadders())
	registerAll(allSandstoneStairs())
	registerAll(allSeaPickles())
	registerAll(allWood())
	registerAll(allChains())
	registerAll(allHayBales())
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
	world.RegisterItem(Chest{})
	world.RegisterItem(Cobblestone{Mossy: true})
	world.RegisterItem(Obsidian{})
	world.RegisterItem(Obsidian{Crying: true})
	world.RegisterItem(DiamondBlock{})
	world.RegisterItem(Glass{})
	world.RegisterItem(Glowstone{})
	world.RegisterItem(EmeraldBlock{})
	world.RegisterItem(EndBricks{})
	world.RegisterItem(EndBrickStairs{})
	world.RegisterItem(NetheriteBlock{})
	world.RegisterItem(GoldBlock{})
	world.RegisterItem(IronBlock{})
	world.RegisterItem(ItemFrame{})
	world.RegisterItem(ItemFrame{Glowing: true})
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
	world.RegisterItem(NetherBrickFence{})
	world.RegisterItem(Barrel{})
	world.RegisterItem(Pumpkin{})
	world.RegisterItem(LitPumpkin{})
	world.RegisterItem(Pumpkin{Carved: true})
	world.RegisterItem(EndStone{})
	world.RegisterItem(Netherrack{})
	world.RegisterItem(Clay{})
	world.RegisterItem(BoneBlock{})
	world.RegisterItem(AncientDebris{})
	world.RegisterItem(NetherGoldOre{})
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
	world.RegisterItem(SeaLantern{})
	world.RegisterItem(SoulSoil{})
	world.RegisterItem(BlueIce{})
	world.RegisterItem(GildedBlackstone{})
	world.RegisterItem(Shroomlight{})
	world.RegisterItem(Cake{})
	world.RegisterItem(NetherWart{})
	world.RegisterItem(InvisibleBedrock{})
	world.RegisterItem(NoteBlock{Pitch: 24})
	world.RegisterItem(DragonEgg{})
	world.RegisterItem(NetherSprouts{})
	world.RegisterItem(Farmland{})
	world.RegisterItem(Tuff{})
	world.RegisterItem(Calcite{})
	world.RegisterItem(RawIronBlock{})
	world.RegisterItem(RawGoldBlock{})
	world.RegisterItem(RawCopperBlock{})
	world.RegisterItem(MossCarpet{})
	world.RegisterItem(SporeBlossom{})
	world.RegisterItem(Dripstone{})
	world.RegisterItem(DriedKelpBlock{})
	world.RegisterItem(HoneycombBlock{})
	world.RegisterItem(Podzol{})
	world.RegisterItem(Ladder{})
	world.RegisterItem(AmethystBlock{})
	world.RegisterItem(PackedIce{})
	world.RegisterItem(DeadBush{})
	world.RegisterItem(SeaPickle{})
	world.RegisterItem(Snow{})
	world.RegisterItem(Bookshelf{})
	world.RegisterItem(Chain{})

	world.RegisterItem(item.Bucket{Content: Water{}})
	world.RegisterItem(item.Bucket{Content: Lava{}})

	for _, b := range allLight() {
		world.RegisterItem(b.(world.Item))
	}
	for _, c := range allCoral() {
		world.RegisterItem(c.(world.Item))
	}
	for _, c := range allCoralBlocks() {
		world.RegisterItem(c.(world.Item))
	}
	for _, s := range allSandstones() {
		world.RegisterItem(s.(world.Item))
	}
	for _, s := range allStoneBricks() {
		world.RegisterItem(s.(world.Item))
	}
	for _, c := range item.Colours() {
		world.RegisterItem(Concrete{Colour: c})
		world.RegisterItem(ConcretePowder{Colour: c})
		world.RegisterItem(StainedTerracotta{Colour: c})
		world.RegisterItem(Carpet{Colour: c})
		world.RegisterItem(Wool{Colour: c})
		world.RegisterItem(StainedGlass{Colour: c})
		world.RegisterItem(StainedGlassPane{Colour: c})
		world.RegisterItem(GlazedTerracotta{Colour: c})
	}
	for _, w := range WoodTypes() {
		world.RegisterItem(Log{Wood: w})
		world.RegisterItem(Log{Wood: w, Stripped: true})
		if w != WarpedWood() && w != CrimsonWood() {
			world.RegisterItem(Leaves{Wood: w, Persistent: true})
		}
		world.RegisterItem(Planks{Wood: w})
		world.RegisterItem(WoodStairs{Wood: w})
		world.RegisterItem(WoodSlab{Wood: w})
		world.RegisterItem(WoodSlab{Wood: w, Double: true})
		world.RegisterItem(WoodFence{Wood: w})
		world.RegisterItem(WoodFenceGate{Wood: w})
		world.RegisterItem(WoodTrapdoor{Wood: w})
		world.RegisterItem(WoodDoor{Wood: w})
		world.RegisterItem(Sign{Wood: w})
		world.RegisterItem(Wood{Wood: w})
		world.RegisterItem(Wood{Wood: w, Stripped: true})
	}
	for _, ore := range OreTypes() {
		world.RegisterItem(CoalOre{Type: ore})
		world.RegisterItem(IronOre{Type: ore})
		world.RegisterItem(GoldOre{Type: ore})
		world.RegisterItem(CopperOre{Type: ore})
		world.RegisterItem(LapisOre{Type: ore})
		world.RegisterItem(DiamondOre{Type: ore})
		world.RegisterItem(EmeraldOre{Type: ore})
	}
	for _, f := range FireTypes() {
		world.RegisterItem(Lantern{Type: f})
		world.RegisterItem(Torch{Type: f})
	}
	for _, f := range FlowerTypes() {
		world.RegisterItem(Flower{Type: f})
	}
	for _, f := range DoubleFlowerTypes() {
		world.RegisterItem(DoubleFlower{Type: f})
	}
	for _, g := range GrassTypes() {
		world.RegisterItem(TallGrass{Type: g})
		world.RegisterItem(DoubleTallGrass{Type: g})
	}
	for _, p := range PrismarineTypes() {
		world.RegisterItem(Prismarine{Type: p})
	}
}

func registerAll(blocks []world.Block) {
	for _, b := range blocks {
		world.RegisterBlock(b)
	}
}
