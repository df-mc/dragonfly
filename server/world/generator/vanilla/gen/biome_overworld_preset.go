package gen

var (
	overworldFullRange = climateSpan(-1.0, 1.0)

	overworldTemperatures = [5]climateParameter{
		climateSpan(-1.0, -0.45),
		climateSpan(-0.45, -0.15),
		climateSpan(-0.15, 0.2),
		climateSpan(0.2, 0.55),
		climateSpan(0.55, 1.0),
	}
	overworldHumidities = [5]climateParameter{
		climateSpan(-1.0, -0.35),
		climateSpan(-0.35, -0.1),
		climateSpan(-0.1, 0.1),
		climateSpan(0.1, 0.3),
		climateSpan(0.3, 1.0),
	}
	overworldErosions = [7]climateParameter{
		climateSpan(-1.0, -0.78),
		climateSpan(-0.78, -0.375),
		climateSpan(-0.375, -0.2225),
		climateSpan(-0.2225, 0.05),
		climateSpan(0.05, 0.45),
		climateSpan(0.45, 0.55),
		climateSpan(0.55, 1.0),
	}

	overworldFrozenRange   = overworldTemperatures[0]
	overworldUnfrozenRange = climateCombine(overworldTemperatures[1], overworldTemperatures[4])

	overworldMushroomFieldsContinentalness = climateSpan(-1.2, -1.05)
	overworldDeepOceanContinentalness      = climateSpan(-1.05, -0.455)
	overworldOceanContinentalness          = climateSpan(-0.455, -0.19)
	overworldCoastContinentalness          = climateSpan(-0.19, -0.11)
	overworldInlandContinentalness         = climateSpan(-0.11, 0.55)
	overworldNearInlandContinentalness     = climateSpan(-0.11, 0.03)
	overworldMidInlandContinentalness      = climateSpan(0.03, 0.3)
	overworldFarInlandContinentalness      = climateSpan(0.3, 1.0)
)

var overworldPresetPoints = buildOverworldPresetPoints()

func buildOverworldPresetPoints() []climateParameterPoint {
	points := make([]climateParameterPoint, 0, 5000)
	addOverworldOffCoastBiomes(&points)
	addOverworldInlandBiomes(&points)
	addOverworldUndergroundBiomes(&points)
	return points
}

func addOverworldOffCoastBiomes(points *[]climateParameterPoint) {
	addOverworldSurfaceBiome(points, overworldFullRange, overworldFullRange, overworldMushroomFieldsContinentalness, overworldFullRange, overworldFullRange, 0.0, BiomeMushroomFields)

	for i, temperature := range overworldTemperatures {
		addOverworldSurfaceBiome(points, temperature, overworldFullRange, overworldDeepOceanContinentalness, overworldFullRange, overworldFullRange, 0.0, oceans[0][i])
		addOverworldSurfaceBiome(points, temperature, overworldFullRange, overworldOceanContinentalness, overworldFullRange, overworldFullRange, 0.0, oceans[1][i])
	}
}

func addOverworldInlandBiomes(points *[]climateParameterPoint) {
	addOverworldMidSlice(points, climateSpan(-1.0, -0.93333334))
	addOverworldHighSlice(points, climateSpan(-0.93333334, -0.7666667))
	addOverworldPeaks(points, climateSpan(-0.7666667, -0.56666666))
	addOverworldHighSlice(points, climateSpan(-0.56666666, -0.4))
	addOverworldMidSlice(points, climateSpan(-0.4, -0.26666668))
	addOverworldLowSlice(points, climateSpan(-0.26666668, -0.05))
	addOverworldValleys(points, climateSpan(-0.05, 0.05))
	addOverworldLowSlice(points, climateSpan(0.05, 0.26666668))
	addOverworldMidSlice(points, climateSpan(0.26666668, 0.4))
	addOverworldHighSlice(points, climateSpan(0.4, 0.56666666))
	addOverworldPeaks(points, climateSpan(0.56666666, 0.7666667))
	addOverworldHighSlice(points, climateSpan(0.7666667, 0.93333334))
	addOverworldMidSlice(points, climateSpan(0.93333334, 1.0))
}

