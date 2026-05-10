package world

import (
	"encoding/binary"
	"hash/fnv"
	"maps"
	"slices"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
)

// RedstoneUpdateCause describes the world event that caused a redstone update to be evaluated.
type RedstoneUpdateCause uint8

const (
	// RedstoneUpdateCauseBlockUpdate means a block or liquid change invalidated nearby redstone.
	RedstoneUpdateCauseBlockUpdate RedstoneUpdateCause = iota
	// RedstoneUpdateCauseScheduledTick means a scheduled redstone tick invalidated a component.
	RedstoneUpdateCauseScheduledTick
	// RedstoneUpdateCauseCompilerRebuild means a redstone compiler rebuild invalidated a component.
	RedstoneUpdateCauseCompilerRebuild
)

// RedstoneUpdate represents a redstone state transition proposed by the world redstone engine. Handlers may cancel
// the event to suppress the proposed mutation and any propagation from that mutation.
type RedstoneUpdate struct {
	// Pos is the block position that will receive the update.
	Pos cube.Pos
	// ChangedNeighbour is the neighbouring block that caused the update, if any.
	ChangedNeighbour cube.Pos
	// Before is the block currently at Pos.
	Before Block
	// After is the block that will replace Before, if the update is a block-state update. After is nil for updates
	// that perform side effects instead of replacing the block.
	After Block
	// OldPower is the last redstone power observed by the engine at Pos.
	OldPower int
	// NewPower is the redstone power observed by the engine at Pos for this update.
	NewPower int
	// CurrentTick is the world tick during which the update was evaluated.
	CurrentTick int64
	// NetworkID identifies the compiled dynamic redstone region that produced the update.
	NetworkID uint64
	// Cause identifies why the update was evaluated.
	Cause RedstoneUpdateCause
}

// RedstonePowerSource is implemented by blocks that emit redstone power. The face passed is the face of the source
// block that power is being read from.
type RedstonePowerSource interface {
	RedstonePower(pos cube.Pos, tx *Tx, face cube.Face) int
}

// RedstoneStrongPowerSource is implemented by sources that strongly power blocks from specific faces. Strong power may
// pass through solid blocks, unlike weak redstone wire power.
type RedstoneStrongPowerSource interface {
	RedstoneStrongPower(pos cube.Pos, tx *Tx, face cube.Face) int
}

// RedstonePowerRelayer is implemented by redstone wire-like blocks that relay power through a compiled redstone
// network. The returned value is the signal loss when power enters through from and leaves through to.
type RedstonePowerRelayer interface {
	RedstoneSignalLoss(pos cube.Pos, tx *Tx, from, to cube.Face) int
}

// RedstonePowerRelayerNeighbourer may be implemented by relayers with non-adjacent connections, such as redstone
// wire stepping up and down block edges.
type RedstonePowerRelayerNeighbourer interface {
	RedstoneRelayerNeighbours(pos cube.Pos, tx *Tx) []cube.Pos
}

// RedstonePowerConsumer is implemented by blocks whose block state changes when their input power changes. The
// returned block is written to the world if changed is true and the redstone update event is not cancelled.
type RedstonePowerConsumer interface {
	RedstonePowerUpdate(pos cube.Pos, tx *Tx, power int) (after Block, changed bool)
}

// RedstonePowerTransitionConsumer may be implemented by consumers that need the previous and new input power to
// distinguish a real redstone transition from a same-power block update.
type RedstonePowerTransitionConsumer interface {
	RedstonePowerTransitionUpdate(pos cube.Pos, tx *Tx, oldPower, newPower int) (after Block, changed bool)
}

// RedstonePowerSounder may be implemented by consumers that play a sound after an uncancelled redstone-driven state
// change.
type RedstonePowerSounder interface {
	RedstonePowerUpdateSound(pos cube.Pos, tx *Tx, before, after Block, oldPower, newPower int) Sound
}

// RedstonePowerPostUpdater may be implemented by consumers that need to apply side effects after an uncancelled
// redstone state update, such as syncing the other half of a door.
type RedstonePowerPostUpdater interface {
	RedstonePowerPostUpdate(pos cube.Pos, tx *Tx, before, after Block, oldPower, newPower int)
}

