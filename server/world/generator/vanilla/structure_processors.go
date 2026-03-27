package vanilla

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

type structureProcessor struct {
	kind      string
	blockRot  *structureBlockRotProcessor
	rule      *structureRuleProcessor
	protected *structureProtectedBlocksProcessor
	capped    *structureCappedProcessor
}

type structureBlockRotProcessor struct {
	integrity      float64
	rottableBlocks string
}

type structureRuleProcessor struct {
	rules []structureProcessorRule
}

type structureProcessorRule struct {
	input       structureRuleTest
	location    structureRuleTest
	position    structurePosRuleTest
	output      gen.BlockState
	blockEntity structureBlockEntityModifier
}

type structureProtectedBlocksProcessor struct {
	tag string
}

type structureCappedProcessor struct {
	delegate structureProcessor
	limit    int
}

type structureRuleTest struct {
	kind        string
	block       string
	tag         string
	probability float64
	blockState  gen.BlockState
}

type structurePosRuleTest struct {
	kind      string
	axis      string
	minChance float64
	maxChance float64
	minDist   int
	maxDist   int
}

type structureBlockEntityModifier struct {
	kind      string
	lootTable string
}

type structureProcessedBlock struct {
	templatePos   [3]int
	worldPos      cube.Pos
	originalState gen.BlockState
	state         gen.BlockState
	originalNBT   map[string]any
	nbt           map[string]any
}

type structureRuleProcessorDef struct {
	Rules []structureProcessorRuleDef `json:"rules"`
}

type structureProcessorRuleDef struct {
	InputPredicate      structureRuleTestDef             `json:"input_predicate"`
	LocationPredicate   structureRuleTestDef             `json:"location_predicate"`
	PositionPredicate   *structurePosRuleTestDef         `json:"position_predicate,omitempty"`
	OutputState         gen.BlockState                   `json:"output_state"`
	BlockEntityModifier *structureBlockEntityModifierDef `json:"block_entity_modifier,omitempty"`
}

type structureRuleTestDef struct {
	PredicateType string         `json:"predicate_type"`
	Block         string         `json:"block"`
	Tag           string         `json:"tag"`
	Probability   float64        `json:"probability"`
	BlockState    gen.BlockState `json:"block_state"`
}

type structurePosRuleTestDef struct {
	PredicateType string  `json:"predicate_type"`
	Axis          string  `json:"axis"`
	MinChance     float64 `json:"min_chance"`
	MaxChance     float64 `json:"max_chance"`
	MinDist       int     `json:"min_dist"`
	MaxDist       int     `json:"max_dist"`
}

type structureBlockEntityModifierDef struct {
	Type      string `json:"type"`
	LootTable string `json:"loot_table"`
}

type structureBlockRotProcessorDef struct {
	Integrity      float64 `json:"integrity"`
	RottableBlocks string  `json:"rottable_blocks"`
}

type structureProtectedBlocksProcessorDef struct {
	Value string `json:"value"`
}

type structureCappedProcessorDef struct {
	Delegate gen.StructureProcessorDef `json:"delegate"`
	Limit    int                       `json:"limit"`
}

func compileStructureProcessors(worldgen *gen.WorldgenRegistry, ref gen.ProcessorListRef) []structureProcessor {
	defs := resolveStructureProcessorDefs(worldgen, ref)
	if len(defs) == 0 {
		return nil
	}
	out := make([]structureProcessor, 0, len(defs))
	for _, def := range defs {
		processor, ok := compileStructureProcessorDef(worldgen, def)
		if !ok {
			continue
		}
		out = append(out, processor)
	}
	return out
}

func resolveStructureProcessorDefs(worldgen *gen.WorldgenRegistry, ref gen.ProcessorListRef) []gen.StructureProcessorDef {
	if ref.Inline != nil {
		return append([]gen.StructureProcessorDef(nil), ref.Inline.Processors...)
	}
	if worldgen == nil || ref.Name == "" {
		return nil
	}
	def, err := worldgen.ProcessorList(ref.Name)
	if err != nil {
		return nil
	}
	return append([]gen.StructureProcessorDef(nil), def.Processors...)
}