func addOverworldPeaks(points *[]climateParameterPoint, weirdness climateParameter) {
	for ti, temperature := range overworldTemperatures {
		for hi, humidity := range overworldHumidities {
			middleBiome := overworldPickMiddleBiome(ti, hi, weirdness)
			middleBiomeOrBadlandsIfHot := overworldPickMiddleBiomeOrBadlandsIfHot(ti, hi, weirdness)
			middleBiomeOrBadlandsIfHotOrSlopeIfCold := overworldPickMiddleBiomeOrBadlandsIfHotOrSlopeIfCold(ti, hi, weirdness)
			plateauBiome := overworldPickPlateauBiome(ti, hi, weirdness)
			shatteredBiome := overworldPickShatteredBiome(ti, hi, weirdness)
			shatteredBiomeOrWindsweptSavanna := overworldMaybePickWindsweptSavannaBiome(ti, hi, weirdness, shatteredBiome)
			peakBiome := overworldPickPeakBiome(ti, hi, weirdness)

			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldFarInlandContinentalness), overworldErosions[0], weirdness, 0.0, peakBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldNearInlandContinentalness), overworldErosions[1], weirdness, 0.0, middleBiomeOrBadlandsIfHotOrSlopeIfCold)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[1], weirdness, 0.0, peakBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldNearInlandContinentalness), climateCombine(overworldErosions[2], overworldErosions[3]), weirdness, 0.0, middleBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[2], weirdness, 0.0, plateauBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldMidInlandContinentalness, overworldErosions[3], weirdness, 0.0, middleBiomeOrBadlandsIfHot)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldFarInlandContinentalness, overworldErosions[3], weirdness, 0.0, plateauBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldFarInlandContinentalness), overworldErosions[4], weirdness, 0.0, middleBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldNearInlandContinentalness), overworldErosions[5], weirdness, 0.0, shatteredBiomeOrWindsweptSavanna)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[5], weirdness, 0.0, shatteredBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, middleBiome)
		}
	}
}

func addOverworldHighSlice(points *[]climateParameterPoint, weirdness climateParameter) {
	for ti, temperature := range overworldTemperatures {
		for hi, humidity := range overworldHumidities {
			middleBiome := overworldPickMiddleBiome(ti, hi, weirdness)
			middleBiomeOrBadlandsIfHot := overworldPickMiddleBiomeOrBadlandsIfHot(ti, hi, weirdness)
			middleBiomeOrBadlandsIfHotOrSlopeIfCold := overworldPickMiddleBiomeOrBadlandsIfHotOrSlopeIfCold(ti, hi, weirdness)
			plateauBiome := overworldPickPlateauBiome(ti, hi, weirdness)
			shatteredBiome := overworldPickShatteredBiome(ti, hi, weirdness)
			middleBiomeOrWindsweptSavanna := overworldMaybePickWindsweptSavannaBiome(ti, hi, weirdness, middleBiome)
			slopeBiome := overworldPickSlopeBiome(ti, hi, weirdness)
			peakBiome := overworldPickPeakBiome(ti, hi, weirdness)

			addOverworldSurfaceBiome(points, temperature, humidity, overworldCoastContinentalness, climateCombine(overworldErosions[0], overworldErosions[1]), weirdness, 0.0, middleBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldNearInlandContinentalness, overworldErosions[0], weirdness, 0.0, slopeBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[0], weirdness, 0.0, peakBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldNearInlandContinentalness, overworldErosions[1], weirdness, 0.0, middleBiomeOrBadlandsIfHotOrSlopeIfCold)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[1], weirdness, 0.0, slopeBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldNearInlandContinentalness), climateCombine(overworldErosions[2], overworldErosions[3]), weirdness, 0.0, middleBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[2], weirdness, 0.0, plateauBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldMidInlandContinentalness, overworldErosions[3], weirdness, 0.0, middleBiomeOrBadlandsIfHot)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldFarInlandContinentalness, overworldErosions[3], weirdness, 0.0, plateauBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldFarInlandContinentalness), overworldErosions[4], weirdness, 0.0, middleBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldNearInlandContinentalness), overworldErosions[5], weirdness, 0.0, middleBiomeOrWindsweptSavanna)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[5], weirdness, 0.0, shatteredBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, middleBiome)
		}
	}
}

