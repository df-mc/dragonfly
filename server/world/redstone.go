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
	// HasChangedNeighbour reports whether ChangedNeighbour is set. A zero block position is valid, so callers must not
	// use ChangedNeighbour == cube.Pos{} as an absence check.
	HasChangedNeighbour bool
	// ChangedRedstoneRelevant reports whether ChangedNeighbour was a redstone component before or after the change.
	ChangedRedstoneRelevant bool
	// Source is the original block position that caused this redstone propagation, if known.
	Source cube.Pos
	// HasSource reports whether Source is set.
	HasSource bool
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

// RedstoneWeakBlockPowerer may be implemented by sources whose weak output can make an adjacent conductive block
// weakly powered. Weakly powered blocks activate adjacent mechanisms, but do not power adjacent redstone dust.
type RedstoneWeakBlockPowerer interface {
	RedstoneWeaklyPowersBlocks() bool
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

// RedstonePowerContextAction may be implemented by action blocks that need the proposed update metadata to distinguish
// self-caused redstone changes from external block updates.
type RedstonePowerContextAction interface {
	RedstonePowerActionUpdate(pos cube.Pos, tx *Tx, update RedstoneUpdate) bool
}

// RedstoneComparatorReadable is implemented by blocks that expose an analog signal to a comparator.
type RedstoneComparatorReadable interface {
	RedstoneComparatorOutput(pos cube.Pos, tx *Tx, face cube.Face) int
}

// RedstoneNonConductive may be implemented by solid redstone blocks that should not conduct strong power.
type RedstoneNonConductive interface {
	RedstoneNonConductive()
}

// redstoneEngine owns transient redstone graph state for a world.
type redstoneEngine struct {
	currentTick       int64
	dirty             map[cube.Pos]redstoneDirty
	power             map[cube.Pos]int
	output            map[cube.Pos]int
	evaluating        map[cube.Pos]struct{}
	suppressedSources map[cube.Pos]int
	torchBurnout      map[cube.Pos]redstoneTorchBurnout
}

// redstoneDirty records why a position needs redstone evaluation.
type redstoneDirty struct {
	changed                 cube.Pos
	hasChanged              bool
	changedRedstoneRelevant bool
	source                  cube.Pos
	hasSource               bool
	cause                   RedstoneUpdateCause
}

// redstoneTorchBurnout tracks rapid torch turn-off history and the current burned-out state.
type redstoneTorchBurnout struct {
	offTicks             []int64
	burnedOut            bool
	pendingSelfTriggered bool
}

const (
	redstoneTorchBurnoutThreshold   = 8
	redstoneTorchBurnoutWindowTicks = 60
)

// redstoneGraph is a compiled redstone network for one engine tick.
type redstoneGraph struct {
	id    uint64
	nodes []redstoneNode
	edges []redstoneEdge
}

// redstoneNode represents one redstone-relevant block in a compiled graph.
type redstoneNode struct {
	pos    cube.Pos
	source bool
	sink   bool
}

// redstoneEdge connects two graph nodes with a signal-loss weight.
type redstoneEdge struct {
	from, to int
	weight   int
}

// newRedstoneEngine creates a redstone engine initialized at tick.
func newRedstoneEngine(tick int64) *redstoneEngine {
	return &redstoneEngine{
		currentTick: tick,
		dirty:       make(map[cube.Pos]redstoneDirty),
		power:       make(map[cube.Pos]int),
		output:      make(map[cube.Pos]int),
		evaluating:  make(map[cube.Pos]struct{}),
	}
}

// invalidateAround marks pos and its direct neighbours dirty for redstone evaluation.
func (e *redstoneEngine) invalidateAround(pos, changed cube.Pos, cause RedstoneUpdateCause, r cube.Range) {
	e.invalidateAroundWith(pos, redstoneDirty{changed: changed, hasChanged: true, source: changed, hasSource: true, cause: cause}, r)
}

// invalidateAroundBlockChange marks pos and its direct neighbours dirty for a block change.
func (e *redstoneEngine) invalidateAroundBlockChange(pos cube.Pos, before, after Block, cause RedstoneUpdateCause, r cube.Range) {
	d := redstoneDirty{
		changed:                 pos,
		hasChanged:              true,
		changedRedstoneRelevant: isRedstoneRelevant(before) || isRedstoneRelevant(after),
		source:                  pos,
		hasSource:               true,
		cause:                   cause,
	}
	e.invalidateAroundWith(pos, d, r)
}

// invalidateAroundWith marks pos and its direct neighbours dirty using the same update context.
func (e *redstoneEngine) invalidateAroundWith(pos cube.Pos, d redstoneDirty, r cube.Range) {
	if e == nil || pos.OutOfBounds(r) {
		return
	}
	e.invalidate(pos, d, r)
	pos.Neighbours(func(neighbour cube.Pos) {
		e.invalidate(neighbour, d, r)
	}, r)
}

// invalidate marks a single in-range position dirty for redstone evaluation.
func (e *redstoneEngine) invalidate(pos cube.Pos, d redstoneDirty, r cube.Range) {
	if pos.OutOfBounds(r) {
		return
	}
	if existing, ok := e.dirty[pos]; ok {
		e.dirty[pos] = mergeRedstoneDirty(existing, d)
		return
	}
	e.dirty[pos] = d
}

// mergeRedstoneDirty keeps the strongest cause when multiple invalidations touch the same position before evaluation.
func mergeRedstoneDirty(a, b redstoneDirty) redstoneDirty {
	if redstoneDirtyPriority(b) >= redstoneDirtyPriority(a) {
		b.changedRedstoneRelevant = a.changedRedstoneRelevant || b.changedRedstoneRelevant
		return b
	}
	a.changedRedstoneRelevant = a.changedRedstoneRelevant || b.changedRedstoneRelevant
	return a
}

func redstoneDirtyPriority(d redstoneDirty) int {
	switch d.cause {
	case RedstoneUpdateCauseBlockUpdate:
		return 3
	case RedstoneUpdateCauseScheduledTick:
		return 2
	case RedstoneUpdateCauseCompilerRebuild:
		return 1
	default:
		return 0
	}
}

// removeChunk drops transient redstone state tied to an unloaded chunk.
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
	maps.DeleteFunc(e.torchBurnout, func(pos cube.Pos, _ redstoneTorchBurnout) bool {
		return chunkPosFromBlockPos(pos) == chunkPos
	})
}

