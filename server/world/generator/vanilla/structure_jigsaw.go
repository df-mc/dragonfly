package vanilla

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

type structureResolver struct {
	worldgen  *gen.WorldgenRegistry
	templates *gen.StructureTemplateRegistry

	mu    sync.RWMutex
	pools map[string]resolvedStructurePool
}

type resolvedStructurePool struct {
	fallback string
	entries  []resolvedPoolElement
	weighted []int
}

type resolvedPoolElement struct {
	elementType string
	projection  string
	placements  []structureTemplatePlacement
	features    []structureFeaturePlacement
	jigsaws     []structureJigsaw
	size        [3]int
}

type structureTemplatePlacement struct {
	templateName string
	ignoreAir    bool
	processors   []structureProcessor
}

type structureFeaturePlacement struct {
	featureName string
}

type structureJigsaw struct {
	pos               [3]int
	front             structureDirection
	top               structureDirection
	joint             string
	name              string
	pool              string
	target            string
	placementPriority int
	selectionPriority int
}

type placedStructureJigsaw struct {
	pos               cube.Pos
	localY            int
	front             structureDirection
	top               structureDirection
	joint             string
	name              string
	pool              string
	target            string
	placementPriority int
	selectionPriority int
}

type plannedStructurePiece struct {
	element              resolvedPoolElement
	origin               cube.Pos
	rotation             structureRotation
	mirror               structureMirror
	pivot                cube.Pos
	useTemplateTransform bool
	bounds               structureBox
	manualBlocks         []plannedStructureBlock
	genTag               int
	rootPiece            bool
}

type structureRotation uint8

const (
	structureRotationNone structureRotation = iota
	structureRotationClockwise90
	structureRotationClockwise180
	structureRotationCounterclockwise90
)

type structureDirection uint8

const (
	structureDown structureDirection = iota
	structureUp
	structureNorth
	structureSouth
	structureWest
	structureEast
)

type structureBox struct {
	minX int
	minY int
	minZ int
	maxX int
	maxY int
	maxZ int
}

type listPoolElementDef struct {
	Elements   []gen.TemplatePoolElementDef `json:"elements"`
	Projection string                       `json:"projection"`
}

type featurePoolElementDef struct {
	Feature    string `json:"feature"`
	Projection string `json:"projection"`
}

type pendingStructurePiece struct {
	piece    plannedStructurePiece
	depth    int
	priority int
}

func newStructureResolver(worldgen *gen.WorldgenRegistry, templates *gen.StructureTemplateRegistry) *structureResolver {
	return &structureResolver{
		worldgen:  worldgen,
		templates: templates,
		pools:     make(map[string]resolvedStructurePool),
	}
}

func (r *structureResolver) prewarmJigsawCandidates(planners []structurePlanner) {
	for _, planner := range planners {
		for _, candidate := range planner.candidates {
			if candidate.structureType != "jigsaw" {
				continue
			}
			r.prewarmPoolGraph(candidate.jigsaw.StartPool, candidate.jigsaw.PoolAliases)
		}
	}
}

func (r *structureResolver) prewarmPoolGraph(startPool string, aliases []gen.PoolAliasDef) {
	queue := []string{normalizeIdentifierName(startPool)}
	queue = append(queue, collectPoolAliasTargets(aliases)...)
	visited := make(map[string]struct{}, len(queue))

	for len(queue) > 0 {
		name := normalizeIdentifierName(queue[len(queue)-1])
		queue = queue[:len(queue)-1]
		if name == "" || name == "empty" {
			continue
		}
		if _, ok := visited[name]; ok {
			continue
		}
		visited[name] = struct{}{}

		pool, err := r.Pool(name)
		if err != nil {
			continue
		}
		if fallback := normalizeIdentifierName(pool.fallback); fallback != "" && fallback != "empty" {
			queue = append(queue, fallback)
		}
		for _, entry := range pool.entries {
			for _, jigsaw := range entry.jigsaws {
				if poolName := normalizeIdentifierName(jigsaw.pool); poolName != "" && poolName != "empty" {
					queue = append(queue, poolName)
				}
			}
		}
	}
}

