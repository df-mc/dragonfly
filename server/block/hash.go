// Code generated by cmd/blockhash; DO NOT EDIT.

package block

const (
	hashAir = iota
	hashAmethyst
	hashAncientDebris
	hashAndesite
	hashAnvil
	hashBarrel
	hashBarrier
	hashBasalt
	hashBeacon
	hashBedrock
	hashBeetrootSeeds
	hashBlastFurnace
	hashBlueIce
	hashBone
	hashBookshelf
	hashBricks
	hashCactus
	hashCake
	hashCalcite
	hashCarpet
	hashCarrot
	hashChain
	hashChest
	hashChiseledQuartz
	hashClay
	hashCoal
	hashCoalOre
	hashCobblestone
	hashCobblestoneStairs
	hashCocoaBean
	hashConcrete
	hashConcretePowder
	hashCopperOre
	hashCoral
	hashCoralBlock
	hashCraftingTable
	hashDeadBush
	hashDiamond
	hashDiamondOre
	hashDiorite
	hashDirt
	hashDirtPath
	hashDoubleFlower
	hashDoubleTallGrass
	hashDragonEgg
	hashDriedKelp
	hashDripstone
	hashEmerald
	hashEmeraldOre
	hashEnchantingTable
	hashEndBrickStairs
	hashEndBricks
	hashEndStone
	hashEnderChest
	hashFarmland
	hashFire
	hashFlower
	hashFroglight
	hashFurnace
	hashGildedBlackstone
	hashGlass
	hashGlassPane
	hashGlazedTerracotta
	hashGlowstone
	hashGold
	hashGoldOre
	hashGranite
	hashGrass
	hashGravel
	hashHayBale
	hashHoneycomb
	hashInvisibleBedrock
	hashIron
	hashIronBars
	hashIronOre
	hashItemFrame
	hashKelp
	hashLadder
	hashLantern
	hashLapis
	hashLapisOre
	hashLava
	hashLeaves
	hashLight
	hashLitPumpkin
	hashLog
	hashMelon
	hashMelonSeeds
	hashMossCarpet
	hashMud
	hashMudBricks
	hashMuddyMangroveRoots
	hashNetherBrickFence
	hashNetherBricks
	hashNetherGoldOre
	hashNetherQuartzOre
	hashNetherSprouts
	hashNetherWart
	hashNetherWartBlock
	hashNetherite
	hashNetherrack
	hashNote
	hashObsidian
	hashPackedIce
	hashPackedMud
	hashPlanks
	hashPodzol
	hashPotato
	hashPrismarine
	hashPumpkin
	hashPumpkinSeeds
	hashPurpur
	hashPurpurPillar
	hashQuartz
	hashQuartzBricks
	hashQuartzPillar
	hashQuartzStairs
	hashRawCopper
	hashRawGold
	hashRawIron
	hashReinforcedDeepslate
	hashSand
	hashSandstone
	hashSandstoneSlab
	hashSandstoneStairs
	hashSeaLantern
	hashSeaPickle
	hashShroomlight
	hashSign
	hashSkull
	hashSmithingTable
	hashSmoker
	hashSnow
	hashSoulSand
	hashSoulSoil
	hashSponge
	hashSporeBlossom
	hashStainedGlass
	hashStainedGlassPane
	hashStainedTerracotta
	hashStone
	hashStoneBrickStairs
	hashStoneBricks
	hashTallGrass
	hashTerracotta
	hashTorch
	hashTuff
	hashWater
	hashWheatSeeds
	hashWood
	hashWoodDoor
	hashWoodFence
	hashWoodFenceGate
	hashWoodSlab
	hashWoodStairs
	hashWoodTrapdoor
	hashWool
	hashBase
)

// base represents the base hash for all custom blocks.
var base = uint64(hashBase - 1)

// NextHash returns the next free hash for custom blocks.
func NextHash() uint64 {
	base++
	return base
}

// Hash ...
func (Air) Hash() uint64 {
	return hashAir
}

// Hash ...
func (Amethyst) Hash() uint64 {
	return hashAmethyst
}

// Hash ...
func (AncientDebris) Hash() uint64 {
	return hashAncientDebris
}

// Hash ...
func (a Andesite) Hash() uint64 {
	return hashAndesite | uint64(boolByte(a.Polished))<<8
}

