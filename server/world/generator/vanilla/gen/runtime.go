package gen

import "math"

type FunctionContext struct {
	BlockX int
	BlockY int
	BlockZ int
}

type FunctionContext4 struct {
	BlockX int
	BlockY [4]int
	BlockZ int
}

type NoiseSource interface {
	Sample(noise NoiseRef, x, y, z float64) float64
	SampleBlendedNoise(x, y, z, xzScale, yScale, xzFactor, yFactor, smearScaleMultiplier float64) float64
	SampleEndIslands(blockX, blockZ int) float64
}

type NoiseParamsData struct {
	FirstOctave int
	Amplitudes  []float64
}

type OpCode uint8

const (
	OpConstant OpCode = iota
	OpAdd
	OpMul
	OpMin
	OpMax
	OpClamp
	OpAbs
	OpSquare
	OpCube
	OpSqueeze
	OpHalfNegative
	OpQuarterNegative
	OpNoise
	OpShiftedNoise
	OpYClampedGradient
	OpFlatCache
	OpCache2D
	OpCacheOnce
	OpInterpolated
	OpBlendAlpha
	OpBlendOffset
	OpBlendDensity
	OpRangeChoice
	OpSpline
	OpWeirdScaledSampler
	OpShiftA
	OpShiftB
	OpShift
	OpOldBlendedNoise
	OpEndIslands
	OpInvert
	OpFindTopSurface
)

type RarityType uint8

const (
	RarityType1 RarityType = iota + 1
	RarityType2
)

type ArgRef struct {
	Node  int
	Const float64
}

type SplineValue struct {
	Const  float64
	Nested *Spline
}

type SplinePoint struct {
	Location   float64
	Derivative float64
	Value      SplineValue
}

type Spline struct {
	Coordinate ArgRef
	Points     []SplinePoint
}

type Node struct {
	Op                   OpCode
	Dep0                 int
	Dep1                 int
	Dep2                 int
	Const                float64
	Min                  float64
	Max                  float64
	FromY                int
	ToY                  int
	FromValue            float64
	ToValue              float64
	LowerBound           int
	CellHeight           int
	Noise                NoiseRef
	XZScale              float64
	YScale               float64
	XZFactor             float64
	YFactor              float64
	SmearScaleMultiplier float64
	Rarity               RarityType
	Spline               *Spline
}

type Graph struct {
	Nodes          []Node
	FlatCacheSlots []int
	Cache2DSlots   []int
	InnerToCache2D []int
	FlatCacheOrder []int
	Cache2DOrder   []int
}

type FlatCacheGrid struct {
	firstQuartX int
	firstQuartZ int
	values      [][5][5]float64
}

type ColumnContext struct {
	values []float64
}

type EvalScratch struct {
	seen   []uint32
	values []float64
	mark   uint32
}

func NewEvalScratch(graph *Graph) *EvalScratch {
	return &EvalScratch{
		seen:   make([]uint32, len(graph.Nodes)),
		values: make([]float64, len(graph.Nodes)),
		mark:   1,
	}
}

func (s *EvalScratch) reset() {
	s.mark++
	if s.mark != 0 {
		return
	}
	clear(s.seen)
	s.mark = 1
}

func (s *EvalScratch) get(idx int) (float64, bool) {
	if idx < 0 || idx >= len(s.values) {
		return 0, false
	}
	return s.values[idx], s.seen[idx] == s.mark
}

func (s *EvalScratch) set(idx int, value float64) {
	if idx < 0 || idx >= len(s.values) {
		return
	}
	s.values[idx] = value
	s.seen[idx] = s.mark
}

func (g *Graph) NewFlatCacheGrid(chunkX, chunkZ int, noises NoiseSource) *FlatCacheGrid {
	grid := &FlatCacheGrid{
		firstQuartX: chunkX * 4,
		firstQuartZ: chunkZ * 4,
		values:      make([][5][5]float64, len(g.FlatCacheOrder)),
	}
	scratch := NewEvalScratch(g)
	flatValues := make([]float64, len(g.FlatCacheOrder))

	for qz := 0; qz < 5; qz++ {
		for qx := 0; qx < 5; qx++ {
			blockX := (grid.firstQuartX + qx) * 4
			blockZ := (grid.firstQuartZ + qz) * 4
			ctx := FunctionContext{BlockX: blockX, BlockY: 0, BlockZ: blockZ}

			for slot, nodeIdx := range g.FlatCacheOrder {
				scratch.reset()
				value := g.evalFlat(g.Nodes[nodeIdx].Dep0, ctx, noises, flatValues, scratch)
				flatValues[slot] = value
				grid.values[slot][qz][qx] = value
			}
		}
	}
	return grid
}