func collectPoolAliasTargets(defs []gen.PoolAliasDef) []string {
	if len(defs) == 0 {
		return nil
	}
	var out []string
	var walk func([]gen.PoolAliasDef)
	walk = func(defs []gen.PoolAliasDef) {
		for _, def := range defs {
			switch def.Type {
			case "direct":
				var raw directPoolAliasDef
				if err := json.Unmarshal(def.Raw, &raw); err == nil {
					if target := normalizeIdentifierName(raw.Target); target != "" {
						out = append(out, target)
					}
				}
			case "random":
				var raw randomPoolAliasDef
				if err := json.Unmarshal(def.Raw, &raw); err == nil {
					for _, target := range raw.Targets {
						if name := normalizeIdentifierName(target.Data); name != "" {
							out = append(out, name)
						}
					}
				}
			case "random_group":
				var raw randomGroupPoolAliasDef
				if err := json.Unmarshal(def.Raw, &raw); err == nil {
					for _, group := range raw.Groups {
						walk(group.Data)
					}
				}
			}
		}
	}
	walk(defs)
	return out
}

func (r *structureResolver) Pool(name string) (resolvedStructurePool, error) {
	key := normalizeIdentifierName(name)

	r.mu.RLock()
	if pool, ok := r.pools[key]; ok {
		r.mu.RUnlock()
		return pool, nil
	}
	r.mu.RUnlock()

	def, err := r.worldgen.TemplatePool(key)
	if err != nil {
		return resolvedStructurePool{}, err
	}

	pool := resolvedStructurePool{
		fallback: normalizeIdentifierName(def.Fallback),
		entries:  make([]resolvedPoolElement, 0, len(def.Elements)),
	}
	for _, entry := range def.Elements {
		element, err := r.resolvePoolElement(entry.Element)
		if err != nil {
			continue
		}
		if entry.Weight <= 0 {
			continue
		}
		index := len(pool.entries)
		pool.entries = append(pool.entries, element)
		for i := 0; i < entry.Weight; i++ {
			pool.weighted = append(pool.weighted, index)
		}
	}

	r.mu.Lock()
	r.pools[key] = pool
	r.mu.Unlock()
	return pool, nil
}

func (r *structureResolver) resolvePoolElement(def gen.TemplatePoolElementDef) (resolvedPoolElement, error) {
	switch def.ElementType {
	case "legacy_single_pool_element", "single_pool_element":
		single, err := def.Single()
		if err != nil {
			return resolvedPoolElement{}, err
		}
		template, err := r.templates.Template(single.Location)
		if err != nil {
			return resolvedPoolElement{}, err
		}
		return resolvedPoolElement{
			elementType: def.ElementType,
			projection:  normalizeIdentifierName(single.Projection),
			placements: []structureTemplatePlacement{{
				templateName: single.Location,
				ignoreAir:    def.ElementType == "legacy_single_pool_element",
				processors:   compileStructureProcessors(r.worldgen, single.Processors),
			}},
			jigsaws: extractTemplateJigsaws(template),
			size:    template.Size,
		}, nil
	case "list_pool_element":
		var raw listPoolElementDef
		if err := json.Unmarshal(def.Raw, &raw); err != nil {
			return resolvedPoolElement{}, err
		}
		out := resolvedPoolElement{
			elementType: def.ElementType,
			projection:  normalizeIdentifierName(raw.Projection),
		}
		for i, inner := range raw.Elements {
			resolved, err := r.resolvePoolElement(inner)
			if err != nil {
				continue
			}
			out.placements = append(out.placements, resolved.placements...)
			out.features = append(out.features, resolved.features...)
			out.size = maxStructureSize(out.size, resolved.size)
			if i == 0 {
				out.jigsaws = append(out.jigsaws, resolved.jigsaws...)
			}
		}
		return out, nil
	case "feature_pool_element":
		var raw featurePoolElementDef
		if err := json.Unmarshal(def.Raw, &raw); err != nil {
			return resolvedPoolElement{}, err
		}
		return resolvedPoolElement{
			elementType: def.ElementType,
			projection:  normalizeIdentifierName(raw.Projection),
			features: []structureFeaturePlacement{{
				featureName: normalizeIdentifierName(raw.Feature),
			}},
			size: [3]int{1, 1, 1},
			jigsaws: []structureJigsaw{{
				pos:    [3]int{0, 0, 0},
				front:  structureDown,
				top:    structureSouth,
				joint:  "rollable",
				name:   "bottom",
				pool:   "empty",
				target: "empty",
			}},
		}, nil
	case "empty_pool_element":
		return resolvedPoolElement{elementType: def.ElementType}, nil
	default:
		return resolvedPoolElement{}, fmt.Errorf("unsupported pool element type %q", def.ElementType)
	}
}