// Hash ...
func (a Anvil) Hash() uint64 {
	return hashAnvil | uint64(a.Type.Uint8())<<8 | uint64(a.Facing)<<10
}

// Hash ...
func (b Barrel) Hash() uint64 {
	return hashBarrel | uint64(b.Facing)<<8 | uint64(boolByte(b.Open))<<11
}

// Hash ...
func (Barrier) Hash() uint64 {
	return hashBarrier
}

// Hash ...
func (b Basalt) Hash() uint64 {
	return hashBasalt | uint64(boolByte(b.Polished))<<8 | uint64(b.Axis)<<9
}

// Hash ...
func (Beacon) Hash() uint64 {
	return hashBeacon
}

// Hash ...
func (b Bedrock) Hash() uint64 {
	return hashBedrock | uint64(boolByte(b.InfiniteBurning))<<8
}

// Hash ...
func (b BeetrootSeeds) Hash() uint64 {
	return hashBeetrootSeeds | uint64(b.Growth)<<8
}

// Hash ...
func (b BlastFurnace) Hash() uint64 {
	return hashBlastFurnace | uint64(b.Facing)<<8 | uint64(boolByte(b.Lit))<<11
}

// Hash ...
func (BlueIce) Hash() uint64 {
	return hashBlueIce
}

// Hash ...
func (b Bone) Hash() uint64 {
	return hashBone | uint64(b.Axis)<<8
}

// Hash ...
func (Bookshelf) Hash() uint64 {
	return hashBookshelf
}

// Hash ...
func (Bricks) Hash() uint64 {
	return hashBricks
}

// Hash ...
func (c Cactus) Hash() uint64 {
	return hashCactus | uint64(c.Age)<<8
}

// Hash ...
func (c Cake) Hash() uint64 {
	return hashCake | uint64(c.Bites)<<8
}

// Hash ...
func (Calcite) Hash() uint64 {
	return hashCalcite
}

// Hash ...
func (c Carpet) Hash() uint64 {
	return hashCarpet | uint64(c.Colour.Uint8())<<8
}

// Hash ...
func (c Carrot) Hash() uint64 {
	return hashCarrot | uint64(c.Growth)<<8
}

// Hash ...
func (c Chain) Hash() uint64 {
	return hashChain | uint64(c.Axis)<<8
}

// Hash ...
func (c Chest) Hash() uint64 {
	return hashChest | uint64(c.Facing)<<8
}

// Hash ...
func (ChiseledQuartz) Hash() uint64 {
	return hashChiseledQuartz
}

// Hash ...
func (Clay) Hash() uint64 {
	return hashClay
}

// Hash ...
func (Coal) Hash() uint64 {
	return hashCoal
}

// Hash ...
func (c CoalOre) Hash() uint64 {
	return hashCoalOre | uint64(c.Type.Uint8())<<8
}

// Hash ...
func (c Cobblestone) Hash() uint64 {
	return hashCobblestone | uint64(boolByte(c.Mossy))<<8
}

// Hash ...
func (s CobblestoneStairs) Hash() uint64 {
	return hashCobblestoneStairs | uint64(boolByte(s.Mossy))<<8 | uint64(boolByte(s.UpsideDown))<<9 | uint64(s.Facing)<<10
}

// Hash ...
func (c CocoaBean) Hash() uint64 {
	return hashCocoaBean | uint64(c.Facing)<<8 | uint64(c.Age)<<10
}

// Hash ...
func (c Concrete) Hash() uint64 {
	return hashConcrete | uint64(c.Colour.Uint8())<<8
}

// Hash ...
func (c ConcretePowder) Hash() uint64 {
	return hashConcretePowder | uint64(c.Colour.Uint8())<<8
}

// Hash ...
func (c CopperOre) Hash() uint64 {
	return hashCopperOre | uint64(c.Type.Uint8())<<8
}

// Hash ...
func (c Coral) Hash() uint64 {
	return hashCoral | uint64(c.Type.Uint8())<<8 | uint64(boolByte(c.Dead))<<11
}

// Hash ...
func (c CoralBlock) Hash() uint64 {
	return hashCoralBlock | uint64(c.Type.Uint8())<<8 | uint64(boolByte(c.Dead))<<11
}

