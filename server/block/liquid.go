package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
	"sync"
)

// LiquidRemovable represents a block that may be removed by a liquid flowing into it. When this happens, the
// block's drops are dropped at the position if HasLiquidDrops returns true.
type LiquidRemovable interface {
	HasLiquidDrops() bool
}

// tickLiquid ticks the liquid block passed at a specific position in the world. Depending on the surroundings
// and the liquid block, the liquid will either spread or decrease in depth. Additionally, the liquid might
// be turned into a solid block if a different liquid is next to it.
func tickLiquid(b world.Liquid, pos cube.Pos, w *world.World) {
	if !source(b) && !sourceAround(b, pos, w) {
		if b.LiquidDepth()-4 <= 0 {
			w.SetLiquid(pos, nil)
			return
		}
		w.SetLiquid(pos, b.WithDepth(b.LiquidDepth()-2*b.SpreadDecay(), false))
		return
	}
	displacer, _ := w.Block(pos).(world.LiquidDisplacer)

	canFlowBelow := canFlowInto(b, w, pos.Add(cube.Pos{0, -1}), false)
	if b.LiquidFalling() && !canFlowBelow {
		b = b.WithDepth(8, true)
	} else if canFlowBelow {
		below := pos.Add(cube.Pos{0, -1})
		if displacer == nil || !displacer.SideClosed(pos, below, w) {
			flowInto(b.WithDepth(8, true), pos, below, w, true)
		}
	}

	depth, decay := b.LiquidDepth(), b.SpreadDecay()
	if depth <= decay {
		// Current depth is smaller than the decay, so spreading will result in nothing.
		return
	}
	if source(b) || !canFlowBelow {
		paths := calculateLiquidPaths(b, pos, w, displacer)
		if len(paths) == 0 {
			spreadOutwards(b, pos, w, displacer)
			return
		}

		smallestLen := len(paths[0])
		for _, path := range paths {
			if len(path) <= smallestLen {
				flowInto(b, pos, path[0], w, false)
			}
		}
	}
}

// source checks if a liquid is a source block.
func source(b world.Liquid) bool {
	return b.LiquidDepth() == 8 && !b.LiquidFalling()
}

// spreadOutwards spreads the liquid outwards into the horizontal directions.
func spreadOutwards(b world.Liquid, pos cube.Pos, w *world.World, displacer world.LiquidDisplacer) {
	pos.Neighbours(func(neighbour cube.Pos) {
		if neighbour[1] == pos[1] {
			if displacer == nil || !displacer.SideClosed(pos, neighbour, w) {
				flowInto(b, pos, neighbour, w, false)
			}
		}
	}, w.Range())
}

// sourceAround checks if there is a source in the blocks around the position passed.
func sourceAround(b world.Liquid, pos cube.Pos, w *world.World) (sourcePresent bool) {
	pos.Neighbours(func(neighbour cube.Pos) {
		if neighbour[1] == pos[1]-1 {
			// We don't care about water below this one.
			return
		}
		side, ok := w.Liquid(neighbour)
		if !ok || side.LiquidType() != b.LiquidType() {
			return
		}
		if displacer, ok := w.Block(neighbour).(world.LiquidDisplacer); ok && displacer.SideClosed(neighbour, pos, w) {
			// The side towards this liquid was closed, so this cannot function as a source for this
			// liquid.
			return
		}
		if neighbour[1] == pos[1]+1 || source(side) || side.LiquidDepth() > b.LiquidDepth() {
			sourcePresent = true
		}
	}, w.Range())
	return
}