func compileStructureProcessorDef(worldgen *gen.WorldgenRegistry, def gen.StructureProcessorDef) (structureProcessor, bool) {
	switch def.Type {
	case "rule":
		var raw structureRuleProcessorDef
		if err := json.Unmarshal(def.Raw, &raw); err != nil {
			return structureProcessor{}, false
		}
		rules := make([]structureProcessorRule, 0, len(raw.Rules))
		for _, rule := range raw.Rules {
			position := structurePosRuleTest{kind: "always_true"}
			if rule.PositionPredicate != nil {
				position = compileStructurePosRuleTest(*rule.PositionPredicate)
			}
			modifier := structureBlockEntityModifier{}
			if rule.BlockEntityModifier != nil {
				modifier = compileStructureBlockEntityModifier(*rule.BlockEntityModifier)
			}
			rules = append(rules, structureProcessorRule{
				input:       compileStructureRuleTest(rule.InputPredicate),
				location:    compileStructureRuleTest(rule.LocationPredicate),
				position:    position,
				output:      rule.OutputState,
				blockEntity: modifier,
			})
		}
		return structureProcessor{
			kind: "rule",
			rule: &structureRuleProcessor{rules: rules},
		}, true
	case "block_rot":
		var raw structureBlockRotProcessorDef
		if err := json.Unmarshal(def.Raw, &raw); err != nil {
			return structureProcessor{}, false
		}
		return structureProcessor{
			kind: "block_rot",
			blockRot: &structureBlockRotProcessor{
				integrity:      raw.Integrity,
				rottableBlocks: normalizeStructureTag(raw.RottableBlocks),
			},
		}, true
	case "protected_blocks":
		var raw structureProtectedBlocksProcessorDef
		if err := json.Unmarshal(def.Raw, &raw); err != nil {
			return structureProcessor{}, false
		}
		return structureProcessor{
			kind:      "protected_blocks",
			protected: &structureProtectedBlocksProcessor{tag: normalizeStructureTag(raw.Value)},
		}, true
	case "capped":
		var raw structureCappedProcessorDef
		if err := json.Unmarshal(def.Raw, &raw); err != nil {
			return structureProcessor{}, false
		}
		delegate, ok := compileStructureProcessorDef(worldgen, raw.Delegate)
		if !ok {
			return structureProcessor{}, false
		}
		return structureProcessor{
			kind: "capped",
			capped: &structureCappedProcessor{
				delegate: delegate,
				limit:    raw.Limit,
			},
		}, true
	default:
		return structureProcessor{}, false
	}
}

func compileStructureRuleTest(def structureRuleTestDef) structureRuleTest {
	return structureRuleTest{
		kind:        normalizeIdentifierName(def.PredicateType),
		block:       normalizeIdentifierName(def.Block),
		tag:         normalizeStructureTag(def.Tag),
		probability: def.Probability,
		blockState:  def.BlockState,
	}
}

func compileStructurePosRuleTest(def structurePosRuleTestDef) structurePosRuleTest {
	return structurePosRuleTest{
		kind:      normalizeIdentifierName(def.PredicateType),
		axis:      strings.ToLower(def.Axis),
		minChance: def.MinChance,
		maxChance: def.MaxChance,
		minDist:   def.MinDist,
		maxDist:   def.MaxDist,
	}
}

func compileStructureBlockEntityModifier(def structureBlockEntityModifierDef) structureBlockEntityModifier {
	return structureBlockEntityModifier{
		kind:      normalizeIdentifierName(def.Type),
		lootTable: normalizeIdentifierName(def.LootTable),
	}
}

func normalizeStructureTag(tag string) string {
	tag = strings.TrimPrefix(tag, "#")
	return normalizeIdentifierName(tag)
}