// Hash ...
func (CraftingTable) Hash() uint64 {
	return hashCraftingTable
}

// Hash ...
func (DeadBush) Hash() uint64 {
	return hashDeadBush
}

// Hash ...
func (Diamond) Hash() uint64 {
	return hashDiamond
}

// Hash ...
func (d DiamondOre) Hash() uint64 {
	return hashDiamondOre | uint64(d.Type.Uint8())<<8
}

// Hash ...
func (d Diorite) Hash() uint64 {
	return hashDiorite | uint64(boolByte(d.Polished))<<8
}

// Hash ...
func (d Dirt) Hash() uint64 {
	return hashDirt | uint64(boolByte(d.Coarse))<<8
}

// Hash ...
func (DirtPath) Hash() uint64 {
	return hashDirtPath
}

// Hash ...
func (d DoubleFlower) Hash() uint64 {
	return hashDoubleFlower | uint64(boolByte(d.UpperPart))<<8 | uint64(d.Type.Uint8())<<9
}

// Hash ...
func (d DoubleTallGrass) Hash() uint64 {
	return hashDoubleTallGrass | uint64(boolByte(d.UpperPart))<<8 | uint64(d.Type.Uint8())<<9
}

// Hash ...
func (DragonEgg) Hash() uint64 {
	return hashDragonEgg
}

// Hash ...
func (DriedKelp) Hash() uint64 {
	return hashDriedKelp
}

// Hash ...
func (Dripstone) Hash() uint64 {
	return hashDripstone
}

// Hash ...
func (Emerald) Hash() uint64 {
	return hashEmerald
}

// Hash ...
func (e EmeraldOre) Hash() uint64 {
	return hashEmeraldOre | uint64(e.Type.Uint8())<<8
}

// Hash ...
func (EnchantingTable) Hash() uint64 {
	return hashEnchantingTable
}

// Hash ...
func (s EndBrickStairs) Hash() uint64 {
	return hashEndBrickStairs | uint64(boolByte(s.UpsideDown))<<8 | uint64(s.Facing)<<9
}

// Hash ...
func (EndBricks) Hash() uint64 {
	return hashEndBricks
}

// Hash ...
func (EndStone) Hash() uint64 {
	return hashEndStone
}

// Hash ...
func (c EnderChest) Hash() uint64 {
	return hashEnderChest | uint64(c.Facing)<<8
}

// Hash ...
func (f Farmland) Hash() uint64 {
	return hashFarmland | uint64(f.Hydration)<<8
}

// Hash ...
func (f Fire) Hash() uint64 {
	return hashFire | uint64(f.Type.Uint8())<<8 | uint64(f.Age)<<9
}

// Hash ...
func (f Flower) Hash() uint64 {
	return hashFlower | uint64(f.Type.Uint8())<<8
}

// Hash ...
func (f Froglight) Hash() uint64 {
	return hashFroglight | uint64(f.Type.Uint8())<<8 | uint64(f.Axis)<<10
}

// Hash ...
func (f Furnace) Hash() uint64 {
	return hashFurnace | uint64(f.Facing)<<8 | uint64(boolByte(f.Lit))<<11
}

// Hash ...
func (GildedBlackstone) Hash() uint64 {
	return hashGildedBlackstone
}

// Hash ...
func (Glass) Hash() uint64 {
	return hashGlass
}

// Hash ...
func (GlassPane) Hash() uint64 {
	return hashGlassPane
}

// Hash ...
func (t GlazedTerracotta) Hash() uint64 {
	return hashGlazedTerracotta | uint64(t.Colour.Uint8())<<8 | uint64(t.Facing)<<12
}

// Hash ...
func (Glowstone) Hash() uint64 {
	return hashGlowstone
}

// Hash ...
func (Gold) Hash() uint64 {
	return hashGold
}

// Hash ...
func (g GoldOre) Hash() uint64 {
	return hashGoldOre | uint64(g.Type.Uint8())<<8
}

// Hash ...
func (g Granite) Hash() uint64 {
	return hashGranite | uint64(boolByte(g.Polished))<<8
}

// Hash ...
func (Grass) Hash() uint64 {
	return hashGrass
}

// Hash ...
func (Gravel) Hash() uint64 {
	return hashGravel
}