func maxStructureSize(a, b [3]int) [3]int {
	if b[0] > a[0] {
		a[0] = b[0]
	}
	if b[1] > a[1] {
		a[1] = b[1]
	}
	if b[2] > a[2] {
		a[2] = b[2]
	}
	return a
}

func extractTemplateJigsaws(template gen.StructureTemplate) []structureJigsaw {
	jigsaws := make([]structureJigsaw, 0, 8)
	for _, block := range template.Blocks {
		if block.State < 0 || block.State >= len(template.Palette) {
			continue
		}
		state := template.Palette[block.State]
		if state.Name != "minecraft:jigsaw" {
			continue
		}
		front, top := parseJigsawOrientation(state.Properties)
		joint := parseJigsawJoint(block.NBT, front)
		jigsaws = append(jigsaws, structureJigsaw{
			pos:               block.Pos,
			front:             front,
			top:               top,
			joint:             joint,
			name:              normalizeIdentifierName(anyString(block.NBT["name"])),
			pool:              normalizeIdentifierName(anyString(block.NBT["pool"])),
			target:            normalizeIdentifierName(anyString(block.NBT["target"])),
			placementPriority: anyInt(block.NBT["placement_priority"]),
			selectionPriority: anyInt(block.NBT["selection_priority"]),
		})
	}
	return jigsaws
}

func parseJigsawOrientation(properties map[string]any) (structureDirection, structureDirection) {
	orientation := "north_up"
	if properties != nil {
		if value, ok := properties["orientation"]; ok {
			if s := strings.ToLower(anyString(value)); s != "" {
				orientation = s
			}
		}
	}
	parts := strings.SplitN(orientation, "_", 2)
	if len(parts) != 2 {
		return structureNorth, structureUp
	}
	return parseStructureDirection(parts[0]), parseStructureDirection(parts[1])
}

func parseJigsawJoint(nbt map[string]any, front structureDirection) string {
	if joint := strings.ToLower(anyString(nbt["joint"])); joint != "" {
		return joint
	}
	if front.isHorizontal() {
		return "aligned"
	}
	return "rollable"
}

func parseStructureDirection(value string) structureDirection {
	switch strings.ToLower(value) {
	case "down":
		return structureDown
	case "up":
		return structureUp
	case "south":
		return structureSouth
	case "west":
		return structureWest
	case "east":
		return structureEast
	default:
		return structureNorth
	}
}

func anyString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func anyInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func normalizeIdentifierName(name string) string {
	if strings.HasPrefix(name, "minecraft:") {
		return name[len("minecraft:"):]
	}
	return name
}

func randomStructureRotation(rng *gen.Xoroshiro128) structureRotation {
	return structureRotation(rng.NextInt(4))
}

