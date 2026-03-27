package gen

import (
	"math"
	"testing"
)

func TestComputeFinalDensityMatchesGraphEval(t *testing.T) {
	noises := NewNoiseRegistry(0)
	samples := []struct {
		x int
		y int
		z int
	}{
		{0, 64, 0},
		{12, 72, 8},
		{16, 40, 16},
		{-28, 8, 20},
		{33, -16, -19},
	}

	for _, sample := range samples {
		chunkX := floorDiv(sample.x, 16)
		chunkZ := floorDiv(sample.z, 16)
		flat := OverworldGraph.NewFlatCacheGrid(chunkX, chunkZ, noises)
		col := OverworldGraph.NewColumnContext(sample.x, sample.z, noises, flat)
		ctx := FunctionContext{BlockX: sample.x, BlockY: sample.y, BlockZ: sample.z}

		got := ComputeFinalDensity(ctx, noises, flat, col)
		want := OverworldGraph.Eval(OverworldRootFinalDensity, ctx, noises, flat, col, nil)

		if math.Abs(got-want) > 1e-12 {
			t.Fatalf("density mismatch at %+v: got %v want %v", sample, got, want)
		}
	}
}

func TestComputePreliminarySurfaceLevelMatchesGraphEval(t *testing.T) {
	noises := NewNoiseRegistry(0)
	samples := []struct {
		x int
		z int
	}{
		{0, 0},
		{15, 15},
		{16, 16},
		{-20, 24},
		{48, -32},
	}

	for _, sample := range samples {
		chunkX := floorDiv(sample.x, 16)
		chunkZ := floorDiv(sample.z, 16)
		flat := OverworldGraph.NewFlatCacheGrid(chunkX, chunkZ, noises)
		col := OverworldGraph.NewColumnContext(sample.x, sample.z, noises, flat)
		ctx := FunctionContext{BlockX: sample.x, BlockY: 0, BlockZ: sample.z}

		got := ComputePreliminarySurfaceLevel(ctx, noises, flat, col)
		want := OverworldGraph.Eval(OverworldRootPreliminarySurfaceLevel, ctx, noises, flat, col, nil)

		if math.Abs(got-want) > 1e-12 {
			t.Fatalf("surface mismatch at %+v: got %v want %v", sample, got, want)
		}
	}
}

func TestComputeFinalDensity4MatchesScalar(t *testing.T) {
	noises := NewNoiseRegistry(0)
	x, z := 32, -16
	chunkX := floorDiv(x, 16)
	chunkZ := floorDiv(z, 16)
	flat := OverworldGraph.NewFlatCacheGrid(chunkX, chunkZ, noises)
	col := OverworldGraph.NewColumnContext(x, z, noises, flat)
	ctx4 := FunctionContext4{
		BlockX: x,
		BlockY: [4]int{-64, -8, 64, 128},
		BlockZ: z,
	}

	got := ComputeFinalDensity4(ctx4, noises, flat, col)
	for lane, y := range ctx4.BlockY {
		want := ComputeFinalDensity(
			FunctionContext{BlockX: x, BlockY: y, BlockZ: z},
			noises,
			flat,
			col,
		)
		if math.Abs(got[lane]-want) > 1e-12 {
			t.Fatalf("lane %d mismatch: got %v want %v", lane, got[lane], want)
		}
	}
}

func floorDiv(v, d int) int {
	if v >= 0 {
		return v / d
	}
	return -((-v + d - 1) / d)
}