// Hash ...
func (h HayBale) Hash() uint64 {
	return hashHayBale | uint64(h.Axis)<<8
}

// Hash ...
func (Honeycomb) Hash() uint64 {
	return hashHoneycomb
}

// Hash ...
func (InvisibleBedrock) Hash() uint64 {
	return hashInvisibleBedrock
}

// Hash ...
func (Iron) Hash() uint64 {
	return hashIron
}

// Hash ...
func (IronBars) Hash() uint64 {
	return hashIronBars
}

// Hash ...
func (i IronOre) Hash() uint64 {
	return hashIronOre | uint64(i.Type.Uint8())<<8
}

// Hash ...
func (i ItemFrame) Hash() uint64 {
	return hashItemFrame | uint64(i.Facing)<<8 | uint64(boolByte(i.Glowing))<<11
}

// Hash ...
func (k Kelp) Hash() uint64 {
	return hashKelp | uint64(k.Age)<<8
}

// Hash ...
func (l Ladder) Hash() uint64 {
	return hashLadder | uint64(l.Facing)<<8
}

// Hash ...
func (l Lantern) Hash() uint64 {
	return hashLantern | uint64(boolByte(l.Hanging))<<8 | uint64(l.Type.Uint8())<<9
}

// Hash ...
func (Lapis) Hash() uint64 {
	return hashLapis
}

// Hash ...
func (l LapisOre) Hash() uint64 {
	return hashLapisOre | uint64(l.Type.Uint8())<<8
}

// Hash ...
func (l Lava) Hash() uint64 {
	return hashLava | uint64(boolByte(l.Still))<<8 | uint64(l.Depth)<<9 | uint64(boolByte(l.Falling))<<17
}

// Hash ...
func (l Leaves) Hash() uint64 {
	return hashLeaves | uint64(l.Wood.Uint8())<<8 | uint64(boolByte(l.Persistent))<<12 | uint64(boolByte(l.ShouldUpdate))<<13
}

// Hash ...
func (l Light) Hash() uint64 {
	return hashLight | uint64(l.Level)<<8
}

// Hash ...
func (l LitPumpkin) Hash() uint64 {
	return hashLitPumpkin | uint64(l.Facing)<<8
}

// Hash ...
func (l Log) Hash() uint64 {
	return hashLog | uint64(l.Wood.Uint8())<<8 | uint64(boolByte(l.Stripped))<<12 | uint64(l.Axis)<<13
}

// Hash ...
func (Melon) Hash() uint64 {
	return hashMelon
}

// Hash ...
func (m MelonSeeds) Hash() uint64 {
	return hashMelonSeeds | uint64(m.Growth)<<8 | uint64(m.Direction)<<16
}

// Hash ...
func (MossCarpet) Hash() uint64 {
	return hashMossCarpet
}

// Hash ...
func (Mud) Hash() uint64 {
	return hashMud
}

// Hash ...
func (MudBricks) Hash() uint64 {
	return hashMudBricks
}

// Hash ...
func (MuddyMangroveRoots) Hash() uint64 {
	return hashMuddyMangroveRoots
}

// Hash ...
func (NetherBrickFence) Hash() uint64 {
	return hashNetherBrickFence
}

// Hash ...
func (n NetherBricks) Hash() uint64 {
	return hashNetherBricks | uint64(n.Type.Uint8())<<8
}

// Hash ...
func (NetherGoldOre) Hash() uint64 {
	return hashNetherGoldOre
}

// Hash ...
func (NetherQuartzOre) Hash() uint64 {
	return hashNetherQuartzOre
}

// Hash ...
func (NetherSprouts) Hash() uint64 {
	return hashNetherSprouts
}

// Hash ...
func (n NetherWart) Hash() uint64 {
	return hashNetherWart | uint64(n.Age)<<8
}

// Hash ...
func (n NetherWartBlock) Hash() uint64 {
	return hashNetherWartBlock | uint64(boolByte(n.Warped))<<8
}

// Hash ...
func (Netherite) Hash() uint64 {
	return hashNetherite
}

// Hash ...
func (Netherrack) Hash() uint64 {
	return hashNetherrack
}

// Hash ...
func (Note) Hash() uint64 {
	return hashNote
}

// Hash ...
func (o Obsidian) Hash() uint64 {
	return hashObsidian | uint64(boolByte(o.Crying))<<8
}