func (g Generator) processStructureTemplatePlacement(
	c *chunk.Chunk,
	chunkX, chunkZ int,
	reference cube.Pos,
	rotation structureRotation,
	mirror structureMirror,
	pivot cube.Pos,
	useTemplateTransform bool,
	template gen.StructureTemplate,
	placement structureTemplatePlacement,
) []structureProcessedBlock {
	processed := make([]structureProcessedBlock, 0, len(template.Blocks))
	for _, blockInfo := range template.Blocks {
		if blockInfo.State < 0 || blockInfo.State >= len(template.Palette) {
			continue
		}
		templateState := structureTemplateState(template.Palette[blockInfo.State])
		worldPos := cube.Pos{}
		if useTemplateTransform {
			worldPos = structureTemplateWorldPos(reference, blockInfo.Pos, rotation, mirror, pivot)
		} else {
			rotatedPos := rotateStructurePos(template.Size, blockInfo.Pos, rotation)
			worldPos = cube.Pos{reference[0] + rotatedPos[0], reference[1] + rotatedPos[1], reference[2] + rotatedPos[2]}
		}
		processed = append(processed, structureProcessedBlock{
			templatePos:   blockInfo.Pos,
			worldPos:      worldPos,
			originalState: templateState,
			state:         templateState,
			originalNBT:   cloneStructureNBT(blockInfo.NBT),
			nbt:           cloneStructureNBT(blockInfo.NBT),
		})
	}
	for _, processor := range placement.processors {
		processed = g.applyStructureProcessor(c, chunkX, chunkZ, reference, processed, processor)
		if len(processed) == 0 {
			break
		}
	}
	return processed
}

func (g Generator) applyStructureProcessor(
	c *chunk.Chunk,
	chunkX, chunkZ int,
	reference cube.Pos,
	blocks []structureProcessedBlock,
	processor structureProcessor,
) []structureProcessedBlock {
	switch processor.kind {
	case "rule":
		out := make([]structureProcessedBlock, 0, len(blocks))
		for _, block := range blocks {
			updated, keep := g.applyStructureRuleProcessor(c, chunkX, chunkZ, reference, block, processor.rule)
			if keep {
				out = append(out, updated)
			}
		}
		return out
	case "block_rot":
		out := make([]structureProcessedBlock, 0, len(blocks))
		for _, block := range blocks {
			if g.applyStructureBlockRotProcessor(block, processor.blockRot) {
				out = append(out, block)
			}
		}
		return out
	case "protected_blocks":
		out := make([]structureProcessedBlock, 0, len(blocks))
		for _, block := range blocks {
			if g.applyStructureProtectedBlocksProcessor(c, chunkX, chunkZ, block, processor.protected) {
				out = append(out, block)
			}
		}
		return out
	case "capped":
		return g.applyStructureCappedProcessor(c, chunkX, chunkZ, reference, blocks, processor.capped)
	default:
		return blocks
	}
}

func (g Generator) applyStructureRuleProcessor(
	c *chunk.Chunk,
	chunkX, chunkZ int,
	reference cube.Pos,
	block structureProcessedBlock,
	processor *structureRuleProcessor,
) (structureProcessedBlock, bool) {
	if processor == nil {
		return block, true
	}
	locState := g.structureWorldBlockState(c, chunkX, chunkZ, block.worldPos)
	for _, rule := range processor.rules {
		rng := newStructureSeededRNG(block.worldPos)
		if !matchesStructureRuleTest(block.state, rule.input, &rng) {
			continue
		}
		if !matchesStructureRuleTest(locState, rule.location, &rng) {
			continue
		}
		if !matchesStructurePosRuleTest(block.templatePos, block.worldPos, reference, rule.position, &rng) {
			continue
		}
		block.state = cloneBlockState(rule.output)
		block.nbt = applyStructureBlockEntityModifier(rule.blockEntity, &rng, block.nbt)
		return block, true
	}
	return block, true
}

func (g Generator) applyStructureBlockRotProcessor(block structureProcessedBlock, processor *structureBlockRotProcessor) bool {
	if processor == nil {
		return true
	}
	if processor.rottableBlocks != "" && !matchesStructureBlockTag(block.originalState.Name, processor.rottableBlocks) {
		return true
	}
	rng := newStructureSeededRNG(block.worldPos)
	return rng.NextDouble() <= processor.integrity
}

func (g Generator) applyStructureProtectedBlocksProcessor(c *chunk.Chunk, chunkX, chunkZ int, block structureProcessedBlock, processor *structureProtectedBlocksProcessor) bool {
	if processor == nil || processor.tag == "" {
		return true
	}
	locState := g.structureWorldBlockState(c, chunkX, chunkZ, block.worldPos)
	return matchesStructureProtectedReplaceable(locState.Name, processor.tag)
}