// RedstonePowerAction is implemented by blocks that perform a side effect when their input power changes, such as TNT
// priming on a rising edge. The action is run only if the redstone update event is not cancelled. The returned bool
// reports whether a side effect was performed.
type RedstonePowerAction interface {
	RedstonePowerAction(pos cube.Pos, tx *Tx, oldPower, newPower int) bool
}

// RedstoneComparatorReadable is implemented by blocks that expose an analog signal to a comparator.
type RedstoneComparatorReadable interface {
	RedstoneComparatorOutput(pos cube.Pos, tx *Tx, face cube.Face) int
}

type redstoneEngine struct {
	currentTick int64
	dirty       map[cube.Pos]redstoneDirty
	power       map[cube.Pos]int
	output      map[cube.Pos]int
	evaluating  map[cube.Pos]struct{}
}

type redstoneDirty struct {
	changed cube.Pos
	cause   RedstoneUpdateCause
}

type redstoneGraph struct {
	id    uint64
	nodes []redstoneNode
	edges []redstoneEdge
}

type redstoneNode struct {
	pos    cube.Pos
	power  int
	source bool
	sink   bool
}

type redstoneEdge struct {
	from, to int
	weight   int
}

func newRedstoneEngine(tick int64) *redstoneEngine {
	return &redstoneEngine{
		currentTick: tick,
		dirty:       make(map[cube.Pos]redstoneDirty),
		power:       make(map[cube.Pos]int),
		output:      make(map[cube.Pos]int),
		evaluating:  make(map[cube.Pos]struct{}),
	}
}

func (e *redstoneEngine) invalidateAround(pos, changed cube.Pos, cause RedstoneUpdateCause, r cube.Range) {
	if e == nil || pos.OutOfBounds(r) {
		return
	}
	e.invalidate(pos, changed, cause, r)
	pos.Neighbours(func(neighbour cube.Pos) {
		e.invalidate(neighbour, changed, cause, r)
	}, r)
}

func (e *redstoneEngine) invalidate(pos, changed cube.Pos, cause RedstoneUpdateCause, r cube.Range) {
	if pos.OutOfBounds(r) {
		return
	}
	e.dirty[pos] = redstoneDirty{changed: changed, cause: cause}
}

func (e *redstoneEngine) removeChunk(chunkPos ChunkPos) {
	if e == nil {
		return
	}
	maps.DeleteFunc(e.dirty, func(pos cube.Pos, _ redstoneDirty) bool {
		return chunkPosFromBlockPos(pos) == chunkPos
	})
	maps.DeleteFunc(e.dirty, func(_ cube.Pos, dirty redstoneDirty) bool {
		return chunkPosFromBlockPos(dirty.changed) == chunkPos
	})
	maps.DeleteFunc(e.power, func(pos cube.Pos, _ int) bool {
		return chunkPosFromBlockPos(pos) == chunkPos
	})
	maps.DeleteFunc(e.output, func(pos cube.Pos, _ int) bool {
		return chunkPosFromBlockPos(pos) == chunkPos
	})
}

func (e *redstoneEngine) forget(pos cube.Pos) {
	if e == nil {
		return
	}
	delete(e.power, pos)
	delete(e.output, pos)
	delete(e.evaluating, pos)
}

func (e *redstoneEngine) tick(tx *Tx, tick int64) {
	if e == nil || len(e.dirty) == 0 {
		return
	}
	e.currentTick = tick
	dirty := maps.Clone(e.dirty)
	clear(e.dirty)

	candidates := slices.Collect(maps.Keys(dirty))
	slices.SortFunc(candidates, compareBlockPos)

	graph := e.compile(tx, candidates)
	powers := e.graphPower(tx, graph)
	for i, node := range graph.nodes {
		d := dirty[node.pos]
		if node.sink {
			e.update(tx, node.pos, d.changed, d.cause, graph.id, powers[i])
		}
	}
	for _, node := range graph.nodes {
		d := dirty[node.pos]
		if node.source {
			e.updateSource(tx, node.pos, d.changed, d.cause, graph.id)
		}
	}
}