func fillShuffledStructureRotations(dst *[4]structureRotation, rng *gen.Xoroshiro128) {
	*dst = [4]structureRotation{
		structureRotationNone,
		structureRotationClockwise90,
		structureRotationClockwise180,
		structureRotationCounterclockwise90,
	}
	for i := len(dst) - 1; i > 0; i-- {
		j := int(rng.NextInt(uint32(i + 1)))
		dst[i], dst[j] = dst[j], dst[i]
	}
}

func rotateStructurePos(size, pos [3]int, rotation structureRotation) [3]int {
	switch rotation {
	case structureRotationClockwise90:
		return [3]int{size[2] - 1 - pos[2], pos[1], pos[0]}
	case structureRotationClockwise180:
		return [3]int{size[0] - 1 - pos[0], pos[1], size[2] - 1 - pos[2]}
	case structureRotationCounterclockwise90:
		return [3]int{pos[2], pos[1], size[0] - 1 - pos[0]}
	default:
		return pos
	}
}

func rotateStructureDirection(direction structureDirection, rotation structureRotation) structureDirection {
	switch direction {
	case structureNorth:
		switch rotation {
		case structureRotationClockwise90:
			return structureEast
		case structureRotationClockwise180:
			return structureSouth
		case structureRotationCounterclockwise90:
			return structureWest
		}
	case structureSouth:
		switch rotation {
		case structureRotationClockwise90:
			return structureWest
		case structureRotationClockwise180:
			return structureNorth
		case structureRotationCounterclockwise90:
			return structureEast
		}
	case structureWest:
		switch rotation {
		case structureRotationClockwise90:
			return structureNorth
		case structureRotationClockwise180:
			return structureEast
		case structureRotationCounterclockwise90:
			return structureSouth
		}
	case structureEast:
		switch rotation {
		case structureRotationClockwise90:
			return structureSouth
		case structureRotationClockwise180:
			return structureWest
		case structureRotationCounterclockwise90:
			return structureNorth
		}
	}
	return direction
}

func (d structureDirection) isHorizontal() bool {
	return d >= structureNorth
}

func (d structureDirection) stepX() int {
	switch d {
	case structureWest:
		return -1
	case structureEast:
		return 1
	default:
		return 0
	}
}

func (d structureDirection) stepY() int {
	switch d {
	case structureDown:
		return -1
	case structureUp:
		return 1
	default:
		return 0
	}
}

func (d structureDirection) stepZ() int {
	switch d {
	case structureNorth:
		return -1
	case structureSouth:
		return 1
	default:
		return 0
	}
}

func (d structureDirection) opposite() structureDirection {
	switch d {
	case structureDown:
		return structureUp
	case structureUp:
		return structureDown
	case structureNorth:
		return structureSouth
	case structureSouth:
		return structureNorth
	case structureWest:
		return structureEast
	default:
		return structureWest
	}
}

func rotatedStructureSize(size [3]int, rotation structureRotation) [3]int {
	switch rotation {
	case structureRotationClockwise90, structureRotationCounterclockwise90:
		return [3]int{size[2], size[1], size[0]}
	default:
		return size
	}
}

func (element resolvedPoolElement) worldBox(origin cube.Pos, rotation structureRotation) structureBox {
	if element.size[0] <= 0 || element.size[1] <= 0 || element.size[2] <= 0 {
		return emptyStructureBox()
	}
	size := rotatedStructureSize(element.size, rotation)
	return structureBox{
		minX: origin[0],
		minY: origin[1],
		minZ: origin[2],
		maxX: origin[0] + size[0] - 1,
		maxY: origin[1] + size[1] - 1,
		maxZ: origin[2] + size[2] - 1,
	}
}

func emptyStructureBox() structureBox {
	return structureBox{minX: 1, minY: 1, minZ: 1, maxX: 0, maxY: 0, maxZ: 0}
}

func (b structureBox) empty() bool {
	return b.maxX < b.minX || b.maxY < b.minY || b.maxZ < b.minZ
}