func (g Generator) applyStructureCappedProcessor(
	c *chunk.Chunk,
	chunkX, chunkZ int,
	reference cube.Pos,
	blocks []structureProcessedBlock,
	processor *structureCappedProcessor,
) []structureProcessedBlock {
	if processor == nil || processor.limit <= 0 || len(blocks) == 0 {
		return blocks
	}
	indices := make([]int, len(blocks))
	for i := range indices {
		indices[i] = i
	}
	rng := gen.NewPositionalRandomFactory(g.seed).At(reference[0], reference[1], reference[2])
	shuffleWithRNG(indices, &rng)

	out := append([]structureProcessedBlock(nil), blocks...)
	replaced := 0
	for _, index := range indices {
		if replaced >= processor.limit {
			break
		}
		before := out[index]
		updated := g.applyStructureProcessor(c, chunkX, chunkZ, reference, []structureProcessedBlock{before}, processor.delegate)
		if len(updated) != 1 {
			continue
		}
		if structureProcessedBlocksEqual(before, updated[0]) {
			continue
		}
		out[index] = updated[0]
		replaced++
	}
	return out
}

func structureProcessedBlocksEqual(a, b structureProcessedBlock) bool {
	if a.worldPos != b.worldPos || a.templatePos != b.templatePos || a.originalState.Name != b.originalState.Name || a.state.Name != b.state.Name {
		return false
	}
	if !structureStringMapEqual(a.originalState.Properties, b.originalState.Properties) || !structureStringMapEqual(a.state.Properties, b.state.Properties) {
		return false
	}
	return structureAnyMapEqual(a.nbt, b.nbt)
}

func newStructureSeededRNG(pos cube.Pos) gen.Xoroshiro128 {
	seed := int64(pos[0])*3129871 ^ int64(pos[2])*116129781 ^ int64(pos[1])
	seed = seed*seed*42317861 + seed*11
	return gen.NewXoroshiro128FromSeed(seed >> 16)
}

func matchesStructureRuleTest(state gen.BlockState, test structureRuleTest, rng *gen.Xoroshiro128) bool {
	switch test.kind {
	case "", "always_true":
		return true
	case "block_match":
		return normalizeIdentifierName(state.Name) == test.block
	case "random_block_match":
		return normalizeIdentifierName(state.Name) == test.block && rng.NextDouble() < test.probability
	case "tag_match":
		return matchesStructureBlockTag(state.Name, test.tag)
	case "blockstate_match":
		return structureBlockStatesEqual(state, test.blockState)
	default:
		return false
	}
}

func matchesStructurePosRuleTest(templatePos [3]int, worldPos, reference cube.Pos, test structurePosRuleTest, rng *gen.Xoroshiro128) bool {
	switch test.kind {
	case "", "always_true":
		return true
	case "axis_aligned_linear_pos":
		if test.maxDist <= test.minDist {
			return false
		}
		var dist int
		switch strings.ToLower(test.axis) {
		case "x":
			dist = abs(worldPos[0] - reference[0])
		case "z":
			dist = abs(worldPos[2] - reference[2])
		default:
			dist = abs(worldPos[1] - reference[1])
		}
		f := inverseLerp(float64(dist), float64(test.minDist), float64(test.maxDist))
		chance := clampedLerp(f, test.minChance, test.maxChance)
		return rng.NextDouble() <= chance
	default:
		return false
	}
}

func applyStructureBlockEntityModifier(modifier structureBlockEntityModifier, rng *gen.Xoroshiro128, existing map[string]any) map[string]any {
	switch modifier.kind {
	case "", "passthrough":
		return cloneStructureNBT(existing)
	case "append_loot":
		out := cloneStructureNBT(existing)
		if out == nil {
			out = make(map[string]any, 2)
		}
		out["LootTable"] = "minecraft:" + modifier.lootTable
		out["LootTableSeed"] = int64(rng.NextLong())
		return out
	default:
		return cloneStructureNBT(existing)
	}
}

func matchesStructureBlockTag(name, tag string) bool {
	name = normalizeIdentifierName(name)
	switch normalizeStructureTag(tag) {
	case "doors":
		_, ok := structureBlockTags["doors"][name]
		return ok
	case "trail_ruins_replaceable":
		_, ok := structureBlockTags["trail_ruins_replaceable"][name]
		return ok
	case "ancient_city_replaceable":
		_, ok := structureBlockTags["ancient_city_replaceable"][name]
		return ok
	case "features_cannot_replace":
		_, ok := structureBlockTags["features_cannot_replace"][name]
		return ok
	default:
		return false
	}
}