// forget clears cached redstone input and output power for pos.
func (e *redstoneEngine) forget(pos cube.Pos) {
	if e == nil {
		return
	}
	delete(e.power, pos)
	delete(e.output, pos)
	delete(e.evaluating, pos)
}

// tick evaluates all dirty redstone positions for the current world tick.
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
	cancelledSources, checkedSources := e.updateGraphSources(tx, graph, dirty)
	previousSuppressed := e.suppressedSources
	e.suppressedSources = cancelledSources
	defer func() {
		e.suppressedSources = previousSuppressed
	}()

	powers := e.graphPower(tx, graph)
	for i, node := range graph.nodes {
		d := redstoneDirtyContext(dirty, node.pos)
		if node.sink {
			e.update(tx, node.pos, d, graph.id, powers[i])
		}
	}
	for _, node := range graph.nodes {
		if _, ok := checkedSources[node.pos]; ok {
			continue
		}
		d := redstoneDirtyContext(dirty, node.pos)
		if node.source {
			e.updateSource(tx, node.pos, d, graph.id)
		}
	}
}

// redstoneDirtyContext returns the direct dirty context for pos, or the nearest dirty context that pulled pos into
// the current graph. Graph compilation intentionally includes connected blocks that are not dirty themselves.
func redstoneDirtyContext(dirty map[cube.Pos]redstoneDirty, pos cube.Pos) redstoneDirty {
	if d, ok := dirty[pos]; ok {
		return d
	}
	var (
		bestPos  cube.Pos
		best     redstoneDirty
		bestDist int
		ok       bool
	)
	for dirtyPos, d := range dirty {
		dist := redstoneManhattanDistance(pos, dirtyPos)
		if !ok || dist < bestDist || (dist == bestDist && compareBlockPos(dirtyPos, bestPos) < 0) {
			bestPos, best, bestDist, ok = dirtyPos, d, dist, true
		}
	}
	if !ok {
		return redstoneDirty{changed: pos, hasChanged: true, source: pos, hasSource: true, cause: RedstoneUpdateCauseCompilerRebuild}
	}
	return best
}

// redstoneUpdate builds the public update payload for a dirty context.
func (d redstoneDirty) redstoneUpdate(pos cube.Pos, before Block, oldPower, newPower int, tick int64, graphID uint64) RedstoneUpdate {
	return RedstoneUpdate{
		Pos:                     pos,
		ChangedNeighbour:        d.changed,
		HasChangedNeighbour:     d.hasChanged,
		ChangedRedstoneRelevant: d.changedRedstoneRelevant,
		Source:                  d.source,
		HasSource:               d.hasSource,
		Before:                  before,
		OldPower:                oldPower,
		NewPower:                newPower,
		CurrentTick:             tick,
		NetworkID:               graphID,
		Cause:                   d.cause,
	}
}