func (e *redstoneEngine) compile(tx *Tx, candidates []cube.Pos) redstoneGraph {
	nodes := make([]redstoneNode, 0, len(candidates))
	seen := make(map[cube.Pos]struct{}, len(candidates)*2)
	for _, pos := range candidates {
		e.compileRegion(tx, pos, seen, &nodes)
		pos.Neighbours(func(neighbour cube.Pos) {
			if b, ok := tx.World().blockLoaded(neighbour); ok && isRedstoneRelevant(b) {
				e.compileRegion(tx, neighbour, seen, &nodes)
			}
		}, tx.Range())
	}
	slices.SortFunc(nodes, func(a, b redstoneNode) int {
		return compareBlockPos(a.pos, b.pos)
	})
	edges := e.compileEdges(tx, nodes)
	return redstoneGraph{id: redstoneGraphID(nodes, edges), nodes: nodes, edges: edges}
}

func (e *redstoneEngine) compileRegion(tx *Tx, pos cube.Pos, seen map[cube.Pos]struct{}, nodes *[]redstoneNode) {
	if _, ok := seen[pos]; ok || pos.OutOfBounds(tx.Range()) {
		return
	}
	queue := []cube.Pos{pos}
	for len(queue) != 0 {
		p := queue[0]
		queue = queue[1:]
		if _, ok := seen[p]; ok || p.OutOfBounds(tx.Range()) {
			continue
		}
		seen[p] = struct{}{}

		b, ok := tx.World().blockLoaded(p)
		if !ok {
			continue
		}
		source, consumer, action, relayer := classifyRedstoneBlock(b)
		if !source && !consumer && !action && !relayer {
			continue
		}
		*nodes = append(*nodes, redstoneNode{
			pos:    p,
			source: source,
			sink:   consumer || action,
		})
		if !relayer {
			continue
		}
		e.redstoneRelayerNeighbours(tx, p, b, func(neighbour cube.Pos) {
			if b, ok := tx.World().blockLoaded(neighbour); ok && isRedstoneRelevant(b) {
				queue = append(queue, neighbour)
			}
		})
	}
}

func (e *redstoneEngine) update(tx *Tx, pos, changed cube.Pos, cause RedstoneUpdateCause, graphID uint64, newPower int) {
	b := tx.Block(pos)
	oldPower, newPower := e.power[pos], clampRedstonePower(newPower)

	after, blockChanged := b, false
	if consumer, ok := b.(RedstonePowerTransitionConsumer); ok {
		after, blockChanged = consumer.RedstonePowerTransitionUpdate(pos, tx, oldPower, newPower)
	} else if consumer, ok := b.(RedstonePowerConsumer); ok {
		after, blockChanged = consumer.RedstonePowerUpdate(pos, tx, newPower)
	}
	action, hasAction := b.(RedstonePowerAction)
	actionChanged := hasAction && oldPower != newPower
	if !blockChanged && !actionChanged {
		e.power[pos] = newPower
		return
	}

	update := RedstoneUpdate{
		Pos:              pos,
		ChangedNeighbour: changed,
		Before:           b,
		After:            after,
		OldPower:         oldPower,
		NewPower:         newPower,
		CurrentTick:      e.currentTick,
		NetworkID:        graphID,
		Cause:            cause,
	}
	ctx := event.C(tx)
	if handler, ok := tx.World().Handler().(RedstoneHandler); ok {
		handler.HandleRedstoneUpdate(ctx, update)
	}
	if ctx.Cancelled() {
		return
	}
	if blockChanged {
		tx.SetBlock(pos, after, &SetOpts{DisableRedstoneUpdates: true})
		e.invalidateAround(pos, pos, RedstoneUpdateCauseBlockUpdate, tx.Range())
	}
	if blockChanged {
		if postUpdater, ok := b.(RedstonePowerPostUpdater); ok {
			postUpdater.RedstonePowerPostUpdate(pos, tx, b, after, oldPower, newPower)
		}
		if sounder, ok := b.(RedstonePowerSounder); ok {
			if s := sounder.RedstonePowerUpdateSound(pos, tx, b, after, oldPower, newPower); s != nil {
				tx.PlaySound(pos.Vec3Centre(), s)
			}
		}
	}
	acted := false
	if actionChanged {
		acted = action.RedstonePowerAction(pos, tx, oldPower, newPower)
	}
	if blockChanged || actionChanged || acted {
		e.power[pos] = newPower
	}
}