func (b structureBox) intersectsChunk(chunkX, chunkZ, minY, maxY int) bool {
	if b.empty() {
		return false
	}
	minBlockX := chunkX * 16
	maxBlockX := minBlockX + 15
	minBlockZ := chunkZ * 16
	maxBlockZ := minBlockZ + 15
	return !(b.maxX < minBlockX || b.minX > maxBlockX || b.maxZ < minBlockZ || b.minZ > maxBlockZ || b.maxY < minY || b.minY > maxY)
}

func (b structureBox) intersects(other structureBox) bool {
	if b.empty() || other.empty() {
		return false
	}
	return !(b.maxX < other.minX || b.minX > other.maxX || b.maxY < other.minY || b.minY > other.maxY || b.maxZ < other.minZ || b.minZ > other.maxZ)
}

func unionStructureBoxes(a, b structureBox) structureBox {
	if a.empty() {
		return b
	}
	if b.empty() {
		return a
	}
	return structureBox{
		minX: min(a.minX, b.minX),
		minY: min(a.minY, b.minY),
		minZ: min(a.minZ, b.minZ),
		maxX: max(a.maxX, b.maxX),
		maxY: max(a.maxY, b.maxY),
		maxZ: max(a.maxZ, b.maxZ),
	}
}

func (b structureBox) originAndSize() (cube.Pos, [3]int) {
	if b.empty() {
		return cube.Pos{}, [3]int{}
	}
	return cube.Pos{b.minX, b.minY, b.minZ}, [3]int{b.maxX - b.minX + 1, b.maxY - b.minY + 1, b.maxZ - b.minZ + 1}
}

func shuffleWithRNG[T any](values []T, rng *gen.Xoroshiro128) {
	for i := len(values) - 1; i > 0; i-- {
		j := int(rng.NextInt(uint32(i + 1)))
		values[i], values[j] = values[j], values[i]
	}
}

func (element resolvedPoolElement) appendShuffledJigsaws(dst []placedStructureJigsaw, origin cube.Pos, rotation structureRotation, rng *gen.Xoroshiro128) []placedStructureJigsaw {
	start := len(dst)
	dst = slices.Grow(dst, len(element.jigsaws))
	dst = dst[:start+len(element.jigsaws)]
	for i, jigsaw := range element.jigsaws {
		rotated := rotateStructurePos(element.size, jigsaw.pos, rotation)
		dst[start+i] = placedStructureJigsaw{
			pos: cube.Pos{
				origin[0] + rotated[0],
				origin[1] + rotated[1],
				origin[2] + rotated[2],
			},
			localY:            rotated[1],
			front:             rotateStructureDirection(jigsaw.front, rotation),
			top:               rotateStructureDirection(jigsaw.top, rotation),
			joint:             jigsaw.joint,
			name:              jigsaw.name,
			pool:              jigsaw.pool,
			target:            jigsaw.target,
			placementPriority: jigsaw.placementPriority,
			selectionPriority: jigsaw.selectionPriority,
		}
	}
	ordered := dst[start:]
	shuffleWithRNG(ordered, rng)
	stableSortJigsawsBySelectionPriority(ordered)
	return dst
}

func stableSortJigsawsBySelectionPriority(jigsaws []placedStructureJigsaw) {
	for i := 1; i < len(jigsaws); i++ {
		current := jigsaws[i]
		j := i
		for j > 0 && jigsaws[j-1].selectionPriority < current.selectionPriority {
			jigsaws[j] = jigsaws[j-1]
			j--
		}
		jigsaws[j] = current
	}
}

func appendShuffledPoolElements(dst []resolvedPoolElement, pool resolvedStructurePool, rng *gen.Xoroshiro128) []resolvedPoolElement {
	start := len(dst)
	dst = slices.Grow(dst, len(pool.weighted))
	dst = dst[:start+len(pool.weighted)]
	for i, index := range pool.weighted {
		if index >= 0 && index < len(pool.entries) {
			dst[start+i] = pool.entries[index]
		}
	}
	shuffleWithRNG(dst[start:], rng)
	return dst
}