// flowInto makes the liquid passed flow into the position passed in a world. If successful, the block at that
// position will be broken and the liquid with a lower depth will replace it.
func flowInto(b world.Liquid, src, pos cube.Pos, w *world.World, falling bool) bool {
	newDepth := b.LiquidDepth() - b.SpreadDecay()
	if falling {
		newDepth = b.LiquidDepth()
	}
	if newDepth <= 0 && !falling {
		return false
	}
	existing := w.Block(pos)
	if existingLiquid, alsoLiquid := existing.(world.Liquid); alsoLiquid && existingLiquid.LiquidType() == b.LiquidType() {
		if existingLiquid.LiquidDepth() >= newDepth || existingLiquid.LiquidFalling() {
			// The existing liquid had a higher depth than the one we're propagating, or it was falling
			// (basically considered full depth), so no need to continue.
			return true
		}
		ctx := event.C()
		if w.Handler().HandleLiquidFlow(ctx, src, pos, b.WithDepth(newDepth, falling), existing); ctx.Cancelled() {
			return false
		}
		w.SetLiquid(pos, b.WithDepth(newDepth, falling))
		return true
	} else if alsoLiquid {
		existingLiquid.Harden(pos, w, &src)
		return false
	}
	if _, ok := existing.(world.LiquidDisplacer); ok {
		if _, ok := w.Liquid(pos); ok {
			// We've got a liquid displacer and it's got a liquid within it, so we can't flow into this.
			return false
		}
	}
	removable, ok := existing.(LiquidRemovable)
	if !ok {
		// Can't flow into this block.
		return false
	}
	if _, air := existing.(Air); !air {
		w.SetBlock(pos, nil, nil)
	}
	if removable.HasLiquidDrops() {
		if b, ok := existing.(Breakable); ok {
			for _, d := range b.BreakInfo().Drops(item.ToolNone{}, nil) {
				itemEntity := entity.NewItem(d, pos.Vec3Centre())
				itemEntity.SetVelocity(mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1})
				w.AddEntity(itemEntity)
			}
		} else {
			panic("liquid drops should always implement breakable")
		}
	}
	ctx := event.C()
	if w.Handler().HandleLiquidFlow(ctx, src, pos, b.WithDepth(newDepth, falling), existing); ctx.Cancelled() {
		return false
	}
	w.SetLiquid(pos, b.WithDepth(newDepth, falling))
	return true
}

// liquidPath represents a path to an empty lower block or a block that can be flown into by a liquid, which
// the liquid tends to flow into. All paths with the lowest length will be filled with water.
type liquidPath []cube.Pos

// calculateLiquidPaths calculates paths in the world that the liquid passed can flow in to reach lower
// grounds, starting at the position passed.
// If none of these paths can be found, the returned slice has a length of 0.
func calculateLiquidPaths(b world.Liquid, pos cube.Pos, w *world.World, displacer world.LiquidDisplacer) []liquidPath {
	queue := liquidQueuePool.Get().(*liquidQueue)
	defer func() {
		queue.Reset()
		liquidQueuePool.Put(queue)
	}()
	queue.PushBack(liquidNode{x: pos[0], z: pos[2], depth: int8(b.LiquidDepth())})
	decay := int8(b.SpreadDecay())

	paths := make([]liquidPath, 0, 3)
	first := true

	for {
		if queue.Len() == 0 {
			break
		}
		node := queue.Front()
		neighA, neighB, neighC, neighD := node.neighbours(decay * 2)
		if !first || (displacer == nil || !displacer.SideClosed(pos, cube.Pos{neighA.x, pos[1], neighA.z}, w)) {
			if spreadNeighbour(b, pos, w, neighA, queue) {
				queue.shortestPath = neighA.Len()
				paths = append(paths, neighA.Path(pos))
			}
		}
		if !first || (displacer == nil || !displacer.SideClosed(pos, cube.Pos{neighB.x, pos[1], neighB.z}, w)) {
			if spreadNeighbour(b, pos, w, neighB, queue) {
				queue.shortestPath = neighB.Len()
				paths = append(paths, neighB.Path(pos))
			}
		}
		if !first || (displacer == nil || !displacer.SideClosed(pos, cube.Pos{neighC.x, pos[1], neighC.z}, w)) {
			if spreadNeighbour(b, pos, w, neighC, queue) {
				queue.shortestPath = neighC.Len()
				paths = append(paths, neighC.Path(pos))
			}
		}
		if !first || (displacer == nil || !displacer.SideClosed(pos, cube.Pos{neighD.x, pos[1], neighD.z}, w)) {
			if spreadNeighbour(b, pos, w, neighD, queue) {
				queue.shortestPath = neighD.Len()
				paths = append(paths, neighD.Path(pos))
			}
		}
		first = false
	}
	return paths
}