// propagatedFrom keeps the original source but records the block state that changed during propagation.
func (d redstoneDirty) propagatedFrom(pos cube.Pos) redstoneDirty {
	d.changed = pos
	d.hasChanged = true
	d.changedRedstoneRelevant = true
	return d
}

func redstoneManhattanDistance(a, b cube.Pos) int {
	return abs(a[0]-b[0]) + abs(a[1]-b[1]) + abs(a[2]-b[2])
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// compile builds a deterministic graph around the candidate positions.
func (e *redstoneEngine) compile(tx *Tx, candidates []cube.Pos) redstoneGraph {
	nodes := make([]redstoneNode, 0, len(candidates))
	seen := make(map[cube.Pos]struct{}, len(candidates)*2)
	for _, pos := range candidates {
		nodeCount := len(nodes)
		e.compileRegion(tx, pos, seen, &nodes)
		if len(nodes) == nodeCount {
			e.compileAdjacentRedstone(tx, pos, seen, &nodes)
		}
	}
	for i := 0; i < len(nodes); i++ {
		e.compileAdjacentRedstone(tx, nodes[i].pos, seen, &nodes)
	}
	slices.SortFunc(nodes, func(a, b redstoneNode) int {
		return compareBlockPos(a.pos, b.pos)
	})
	edges := e.compileEdges(tx, nodes)
	return redstoneGraph{id: redstoneGraphID(nodes, edges), nodes: nodes, edges: edges}
}

// compileAdjacentRedstone adds redstone blocks that can interact with pos directly or through an adjacent conductor.
func (e *redstoneEngine) compileAdjacentRedstone(tx *Tx, pos cube.Pos, seen map[cube.Pos]struct{}, nodes *[]redstoneNode) {
	if b, ok := tx.World().blockLoaded(pos); ok && e.redstoneBlockMayConduct(tx, pos, b) {
		pos.Neighbours(func(neighbour cube.Pos) {
			if b, ok := tx.World().blockLoaded(neighbour); ok && isRedstoneRelevant(b) {
				e.compileRegion(tx, neighbour, seen, nodes)
			}
		}, tx.Range())
	}
	pos.Neighbours(func(neighbour cube.Pos) {
		b, ok := tx.World().blockLoaded(neighbour)
		if !ok {
			return
		}
		if isRedstoneRelevant(b) {
			e.compileRegion(tx, neighbour, seen, nodes)
		}
		if e.redstoneBlockMayConduct(tx, neighbour, b) {
			neighbour.Neighbours(func(conductedNeighbour cube.Pos) {
				if b, ok := tx.World().blockLoaded(conductedNeighbour); ok && isRedstoneRelevant(b) {
					e.compileRegion(tx, conductedNeighbour, seen, nodes)
				}
			}, tx.Range())
		}
	}, tx.Range())
}

// redstoneBlockMayConduct reports whether a block should include its adjacent receivers in graph compilation.
func (e *redstoneEngine) redstoneBlockMayConduct(tx *Tx, pos cube.Pos, b Block) bool {
	for _, face := range cube.Faces() {
		if redstoneStrongPowerConductor(pos, b, tx, face) {
			return true
		}
	}
	return false
}

// compileRegion walks a connected redstone-relevant region into nodes.
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
		for _, neighbour := range e.redstoneRelayerConnectedPositions(tx, p, b) {
			if b, ok := tx.World().blockLoaded(neighbour); ok && isRedstoneRelevant(b) {
				queue = append(queue, neighbour)
			}
		}
	}
}