func (g *FlatCacheGrid) Lookup(slot, blockX, blockZ int) float64 {
	if g == nil || slot < 0 || slot >= len(g.values) {
		return 0
	}
	qx := clampInt((blockX>>2)-g.firstQuartX, 0, 4)
	qz := clampInt((blockZ>>2)-g.firstQuartZ, 0, 4)
	return g.values[slot][qz][qx]
}

func (g *Graph) NewColumnContext(blockX, blockZ int, noises NoiseSource, flat *FlatCacheGrid) *ColumnContext {
	values := make([]float64, len(g.Cache2DOrder))
	computed := make([]bool, len(g.Cache2DOrder))
	scratch := NewEvalScratch(g)
	ctx := FunctionContext{BlockX: blockX, BlockY: 0, BlockZ: blockZ}

	for slot, nodeIdx := range g.Cache2DOrder {
		scratch.reset()
		values[slot] = g.evalColumn(g.Nodes[nodeIdx].Dep0, ctx, noises, flat, values, computed, scratch)
		computed[slot] = true
	}

	return &ColumnContext{values: values}
}

func (g *Graph) Eval(root int, ctx FunctionContext, noises NoiseSource, flat *FlatCacheGrid, col *ColumnContext, scratch *EvalScratch) float64 {
	if scratch == nil {
		scratch = NewEvalScratch(g)
	}
	scratch.reset()
	return g.evalNormal(root, ctx, noises, flat, col, scratch)
}

func (g *Graph) evalNormal(idx int, ctx FunctionContext, noises NoiseSource, flat *FlatCacheGrid, col *ColumnContext, scratch *EvalScratch) float64 {
	if idx < 0 || idx >= len(g.Nodes) {
		return 0
	}
	if col != nil {
		if wrapper := g.InnerToCache2D[idx]; wrapper >= 0 {
			if slot := g.Cache2DSlots[wrapper]; slot >= 0 && slot < len(col.values) {
				return col.values[slot]
			}
		}
		if slot := g.Cache2DSlots[idx]; slot >= 0 && slot < len(col.values) {
			return col.values[slot]
		}
	}
	if flat != nil {
		if slot := g.FlatCacheSlots[idx]; slot >= 0 {
			return flat.Lookup(slot, ctx.BlockX, ctx.BlockZ)
		}
	}
	if value, ok := scratch.get(idx); ok {
		return value
	}

	node := g.Nodes[idx]
	var value float64
	if node.Op == OpFindTopSurface {
		upper := int(math.Floor(g.evalNormal(node.Dep1, ctx, noises, flat, col, scratch)))
		innerScratch := NewEvalScratch(g)
		value = FindTopSurface(ctx.BlockX, ctx.BlockZ, node.LowerBound, upper, node.CellHeight, func(y int) float64 {
			innerScratch.reset()
			innerCtx := ctx
			innerCtx.BlockY = y
			return g.evalNormal(node.Dep0, innerCtx, noises, flat, col, innerScratch)
		})
	} else {
		value = g.evalCommon(node, ctx, noises, func(dep int) float64 {
			return g.evalNormal(dep, ctx, noises, flat, col, scratch)
		})
	}

	scratch.set(idx, value)
	return value
}

func (g *Graph) evalFlat(idx int, ctx FunctionContext, noises NoiseSource, flatValues []float64, scratch *EvalScratch) float64 {
	if idx < 0 || idx >= len(g.Nodes) {
		return 0
	}
	if slot := g.FlatCacheSlots[idx]; slot >= 0 && slot < len(flatValues) {
		return flatValues[slot]
	}
	if value, ok := scratch.get(idx); ok {
		return value
	}

	node := g.Nodes[idx]
	var value float64
	switch node.Op {
	case OpCache2D, OpCacheOnce, OpInterpolated, OpBlendDensity:
		value = g.evalFlat(node.Dep0, ctx, noises, flatValues, scratch)
	case OpFindTopSurface:
		upper := int(math.Floor(g.evalFlat(node.Dep1, ctx, noises, flatValues, scratch)))
		innerScratch := NewEvalScratch(g)
		value = FindTopSurface(ctx.BlockX, ctx.BlockZ, node.LowerBound, upper, node.CellHeight, func(y int) float64 {
			innerScratch.reset()
			innerCtx := ctx
			innerCtx.BlockY = y
			return g.evalFlat(node.Dep0, innerCtx, noises, flatValues, innerScratch)
		})
	default:
		value = g.evalCommon(node, ctx, noises, func(dep int) float64 {
			return g.evalFlat(dep, ctx, noises, flatValues, scratch)
		})
	}

	scratch.set(idx, value)
	return value
}