func canAttachJigsaws(source, target placedStructureJigsaw) bool {
	return source.front == target.front.opposite() &&
		(source.joint == "rollable" || source.top == target.top) &&
		source.target == target.name
}

func insertPendingPiece(queue []pendingStructurePiece, pending pendingStructurePiece) []pendingStructurePiece {
	index := len(queue)
	for i, existing := range queue {
		if pending.priority > existing.priority {
			index = i
			break
		}
	}
	queue = append(queue, pendingStructurePiece{})
	copy(queue[index+1:], queue[index:])
	queue[index] = pending
	return queue
}

type structurePoolAliasMapping map[string]string

type directPoolAliasDef struct {
	Alias  string `json:"alias"`
	Target string `json:"target"`
}

type weightedTargetDef struct {
	Data   string `json:"data"`
	Weight int    `json:"weight"`
}

type randomPoolAliasDef struct {
	Alias   string              `json:"alias"`
	Targets []weightedTargetDef `json:"targets"`
}

type weightedAliasGroupDef struct {
	Data   []gen.PoolAliasDef `json:"data"`
	Weight int                `json:"weight"`
}

type randomGroupPoolAliasDef struct {
	Groups []weightedAliasGroupDef `json:"groups"`
}

func resolveStructurePoolAliases(defs []gen.PoolAliasDef, pos cube.Pos, seed int64) structurePoolAliasMapping {
	if len(defs) == 0 {
		return nil
	}
	rng := gen.NewPositionalRandomFactory(seed).At(pos[0], pos[1], pos[2])
	out := make(structurePoolAliasMapping, len(defs))
	for _, def := range defs {
		applyStructurePoolAlias(out, def, &rng)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func applyStructurePoolAlias(out structurePoolAliasMapping, def gen.PoolAliasDef, rng *gen.Xoroshiro128) {
	switch def.Type {
	case "direct":
		var raw directPoolAliasDef
		if err := json.Unmarshal(def.Raw, &raw); err != nil {
			return
		}
		alias := normalizeIdentifierName(raw.Alias)
		target := normalizeIdentifierName(raw.Target)
		if alias != "" && target != "" {
			out[alias] = target
		}
	case "random":
		var raw randomPoolAliasDef
		if err := json.Unmarshal(def.Raw, &raw); err != nil {
			return
		}
		alias := normalizeIdentifierName(raw.Alias)
		target := normalizeIdentifierName(weightedStringChoice(raw.Targets, rng))
		if alias != "" && target != "" {
			out[alias] = target
		}
	case "random_group":
		var raw randomGroupPoolAliasDef
		if err := json.Unmarshal(def.Raw, &raw); err != nil {
			return
		}
		group := weightedAliasGroupChoice(raw.Groups, rng)
		for _, inner := range group {
			applyStructurePoolAlias(out, inner, rng)
		}
	}
}

func weightedStringChoice(entries []weightedTargetDef, rng *gen.Xoroshiro128) string {
	total := 0
	for _, entry := range entries {
		if entry.Weight > 0 {
			total += entry.Weight
		}
	}
	if total <= 0 {
		return ""
	}
	pick := int(rng.NextInt(uint32(total)))
	for _, entry := range entries {
		if entry.Weight <= 0 {
			continue
		}
		if pick < entry.Weight {
			return entry.Data
		}
		pick -= entry.Weight
	}
	return ""
}

func weightedAliasGroupChoice(entries []weightedAliasGroupDef, rng *gen.Xoroshiro128) []gen.PoolAliasDef {
	total := 0
	for _, entry := range entries {
		if entry.Weight > 0 {
			total += entry.Weight
		}
	}
	if total <= 0 {
		return nil
	}
	pick := int(rng.NextInt(uint32(total)))
	for _, entry := range entries {
		if entry.Weight <= 0 {
			continue
		}
		if pick < entry.Weight {
			return entry.Data
		}
		pick -= entry.Weight
	}
	return nil
}

func (m structurePoolAliasMapping) lookup(name string) string {
	key := normalizeIdentifierName(name)
	if len(m) == 0 {
		return key
	}
	if target, ok := m[key]; ok && target != "" {
		return target
	}
	return key
}

func boxWithinHorizontalRange(box structureBox, centerX, centerZ, maxDistance int) bool {
	if box.empty() {
		return true
	}
	return box.minX >= centerX-maxDistance &&
		box.maxX <= centerX+maxDistance &&
		box.minZ >= centerZ-maxDistance &&
		box.maxZ <= centerZ+maxDistance
}

func (g Generator) buildPlannedStructure(
	candidate structurePlannerCandidate,
	start weightedStartTemplate,
	startX, startZ, minY, maxY int,
	rng *gen.Xoroshiro128,
) ([]plannedStructurePiece, structureBox, cube.Pos, [3]int, bool) {
	rootElement := resolvedPoolElement{
		elementType: "start",
		projection:  start.projection,
		placements: []structureTemplatePlacement{{
			templateName: start.name,
			ignoreAir:    start.ignoreAir,
			processors:   start.processors,
		}},
		jigsaws: start.jigsaws,
		size:    start.size,
	}
	if len(rootElement.placements) == 0 {
		return nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}

	rootRotation := randomStructureRotation(rng)
	anchor := [3]int{}
	if candidate.jigsaw.StartJigsawName != "" {
		targetName := normalizeIdentifierName(candidate.jigsaw.StartJigsawName)
		found := false
		var rootJigsaws []placedStructureJigsaw
		rootJigsaws = rootElement.appendShuffledJigsaws(rootJigsaws[:0], cube.Pos{}, rootRotation, rng)
		for _, jigsaw := range rootJigsaws {
			if jigsaw.name == targetName {
				anchor = [3]int{jigsaw.pos[0], jigsaw.pos[1], jigsaw.pos[2]}
				found = true
				break
			}
		}
		if !found {
			return nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
		}
	}

	rootOrigin := cube.Pos{startX - anchor[0], 0, startZ - anchor[2]}
	rootBox := rootElement.worldBox(rootOrigin, rootRotation)
	centerX := (rootBox.minX + rootBox.maxX) / 2
	centerZ := (rootBox.minZ + rootBox.maxZ) / 2
	baseY := g.sampleStructureHeightProvider(candidate.jigsaw.StartHeight, minY, maxY, rng)
	if candidate.jigsaw.ProjectStartToHeight != "" {
		baseY += g.preliminarySurfaceLevelAt(centerX, centerZ, minY, maxY)
	}
	rootOrigin[1] = baseY - 1
	rootBox = rootElement.worldBox(rootOrigin, rootRotation)

	rootPiece := plannedStructurePiece{
		element:   rootElement,
		origin:    rootOrigin,
		rotation:  rootRotation,
		bounds:    rootBox,
		rootPiece: true,
	}

	pieces := []plannedStructurePiece{rootPiece}
	occupied := make([]structureBox, 0, candidate.jigsaw.Size+1)
	if !rootBox.empty() {
		occupied = append(occupied, rootBox)
	}
	overall := rootBox

	if candidate.jigsaw.Size <= 0 {
		return pieces, overall, rootOrigin, rotatedStructureSize(rootElement.size, rootRotation), true
	}

	queue := []pendingStructurePiece{{piece: rootPiece, depth: 0}}
	aliasLookup := resolveStructurePoolAliases(candidate.jigsaw.PoolAliases, cube.Pos{startX, baseY, startZ}, g.seed)
	var (
		candidates      []resolvedPoolElement
		sourceJigsaws   []placedStructureJigsaw
		targetJigsaws   []placedStructureJigsaw
		targetRotations [4]structureRotation
	)
	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]
		if state.depth >= candidate.jigsaw.Size {
			continue
		}

		sourceRigid := normalizeIdentifierName(state.piece.element.projection) == "rigid"
		sourceSurfaceY := 0
		sourceSurfaceLoaded := false
		sourceJigsaws = state.piece.element.appendShuffledJigsaws(sourceJigsaws[:0], state.piece.origin, state.piece.rotation, rng)

	sourceJigsawLoop:
		for _, sourceJigsaw := range sourceJigsaws {
			pool, err := g.structureResolver.Pool(aliasLookup.lookup(sourceJigsaw.pool))
			if err != nil {
				continue
			}

			candidates = candidates[:0]
			if state.depth != candidate.jigsaw.Size {
				candidates = appendShuffledPoolElements(candidates, pool, rng)
			}
			if pool.fallback != "" {
				fallback, err := g.structureResolver.Pool(aliasLookup.lookup(pool.fallback))
				if err == nil {
					candidates = appendShuffledPoolElements(candidates, fallback, rng)
				}
			}

			for _, targetElement := range candidates {
				if targetElement.elementType == "empty_pool_element" {
					break
				}
				fillShuffledStructureRotations(&targetRotations, rng)
				for _, targetRotation := range targetRotations {
					targetJigsaws = targetElement.appendShuffledJigsaws(targetJigsaws[:0], cube.Pos{}, targetRotation, rng)
					targetRigid := normalizeIdentifierName(targetElement.projection) == "rigid"

					for _, targetJigsaw := range targetJigsaws {
						if !canAttachJigsaws(sourceJigsaw, targetJigsaw) {
							continue
						}

						targetPos := sourceJigsaw.pos.Add(cube.Pos{
							sourceJigsaw.front.stepX(),
							sourceJigsaw.front.stepY(),
							sourceJigsaw.front.stepZ(),
						})
						targetOrigin := cube.Pos{
							targetPos[0] - targetJigsaw.pos[0],
							targetPos[1] - targetJigsaw.pos[1],
							targetPos[2] - targetJigsaw.pos[2],
						}

						deltaY := sourceJigsaw.localY - targetJigsaw.localY + sourceJigsaw.front.stepY()
						if sourceRigid && targetRigid {
							targetOrigin[1] = state.piece.bounds.minY + deltaY
						} else {
							if !sourceSurfaceLoaded {
								sourceSurfaceY = g.preliminarySurfaceLevelAt(sourceJigsaw.pos[0], sourceJigsaw.pos[2], minY, maxY)
								sourceSurfaceLoaded = true
							}
							targetOrigin[1] = sourceSurfaceY - targetJigsaw.localY
						}

						targetBox := targetElement.worldBox(targetOrigin, targetRotation)
						if candidate.jigsaw.MaxDistanceFromCenter > 0 && !boxWithinHorizontalRange(targetBox, centerX, centerZ, candidate.jigsaw.MaxDistanceFromCenter) {
							continue
						}
						collides := false
						for _, occupiedBox := range occupied {
							if targetBox.intersects(occupiedBox) {
								collides = true
								break
							}
						}
						if collides {
							continue
						}

						targetPiece := plannedStructurePiece{
							element:  targetElement,
							origin:   targetOrigin,
							rotation: targetRotation,
							bounds:   targetBox,
						}
						pieces = append(pieces, targetPiece)
						if !targetBox.empty() {
							occupied = append(occupied, targetBox)
							overall = unionStructureBoxes(overall, targetBox)
						}
						if len(targetElement.jigsaws) != 0 && state.depth+1 <= candidate.jigsaw.Size {
							queue = insertPendingPiece(queue, pendingStructurePiece{
								piece:    targetPiece,
								depth:    state.depth + 1,
								priority: sourceJigsaw.placementPriority,
							})
						}
						continue sourceJigsawLoop
					}
				}
			}
		}
	}

	rootSize := rotatedStructureSize(rootElement.size, rootRotation)
	return pieces, overall, rootOrigin, rootSize, true
}
