// Code generated by cmd/blockhash; DO NOT EDIT.

package block

const hashSoulSand = 0
const hashCarrot = 1
const hashDiorite = 2
const hashPlanks = 3
const hashStone = 4
const hashGoldBlock = 5
const hashGildedBlackstone = 6
const hashEndStone = 7
const hashFarmland = 8
const hashBarrier = 9
const hashSoulSoil = 10
const hashDiamondOre = 11
const hashEndBrickStairs = 12
const hashStainedGlass = 13
const hashCoalBlock = 14
const hashGlass = 15
const hashLava = 16
const hashWool = 17
const hashSandstone = 18
const hashMelonSeeds = 19
const hashNetherrack = 20
const hashWoodTrapdoor = 21
const hashCake = 22
const hashConcrete = 23
const hashChiseledSandstone = 24
const hashPumpkin = 25
const hashAir = 26
const hashLapisBlock = 27
const hashEndBricks = 28
const hashTerracotta = 29
const hashGlassPane = 30
const hashAndesite = 31
const hashQuartzPillar = 32
const hashGlazedTerracotta = 33
const hashEmeraldOre = 34
const hashStainedGlassPane = 35
const hashWoodStairs = 36
const hashChiseledQuartz = 37
const hashNetherGoldOre = 38
const hashWheatSeeds = 39
const hashKelp = 40
const hashFire = 41
const hashQuartzBricks = 42
const hashNetherQuartzOre = 43
const hashSand = 44
const hashClay = 45
const hashGoldOre = 46
const hashInvisibleBedrock = 47
const hashBeacon = 48
const hashTorch = 49
const hashGrass = 50
const hashCobblestone = 51
const hashBeetrootSeeds = 52
const hashLog = 53
const hashLeaves = 54
const hashCutSandstone = 55
const hashIronOre = 56
const hashCoral = 57
const hashBedrock = 58
const hashCarpet = 59
const hashDiamondBlock = 60
const hashSeaLantern = 61
const hashCocoaBean = 62
const hashWater = 63
const hashWoodSlab = 64
const hashBricks = 65
const hashLantern = 66
const hashDragonEgg = 67
const hashStainedTerracotta = 68
const hashDirtPath = 69
const hashCoralBlock = 70
const hashEmeraldBlock = 71
const hashNetherBrickFence = 72
const hashObsidian = 73
const hashGranite = 74
const hashWoodDoor = 75
const hashLapisOre = 76
const hashGrassPlant = 77
const hashNetherWart = 78
const hashWoodFence = 79
const hashMelon = 80
const hashChest = 81
const hashLight = 82
const hashGravel = 83
const hashAncientDebris = 84
const hashWoodFenceGate = 85
const hashNetheriteBlock = 86
const hashQuartz = 87
const hashIronBars = 88
const hashGlowstone = 89
const hashCoalOre = 90
const hashDirt = 91
const hashBoneBlock = 92
const hashBlueIce = 93
const hashConcretePowder = 94
const hashPumpkinSeeds = 95
const hashLitPumpkin = 96
const hashNoteBlock = 97
const hashPotato = 98
const hashSponge = 99
const hashSmoothSandstone = 100
const hashShroomlight = 101
const hashBasalt = 102
const hashIronBlock = 103

func (s EndBrickStairs) Hash() uint64 {
	return hashEndBrickStairs | uint64(boolByte(s.UpsideDown))<<7 | uint64(s.Facing)<<8
}

func (g StainedGlass) Hash() uint64 {
	return hashStainedGlass | uint64(g.Colour.Uint8())<<7
}

func (CoalBlock) Hash() uint64 {
	return hashCoalBlock
}

func (Glass) Hash() uint64 {
	return hashGlass
}

func (l Lava) Hash() uint64 {
	return hashLava | uint64(boolByte(l.Still))<<7 | uint64(l.Depth)<<8 | uint64(boolByte(l.Falling))<<16
}