// Hash ...
func (PackedIce) Hash() uint64 {
	return hashPackedIce
}

// Hash ...
func (PackedMud) Hash() uint64 {
	return hashPackedMud
}

// Hash ...
func (p Planks) Hash() uint64 {
	return hashPlanks | uint64(p.Wood.Uint8())<<8
}

// Hash ...
func (Podzol) Hash() uint64 {
	return hashPodzol
}

// Hash ...
func (p Potato) Hash() uint64 {
	return hashPotato | uint64(p.Growth)<<8
}

// Hash ...
func (p Prismarine) Hash() uint64 {
	return hashPrismarine | uint64(p.Type.Uint8())<<8
}

// Hash ...
func (p Pumpkin) Hash() uint64 {
	return hashPumpkin | uint64(boolByte(p.Carved))<<8 | uint64(p.Facing)<<9
}

// Hash ...
func (p PumpkinSeeds) Hash() uint64 {
	return hashPumpkinSeeds | uint64(p.Growth)<<8 | uint64(p.Direction)<<16
}

// Hash ...
func (Purpur) Hash() uint64 {
	return hashPurpur
}

// Hash ...
func (p PurpurPillar) Hash() uint64 {
	return hashPurpurPillar | uint64(p.Axis)<<8
}

// Hash ...
func (q Quartz) Hash() uint64 {
	return hashQuartz | uint64(boolByte(q.Smooth))<<8
}

// Hash ...
func (QuartzBricks) Hash() uint64 {
	return hashQuartzBricks
}

// Hash ...
func (q QuartzPillar) Hash() uint64 {
	return hashQuartzPillar | uint64(q.Axis)<<8
}

// Hash ...
func (s QuartzStairs) Hash() uint64 {
	return hashQuartzStairs | uint64(boolByte(s.UpsideDown))<<8 | uint64(s.Facing)<<9 | uint64(boolByte(s.Smooth))<<11
}

// Hash ...
func (RawCopper) Hash() uint64 {
	return hashRawCopper
}

// Hash ...
func (RawGold) Hash() uint64 {
	return hashRawGold
}

// Hash ...
func (RawIron) Hash() uint64 {
	return hashRawIron
}

// Hash ...
func (ReinforcedDeepslate) Hash() uint64 {
	return hashReinforcedDeepslate
}

// Hash ...
func (s Sand) Hash() uint64 {
	return hashSand | uint64(boolByte(s.Red))<<8
}

// Hash ...
func (s Sandstone) Hash() uint64 {
	return hashSandstone | uint64(s.Type.Uint8())<<8 | uint64(boolByte(s.Red))<<10
}

// Hash ...
func (s SandstoneSlab) Hash() uint64 {
	return hashSandstoneSlab | uint64(s.Type.Uint8())<<8 | uint64(boolByte(s.Red))<<10 | uint64(boolByte(s.Top))<<11 | uint64(boolByte(s.Double))<<12
}

// Hash ...
func (s SandstoneStairs) Hash() uint64 {
	return hashSandstoneStairs | uint64(s.Type.Uint8())<<8 | uint64(boolByte(s.Red))<<10 | uint64(boolByte(s.UpsideDown))<<11 | uint64(s.Facing)<<12
}

// Hash ...
func (SeaLantern) Hash() uint64 {
	return hashSeaLantern
}

// Hash ...
func (s SeaPickle) Hash() uint64 {
	return hashSeaPickle | uint64(s.AdditionalCount)<<8 | uint64(boolByte(s.Dead))<<16
}

// Hash ...
func (Shroomlight) Hash() uint64 {
	return hashShroomlight
}

// Hash ...
func (s Sign) Hash() uint64 {
	return hashSign | uint64(s.Wood.Uint8())<<8 | uint64(s.Attach.Uint8())<<12
}

// Hash ...
func (s Skull) Hash() uint64 {
	return hashSkull | uint64(s.Attach.FaceUint8())<<8
}

// Hash ...
func (SmithingTable) Hash() uint64 {
	return hashSmithingTable
}

// Hash ...
func (s Smoker) Hash() uint64 {
	return hashSmoker | uint64(s.Facing)<<8 | uint64(boolByte(s.Lit))<<11
}