func matchesStructureProtectedReplaceable(name, tag string) bool {
	return !matchesStructureBlockTag(name, tag)
}

func structureTemplateState(state gen.StructureTemplateBlockState) gen.BlockState {
	props := make(map[string]string, len(state.Properties))
	keys := make([]string, 0, len(state.Properties))
	for key := range state.Properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		props[key] = structurePropString(state.Properties[key])
	}
	if len(props) == 0 {
		props = nil
	}
	return gen.BlockState{
		Name:       normalizeIdentifierName(state.Name),
		Properties: props,
	}
}

func structureLookupName(name string) string {
	normalized := normalizeIdentifierName(name)
	if strings.HasPrefix(normalized, "potted_") {
		return "minecraft:flower_pot"
	}
	if strings.HasPrefix(name, "minecraft:") {
		return name
	}
	return "minecraft:" + normalized
}

func structureLookupProperties(name string, properties map[string]string) map[string]any {
	normalizedName := normalizeIdentifierName(name)
	if len(properties) == 0 {
		switch {
		case normalizedName == "bone_block", normalizedName == "hay_block":
			return map[string]any{"deprecated": int32(0)}
		case normalizedName == "cauldron":
			return map[string]any{"cauldron_liquid": "water", "fill_level": int32(0)}
		case normalizedName == "flower_pot", strings.HasPrefix(normalizedName, "potted_"):
			return map[string]any{"update_bit": false}
		default:
			return nil
		}
	}
	normalized := make(map[string]string, len(properties)+1)
	for key, value := range properties {
		normalized[key] = value
	}

	switch {
	case strings.HasSuffix(normalizedName, "_log"),
		strings.HasSuffix(normalizedName, "_wood"),
		strings.HasSuffix(normalizedName, "_stem"),
		strings.HasSuffix(normalizedName, "_hyphae"),
		normalizedName == "muddy_mangrove_roots",
		normalizedName == "basalt",
		normalizedName == "deepslate",
		normalizedName == "bone_block",
		normalizedName == "hay_block",
		normalizedName == "purpur_pillar",
		normalizedName == "quartz_pillar":
		if axis, ok := normalized["axis"]; ok {
			delete(normalized, "axis")
			normalized["pillar_axis"] = axis
		}
	}
	switch normalizedName {
	case "bone_block", "hay_block":
		if _, ok := normalized["deprecated"]; !ok {
			normalized["deprecated"] = "0"
		}
	}
	if strings.HasSuffix(normalizedName, "_stairs") {
		if _, ok := normalized["upside_down_bit"]; !ok {
			normalized["upside_down_bit"] = "false"
		}
		if facing, ok := normalized["facing"]; ok {
			delete(normalized, "facing")
			normalized["weirdo_direction"] = structureLookupStairsDirection(facing)
		} else if facing, ok := normalized["facing_direction"]; ok {
			delete(normalized, "facing_direction")
			normalized["weirdo_direction"] = structureLookupStairsDirection(facing)
		}
	}

	out := make(map[string]any, len(normalized))
	for key, value := range normalized {
		out[key] = structureLookupPropertyValue(value)
	}
	return out
}

func structureLookupStairsDirection(facing string) string {
	switch facing {
	case "north":
		return "3"
	case "south":
		return "2"
	case "west":
		return "1"
	case "east":
		return "0"
	default:
		return facing
	}
}

func structureLookupPropertyValue(value string) any {
	switch value {
	case "true":
		return true
	case "false":
		return false
	}
	if i, err := strconv.Atoi(value); err == nil {
		return int32(i)
	}
	return value
}

func structurePropString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return strings.TrimSpace(strings.ToLower(strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(toString(v)), "minecraft:"), "#")))
	}
}

func toString(value any) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		if v == math.Trunc(v) {
			return strconv.Itoa(int(v))
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return strings.TrimSpace(strings.ToLower(strings.TrimSpace(fmt.Sprint(v))))
	}
}