func (w Wool) Hash() uint64 {
	return hashWool | uint64(w.Colour.Uint8())<<7
}

func (s Sandstone) Hash() uint64 {
	return hashSandstone | uint64(boolByte(s.Red))<<7
}

func (DiamondOre) Hash() uint64 {
	return hashDiamondOre
}

func (m MelonSeeds) Hash() uint64 {
	return hashMelonSeeds | uint64(m.Growth)<<7 | uint64(m.Direction)<<15
}

func (Netherrack) Hash() uint64 {
	return hashNetherrack
}

func (t WoodTrapdoor) Hash() uint64 {
	return hashWoodTrapdoor | uint64(t.Wood.Uint8())<<7 | uint64(t.Facing)<<11 | uint64(boolByte(t.Open))<<13 | uint64(boolByte(t.Top))<<14
}

func (c Concrete) Hash() uint64 {
	return hashConcrete | uint64(c.Colour.Uint8())<<7
}

func (s ChiseledSandstone) Hash() uint64 {
	return hashChiseledSandstone | uint64(boolByte(s.Red))<<7
}

func (p Pumpkin) Hash() uint64 {
	return hashPumpkin | uint64(boolByte(p.Carved))<<7 | uint64(p.Facing)<<8
}

func (Air) Hash() uint64 {
	return hashAir
}

func (LapisBlock) Hash() uint64 {
	return hashLapisBlock
}

func (EndBricks) Hash() uint64 {
	return hashEndBricks
}

func (Terracotta) Hash() uint64 {
	return hashTerracotta
}

func (c Cake) Hash() uint64 {
	return hashCake | uint64(c.Bites)<<7
}

func (GlassPane) Hash() uint64 {
	return hashGlassPane
}

func (a Andesite) Hash() uint64 {
	return hashAndesite | uint64(boolByte(a.Polished))<<7
}

func (q QuartzPillar) Hash() uint64 {
	return hashQuartzPillar | uint64(q.Axis)<<7
}

func (t GlazedTerracotta) Hash() uint64 {
	return hashGlazedTerracotta | uint64(t.Colour.Uint8())<<7 | uint64(t.Facing)<<11
}

func (EmeraldOre) Hash() uint64 {
	return hashEmeraldOre
}

func (p StainedGlassPane) Hash() uint64 {
	return hashStainedGlassPane | uint64(p.Colour.Uint8())<<7
}

func (s WoodStairs) Hash() uint64 {
	return hashWoodStairs | uint64(s.Wood.Uint8())<<7 | uint64(boolByte(s.UpsideDown))<<11 | uint64(s.Facing)<<12
}

func (ChiseledQuartz) Hash() uint64 {
	return hashChiseledQuartz
}

func (NetherGoldOre) Hash() uint64 {
	return hashNetherGoldOre
}

func (s WheatSeeds) Hash() uint64 {
	return hashWheatSeeds | uint64(s.Growth)<<7
}

func (k Kelp) Hash() uint64 {
	return hashKelp | uint64(k.Age)<<7
}

func (f Fire) Hash() uint64 {
	return hashFire | uint64(f.Type.Uint8())<<7 | uint64(f.Age)<<11
}

func (QuartzBricks) Hash() uint64 {
	return hashQuartzBricks
}

func (NetherQuartzOre) Hash() uint64 {
	return hashNetherQuartzOre
}

func (s Sand) Hash() uint64 {
	return hashSand | uint64(boolByte(s.Red))<<7
}

func (c Clay) Hash() uint64 {
	return hashClay
}

func (GoldOre) Hash() uint64 {
	return hashGoldOre
}

func (InvisibleBedrock) Hash() uint64 {
	return hashInvisibleBedrock
}

func (Beacon) Hash() uint64 {
	return hashBeacon
}

func (Grass) Hash() uint64 {
	return hashGrass
}

