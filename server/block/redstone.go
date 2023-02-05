package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	"golang.org/x/exp/slices"
)

// RedstoneUpdater represents a block that can be updated through a change in redstone signal.
type RedstoneUpdater interface {
	// RedstoneUpdate is called when a change in redstone signal is computed.
	RedstoneUpdate(pos cube.Pos, w *world.World)
}

// RedstoneBlocking represents a block that blocks redstone signals.
type RedstoneBlocking interface {
	// RedstoneBlocking returns true if the block blocks redstone signals.
	RedstoneBlocking() bool
}

// wireNetwork implements a minimally-invasive bolt-on accelerator that performs a breadth-first search through redstone
// wires in order to more efficiently and compute new redstone wire power levels and determine the order in which other
// blocks should be updated. This implementation is heavily based off of RedstoneWireTurbo and MCHPRS.
type wireNetwork struct {
	nodes            []*wireNode
	nodeCache        map[cube.Pos]*wireNode
	updateQueue      [3][]*wireNode
	currentWalkLayer uint32
}

// wireNode is a data structure to keep track of redstone wires and neighbours that will receive updates.
type wireNode struct {
	visited bool

	pos   cube.Pos
	block world.Block

	neighbours []*wireNode
	oriented   bool

	xBias int32
	zBias int32

	layer uint32
}

const (
	wireHeadingNorth = 0
	wireHeadingEast  = 1
	wireHeadingSouth = 2
	wireHeadingWest  = 3
)

// updateStrongRedstone sets off the breadth-first walk through all redstone wires connected to the initial position
// triggered. This is the main entry point for the redstone update algorithm.
func updateStrongRedstone(pos cube.Pos, w *world.World) {
	n := &wireNetwork{
		nodeCache:   make(map[cube.Pos]*wireNode),
		updateQueue: [3][]*wireNode{},
	}

	root := &wireNode{
		block:   w.Block(pos),
		pos:     pos,
		visited: true,
	}
	n.nodeCache[pos] = root
	n.nodes = append(n.nodes, root)

	n.propagateChanges(w, root, 0)
	n.breadthFirstWalk(w)
}

// updateAroundRedstone updates redstone components around the given centre position. It will also ignore any faces
// provided within the ignoredFaces parameter. This implementation is based off of RedstoneCircuit and Java 1.19.
func updateAroundRedstone(centre cube.Pos, w *world.World, ignoredFaces ...cube.Face) {
	for _, face := range []cube.Face{
		cube.FaceWest,
		cube.FaceEast,
		cube.FaceDown,
		cube.FaceUp,
		cube.FaceNorth,
		cube.FaceSouth,
	} {
		if slices.Contains(ignoredFaces, face) {
			continue
		}

		pos := centre.Side(face)
		if r, ok := w.Block(pos).(RedstoneUpdater); ok {
			r.RedstoneUpdate(pos, w)
		}
	}
}

// updateDirectionalRedstone updates redstone components through the given face. This implementation is based off of
// RedstoneCircuit and Java 1.19.
func updateDirectionalRedstone(pos cube.Pos, w *world.World, face cube.Face) {
	updateAroundRedstone(pos, w)
	updateAroundRedstone(pos.Side(face), w, face.Opposite())
}

// updateGateRedstone is used to update redstone gates on each face of the given offset centre position.
func updateGateRedstone(centre cube.Pos, w *world.World, face cube.Face) {
	pos := centre.Side(face.Opposite())
	if r, ok := w.Block(pos).(RedstoneUpdater); ok {
		r.RedstoneUpdate(pos, w)
	}

	updateAroundRedstone(pos, w, face)
}

// receivedRedstonePower returns true if the given position is receiving power from any faces that aren't ignored.
func receivedRedstonePower(pos cube.Pos, w *world.World, ignoredFaces ...cube.Face) bool {
	for _, face := range cube.Faces() {
		if slices.Contains(ignoredFaces, face) {
			continue
		}
		if w.RedstonePower(pos.Side(face), face, true) > 0 {
			return true
		}
	}
	return false
}