// update applies a computed input power to a consumer or action block.
func (e *redstoneEngine) update(tx *Tx, pos cube.Pos, d redstoneDirty, graphID uint64, newPower int) {
	b := tx.Block(pos)
	oldPower, newPower := e.power[pos], clampRedstonePower(newPower)

	after, blockChanged := b, false
	if consumer, ok := b.(RedstonePowerTransitionConsumer); ok {
		after, blockChanged = consumer.RedstonePowerTransitionUpdate(pos, tx, oldPower, newPower)
	} else if consumer, ok := b.(RedstonePowerConsumer); ok {
		after, blockChanged = consumer.RedstonePowerUpdate(pos, tx, newPower)
	}
	action, hasAction := b.(RedstonePowerAction)
	contextAction, hasContextAction := b.(RedstonePowerContextAction)

	update := d.redstoneUpdate(pos, b, oldPower, newPower, e.currentTick, graphID)
	if blockChanged {
		update.After = after
	}
	shouldRunAction := hasContextAction || (hasAction && oldPower != newPower)
	if oldPower != newPower || blockChanged || shouldRunAction {
		if !e.redstoneUpdateAllowed(tx, update) {
			return
		}
	}

	if !blockChanged && !shouldRunAction {
		storeRedstonePower(e.power, pos, newPower)
		return
	}

	if blockChanged {
		tx.SetBlock(pos, after, &SetOpts{DisableRedstoneUpdates: true})
		e.invalidateAroundWith(pos, d.propagatedFrom(pos), tx.Range())
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
	if hasContextAction {
		acted = contextAction.RedstonePowerActionUpdate(pos, tx, update)
	} else if shouldRunAction {
		acted = action.RedstonePowerAction(pos, tx, oldPower, newPower)
	}
	if blockChanged || shouldRunAction || acted {
		storeRedstonePower(e.power, pos, newPower)
	}
}

// updateGraphSources updates non-relayer source outputs before graph propagation.
func (e *redstoneEngine) updateGraphSources(tx *Tx, graph redstoneGraph, dirty map[cube.Pos]redstoneDirty) (map[cube.Pos]int, map[cube.Pos]struct{}) {
	var cancelled map[cube.Pos]int
	var checked map[cube.Pos]struct{}
	for _, node := range graph.nodes {
		if !node.source {
			continue
		}
		b, ok := tx.World().blockLoaded(node.pos)
		if !ok {
			continue
		}
		if _, ok := b.(RedstonePowerRelayer); ok {
			continue
		}
		if checked == nil {
			checked = make(map[cube.Pos]struct{})
		}
		checked[node.pos] = struct{}{}

		d := redstoneDirtyContext(dirty, node.pos)
		if !e.updateSource(tx, node.pos, d, graph.id) {
			if cancelled == nil {
				cancelled = make(map[cube.Pos]int)
			}
			cancelled[node.pos] = e.output[node.pos]
		}
	}
	return cancelled, checked
}

// updateSource updates cached output power for a source and reports whether it was allowed.
func (e *redstoneEngine) updateSource(tx *Tx, pos cube.Pos, d redstoneDirty, graphID uint64) bool {
	b := tx.Block(pos)
	oldPower, newPower := e.output[pos], e.sourcePower(pos, tx)
	if oldPower == newPower {
		return true
	}
	update := d.redstoneUpdate(pos, b, oldPower, newPower, e.currentTick, graphID)
	if !e.redstoneUpdateAllowed(tx, update) {
		return false
	}
	storeRedstonePower(e.output, pos, newPower)
	e.invalidateAroundWith(pos, d.propagatedFrom(pos), tx.Range())
	return true
}

// storeRedstonePower stores non-zero redstone power and removes zero entries from cache maps.
func storeRedstonePower(cache map[cube.Pos]int, pos cube.Pos, power int) {
	if power == 0 {
		delete(cache, pos)
		return
	}
	cache[pos] = power
}

// directPower returns the strongest direct power reaching pos from any side.
func (e *redstoneEngine) directPower(pos cube.Pos, tx *Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, e.directPowerFrom(pos, tx, face))
	}
	return power
}

// directPowerFrom returns direct power reaching pos through face.
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

// strongPower returns the strongest strong power reaching pos from any side.
func (e *redstoneEngine) strongPower(pos cube.Pos, tx *Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, e.strongPowerFrom(pos, tx, face))
	}
	return power
}

// strongPowerFrom returns strong power reaching pos through face.
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
		if power, ok := e.suppressedSources[neighbour]; ok {
			return clampRedstonePower(power)
		}
		return clampRedstonePower(source.RedstoneStrongPower(neighbour, tx, face.Opposite()))
	}
	return 0
}

// conductedStrongPower returns strong power conducted through adjacent conductive blocks.
func (e *redstoneEngine) conductedStrongPower(pos cube.Pos, tx *Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, e.conductedStrongPowerFrom(pos, tx, face))
	}
	return power
}