func (e *redstoneEngine) updateSource(tx *Tx, pos, changed cube.Pos, cause RedstoneUpdateCause, graphID uint64) {
	b := tx.Block(pos)
	oldPower, newPower := e.output[pos], e.sourcePower(pos, tx)
	if oldPower == newPower {
		return
	}
	update := RedstoneUpdate{
		Pos:              pos,
		ChangedNeighbour: changed,
		Before:           b,
		OldPower:         oldPower,
		NewPower:         newPower,
		CurrentTick:      e.currentTick,
		NetworkID:        graphID,
		Cause:            cause,
	}
	ctx := event.C(tx)
	if handler, ok := tx.World().Handler().(RedstoneHandler); ok {
		handler.HandleRedstoneUpdate(ctx, update)
	}
	if ctx.Cancelled() {
		return
	}
	e.output[pos] = newPower
	e.invalidateAround(pos, pos, RedstoneUpdateCauseBlockUpdate, tx.Range())
}

func (e *redstoneEngine) directPower(pos cube.Pos, tx *Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, e.directPowerFrom(pos, tx, face))
	}
	return power
}

func (e *redstoneEngine) directPowerFrom(pos cube.Pos, tx *Tx, face cube.Face) int {
	neighbour := pos.Side(face)
	if neighbour.OutOfBounds(tx.Range()) {
		return 0
	}
	b, ok := tx.World().blockLoaded(neighbour)
	if !ok {
		return 0
	}
	if source, ok := b.(RedstonePowerSource); ok {
		return clampRedstonePower(e.redstonePower(source, neighbour, tx, face.Opposite()))
	}
	return 0
}

func (e *redstoneEngine) strongPower(pos cube.Pos, tx *Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, e.strongPowerFrom(pos, tx, face))
	}
	return power
}

func (e *redstoneEngine) strongPowerFrom(pos cube.Pos, tx *Tx, face cube.Face) int {
	neighbour := pos.Side(face)
	if neighbour.OutOfBounds(tx.Range()) {
		return 0
	}
	b, ok := tx.World().blockLoaded(neighbour)
	if !ok {
		return 0
	}
	if source, ok := b.(RedstoneStrongPowerSource); ok {
		return clampRedstonePower(source.RedstoneStrongPower(neighbour, tx, face.Opposite()))
	}
	return 0
}

func (e *redstoneEngine) conductedStrongPower(pos cube.Pos, tx *Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, e.conductedStrongPowerFrom(pos, tx, face))
	}
	return power
}

func (e *redstoneEngine) conductedStrongPowerFrom(pos cube.Pos, tx *Tx, face cube.Face) int {
	conductorPos := pos.Side(face)
	if conductorPos.OutOfBounds(tx.Range()) {
		return 0
	}
	conductor, ok := tx.World().blockLoaded(conductorPos)
	if !ok || !conductor.Model().FaceSolid(conductorPos, face.Opposite(), tx) {
		return 0
	}
	power := 0
	for _, sourceFace := range cube.Faces() {
		power = max(power, e.strongPowerFrom(conductorPos, tx, sourceFace))
	}
	return power
}

func (e *redstoneEngine) sourcePower(pos cube.Pos, tx *Tx) int {
	b, ok := tx.World().blockLoaded(pos)
	if !ok {
		return 0
	}
	source, ok := b.(RedstonePowerSource)
	if !ok {
		return 0
	}
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, clampRedstonePower(e.redstonePower(source, pos, tx, face)))
	}
	return power
}