func (g *Graph) evalColumn(idx int, ctx FunctionContext, noises NoiseSource, flat *FlatCacheGrid, values []float64, computed []bool, scratch *EvalScratch) float64 {
	if idx < 0 || idx >= len(g.Nodes) {
		return 0
	}
	if wrapper := g.InnerToCache2D[idx]; wrapper >= 0 {
		if slot := g.Cache2DSlots[wrapper]; slot >= 0 && slot < len(values) && computed[slot] {
			return values[slot]
		}
	}
	if slot := g.Cache2DSlots[idx]; slot >= 0 && slot < len(values) && computed[slot] {
		return values[slot]
	}
	if flat != nil {
		if slot := g.FlatCacheSlots[idx]; slot >= 0 {
			return flat.Lookup(slot, ctx.BlockX, ctx.BlockZ)
		}
	}
	if value, ok := scratch.get(idx); ok {
		return value
	}

	node := g.Nodes[idx]
	var value float64
	switch node.Op {
	case OpCache2D, OpCacheOnce, OpInterpolated, OpBlendDensity:
		value = g.evalColumn(node.Dep0, ctx, noises, flat, values, computed, scratch)
	case OpFindTopSurface:
		upper := int(math.Floor(g.evalColumn(node.Dep1, ctx, noises, flat, values, computed, scratch)))
		innerScratch := NewEvalScratch(g)
		value = FindTopSurface(ctx.BlockX, ctx.BlockZ, node.LowerBound, upper, node.CellHeight, func(y int) float64 {
			innerScratch.reset()
			innerCtx := ctx
			innerCtx.BlockY = y
			return g.evalColumn(node.Dep0, innerCtx, noises, flat, values, computed, innerScratch)
		})
	default:
		value = g.evalCommon(node, ctx, noises, func(dep int) float64 {
			return g.evalColumn(dep, ctx, noises, flat, values, computed, scratch)
		})
	}

	scratch.set(idx, value)
	return value
}

func (g *Graph) evalCommon(node Node, ctx FunctionContext, noises NoiseSource, dep func(int) float64) float64 {
	switch node.Op {
	case OpConstant:
		return node.Const
	case OpAdd:
		return dep(node.Dep0) + dep(node.Dep1)
	case OpMul:
		return dep(node.Dep0) * dep(node.Dep1)
	case OpMin:
		return minFloat(dep(node.Dep0), dep(node.Dep1))
	case OpMax:
		return maxFloat(dep(node.Dep0), dep(node.Dep1))
	case OpClamp:
		return clampFloat(dep(node.Dep0), node.Min, node.Max)
	case OpAbs:
		return math.Abs(dep(node.Dep0))
	case OpSquare:
		v := dep(node.Dep0)
		return v * v
	case OpCube:
		v := dep(node.Dep0)
		return v * v * v
	case OpSqueeze:
		return squeeze(dep(node.Dep0))
	case OpHalfNegative:
		return halfNegative(dep(node.Dep0))
	case OpQuarterNegative:
		return quarterNegative(dep(node.Dep0))
	case OpNoise:
		return noises.Sample(node.Noise, float64(ctx.BlockX)*node.XZScale, float64(ctx.BlockY)*node.YScale, float64(ctx.BlockZ)*node.XZScale)
	case OpShiftedNoise:
		return noises.Sample(
			node.Noise,
			(float64(ctx.BlockX)+dep(node.Dep0))*node.XZScale,
			(float64(ctx.BlockY)+dep(node.Dep1))*node.YScale,
			(float64(ctx.BlockZ)+dep(node.Dep2))*node.XZScale,
		)
	case OpYClampedGradient:
		return yClampedGradient(ctx.BlockY, node.FromY, node.ToY, node.FromValue, node.ToValue)
	case OpFlatCache, OpCache2D, OpCacheOnce, OpInterpolated, OpBlendDensity:
		return dep(node.Dep0)
	case OpBlendAlpha:
		return 1
	case OpBlendOffset:
		return 0
	case OpRangeChoice:
		input := dep(node.Dep0)
		if input >= node.Min && input < node.Max {
			return dep(node.Dep1)
		}
		return dep(node.Dep2)
	case OpSpline:
		return evalSpline(node.Spline, func(arg ArgRef) float64 {
			if arg.Node >= 0 {
				return dep(arg.Node)
			}
			return arg.Const
		})
	case OpWeirdScaledSampler:
		input := dep(node.Dep0)
		rarity := rarityValue(node.Rarity, input)
		return rarity * math.Abs(noises.Sample(node.Noise, float64(ctx.BlockX)/rarity, float64(ctx.BlockY)/rarity, float64(ctx.BlockZ)/rarity))
	case OpShiftA:
		return noises.Sample(node.Noise, float64(ctx.BlockX), 0, float64(ctx.BlockZ)) * 4
	case OpShiftB:
		return noises.Sample(node.Noise, float64(ctx.BlockZ), float64(ctx.BlockX), 0) * 4
	case OpShift:
		return noises.Sample(node.Noise, float64(ctx.BlockX), 0, float64(ctx.BlockZ)) * 4
	case OpOldBlendedNoise:
		return noises.SampleBlendedNoise(
			float64(ctx.BlockX),
			float64(ctx.BlockY),
			float64(ctx.BlockZ),
			node.XZScale,
			node.YScale,
			node.XZFactor,
			node.YFactor,
			node.SmearScaleMultiplier,
		)
	case OpEndIslands:
		return noises.SampleEndIslands(ctx.BlockX, ctx.BlockZ)
	case OpInvert:
		return -dep(node.Dep0)
	default:
		return 0
	}
}