// conductedStrongPowerFrom returns strong power conducted through the block on face.
func (e *redstoneEngine) conductedStrongPowerFrom(pos cube.Pos, tx *Tx, face cube.Face) int {
	conductorPos := pos.Side(face)
	if conductorPos.OutOfBounds(tx.Range()) {
		return 0
	}
	conductor, ok := tx.World().blockLoaded(conductorPos)
	if !ok || !redstoneStrongPowerConductor(conductorPos, conductor, tx, face.Opposite()) {
		return 0
	}
	power := 0
	for _, sourceFace := range cube.Faces() {
		power = max(power, e.strongPowerFrom(conductorPos, tx, sourceFace))
	}
	return power
}

// weakBlockPower returns weak power directly applied to a conductive block by sources that weak-power blocks.
func (e *redstoneEngine) weakBlockPower(pos cube.Pos, tx *Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, e.weakBlockPowerFrom(pos, tx, face))
	}
	return power
}

// weakBlockPowerFrom returns weak power applied to pos from face by a source that weak-powers conductive blocks.
func (e *redstoneEngine) weakBlockPowerFrom(pos cube.Pos, tx *Tx, face cube.Face) int {
	sourcePos := pos.Side(face)
	if sourcePos.OutOfBounds(tx.Range()) {
		return 0
	}
	b, ok := tx.World().blockLoaded(sourcePos)
	if !ok {
		return 0
	}
	if source, ok := b.(RedstonePowerSource); ok && e.redstoneWeaklyPowersBlocks(b) {
		return clampRedstonePower(e.redstonePower(source, sourcePos, tx, face.Opposite()))
	}
	return 0
}

// conductedWeakPower returns weak power conducted through adjacent conductive blocks for mechanism activation.
func (e *redstoneEngine) conductedWeakPower(pos cube.Pos, tx *Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, e.conductedWeakPowerFrom(pos, tx, face))
	}
	return power
}

// conductedWeakPowerFrom returns weak power conducted through the block on face for mechanism activation. It excludes
// strong power and only counts sources that explicitly weak-power conductive blocks, such as redstone dust.
func (e *redstoneEngine) conductedWeakPowerFrom(pos cube.Pos, tx *Tx, face cube.Face) int {
	conductorPos := pos.Side(face)
	if conductorPos.OutOfBounds(tx.Range()) {
		return 0
	}
	conductor, ok := tx.World().blockLoaded(conductorPos)
	if !ok || !redstoneStrongPowerConductor(conductorPos, conductor, tx, face.Opposite()) {
		return 0
	}
	return e.weakBlockPower(conductorPos, tx)
}

// conductedActivationPower returns power that can activate a non-relayer mechanism behind adjacent conductive blocks.
func (e *redstoneEngine) conductedActivationPower(pos cube.Pos, tx *Tx) int {
	return max(e.conductedStrongPower(pos, tx), e.conductedWeakPower(pos, tx))
}

// conductedActivationPowerFrom returns mechanism activation power conducted through the block on face.
func (e *redstoneEngine) conductedActivationPowerFrom(pos cube.Pos, tx *Tx, face cube.Face) int {
	return max(e.conductedStrongPowerFrom(pos, tx, face), e.conductedWeakPowerFrom(pos, tx, face))
}

// conductivePowerTo returns power held by pos as a conductive block, excluding direct component activation.
func (e *redstoneEngine) conductivePowerTo(pos cube.Pos, tx *Tx) int {
	b, ok := tx.World().blockLoaded(pos)
	if !ok || !redstoneFullPowerConductor(pos, b, tx) {
		return 0
	}
	return max(e.strongPower(pos, tx), e.weakBlockPower(pos, tx))
}

// acceptsDirectSourcePower reports whether direct source output should activate the block at pos.
func (e *redstoneEngine) acceptsDirectSourcePower(pos cube.Pos, tx *Tx) bool {
	b, ok := tx.World().blockLoaded(pos)
	if !ok {
		return true
	}
	if isRedstoneRelevant(b) {
		return true
	}
	return !redstoneFullPowerConductor(pos, b, tx)
}

// acceptsWeakConductedPower reports whether the block at pos may be activated by a weakly powered conductor.
func (e *redstoneEngine) acceptsWeakConductedPower(pos cube.Pos, tx *Tx) bool {
	b, ok := tx.World().blockLoaded(pos)
	if !ok {
		return true
	}
	_, relayer := b.(RedstonePowerRelayer)
	return !relayer
}