func (e *redstoneEngine) graphPower(tx *Tx, graph redstoneGraph) []int {
	powers := make([]int, len(graph.nodes))
	if len(graph.nodes) == 0 {
		return powers
	}

	index := make(map[cube.Pos]int, len(graph.nodes))
	sources := make([]RedstonePowerSource, len(graph.nodes))
	relayers := make([]RedstonePowerRelayer, len(graph.nodes))
	edges := make([][]redstoneEdge, len(graph.nodes))
	for i, node := range graph.nodes {
		index[node.pos] = i
		if b, ok := tx.World().blockLoaded(node.pos); ok {
			sources[i], _ = b.(RedstonePowerSource)
			relayers[i], _ = b.(RedstonePowerRelayer)
		}
	}
	for _, edge := range graph.edges {
		edges[edge.from] = append(edges[edge.from], edge)
	}

	queue := make([]int, 0, len(graph.nodes))
	push := func(i, power int) {
		power = clampRedstonePower(power)
		if power <= powers[i] {
			return
		}
		powers[i] = power
		queue = append(queue, i)
	}

	for i, node := range graph.nodes {
		push(i, e.conductedStrongPower(node.pos, tx))
	}

	for i, source := range sources {
		// Relayers such as redstone wire store their previous output as RedstonePower.
		// They must be recomputed from real sources, not used as seeds themselves.
		if source == nil || relayers[i] != nil {
			continue
		}
		pos := graph.nodes[i].pos
		for _, face := range cube.Faces() {
			j, ok := index[pos.Side(face)]
			if !ok {
				continue
			}
			power := clampRedstonePower(e.redstonePower(source, pos, tx, face))
			push(j, power)
		}
	}

	for head := 0; head < len(queue); head++ {
		i := queue[head]
		if relayers[i] == nil {
			continue
		}
		for _, edge := range edges[i] {
			push(edge.to, powers[i]-edge.weight)
		}
	}
	return powers
}

func (e *redstoneEngine) powerTo(pos cube.Pos, tx *Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, e.powerFrom(pos, tx, face, false))
	}
	power = max(power, e.conductedStrongPower(pos, tx))
	return clampRedstonePower(power)
}

func (e *redstoneEngine) powerFrom(pos cube.Pos, tx *Tx, face cube.Face, relayerSources bool) int {
	power := e.conductedStrongPowerFrom(pos, tx, face)
	type step struct {
		pos   cube.Pos
		from  cube.Face
		loss  int
		depth int
	}
	queue := []step{{pos: pos.Side(face), from: face.Opposite(), loss: 0, depth: 0}}
	seen := make(map[cube.Pos]int, 16)
	for len(queue) != 0 {
		s := queue[0]
		queue = queue[1:]
		if s.pos.OutOfBounds(tx.Range()) || s.loss >= 15 || s.depth >= 15 {
			continue
		}
		if s.pos == pos {
			continue
		}
		if loss, ok := seen[s.pos]; ok && loss <= s.loss {
			continue
		}
		seen[s.pos] = s.loss

		b, ok := tx.World().blockLoaded(s.pos)
		if !ok {
			continue
		}
		relayer, isRelayer := b.(RedstonePowerRelayer)
		// See graphPower: relayers carry recomputed power through edges, so their
		// stored RedstonePower should not count as an independent source here.
		if source, ok := b.(RedstonePowerSource); ok && (!isRelayer || relayerSources) {
			power = max(power, clampRedstonePower(e.redstonePower(source, s.pos, tx, s.from)-s.loss))
		}
		if !isRelayer {
			continue
		}
		for _, next := range e.redstoneRelayerNeighbourPositions(tx, s.pos, b) {
			to := redstoneStepFace(s.pos, next)
			if to == s.from {
				continue
			}
			loss := s.loss + max(relayer.RedstoneSignalLoss(s.pos, tx, s.from, to), 1)
			if loss <= 15 {
				queue = append(queue, step{pos: next, from: to.Opposite(), loss: loss, depth: s.depth + 1})
			}
		}
	}
	return clampRedstonePower(power)
}

func (e *redstoneEngine) redstonePower(source RedstonePowerSource, pos cube.Pos, tx *Tx, face cube.Face) int {
	if _, ok := e.evaluating[pos]; ok {
		return 0
	}
	e.evaluating[pos] = struct{}{}
	defer delete(e.evaluating, pos)
	return source.RedstonePower(pos, tx, face)
}

func (e *redstoneEngine) compileEdges(tx *Tx, nodes []redstoneNode) []redstoneEdge {
	index := make(map[cube.Pos]int, len(nodes))
	for i, node := range nodes {
		index[node.pos] = i
	}
	edges := make([]redstoneEdge, 0, len(nodes))
	for i, node := range nodes {
		b, loaded := tx.World().blockLoaded(node.pos)
		if !loaded {
			continue
		}
		relayer, ok := b.(RedstonePowerRelayer)
		if !ok {
			continue
		}
		for _, neighbour := range e.redstoneRelayerNeighbourPositions(tx, node.pos, b) {
			j, ok := index[neighbour]
			if !ok {
				continue
			}
			face := redstoneStepFace(node.pos, neighbour)
			edges = append(edges, redstoneEdge{from: i, to: j, weight: max(relayer.RedstoneSignalLoss(node.pos, tx, face.Opposite(), face), 1)})
		}
	}
	slices.SortFunc(edges, compareRedstoneEdge)
	return edges
}