func addOverworldMidSlice(points *[]climateParameterPoint, weirdness climateParameter) {
	addOverworldSurfaceBiome(points, overworldFullRange, overworldFullRange, overworldCoastContinentalness, climateCombine(overworldErosions[0], overworldErosions[2]), weirdness, 0.0, BiomeStonyShore)
	addOverworldSurfaceBiome(points, climateCombine(overworldTemperatures[1], overworldTemperatures[2]), overworldFullRange, climateCombine(overworldNearInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, BiomeSwamp)
	addOverworldSurfaceBiome(points, climateCombine(overworldTemperatures[3], overworldTemperatures[4]), overworldFullRange, climateCombine(overworldNearInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, BiomeMangroveSwamp)

	for ti, temperature := range overworldTemperatures {
		for hi, humidity := range overworldHumidities {
			middleBiome := overworldPickMiddleBiome(ti, hi, weirdness)
			middleBiomeOrBadlandsIfHot := overworldPickMiddleBiomeOrBadlandsIfHot(ti, hi, weirdness)
			middleBiomeOrBadlandsIfHotOrSlopeIfCold := overworldPickMiddleBiomeOrBadlandsIfHotOrSlopeIfCold(ti, hi, weirdness)
			shatteredBiome := overworldPickShatteredBiome(ti, hi, weirdness)
			plateauBiome := overworldPickPlateauBiome(ti, hi, weirdness)
			beachBiome := overworldPickBeachBiome(ti)
			middleBiomeOrWindsweptSavanna := overworldMaybePickWindsweptSavannaBiome(ti, hi, weirdness, middleBiome)
			shatteredCoastBiome := overworldPickShatteredCoastBiome(ti, hi, weirdness)
			slopeBiome := overworldPickSlopeBiome(ti, hi, weirdness)

			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldNearInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[0], weirdness, 0.0, slopeBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldNearInlandContinentalness, overworldMidInlandContinentalness), overworldErosions[1], weirdness, 0.0, middleBiomeOrBadlandsIfHotOrSlopeIfCold)
			if ti == 0 {
				addOverworldSurfaceBiome(points, temperature, humidity, overworldFarInlandContinentalness, overworldErosions[1], weirdness, 0.0, slopeBiome)
			} else {
				addOverworldSurfaceBiome(points, temperature, humidity, overworldFarInlandContinentalness, overworldErosions[1], weirdness, 0.0, plateauBiome)
			}
			addOverworldSurfaceBiome(points, temperature, humidity, overworldNearInlandContinentalness, overworldErosions[2], weirdness, 0.0, middleBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldMidInlandContinentalness, overworldErosions[2], weirdness, 0.0, middleBiomeOrBadlandsIfHot)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldFarInlandContinentalness, overworldErosions[2], weirdness, 0.0, plateauBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldNearInlandContinentalness), overworldErosions[3], weirdness, 0.0, middleBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[3], weirdness, 0.0, middleBiomeOrBadlandsIfHot)
			if weirdness.max < 0 {
				addOverworldSurfaceBiome(points, temperature, humidity, overworldCoastContinentalness, overworldErosions[4], weirdness, 0.0, beachBiome)
				addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldNearInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[4], weirdness, 0.0, middleBiome)
			} else {
				addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldCoastContinentalness, overworldFarInlandContinentalness), overworldErosions[4], weirdness, 0.0, middleBiome)
			}
			addOverworldSurfaceBiome(points, temperature, humidity, overworldCoastContinentalness, overworldErosions[5], weirdness, 0.0, shatteredCoastBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldNearInlandContinentalness, overworldErosions[5], weirdness, 0.0, middleBiomeOrWindsweptSavanna)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[5], weirdness, 0.0, shatteredBiome)
			if weirdness.max < 0 {
				addOverworldSurfaceBiome(points, temperature, humidity, overworldCoastContinentalness, overworldErosions[6], weirdness, 0.0, beachBiome)
			} else {
				addOverworldSurfaceBiome(points, temperature, humidity, overworldCoastContinentalness, overworldErosions[6], weirdness, 0.0, middleBiome)
			}
			if ti == 0 {
				addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldNearInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, middleBiome)
			}
		}
	}
}