// identifyNeighbours identifies the neighbouring positions of a given node, determines their types, and links them into
// the graph. After that, based on what nodes in the graph have been visited, the neighbours are reordered left-to-right
// relative to the direction of information flow.
func (n *wireNetwork) identifyNeighbours(w *world.World, node *wireNode) {
	neighbours := computeRedstoneNeighbours(node.pos)
	neighboursVisited := make([]bool, 0, 24)
	neighbourNodes := make([]*wireNode, 0, 24)
	for _, neighbourPos := range neighbours[:24] {
		neighbour, ok := n.nodeCache[neighbourPos]
		if !ok {
			neighbour = &wireNode{
				pos:   neighbourPos,
				block: w.Block(neighbourPos),
			}
			n.nodeCache[neighbourPos] = neighbour
			n.nodes = append(n.nodes, neighbour)
		}

		neighbourNodes = append(neighbourNodes, neighbour)
		neighboursVisited = append(neighboursVisited, neighbour.visited)
	}

	fromWest := neighboursVisited[0] || neighboursVisited[7] || neighboursVisited[8]
	fromEast := neighboursVisited[1] || neighboursVisited[12] || neighboursVisited[13]
	fromNorth := neighboursVisited[4] || neighboursVisited[17] || neighboursVisited[20]
	fromSouth := neighboursVisited[5] || neighboursVisited[18] || neighboursVisited[21]

	var cX, cZ int32
	if fromWest {
		cX++
	}
	if fromEast {
		cX--
	}
	if fromNorth {
		cZ++
	}
	if fromSouth {
		cZ--
	}

	var heading uint32
	if cX == 0 && cZ == 0 {
		heading = computeRedstoneHeading(node.xBias, node.zBias)
		for _, neighbourNode := range neighbourNodes {
			neighbourNode.xBias = node.xBias
			neighbourNode.zBias = node.zBias
		}
	} else {
		if cX != 0 && cZ != 0 {
			if node.xBias != 0 {
				cZ = 0
			}
			if node.zBias != 0 {
				cX = 0
			}
		}
		heading = computeRedstoneHeading(cX, cZ)
		for _, neighbourNode := range neighbourNodes {
			neighbourNode.xBias = cX
			neighbourNode.zBias = cZ
		}
	}

	n.orientNeighbours(&neighbourNodes, node, heading)
}

// reordering contains lookup tables that completely remap neighbour positions into a left-to-right ordering, based on
// the cardinal direction that is determined to be forward.
var reordering = [][]uint32{
	{2, 3, 16, 19, 0, 4, 1, 5, 7, 8, 17, 20, 12, 13, 18, 21, 6, 9, 22, 14, 11, 10, 23, 15},
	{2, 3, 16, 19, 4, 1, 5, 0, 17, 20, 12, 13, 18, 21, 7, 8, 22, 14, 11, 15, 23, 9, 6, 10},
	{2, 3, 16, 19, 1, 5, 0, 4, 12, 13, 18, 21, 7, 8, 17, 20, 11, 15, 23, 10, 6, 14, 22, 9},
	{2, 3, 16, 19, 5, 0, 4, 1, 18, 21, 7, 8, 17, 20, 12, 13, 23, 10, 6, 9, 22, 15, 11, 14},
}

// orientNeighbours reorders the neighbours of a node based on the direction that is determined to be forward.
func (n *wireNetwork) orientNeighbours(src *[]*wireNode, dst *wireNode, heading uint32) {
	dst.oriented = true
	dst.neighbours = make([]*wireNode, 0, 24)
	for _, i := range reordering[heading] {
		dst.neighbours = append(dst.neighbours, (*src)[i])
	}
}

// propagateChanges propagates changes for any redstone wire in layer N, informing the neighbours to recompute their
// states in layers N + 1 and N + 2.
func (n *wireNetwork) propagateChanges(w *world.World, node *wireNode, layer uint32) {
	if !node.oriented {
		n.identifyNeighbours(w, node)
	}

	layerOne := layer + 1
	for _, neighbour := range node.neighbours[:24] {
		if layerOne > neighbour.layer {
			neighbour.layer = layerOne
			n.updateQueue[1] = append(n.updateQueue[1], neighbour)
		}
	}

	layerTwo := layer + 2
	for _, neighbour := range node.neighbours[:4] {
		if layerTwo > neighbour.layer {
			neighbour.layer = layerTwo
			n.updateQueue[2] = append(n.updateQueue[2], neighbour)
		}
	}
}

// breadthFirstWalk performs a breadth-first (layer by layer) traversal through redstone wires, propagating value
// changes to neighbours in the order that they are visited.
func (n *wireNetwork) breadthFirstWalk(w *world.World) {
	n.shiftQueue()
	n.currentWalkLayer = 1

	for len(n.updateQueue[0]) > 0 || len(n.updateQueue[1]) > 0 {
		for _, node := range n.updateQueue[0] {
			if _, ok := node.block.(RedstoneWire); ok {
				n.updateNode(w, node, n.currentWalkLayer)
				continue
			}
			if t, ok := node.block.(RedstoneUpdater); ok {
				t.RedstoneUpdate(node.pos, w)
			}
		}

		n.shiftQueue()
		n.currentWalkLayer++
	}

	n.currentWalkLayer = 0
}

// shiftQueue shifts the update queue, moving all nodes from the current layer to the next layer. The last queue is then
// simply invalidated.
func (n *wireNetwork) shiftQueue() {
	n.updateQueue[0] = n.updateQueue[1]
	n.updateQueue[1] = n.updateQueue[2]
	n.updateQueue[2] = nil
}

// updateNode processes a node which has had neighbouring redstone wires that have experienced value changes.
func (n *wireNetwork) updateNode(w *world.World, node *wireNode, layer uint32) {
	node.visited = true

	oldWire := node.block.(RedstoneWire)
	newWire := n.calculateCurrentChanges(w, node)
	if oldWire.Power != newWire.Power {
		node.block = newWire

		n.propagateChanges(w, node, layer)
	}
}