// redstoneWeaklyPowersBlocks reports whether b's weak source output can weak-power adjacent conductive blocks.
func (e *redstoneEngine) redstoneWeaklyPowersBlocks(b Block) bool {
	weakBlockPowerer, ok := b.(RedstoneWeakBlockPowerer)
	return ok && weakBlockPowerer.RedstoneWeaklyPowersBlocks()
}

// sourcePower returns the strongest output emitted by a source block.
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

// graphPower computes propagated graph power for every graph node.
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
		if node.sink && relayers[i] == nil {
			push(i, e.conductedActivationPower(node.pos, tx))
			continue
		}
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

// powerTo returns the strongest redstone power currently reaching pos.
func (e *redstoneEngine) powerTo(pos cube.Pos, tx *Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		power = max(power, e.powerFrom(pos, tx, face, false))
	}
	if e.acceptsWeakConductedPower(pos, tx) {
		power = max(power, e.conductedActivationPower(pos, tx))
	} else {
		power = max(power, e.conductedStrongPower(pos, tx))
	}
	return clampRedstonePower(power)
}

// powerFrom returns redstone power reaching pos through face.
func (e *redstoneEngine) powerFrom(pos cube.Pos, tx *Tx, face cube.Face, relayerSources bool) int {
	power := e.conductedStrongPowerFrom(pos, tx, face)
	if e.acceptsWeakConductedPower(pos, tx) {
		power = e.conductedActivationPowerFrom(pos, tx, face)
	}
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
		if source, ok := b.(RedstonePowerSource); ok && (!isRelayer || relayerSources) && e.acceptsDirectSourcePower(pos, tx) {
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
			nextBlock, ok := tx.World().blockLoaded(next)
			if !ok {
				continue
			}
			loss := s.loss
			if _, nextRelayer := nextBlock.(RedstonePowerRelayer); nextRelayer {
				loss += max(relayer.RedstoneSignalLoss(s.pos, tx, s.from, to), 1)
			}
			if loss <= 15 {
				queue = append(queue, step{pos: next, from: to.Opposite(), loss: loss, depth: s.depth + 1})
			}
		}
	}
	return clampRedstonePower(power)
}

// torchBurnoutStatus returns whether the torch at pos is burned out and whether it can recover at currentTick.
func (e *redstoneEngine) torchBurnoutStatus(pos cube.Pos, currentTick int64) (burnedOut, recoverable bool) {
	data, ok := e.pruneTorchBurnout(pos, currentTick)
	if !ok {
		return false, true
	}
	if !data.burnedOut {
		return false, true
	}
	return true, len(data.offTicks) < redstoneTorchBurnoutThreshold
}

// recordTorchTurnOff records a torch being forced off and reports whether that torch should burn out.
func (e *redstoneEngine) recordTorchTurnOff(pos cube.Pos, currentTick int64) bool {
	if e.torchBurnout == nil {
		e.torchBurnout = make(map[cube.Pos]redstoneTorchBurnout)
	}
	data, _ := e.pruneTorchBurnout(pos, currentTick)
	data.offTicks = append(data.offTicks, currentTick)
	if len(data.offTicks) >= redstoneTorchBurnoutThreshold {
		data.burnedOut = true
	}
	e.torchBurnout[pos] = data
	return data.burnedOut
}

// clearTorchBurnout removes transient burnout state for a redstone torch.
func (e *redstoneEngine) clearTorchBurnout(pos cube.Pos) {
	if e == nil {
		return
	}
	delete(e.torchBurnout, pos)
}

// markTorchSelfTriggered records that the next torch turn-off at pos was caused by that torch's own output loop.
func (e *redstoneEngine) markTorchSelfTriggered(pos cube.Pos) {
	if e == nil {
		return
	}
	if e.torchBurnout == nil {
		e.torchBurnout = make(map[cube.Pos]redstoneTorchBurnout)
	}
	data := e.torchBurnout[pos]
	data.pendingSelfTriggered = true
	e.torchBurnout[pos] = data
}

// consumeTorchSelfTriggered reports and clears whether the next torch turn-off at pos was self-triggered.
func (e *redstoneEngine) consumeTorchSelfTriggered(pos cube.Pos) bool {
	if e == nil || e.torchBurnout == nil {
		return false
	}
	data, ok := e.torchBurnout[pos]
	if !ok {
		return false
	}
	selfTriggered := data.pendingSelfTriggered
	data.pendingSelfTriggered = false
	if len(data.offTicks) == 0 && !data.burnedOut {
		delete(e.torchBurnout, pos)
	} else {
		e.torchBurnout[pos] = data
	}
	return selfTriggered
}