func (c Cobblestone) Hash() uint64 {
	return hashCobblestone | uint64(boolByte(c.Mossy))<<7
}

func (b BeetrootSeeds) Hash() uint64 {
	return hashBeetrootSeeds | uint64(b.Growth)<<7
}

func (l Log) Hash() uint64 {
	return hashLog | uint64(l.Wood.Uint8())<<7 | uint64(boolByte(l.Stripped))<<11 | uint64(l.Axis)<<12
}

func (l Leaves) Hash() uint64 {
	return hashLeaves | uint64(l.Wood.Uint8())<<7 | uint64(boolByte(l.Persistent))<<11 | uint64(boolByte(l.ShouldUpdate))<<12
}

func (s CutSandstone) Hash() uint64 {
	return hashCutSandstone | uint64(boolByte(s.Red))<<7
}

func (IronOre) Hash() uint64 {
	return hashIronOre
}

func (t Torch) Hash() uint64 {
	return hashTorch | uint64(t.Facing)<<7 | uint64(t.Type.Uint8())<<10
}

func (c Coral) Hash() uint64 {
	return hashCoral | uint64(c.Type.Uint8())<<7 | uint64(boolByte(c.Dead))<<11
}

func (b Bedrock) Hash() uint64 {
	return hashBedrock | uint64(boolByte(b.InfiniteBurning))<<7
}

func (c Carpet) Hash() uint64 {
	return hashCarpet | uint64(c.Colour.Uint8())<<7
}

func (DiamondBlock) Hash() uint64 {
	return hashDiamondBlock
}

func (SeaLantern) Hash() uint64 {
	return hashSeaLantern
}

func (w Water) Hash() uint64 {
	return hashWater | uint64(boolByte(w.Still))<<7 | uint64(w.Depth)<<8 | uint64(boolByte(w.Falling))<<16
}

func (s WoodSlab) Hash() uint64 {
	return hashWoodSlab | uint64(s.Wood.Uint8())<<7 | uint64(boolByte(s.Top))<<11 | uint64(boolByte(s.Double))<<12
}

func (Bricks) Hash() uint64 {
	return hashBricks
}

func (l Lantern) Hash() uint64 {
	return hashLantern | uint64(boolByte(l.Hanging))<<7 | uint64(l.Type.Uint8())<<8
}

func (DragonEgg) Hash() uint64 {
	return hashDragonEgg
}

func (t StainedTerracotta) Hash() uint64 {
	return hashStainedTerracotta | uint64(t.Colour.Uint8())<<7
}

func (DirtPath) Hash() uint64 {
	return hashDirtPath
}

func (c CocoaBean) Hash() uint64 {
	return hashCocoaBean | uint64(c.Facing)<<7 | uint64(c.Age)<<9
}

func (c CoralBlock) Hash() uint64 {
	return hashCoralBlock | uint64(c.Type.Uint8())<<7 | uint64(boolByte(c.Dead))<<11
}

func (EmeraldBlock) Hash() uint64 {
	return hashEmeraldBlock
}

func (NetherBrickFence) Hash() uint64 {
	return hashNetherBrickFence
}

func (o Obsidian) Hash() uint64 {
	return hashObsidian | uint64(boolByte(o.Crying))<<7
}

func (g Granite) Hash() uint64 {
	return hashGranite | uint64(boolByte(g.Polished))<<7
}

func (d WoodDoor) Hash() uint64 {
	return hashWoodDoor | uint64(d.Wood.Uint8())<<7 | uint64(d.Facing)<<11 | uint64(boolByte(d.Open))<<13 | uint64(boolByte(d.Top))<<14 | uint64(boolByte(d.Right))<<15
}

func (LapisOre) Hash() uint64 {
	return hashLapisOre
}

func (g GrassPlant) Hash() uint64 {
	return hashGrassPlant | uint64(boolByte(g.UpperPart))<<7 | uint64(g.Type.Uint8())<<8
}