var (
	rsNeighbours   = [...]uint32{4, 5, 6, 7}
	rsNeighboursUp = [...]uint32{9, 11, 13, 15}
	rsNeighboursDn = [...]uint32{8, 10, 12, 14}
)

// calculateCurrentChanges computes redstone wire power levels from neighboring blocks. Modifications cut the number of
// power level changes by about 45% from vanilla, and also synergies well with the breadth-first search implementation.
func (n *wireNetwork) calculateCurrentChanges(w *world.World, node *wireNode) RedstoneWire {
	wire := node.block.(RedstoneWire)
	i := wire.Power

	var blockPower int
	if !node.oriented {
		n.identifyNeighbours(w, node)
	}

	var wirePower int
	for _, face := range cube.Faces() {
		wirePower = max(wirePower, w.RedstonePower(node.pos.Side(face), face, false))
	}

	if wirePower < 15 {
		centerUp := node.neighbours[1].block
		_, centerUpSolid := centerUp.Model().(model.Solid)
		for m := 0; m < 4; m++ {
			neighbour := node.neighbours[rsNeighbours[m]].block
			_, neighbourSolid := neighbour.Model().(model.Solid)

			blockPower = n.maxCurrentStrength(neighbour, blockPower)
			if !neighbourSolid {
				neighbourDown := node.neighbours[rsNeighboursDn[m]].block
				blockPower = n.maxCurrentStrength(neighbourDown, blockPower)
			} else if d, ok := neighbour.(LightDiffuser); (!ok || d.LightDiffusionLevel() > 0) && !centerUpSolid {
				neighbourUp := node.neighbours[rsNeighboursUp[m]].block
				blockPower = n.maxCurrentStrength(neighbourUp, blockPower)
			}
		}
	}

	j := blockPower - 1
	if wirePower > j {
		j = wirePower
	}

	if i != j {
		wire.Power = j
		w.SetBlock(node.pos, wire, &world.SetOpts{DisableBlockUpdates: true})
	}
	return wire
}

// maxCurrentStrength computes a redstone wire's power level based on a cached state.
func (n *wireNetwork) maxCurrentStrength(neighbour world.Block, strength int) int {
	if wire, ok := neighbour.(RedstoneWire); ok {
		return max(wire.Power, strength)
	}
	return strength
}

// computeRedstoneNeighbours computes the neighbours of a redstone wire node, ignoring neighbours that don't necessarily
// need to be updated, but are in vanilla.
func computeRedstoneNeighbours(pos cube.Pos) []cube.Pos {
	return []cube.Pos{
		// Immediate neighbours, in the order of west, east, down, up, north, and finally south.
		pos.Side(cube.FaceWest),
		pos.Side(cube.FaceEast),
		pos.Side(cube.FaceDown),
		pos.Side(cube.FaceUp),
		pos.Side(cube.FaceNorth),
		pos.Side(cube.FaceSouth),

		// Neighbours of neighbours, in the same order, except that duplicates are not included.
		pos.Side(cube.FaceWest).Side(cube.FaceWest),
		pos.Side(cube.FaceWest).Side(cube.FaceDown),
		pos.Side(cube.FaceWest).Side(cube.FaceUp),
		pos.Side(cube.FaceWest).Side(cube.FaceNorth),
		pos.Side(cube.FaceWest).Side(cube.FaceSouth),

		pos.Side(cube.FaceEast).Side(cube.FaceEast),
		pos.Side(cube.FaceEast).Side(cube.FaceDown),
		pos.Side(cube.FaceEast).Side(cube.FaceUp),
		pos.Side(cube.FaceEast).Side(cube.FaceNorth),
		pos.Side(cube.FaceEast).Side(cube.FaceSouth),

		pos.Side(cube.FaceDown).Side(cube.FaceDown),
		pos.Side(cube.FaceDown).Side(cube.FaceNorth),
		pos.Side(cube.FaceDown).Side(cube.FaceSouth),

		pos.Side(cube.FaceUp).Side(cube.FaceUp),
		pos.Side(cube.FaceUp).Side(cube.FaceNorth),
		pos.Side(cube.FaceUp).Side(cube.FaceSouth),

		pos.Side(cube.FaceNorth).Side(cube.FaceNorth),
		pos.Side(cube.FaceSouth).Side(cube.FaceSouth),
	}
}

// computeRedstoneHeading computes the cardinal direction that is "forward" given which redstone wires have been visited
// and which have not around the position currently being processed.
func computeRedstoneHeading(rX, rZ int32) uint32 {
	code := (rX + 1) + 3*(rZ+1)
	switch code {
	case 0:
		return wireHeadingNorth
	case 1:
		return wireHeadingNorth
	case 2:
		return wireHeadingEast
	case 3:
		return wireHeadingWest
	case 4:
		return wireHeadingWest
	case 5:
		return wireHeadingEast
	case 6:
		return wireHeadingSouth
	case 7:
		return wireHeadingSouth
	case 8:
		return wireHeadingSouth
	}
	panic("should never happen")
}