// pruneTorchBurnout removes expired turn-off entries and returns the remaining burnout data.
func (e *redstoneEngine) pruneTorchBurnout(pos cube.Pos, currentTick int64) (redstoneTorchBurnout, bool) {
	if e == nil || e.torchBurnout == nil {
		return redstoneTorchBurnout{}, false
	}
	data, ok := e.torchBurnout[pos]
	if !ok {
		return redstoneTorchBurnout{}, false
	}
	data.offTicks = slices.DeleteFunc(data.offTicks, func(tick int64) bool {
		return currentTick-tick >= redstoneTorchBurnoutWindowTicks
	})
	if len(data.offTicks) == 0 && !data.burnedOut && !data.pendingSelfTriggered {
		delete(e.torchBurnout, pos)
		return redstoneTorchBurnout{}, false
	}
	e.torchBurnout[pos] = data
	return data, true
}

// redstonePower reads source power while guarding against recursive source evaluation.
func (e *redstoneEngine) redstonePower(source RedstonePowerSource, pos cube.Pos, tx *Tx, face cube.Face) int {
	if power, ok := e.suppressedSources[pos]; ok {
		return clampRedstonePower(power)
	}
	if _, ok := e.evaluating[pos]; ok {
		return 0
	}
	e.evaluating[pos] = struct{}{}
	defer delete(e.evaluating, pos)
	return source.RedstonePower(pos, tx, face)
}

// redstoneUpdateAllowed dispatches redstone callbacks and reports whether the update was cancelled.
func (e *redstoneEngine) redstoneUpdateAllowed(tx *Tx, update RedstoneUpdate) bool {
	ctx := event.C(tx)
	tx.World().Handler().HandleRedstoneUpdate(ctx, update)
	return !ctx.Cancelled()
}

// redstoneLightDiffuser is the local subset needed to exclude transparent conductors.
type redstoneLightDiffuser interface {
	LightDiffusionLevel() uint8
}

// redstoneStrongPowerConductor reports whether b can conduct strong power through face.
func redstoneStrongPowerConductor(pos cube.Pos, b Block, tx *Tx, face cube.Face) bool {
	if !b.Model().FaceSolid(pos, face, tx) {
		return false
	}
	if _, ok := b.(RedstoneNonConductive); ok {
		return false
	}
	if diffuser, ok := b.(redstoneLightDiffuser); ok && diffuser.LightDiffusionLevel() == 0 {
		return false
	}
	return true
}

// redstoneFullPowerConductor reports whether b is a full solid redstone conductor.
func redstoneFullPowerConductor(pos cube.Pos, b Block, tx *Tx) bool {
	for _, face := range cube.Faces() {
		if !redstoneStrongPowerConductor(pos, b, tx, face) {
			return false
		}
	}
	return true
}

// compileEdges builds deterministic weighted edges between relayer-connected nodes.
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
			weight := 0
			if neighbourBlock, ok := tx.World().blockLoaded(neighbour); ok {
				if _, neighbourRelayer := neighbourBlock.(RedstonePowerRelayer); neighbourRelayer {
					weight = max(relayer.RedstoneSignalLoss(node.pos, tx, face.Opposite(), face), 1)
				}
			}
			edges = append(edges, redstoneEdge{from: i, to: j, weight: weight})
		}
	}
	slices.SortFunc(edges, compareRedstoneEdge)
	return edges
}

// redstoneRelayerNeighbourPositions returns sorted relayer neighbours for b at pos.
func (e *redstoneEngine) redstoneRelayerNeighbourPositions(tx *Tx, pos cube.Pos, b Block) []cube.Pos {
	if neighbourer, ok := b.(RedstonePowerRelayerNeighbourer); ok {
		neighbours := slices.Clone(neighbourer.RedstoneRelayerNeighbours(pos, tx))
		slices.SortFunc(neighbours, compareBlockPos)
		return neighbours
	}
	neighbours := make([]cube.Pos, 0, len(cube.Faces()))
	e.redstoneRelayerNeighbours(tx, pos, func(neighbour cube.Pos) {
		neighbours = append(neighbours, neighbour)
	})
	slices.SortFunc(neighbours, compareBlockPos)
	return neighbours
}