func addOverworldLowSlice(points *[]climateParameterPoint, weirdness climateParameter) {
	addOverworldSurfaceBiome(points, overworldFullRange, overworldFullRange, overworldCoastContinentalness, climateCombine(overworldErosions[0], overworldErosions[2]), weirdness, 0.0, BiomeStonyShore)
	addOverworldSurfaceBiome(points, climateCombine(overworldTemperatures[1], overworldTemperatures[2]), overworldFullRange, climateCombine(overworldNearInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, BiomeSwamp)
	addOverworldSurfaceBiome(points, climateCombine(overworldTemperatures[3], overworldTemperatures[4]), overworldFullRange, climateCombine(overworldNearInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, BiomeMangroveSwamp)

	for ti, temperature := range overworldTemperatures {
		for hi, humidity := range overworldHumidities {
			middleBiome := overworldPickMiddleBiome(ti, hi, weirdness)
			middleBiomeOrBadlandsIfHot := overworldPickMiddleBiomeOrBadlandsIfHot(ti, hi, weirdness)
			middleBiomeOrBadlandsIfHotOrSlopeIfCold := overworldPickMiddleBiomeOrBadlandsIfHotOrSlopeIfCold(ti, hi, weirdness)
			beachBiome := overworldPickBeachBiome(ti)
			middleBiomeOrWindsweptSavanna := overworldMaybePickWindsweptSavannaBiome(ti, hi, weirdness, middleBiome)
			shatteredCoastBiome := overworldPickShatteredCoastBiome(ti, hi, weirdness)

			addOverworldSurfaceBiome(points, temperature, humidity, overworldNearInlandContinentalness, climateCombine(overworldErosions[0], overworldErosions[1]), weirdness, 0.0, middleBiomeOrBadlandsIfHot)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), climateCombine(overworldErosions[0], overworldErosions[1]), weirdness, 0.0, middleBiomeOrBadlandsIfHotOrSlopeIfCold)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldNearInlandContinentalness, climateCombine(overworldErosions[2], overworldErosions[3]), weirdness, 0.0, middleBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), climateCombine(overworldErosions[2], overworldErosions[3]), weirdness, 0.0, middleBiomeOrBadlandsIfHot)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldCoastContinentalness, climateCombine(overworldErosions[3], overworldErosions[4]), weirdness, 0.0, beachBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldNearInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[4], weirdness, 0.0, middleBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldCoastContinentalness, overworldErosions[5], weirdness, 0.0, shatteredCoastBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldNearInlandContinentalness, overworldErosions[5], weirdness, 0.0, middleBiomeOrWindsweptSavanna)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[5], weirdness, 0.0, middleBiome)
			addOverworldSurfaceBiome(points, temperature, humidity, overworldCoastContinentalness, overworldErosions[6], weirdness, 0.0, beachBiome)
			if ti == 0 {
				addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldNearInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, middleBiome)
			}
		}
	}
}

