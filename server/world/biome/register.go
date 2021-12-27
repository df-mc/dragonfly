package biome

// init registers all biomes that can be used in a world.
func init() {
	RegisterBiome(Ocean{})
	RegisterBiome(LegacyFrozenOcean{})
	RegisterBiome(DeepOcean{})
	RegisterBiome(FrozenOcean{})
	RegisterBiome(DeepFrozenOcean{})
	RegisterBiome(ColdOcean{})
	RegisterBiome(DeepColdOcean{})
	RegisterBiome(LukewarmOcean{})
	RegisterBiome(DeepLukewarmOcean{})
	RegisterBiome(WarmOcean{})
	RegisterBiome(DeepWarmOcean{})
	RegisterBiome(River{})
	RegisterBiome(FrozenRiver{})
	RegisterBiome(Beach{})
	RegisterBiome(StonyShore{})
	RegisterBiome(SnowyBeach{})
	RegisterBiome(Forest{})
	RegisterBiome(WoodedHills{})
	RegisterBiome(FlowerForest{})
	RegisterBiome(BirchForest{})
	RegisterBiome(BirchForestHills{})
	RegisterBiome(OldGrowthBirchForest{})
	RegisterBiome(TallBirchHills{})
	RegisterBiome(DarkForest{})
	RegisterBiome(DarkForestHills{})
	RegisterBiome(Jungle{})
	RegisterBiome(JungleHills{})
	RegisterBiome(ModifiedJungle{})
	RegisterBiome(JungleEdge{})
	RegisterBiome(ModifiedJungleEdge{})
	RegisterBiome(BambooJungle{})
	RegisterBiome(BambooJungleHills{})
	RegisterBiome(Taiga{})
	RegisterBiome(TaigaHills{})
	RegisterBiome(TaigaMountains{})
	RegisterBiome(SnowyTaiga{})
	RegisterBiome(SnowyTaigaHills{})
	RegisterBiome(SnowyTaigaMountains{})
	RegisterBiome(OldGrowthPineTaiga{})
	RegisterBiome(GiantTreeTaigaHills{})
	RegisterBiome(OldGrowthSpruceTaiga{})
	RegisterBiome(GiantSpruceTaigaHills{})
	RegisterBiome(MushroomFields{})
	RegisterBiome(MushroomFieldShore{})
	RegisterBiome(Swamp{})
	RegisterBiome(SwampHills{})
	RegisterBiome(Savanna{})
	RegisterBiome(SavannaPlateau{})
	RegisterBiome(WindsweptSavanna{})
	RegisterBiome(ShatteredSavannaPlateau{})
	RegisterBiome(Plains{})
	RegisterBiome(SunflowerPlains{})
	RegisterBiome(Desert{})
	RegisterBiome(DesertHills{})
	RegisterBiome(DesertLakes{})
	RegisterBiome(SnowyPlains{})
	RegisterBiome(SnowyMountains{})
	RegisterBiome(IceSpikes{})
	RegisterBiome(GravellyMountainsPlus{})
	RegisterBiome(MountainEdge{})
	RegisterBiome(Badlands{})
	RegisterBiome(BadlandsPlateau{})
	RegisterBiome(ModifiedBadlandsPlateau{})
	RegisterBiome(WoodedBadlandsPlateau{})
	RegisterBiome(ModifiedWoodedBadlandsPlateau{})
	RegisterBiome(ErodedBadlands{})
	RegisterBiome(Meadow{})
	RegisterBiome(Grove{})
	RegisterBiome(SnowySlopes{})
	RegisterBiome(JaggedPeaks{})
	RegisterBiome(FrozenPeaks{})
	RegisterBiome(StonyPeaks{})
	RegisterBiome(LushCaves{})
	RegisterBiome(DripstoneCaves{})
	RegisterBiome(NetherWastes{})
	RegisterBiome(CrimsonForest{})
	RegisterBiome(WarpedForest{})
	RegisterBiome(SoulSandValley{})
	RegisterBiome(BasaltDeltas{})
	RegisterBiome(End{})
}