// Hash ...
func (Snow) Hash() uint64 {
	return hashSnow
}

// Hash ...
func (SoulSand) Hash() uint64 {
	return hashSoulSand
}

// Hash ...
func (SoulSoil) Hash() uint64 {
	return hashSoulSoil
}

// Hash ...
func (s Sponge) Hash() uint64 {
	return hashSponge | uint64(boolByte(s.Wet))<<8
}

// Hash ...
func (SporeBlossom) Hash() uint64 {
	return hashSporeBlossom
}

// Hash ...
func (g StainedGlass) Hash() uint64 {
	return hashStainedGlass | uint64(g.Colour.Uint8())<<8
}

// Hash ...
func (p StainedGlassPane) Hash() uint64 {
	return hashStainedGlassPane | uint64(p.Colour.Uint8())<<8
}

// Hash ...
func (t StainedTerracotta) Hash() uint64 {
	return hashStainedTerracotta | uint64(t.Colour.Uint8())<<8
}

// Hash ...
func (s Stone) Hash() uint64 {
	return hashStone | uint64(boolByte(s.Smooth))<<8
}

// Hash ...
func (s StoneBrickStairs) Hash() uint64 {
	return hashStoneBrickStairs | uint64(boolByte(s.Mossy))<<8 | uint64(boolByte(s.UpsideDown))<<9 | uint64(s.Facing)<<10
}

// Hash ...
func (s StoneBricks) Hash() uint64 {
	return hashStoneBricks | uint64(s.Type.Uint8())<<8
}

// Hash ...
func (g TallGrass) Hash() uint64 {
	return hashTallGrass | uint64(g.Type.Uint8())<<8
}

// Hash ...
func (Terracotta) Hash() uint64 {
	return hashTerracotta
}

// Hash ...
func (t Torch) Hash() uint64 {
	return hashTorch | uint64(t.Facing)<<8 | uint64(t.Type.Uint8())<<11
}

// Hash ...
func (Tuff) Hash() uint64 {
	return hashTuff
}

// Hash ...
func (w Water) Hash() uint64 {
	return hashWater | uint64(boolByte(w.Still))<<8 | uint64(w.Depth)<<9 | uint64(boolByte(w.Falling))<<17
}

// Hash ...
func (s WheatSeeds) Hash() uint64 {
	return hashWheatSeeds | uint64(s.Growth)<<8
}

// Hash ...
func (w Wood) Hash() uint64 {
	return hashWood | uint64(w.Wood.Uint8())<<8 | uint64(boolByte(w.Stripped))<<12 | uint64(w.Axis)<<13
}

// Hash ...
func (d WoodDoor) Hash() uint64 {
	return hashWoodDoor | uint64(d.Wood.Uint8())<<8 | uint64(d.Facing)<<12 | uint64(boolByte(d.Open))<<14 | uint64(boolByte(d.Top))<<15 | uint64(boolByte(d.Right))<<16
}

// Hash ...
func (w WoodFence) Hash() uint64 {
	return hashWoodFence | uint64(w.Wood.Uint8())<<8
}

// Hash ...
func (f WoodFenceGate) Hash() uint64 {
	return hashWoodFenceGate | uint64(f.Wood.Uint8())<<8 | uint64(f.Facing)<<12 | uint64(boolByte(f.Open))<<14 | uint64(boolByte(f.Lowered))<<15
}

// Hash ...
func (s WoodSlab) Hash() uint64 {
	return hashWoodSlab | uint64(s.Wood.Uint8())<<8 | uint64(boolByte(s.Top))<<12 | uint64(boolByte(s.Double))<<13
}

// Hash ...
func (s WoodStairs) Hash() uint64 {
	return hashWoodStairs | uint64(s.Wood.Uint8())<<8 | uint64(boolByte(s.UpsideDown))<<12 | uint64(s.Facing)<<13
}

// Hash ...
func (t WoodTrapdoor) Hash() uint64 {
	return hashWoodTrapdoor | uint64(t.Wood.Uint8())<<8 | uint64(t.Facing)<<12 | uint64(boolByte(t.Open))<<14 | uint64(boolByte(t.Top))<<15
}

// Hash ...
func (w Wool) Hash() uint64 {
	return hashWool | uint64(w.Colour.Uint8())<<8
}