func (n NetherWart) Hash() uint64 {
	return hashNetherWart | uint64(n.Age)<<7
}

func (w WoodFence) Hash() uint64 {
	return hashWoodFence | uint64(w.Wood.Uint8())<<7
}

func (Melon) Hash() uint64 {
	return hashMelon
}

func (c Chest) Hash() uint64 {
	return hashChest | uint64(c.Facing)<<7
}

func (l Light) Hash() uint64 {
	return hashLight | uint64(l.Level)<<7
}

func (Gravel) Hash() uint64 {
	return hashGravel
}

func (AncientDebris) Hash() uint64 {
	return hashAncientDebris
}

func (NetheriteBlock) Hash() uint64 {
	return hashNetheriteBlock
}

func (q Quartz) Hash() uint64 {
	return hashQuartz | uint64(boolByte(q.Smooth))<<7
}

func (IronBars) Hash() uint64 {
	return hashIronBars
}

func (Glowstone) Hash() uint64 {
	return hashGlowstone
}

func (CoalOre) Hash() uint64 {
	return hashCoalOre
}

func (d Dirt) Hash() uint64 {
	return hashDirt | uint64(boolByte(d.Coarse))<<7
}

func (b BoneBlock) Hash() uint64 {
	return hashBoneBlock | uint64(b.Axis)<<7
}

func (f WoodFenceGate) Hash() uint64 {
	return hashWoodFenceGate | uint64(f.Wood.Uint8())<<7 | uint64(f.Facing)<<11 | uint64(boolByte(f.Open))<<13 | uint64(boolByte(f.Lowered))<<14
}

func (BlueIce) Hash() uint64 {
	return hashBlueIce
}

func (c ConcretePowder) Hash() uint64 {
	return hashConcretePowder | uint64(c.Colour.Uint8())<<7
}

func (p PumpkinSeeds) Hash() uint64 {
	return hashPumpkinSeeds | uint64(p.Growth)<<7 | uint64(p.Direction)<<15
}

func (l LitPumpkin) Hash() uint64 {
	return hashLitPumpkin | uint64(l.Facing)<<7
}

func (n NoteBlock) Hash() uint64 {
	return hashNoteBlock
}

func (p Potato) Hash() uint64 {
	return hashPotato | uint64(p.Growth)<<7
}

func (s Sponge) Hash() uint64 {
	return hashSponge | uint64(boolByte(s.Wet))<<7
}

func (s SmoothSandstone) Hash() uint64 {
	return hashSmoothSandstone | uint64(boolByte(s.Red))<<7
}

func (Shroomlight) Hash() uint64 {
	return hashShroomlight
}

func (b Basalt) Hash() uint64 {
	return hashBasalt | uint64(boolByte(b.Polished))<<7 | uint64(b.Axis)<<8
}

func (IronBlock) Hash() uint64 {
	return hashIronBlock
}

func (SoulSand) Hash() uint64 {
	return hashSoulSand
}

func (c Carrot) Hash() uint64 {
	return hashCarrot | uint64(c.Growth)<<7
}

func (d Diorite) Hash() uint64 {
	return hashDiorite | uint64(boolByte(d.Polished))<<7
}

func (p Planks) Hash() uint64 {
	return hashPlanks | uint64(p.Wood.Uint8())<<7
}

func (s Stone) Hash() uint64 {
	return hashStone | uint64(boolByte(s.Smooth))<<7
}

func (GoldBlock) Hash() uint64 {
	return hashGoldBlock
}

func (GildedBlackstone) Hash() uint64 {
	return hashGildedBlackstone
}

func (EndStone) Hash() uint64 {
	return hashEndStone
}

func (f Farmland) Hash() uint64 {
	return hashFarmland | uint64(f.Hydration)<<7
}

func (Barrier) Hash() uint64 {
	return hashBarrier
}

func (SoulSoil) Hash() uint64 {
	return hashSoulSoil
}