func addOverworldValleys(points *[]climateParameterPoint, weirdness climateParameter) {
	frozenValleyCoast := BiomeStonyShore
	unfrozenValleyCoast := BiomeStonyShore
	if weirdness.max >= 0 {
		frozenValleyCoast = BiomeFrozenRiver
		unfrozenValleyCoast = BiomeRiver
	}
	addOverworldSurfaceBiome(points, overworldFrozenRange, overworldFullRange, overworldCoastContinentalness, climateCombine(overworldErosions[0], overworldErosions[1]), weirdness, 0.0, frozenValleyCoast)
	addOverworldSurfaceBiome(points, overworldUnfrozenRange, overworldFullRange, overworldCoastContinentalness, climateCombine(overworldErosions[0], overworldErosions[1]), weirdness, 0.0, unfrozenValleyCoast)
	addOverworldSurfaceBiome(points, overworldFrozenRange, overworldFullRange, overworldNearInlandContinentalness, climateCombine(overworldErosions[0], overworldErosions[1]), weirdness, 0.0, BiomeFrozenRiver)
	addOverworldSurfaceBiome(points, overworldUnfrozenRange, overworldFullRange, overworldNearInlandContinentalness, climateCombine(overworldErosions[0], overworldErosions[1]), weirdness, 0.0, BiomeRiver)
	addOverworldSurfaceBiome(points, overworldFrozenRange, overworldFullRange, climateCombine(overworldCoastContinentalness, overworldFarInlandContinentalness), climateCombine(overworldErosions[2], overworldErosions[5]), weirdness, 0.0, BiomeFrozenRiver)
	addOverworldSurfaceBiome(points, overworldUnfrozenRange, overworldFullRange, climateCombine(overworldCoastContinentalness, overworldFarInlandContinentalness), climateCombine(overworldErosions[2], overworldErosions[5]), weirdness, 0.0, BiomeRiver)
	addOverworldSurfaceBiome(points, overworldFrozenRange, overworldFullRange, overworldCoastContinentalness, overworldErosions[6], weirdness, 0.0, BiomeFrozenRiver)
	addOverworldSurfaceBiome(points, overworldUnfrozenRange, overworldFullRange, overworldCoastContinentalness, overworldErosions[6], weirdness, 0.0, BiomeRiver)
	addOverworldSurfaceBiome(points, climateCombine(overworldTemperatures[1], overworldTemperatures[2]), overworldFullRange, climateCombine(overworldInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, BiomeSwamp)
	addOverworldSurfaceBiome(points, climateCombine(overworldTemperatures[3], overworldTemperatures[4]), overworldFullRange, climateCombine(overworldInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, BiomeMangroveSwamp)
	addOverworldSurfaceBiome(points, overworldFrozenRange, overworldFullRange, climateCombine(overworldInlandContinentalness, overworldFarInlandContinentalness), overworldErosions[6], weirdness, 0.0, BiomeFrozenRiver)

	for ti, temperature := range overworldTemperatures {
		for hi, humidity := range overworldHumidities {
			middleBiomeOrBadlandsIfHot := overworldPickMiddleBiomeOrBadlandsIfHot(ti, hi, weirdness)
			addOverworldSurfaceBiome(points, temperature, humidity, climateCombine(overworldMidInlandContinentalness, overworldFarInlandContinentalness), climateCombine(overworldErosions[0], overworldErosions[1]), weirdness, 0.0, middleBiomeOrBadlandsIfHot)
		}
	}
}

func addOverworldUndergroundBiomes(points *[]climateParameterPoint) {
	addOverworldUndergroundBiome(points, overworldFullRange, overworldFullRange, climateSpan(0.8, 1.0), overworldFullRange, overworldFullRange, 0.0, BiomeDripstoneCaves)
	addOverworldUndergroundBiome(points, overworldFullRange, climateSpan(0.7, 1.0), overworldFullRange, overworldFullRange, overworldFullRange, 0.0, BiomeLushCaves)
	addOverworldBottomBiome(points, overworldFullRange, overworldFullRange, overworldFullRange, climateCombine(overworldErosions[0], overworldErosions[1]), overworldFullRange, 0.0, BiomeDeepDark)
}

func addOverworldSurfaceBiome(points *[]climateParameterPoint, temperature, humidity, continentalness, erosion, weirdness climateParameter, offset float64, biome Biome) {
	*points = append(*points, climateParameterPoint{
		params: [6]climateParameter{
			temperature,
			humidity,
			continentalness,
			erosion,
			climatePoint(0.0),
			weirdness,
		},
		offset: int64(offset * 10000.0),
		biome:  biome,
	})
	*points = append(*points, climateParameterPoint{
		params: [6]climateParameter{
			temperature,
			humidity,
			continentalness,
			erosion,
			climatePoint(1.0),
			weirdness,
		},
		offset: int64(offset * 10000.0),
		biome:  biome,
	})
}

func addOverworldUndergroundBiome(points *[]climateParameterPoint, temperature, humidity, continentalness, erosion, weirdness climateParameter, offset float64, biome Biome) {
	*points = append(*points, climateParameterPoint{
		params: [6]climateParameter{
			temperature,
			humidity,
			continentalness,
			erosion,
			climateSpan(0.2, 0.9),
			weirdness,
		},
		offset: int64(offset * 10000.0),
		biome:  biome,
	})
}

func addOverworldBottomBiome(points *[]climateParameterPoint, temperature, humidity, continentalness, erosion, weirdness climateParameter, offset float64, biome Biome) {
	*points = append(*points, climateParameterPoint{
		params: [6]climateParameter{
			temperature,
			humidity,
			continentalness,
			erosion,
			climatePoint(1.1),
			weirdness,
		},
		offset: int64(offset * 10000.0),
		biome:  biome,
	})
}

func overworldPickMiddleBiome(tempIndex, humidityIndex int, weirdness climateParameter) Biome {
	if weirdness.max < 0 {
		return middleBiomes[tempIndex][humidityIndex]
	}
	if biome := middleBiomeVariants[tempIndex][humidityIndex]; biome != BiomeTheVoid {
		return biome
	}
	return middleBiomes[tempIndex][humidityIndex]
}

func overworldPickMiddleBiomeOrBadlandsIfHot(tempIndex, humidityIndex int, weirdness climateParameter) Biome {
	if tempIndex == 4 {
		return overworldPickBadlandsBiome(humidityIndex, weirdness)
	}
	return overworldPickMiddleBiome(tempIndex, humidityIndex, weirdness)
}

func overworldPickMiddleBiomeOrBadlandsIfHotOrSlopeIfCold(tempIndex, humidityIndex int, weirdness climateParameter) Biome {
	if tempIndex == 0 {
		return overworldPickSlopeBiome(tempIndex, humidityIndex, weirdness)
	}
	return overworldPickMiddleBiomeOrBadlandsIfHot(tempIndex, humidityIndex, weirdness)
}

func overworldMaybePickWindsweptSavannaBiome(tempIndex, humidityIndex int, weirdness climateParameter, underlyingBiome Biome) Biome {
	if tempIndex > 1 && humidityIndex < 4 && weirdness.max >= 0 {
		return BiomeWindsweptSavanna
	}
	return underlyingBiome
}

func overworldPickShatteredCoastBiome(tempIndex, humidityIndex int, weirdness climateParameter) Biome {
	beachOrMiddleBiome := overworldPickBeachBiome(tempIndex)
	if weirdness.max >= 0 {
		beachOrMiddleBiome = overworldPickMiddleBiome(tempIndex, humidityIndex, weirdness)
	}
	return overworldMaybePickWindsweptSavannaBiome(tempIndex, humidityIndex, weirdness, beachOrMiddleBiome)
}

func overworldPickBeachBiome(tempIndex int) Biome {
	switch tempIndex {
	case 0:
		return BiomeSnowyBeach
	case 4:
		return BiomeDesert
	default:
		return BiomeBeach
	}
}

func overworldPickBadlandsBiome(humidityIndex int, weirdness climateParameter) Biome {
	if humidityIndex < 2 {
		if weirdness.max < 0 {
			return BiomeBadlands
		}
		return BiomeErodedBadlands
	}
	if humidityIndex < 3 {
		return BiomeBadlands
	}
	return BiomeWoodedBadlands
}

func overworldPickPlateauBiome(tempIndex, humidityIndex int, weirdness climateParameter) Biome {
	if weirdness.max >= 0 {
		if biome := plateauBiomeVariants[tempIndex][humidityIndex]; biome != BiomeTheVoid {
			return biome
		}
	}
	return plateauBiomes[tempIndex][humidityIndex]
}

func overworldPickPeakBiome(tempIndex, humidityIndex int, weirdness climateParameter) Biome {
	if tempIndex <= 2 {
		if weirdness.max < 0 {
			return BiomeJaggedPeaks
		}
		return BiomeFrozenPeaks
	}
	if tempIndex == 3 {
		return BiomeStonyPeaks
	}
	return overworldPickBadlandsBiome(humidityIndex, weirdness)
}

func overworldPickSlopeBiome(tempIndex, humidityIndex int, weirdness climateParameter) Biome {
	if tempIndex >= 3 {
		return overworldPickPlateauBiome(tempIndex, humidityIndex, weirdness)
	}
	if humidityIndex <= 1 {
		return BiomeSnowySlopes
	}
	return BiomeGrove
}

func overworldPickShatteredBiome(tempIndex, humidityIndex int, weirdness climateParameter) Biome {
	if biome := shatteredBiomes[tempIndex][humidityIndex]; biome != BiomeTheVoid {
		return biome
	}
	return overworldPickMiddleBiome(tempIndex, humidityIndex, weirdness)
}

func climateCombine(a, b climateParameter) climateParameter {
	return climateParameter{min: min(a.min, b.min), max: max(a.max, b.max)}
}