func (e *redstoneEngine) redstoneRelayerNeighbourPositions(tx *Tx, pos cube.Pos, b Block) []cube.Pos {
	if neighbourer, ok := b.(RedstonePowerRelayerNeighbourer); ok {
		neighbours := slices.Clone(neighbourer.RedstoneRelayerNeighbours(pos, tx))
		slices.SortFunc(neighbours, compareBlockPos)
		return neighbours
	}
	neighbours := make([]cube.Pos, 0, len(cube.Faces()))
	e.redstoneRelayerNeighbours(tx, pos, b, func(neighbour cube.Pos) {
		neighbours = append(neighbours, neighbour)
	})
	slices.SortFunc(neighbours, compareBlockPos)
	return neighbours
}

func (e *redstoneEngine) redstoneRelayerNeighbours(tx *Tx, pos cube.Pos, _ Block, f func(cube.Pos)) {
	for _, face := range cube.Faces() {
		neighbour := pos.Side(face)
		if !neighbour.OutOfBounds(tx.Range()) {
			f(neighbour)
		}
	}
}

func redstoneStepFace(from, to cube.Pos) cube.Face {
	dx, dy, dz := to[0]-from[0], to[1]-from[1], to[2]-from[2]
	switch {
	case dx > 0:
		return cube.FaceEast
	case dx < 0:
		return cube.FaceWest
	case dz > 0:
		return cube.FaceSouth
	case dz < 0:
		return cube.FaceNorth
	case dy > 0:
		return cube.FaceUp
	case dy < 0:
		return cube.FaceDown
	default:
		return cube.FaceUp
	}
}

func redstoneGraphID(nodes []redstoneNode, edges []redstoneEdge) uint64 {
	if len(nodes) == 0 {
		return 0
	}
	h := fnv.New64a()
	var buf [32]byte
	for _, node := range nodes {
		binary.LittleEndian.PutUint64(buf[0:], uint64(node.pos[0]))
		binary.LittleEndian.PutUint64(buf[8:], uint64(node.pos[1]))
		binary.LittleEndian.PutUint64(buf[16:], uint64(node.pos[2]))
		binary.LittleEndian.PutUint64(buf[24:], uint64(boolInt(node.source)<<1|boolInt(node.sink)))
		_, _ = h.Write(buf[:])
	}
	for _, edge := range edges {
		binary.LittleEndian.PutUint64(buf[0:], uint64(edge.from))
		binary.LittleEndian.PutUint64(buf[8:], uint64(edge.to))
		binary.LittleEndian.PutUint64(buf[16:], uint64(edge.weight))
		_, _ = h.Write(buf[:24])
	}
	return h.Sum64()
}

func classifyRedstoneBlock(b Block) (source, consumer, action, relayer bool) {
	_, source = b.(RedstonePowerSource)
	_, consumer = b.(RedstonePowerConsumer)
	if !consumer {
		_, consumer = b.(RedstonePowerTransitionConsumer)
	}
	_, action = b.(RedstonePowerAction)
	_, relayer = b.(RedstonePowerRelayer)
	return
}

func isRedstoneRelevant(b Block) bool {
	source, consumer, action, relayer := classifyRedstoneBlock(b)
	return source || consumer || action || relayer
}

func compareBlockPos(a, b cube.Pos) int {
	if a[1] != b[1] {
		return a[1] - b[1]
	}
	if a[2] != b[2] {
		return a[2] - b[2]
	}
	return a[0] - b[0]
}

func compareRedstoneEdge(a, b redstoneEdge) int {
	if a.from != b.from {
		return a.from - b.from
	}
	if a.to != b.to {
		return a.to - b.to
	}
	return a.weight - b.weight
}

func clampRedstonePower(power int) int {
	if power < 0 {
		return 0
	}
	if power > 15 {
		return 15
	}
	return power
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