func (g Generator) structureWorldBlockState(c *chunk.Chunk, chunkX, chunkZ int, pos cube.Pos) gen.BlockState {
	if c == nil {
		return gen.BlockState{Name: "air"}
	}
	if pos[0] < chunkX*16 || pos[0] >= chunkX*16+16 || pos[2] < chunkZ*16 || pos[2] >= chunkZ*16+16 {
		return gen.BlockState{Name: "air"}
	}
	localX := pos[0] - chunkX*16
	localZ := pos[2] - chunkZ*16
	if pos[1] < c.Range().Min() || pos[1] > c.Range().Max() {
		return gen.BlockState{Name: "air"}
	}
	rid := c.Block(uint8(localX), int16(pos[1]), uint8(localZ), 0)
	b, ok := world.BlockByRuntimeID(rid)
	if !ok {
		return gen.BlockState{Name: "air"}
	}
	name, properties := b.EncodeBlock()
	out := gen.BlockState{Name: normalizeIdentifierName(name)}
	if len(properties) != 0 {
		out.Properties = make(map[string]string, len(properties))
		keys := make([]string, 0, len(properties))
		for key := range properties {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			out.Properties[key] = structurePropString(properties[key])
		}
	}
	return out
}

func cloneBlockState(state gen.BlockState) gen.BlockState {
	out := gen.BlockState{Name: normalizeIdentifierName(state.Name)}
	if len(state.Properties) != 0 {
		out.Properties = make(map[string]string, len(state.Properties))
		for key, value := range state.Properties {
			out.Properties[key] = value
		}
	}
	return out
}

func rotatePlacedStructureState(state gen.BlockState, rotation structureRotation) gen.BlockState {
	if rotation == structureRotationNone || len(state.Properties) == 0 {
		return cloneBlockState(state)
	}
	out := cloneBlockState(state)
	props := out.Properties

	if hasAnyDirectionProperty(props) {
		north := props["north"]
		east := props["east"]
		south := props["south"]
		west := props["west"]
		switch rotation {
		case structureRotationClockwise90:
			props["north"] = west
			props["east"] = north
			props["south"] = east
			props["west"] = south
		case structureRotationClockwise180:
			props["north"] = south
			props["east"] = west
			props["south"] = north
			props["west"] = east
		case structureRotationCounterclockwise90:
			props["north"] = east
			props["east"] = south
			props["south"] = west
			props["west"] = north
		}
	}

	for _, key := range []string{"facing", "direction", "horizontal_facing"} {
		if value, ok := props[key]; ok {
			props[key] = rotateHorizontalDirectionName(value, rotation)
		}
	}
	if value, ok := props["axis"]; ok {
		props["axis"] = rotateAxisName(value, rotation)
	}
	if value, ok := props["rotation"]; ok {
		props["rotation"] = rotateStructureRotationProperty(value, rotation)
	}
	if value, ok := props["shape"]; ok {
		props["shape"] = rotateShapeProperty(value, rotation)
	}
	if value, ok := props["orientation"]; ok {
		props["orientation"] = rotateOrientationProperty(value, rotation)
	}

	return out
}

func hasAnyDirectionProperty(properties map[string]string) bool {
	_, north := properties["north"]
	_, east := properties["east"]
	_, south := properties["south"]
	_, west := properties["west"]
	return north || east || south || west
}

func rotateHorizontalDirectionName(value string, rotation structureRotation) string {
	switch rotation {
	case structureRotationClockwise90:
		switch value {
		case "north":
			return "east"
		case "east":
			return "south"
		case "south":
			return "west"
		case "west":
			return "north"
		}
	case structureRotationClockwise180:
		switch value {
		case "north":
			return "south"
		case "east":
			return "west"
		case "south":
			return "north"
		case "west":
			return "east"
		}
	case structureRotationCounterclockwise90:
		switch value {
		case "north":
			return "west"
		case "east":
			return "north"
		case "south":
			return "east"
		case "west":
			return "south"
		}
	}
	return value
}

func rotateAxisName(value string, rotation structureRotation) string {
	switch rotation {
	case structureRotationClockwise90, structureRotationCounterclockwise90:
		switch value {
		case "x":
			return "z"
		case "z":
			return "x"
		}
	}
	return value
}

func rotateStructureRotationProperty(value string, rotation structureRotation) string {
	n, err := strconv.Atoi(value)
	if err != nil {
		return value
	}
	switch rotation {
	case structureRotationClockwise90:
		n = (n + 4) % 16
	case structureRotationClockwise180:
		n = (n + 8) % 16
	case structureRotationCounterclockwise90:
		n = (n + 12) % 16
	}
	return strconv.Itoa(n)
}