// redstoneRelayerConnectedPositions returns relayers connected to pos in either direction. Graph membership must be
// weakly connected so a one-way relayer, such as dust climbing glowstone, still compiles with its lower input path.
func (e *redstoneEngine) redstoneRelayerConnectedPositions(tx *Tx, pos cube.Pos, b Block) []cube.Pos {
	neighbours := e.redstoneRelayerNeighbourPositions(tx, pos, b)
	seen := make(map[cube.Pos]struct{}, len(neighbours)+8)
	for _, neighbour := range neighbours {
		seen[neighbour] = struct{}{}
	}
	for _, candidate := range redstoneRelayerIncomingCandidates(pos, tx.Range()) {
		if _, ok := seen[candidate]; ok {
			continue
		}
		candidateBlock, ok := tx.World().blockLoaded(candidate)
		if !ok {
			continue
		}
		if _, ok := candidateBlock.(RedstonePowerRelayer); !ok {
			continue
		}
		if slices.Contains(e.redstoneRelayerNeighbourPositions(tx, candidate, candidateBlock), pos) {
			neighbours = append(neighbours, candidate)
			seen[candidate] = struct{}{}
		}
	}
	slices.SortFunc(neighbours, compareBlockPos)
	return neighbours
}

// redstoneRelayerIncomingCandidates returns nearby positions that can point at pos with the built-in relayer geometry.
func redstoneRelayerIncomingCandidates(pos cube.Pos, r cube.Range) []cube.Pos {
	candidates := make([]cube.Pos, 0, 26)
	for x := pos[0] - 1; x <= pos[0]+1; x++ {
		for y := pos[1] - 1; y <= pos[1]+1; y++ {
			for z := pos[2] - 1; z <= pos[2]+1; z++ {
				candidate := cube.Pos{x, y, z}
				if candidate == pos || candidate.OutOfBounds(r) {
					continue
				}
				candidates = append(candidates, candidate)
			}
		}
	}
	return candidates
}

// redstoneRelayerNeighbours visits the six default adjacent relayer neighbours.
func (e *redstoneEngine) redstoneRelayerNeighbours(tx *Tx, pos cube.Pos, f func(cube.Pos)) {
	for _, face := range cube.Faces() {
		neighbour := pos.Side(face)
		if !neighbour.OutOfBounds(tx.Range()) {
			f(neighbour)
		}
	}
}

// redstoneStepFace returns the dominant face direction from one relayer position to another.
func redstoneStepFace(from, to cube.Pos) cube.Face {
	dx, dy, dz := to[0]-from[0], to[1]-from[1], to[2]-from[2]
	switch {
	case dy > 0:
		return cube.FaceUp
	case dy < 0:
		return cube.FaceDown
	case dx > 0:
		return cube.FaceEast
	case dx < 0:
		return cube.FaceWest
	case dz > 0:
		return cube.FaceSouth
	case dz < 0:
		return cube.FaceNorth
	default:
		return cube.FaceUp
	}
}

// redstoneGraphID returns a stable hash for a compiled graph topology.
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

// classifyRedstoneBlock reports the redstone capabilities implemented by b.
func classifyRedstoneBlock(b Block) (source, consumer, action, relayer bool) {
	_, source = b.(RedstonePowerSource)
	_, consumer = b.(RedstonePowerConsumer)
	if !consumer {
		_, consumer = b.(RedstonePowerTransitionConsumer)
	}
	_, action = b.(RedstonePowerAction)
	if !action {
		_, action = b.(RedstonePowerContextAction)
	}
	_, relayer = b.(RedstonePowerRelayer)
	return
}

// isRedstoneRelevant reports whether b should be included in redstone graph compilation.
func isRedstoneRelevant(b Block) bool {
	source, consumer, action, relayer := classifyRedstoneBlock(b)
	return source || consumer || action || relayer
}

// compareBlockPos orders positions deterministically by Y, Z, then X.
func compareBlockPos(a, b cube.Pos) int {
	if a[1] != b[1] {
		return a[1] - b[1]
	}
	if a[2] != b[2] {
		return a[2] - b[2]
	}
	return a[0] - b[0]
}

// compareRedstoneEdge orders edges deterministically by endpoints and weight.
func compareRedstoneEdge(a, b redstoneEdge) int {
	if a.from != b.from {
		return a.from - b.from
	}
	if a.to != b.to {
		return a.to - b.to
	}
	return a.weight - b.weight
}

// clampRedstonePower clamps power to the vanilla 0-15 redstone range.
func clampRedstonePower(power int) int {
	if power < 0 {
		return 0
	}
	if power > 15 {
		return 15
	}
	return power
}

// boolInt converts b to 1 for true and 0 for false.
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