func FindTopSurface(blockX, blockZ, lowerBound, upperBound, cellHeight int, densityFn func(y int) float64) float64 {
	y := (upperBound / cellHeight) * cellHeight
	if y <= lowerBound {
		return float64(lowerBound)
	}
	for y >= lowerBound {
		if densityFn(y) > 0 {
			return float64(y)
		}
		y -= cellHeight
	}
	return float64(lowerBound)
}

func evalSpline(spline *Spline, evalArg func(ArgRef) float64) float64 {
	if spline == nil || len(spline.Points) == 0 {
		return 0
	}
	coord := evalArg(spline.Coordinate)
	if coord <= spline.Points[0].Location {
		return evalSplineValue(spline.Points[0].Value, evalArg)
	}

	last := spline.Points[len(spline.Points)-1]
	if coord >= last.Location {
		return evalSplineValue(last.Value, evalArg)
	}

	for i := 0; i < len(spline.Points)-1; i++ {
		p0 := spline.Points[i]
		p1 := spline.Points[i+1]
		if coord >= p1.Location && i < len(spline.Points)-2 {
			continue
		}

		v0 := evalSplineValue(p0.Value, evalArg)
		v1 := evalSplineValue(p1.Value, evalArg)
		dt := p1.Location - p0.Location
		if dt == 0 {
			return v1
		}
		t := (coord - p0.Location) / dt
		t2 := t * t
		t3 := t2 * t
		h00 := 2*t3 - 3*t2 + 1
		h10 := t3 - 2*t2 + t
		h01 := -2*t3 + 3*t2
		h11 := t3 - t2
		return h00*v0 + h10*dt*p0.Derivative + h01*v1 + h11*dt*p1.Derivative
	}

	return evalSplineValue(last.Value, evalArg)
}

func evalSplineValue(value SplineValue, evalArg func(ArgRef) float64) float64 {
	if value.Nested != nil {
		return evalSpline(value.Nested, evalArg)
	}
	return value.Const
}

func yClampedGradient(y, fromY, toY int, fromValue, toValue float64) float64 {
	if y <= fromY {
		return fromValue
	}
	if y >= toY {
		return toValue
	}
	t := float64(y-fromY) / float64(toY-fromY)
	return fromValue + t*(toValue-fromValue)
}

func halfNegative(v float64) float64 {
	if v > 0 {
		return v
	}
	return v * 0.5
}

func quarterNegative(v float64) float64 {
	if v > 0 {
		return v
	}
	return v * 0.25
}

func squeeze(v float64) float64 {
	c := clampFloat(v, -1, 1)
	return c/2 - c*c*c/24
}

func rarityValue(kind RarityType, input float64) float64 {
	if kind == RarityType1 {
		if input < -0.5 {
			return 0.75
		}
		if input < 0 {
			return 1
		}
		if input < 0.5 {
			return 1.5
		}
		return 2
	}
	if input < -0.75 {
		return 0.5
	}
	if input < -0.5 {
		return 0.75
	}
	if input < 0.5 {
		return 1
	}
	if input < 0.75 {
		return 2
	}
	return 3
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func clampFloat(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