func rotateShapeProperty(value string, rotation structureRotation) string {
	switch value {
	case "north_south":
		if rotation == structureRotationClockwise90 || rotation == structureRotationCounterclockwise90 {
			return "east_west"
		}
	case "east_west":
		if rotation == structureRotationClockwise90 || rotation == structureRotationCounterclockwise90 {
			return "north_south"
		}
	case "ascending_north", "ascending_east", "ascending_south", "ascending_west":
		return "ascending_" + rotateHorizontalDirectionName(strings.TrimPrefix(value, "ascending_"), rotation)
	case "south_east":
		switch rotation {
		case structureRotationClockwise90:
			return "south_west"
		case structureRotationClockwise180:
			return "north_west"
		case structureRotationCounterclockwise90:
			return "north_east"
		}
	case "south_west":
		switch rotation {
		case structureRotationClockwise90:
			return "north_west"
		case structureRotationClockwise180:
			return "north_east"
		case structureRotationCounterclockwise90:
			return "south_east"
		}
	case "north_west":
		switch rotation {
		case structureRotationClockwise90:
			return "north_east"
		case structureRotationClockwise180:
			return "south_east"
		case structureRotationCounterclockwise90:
			return "south_west"
		}
	case "north_east":
		switch rotation {
		case structureRotationClockwise90:
			return "south_east"
		case structureRotationClockwise180:
			return "south_west"
		case structureRotationCounterclockwise90:
			return "north_west"
		}
	}
	return value
}

func rotateOrientationProperty(value string, rotation structureRotation) string {
	parts := strings.SplitN(value, "_", 2)
	if len(parts) != 2 {
		return value
	}
	return rotateHorizontalDirectionName(parts[0], rotation) + "_" + parts[1]
}

func cloneStructureNBT(nbt map[string]any) map[string]any {
	if len(nbt) == 0 {
		return nil
	}
	out := make(map[string]any, len(nbt))
	for key, value := range nbt {
		out[key] = value
	}
	return out
}

func structureBlockStatesEqual(a, b gen.BlockState) bool {
	if normalizeIdentifierName(a.Name) != normalizeIdentifierName(b.Name) {
		return false
	}
	return structureStringMapEqual(a.Properties, b.Properties)
}

func structureStringMapEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}

func structureAnyMapEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if !structureAnyValueEqual(value, b[key]) {
			return false
		}
	}
	return true
}

func structureAnyValueEqual(a, b any) bool {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		return ok && structureAnyMapEqual(av, bv)
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !structureAnyValueEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a, b)
	}
}

func inverseLerp(value, min, max float64) float64 {
	if min == max {
		return 0
	}
	return (value - min) / (max - min)
}

func clampedLerp(t, start, end float64) float64 {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return start + (end-start)*t
}

var structureBlockTags = map[string]map[string]struct{}{
	"doors": {
		"oak_door": {}, "spruce_door": {}, "birch_door": {}, "jungle_door": {}, "acacia_door": {}, "dark_oak_door": {},
		"pale_oak_door": {}, "crimson_door": {}, "warped_door": {}, "mangrove_door": {}, "bamboo_door": {}, "cherry_door": {},
		"copper_door": {}, "exposed_copper_door": {}, "weathered_copper_door": {}, "oxidized_copper_door": {},
		"waxed_copper_door": {}, "waxed_exposed_copper_door": {}, "waxed_weathered_copper_door": {}, "waxed_oxidized_copper_door": {},
		"iron_door": {},
	},
	"trail_ruins_replaceable": {
		"gravel": {},
	},
	"features_cannot_replace": {
		"bedrock": {}, "spawner": {}, "chest": {}, "end_portal_frame": {}, "reinforced_deepslate": {}, "trial_spawner": {}, "vault": {},
	},
	"ancient_city_replaceable": {
		"deepslate": {}, "deepslate_bricks": {}, "deepslate_tiles": {}, "deepslate_brick_slab": {}, "deepslate_tile_slab": {},
		"deepslate_brick_stairs": {}, "deepslate_tile_wall": {}, "deepslate_brick_wall": {}, "cobbled_deepslate": {},
		"cracked_deepslate_bricks": {}, "cracked_deepslate_tiles": {}, "gray_wool": {},
	},
}