// spreadNeighbour attempts to spread a path node into the neighbour passed. Note that this does not spread
// the liquid, it only spreads the node used to calculate flow paths.
func spreadNeighbour(b world.Liquid, src cube.Pos, w *world.World, node liquidNode, queue *liquidQueue) bool {
	if node.depth+3 <= 0 {
		// Depth has reached zero or below, can't spread any further.
		return false
	}
	if node.Len() > queue.shortestPath {
		// This path is longer than any existing path, so don't spread any further.
		return false
	}
	pos := cube.Pos{node.x, src[1], node.z}
	if !canFlowInto(b, w, pos, true) {
		// Can't flow into this block, can't spread any further.
		return false
	}
	pos[1]--
	if canFlowInto(b, w, pos, false) {
		return true
	}
	queue.PushBack(node)
	return false
}

// canFlowInto checks if a liquid can flow into the block present in the world at a specific block position.
func canFlowInto(b world.Liquid, w *world.World, pos cube.Pos, sideways bool) bool {
	bl := w.Block(pos)
	if _, air := bl.(Air); air {
		// Fast route for air: A type assert to a concrete type is much faster than a type assert to an interface.
		return true
	}
	if _, ok := bl.(LiquidRemovable); ok {
		if liq, ok := bl.(world.Liquid); ok && sideways {
			if (liq.LiquidDepth() == 8 && !liq.LiquidFalling()) || liq.LiquidType() != b.LiquidType() {
				// Can't flow into a liquid if it has a depth of 8 or if it doesn't have the same type.
				return false
			}
		}
		return true
	}
	return false
}

// liquidNode represents a position that is part of a flow path for a liquid.
type liquidNode struct {
	x, z     int
	depth    int8
	previous *liquidNode
}

// neighbours returns the four horizontal neighbours of the node with decreased depth.
func (node liquidNode) neighbours(decay int8) (a, b, c, d liquidNode) {
	return liquidNode{x: node.x - 1, z: node.z, depth: node.depth - decay, previous: &node},
		liquidNode{x: node.x + 1, z: node.z, depth: node.depth - decay, previous: &node},
		liquidNode{x: node.x, z: node.z - 1, depth: node.depth - decay, previous: &node},
		liquidNode{x: node.x, z: node.z + 1, depth: node.depth - decay, previous: &node}
}

// Len returns the length of the path created by the node.
func (node liquidNode) Len() int {
	i := 1
	for {
		if node.previous == nil {
			return i - 1
		}
		//noinspection GoAssignmentToReceiver
		node = *node.previous
		i++
	}
}

// Path converts the liquid node into a path.
func (node liquidNode) Path(src cube.Pos) liquidPath {
	l := node.Len()
	path := make(liquidPath, l)
	i := l - 1
	for {
		if node.previous == nil {
			return path
		}
		path[i] = cube.Pos{node.x, src[1], node.z}

		//noinspection GoAssignmentToReceiver
		node = *node.previous
		i--
	}
}

// liquidQueuePool is use to re-use liquid node queues.
var liquidQueuePool = sync.Pool{
	New: func() any {
		return &liquidQueue{
			nodes:        make([]liquidNode, 0, 64),
			shortestPath: math.MaxInt8,
		}
	},
}

// liquidQueue represents a queue that may be used to push nodes into and take them out of it.
type liquidQueue struct {
	nodes        []liquidNode
	i            int
	shortestPath int
}

func (q *liquidQueue) PushBack(node liquidNode) {
	q.nodes = append(q.nodes, node)
}

func (q *liquidQueue) Front() liquidNode {
	v := q.nodes[q.i]
	q.i++
	return v
}

func (q *liquidQueue) Len() int {
	return len(q.nodes) - q.i
}

func (q *liquidQueue) Reset() {
	q.nodes = q.nodes[:0]
	q.i = 0
	q.shortestPath = math.MaxInt8
}
