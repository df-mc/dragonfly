package vanilla

import (
	"encoding/json"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

func TestProcessStructurePlacementAppliesRuleProcessor(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	template := singleBlockStructureTemplate("minecraft:wheat")
	placement := structureTemplatePlacement{
		processors: compileStructureProcessors(nil, inlineProcessorList(
			t,
			"rule",
			`{
				"processor_type":"minecraft:rule",
				"rules":[
					{
						"input_predicate":{"predicate_type":"minecraft:block_match","block":"minecraft:wheat"},
						"location_predicate":{"predicate_type":"minecraft:always_true"},
						"output_state":{"Name":"minecraft:carrots","Properties":{"age":"0"}}
					}
				]
			}`,
		)),
	}

	blocks := g.processStructureTemplatePlacement(c, 0, 0, cube.Pos{8, 1, 8}, structureRotationNone, structureMirrorNone, cube.Pos{}, false, template, placement)
	if len(blocks) != 1 {
		t.Fatalf("expected one processed block, got %d", len(blocks))
	}
	if blocks[0].state.Name != "carrots" {
		t.Fatalf("expected rule processor to replace wheat with carrots, got %q", blocks[0].state.Name)
	}
	if blocks[0].state.Properties["age"] != "0" {
		t.Fatalf("expected rule processor to preserve output state properties, got %#v", blocks[0].state.Properties)
	}
	if _, ok := g.lookupTemplateBlock(structureLookupName(blocks[0].state.Name), structureLookupProperties(blocks[0].state.Name, blocks[0].state.Properties)); !ok {
		t.Fatal("expected processed rule output state to resolve to a Dragonfly block")
	}
}

func TestStructureLookupPropertiesNormalizesJavaAxisStates(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := NewForDimension(0, world.Nether)
	props := structureLookupProperties("bone_block", map[string]string{"axis": "z"})
	rid, ok := g.lookupTemplateBlock("minecraft:bone_block", props)
	if !ok {
		t.Fatal("expected Java bone_block axis state to resolve to a Dragonfly block")
	}
	b, ok := world.BlockByRuntimeID(rid)
	if !ok {
		t.Fatal("expected runtime ID to resolve back to a block")
	}
	name, resolved := b.EncodeBlock()
	if name != "minecraft:bone_block" {
		t.Fatalf("expected normalized lookup to resolve bone_block, got %q", name)
	}
	if resolved["pillar_axis"] != "z" {
		t.Fatalf("expected normalized lookup to preserve axis as pillar_axis=z, got %#v", resolved)
	}
}

func TestProcessStructurePlacementAppliesBlockRotProcessor(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	template := singleBlockStructureTemplate("minecraft:stone")
	placement := structureTemplatePlacement{
		processors: compileStructureProcessors(nil, inlineProcessorList(
			t,
			"block_rot",
			`{"processor_type":"minecraft:block_rot","integrity":0.0}`,
		)),
	}

	blocks := g.processStructureTemplatePlacement(c, 0, 0, cube.Pos{8, 1, 8}, structureRotationNone, structureMirrorNone, cube.Pos{}, false, template, placement)
	if len(blocks) != 0 {
		t.Fatalf("expected block rot processor with integrity 0 to remove block, got %d block(s)", len(blocks))
	}
}

func TestProcessStructurePlacementAppliesProtectedBlocksProcessor(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	c.SetBlock(8, 1, 8, 0, world.BlockRuntimeID(block.Chest{}))
	template := singleBlockStructureTemplate("minecraft:stone")
	placement := structureTemplatePlacement{
		processors: compileStructureProcessors(nil, inlineProcessorList(
			t,
			"protected_blocks",
			`{"processor_type":"minecraft:protected_blocks","value":"#minecraft:features_cannot_replace"}`,
		)),
	}

	blocks := g.processStructureTemplatePlacement(c, 0, 0, cube.Pos{8, 1, 8}, structureRotationNone, structureMirrorNone, cube.Pos{}, false, template, placement)
	if len(blocks) != 0 {
		t.Fatalf("expected protected block processor to skip replacing chest, got %d block(s)", len(blocks))
	}
}

func TestProcessStructurePlacementAppliesCappedProcessor(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	template := gen.StructureTemplate{
		Size: [3]int{3, 1, 1},
		Palette: []gen.StructureTemplateBlockState{
			{Name: "minecraft:gravel"},
		},
		Blocks: []gen.StructureTemplateBlock{
			{Pos: [3]int{0, 0, 0}, State: 0},
			{Pos: [3]int{1, 0, 0}, State: 0},
			{Pos: [3]int{2, 0, 0}, State: 0},
		},
	}
	placement := structureTemplatePlacement{
		processors: compileStructureProcessors(nil, inlineProcessorList(
			t,
			"capped",
			`{
				"processor_type":"minecraft:capped",
				"limit":1,
				"delegate":{
					"processor_type":"minecraft:rule",
					"rules":[
						{
							"input_predicate":{"predicate_type":"minecraft:tag_match","tag":"#minecraft:trail_ruins_replaceable"},
							"location_predicate":{"predicate_type":"minecraft:always_true"},
							"output_state":{"Name":"minecraft:suspicious_gravel","Properties":{"dusted":"0"}}
						}
					]
				}
			}`,
		)),
	}

	blocks := g.processStructureTemplatePlacement(c, 0, 0, cube.Pos{8, 1, 8}, structureRotationNone, structureMirrorNone, cube.Pos{}, false, template, placement)
	if len(blocks) != 3 {
		t.Fatalf("expected capped processor to retain block count, got %d", len(blocks))
	}
	converted := 0
	for _, block := range blocks {
		if block.state.Name == "suspicious_gravel" {
			converted++
		}
	}
	if converted != 1 {
		t.Fatalf("expected capped processor to convert exactly one block, got %d", converted)
	}
}

func TestApplyStructureCappedProcessorHandlesNestedNBTLists(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	blocks := []structureProcessedBlock{{
		templatePos: [3]int{0, 0, 0},
		worldPos:    cube.Pos{8, 1, 8},
		originalState: gen.BlockState{
			Name: "stone",
		},
		state: gen.BlockState{
			Name: "stone",
		},
		originalNBT: map[string]any{
			"items": []any{
				map[string]any{"slot": int32(0), "id": "minecraft:stone"},
			},
		},
		nbt: map[string]any{
			"items": []any{
				map[string]any{"slot": int32(0), "id": "minecraft:stone"},
			},
		},
	}}

	processed := g.applyStructureCappedProcessor(c, 0, 0, cube.Pos{8, 1, 8}, blocks, &structureCappedProcessor{
		limit: 1,
		delegate: structureProcessor{
			kind: "block_rot",
			blockRot: &structureBlockRotProcessor{
				integrity: 1,
			},
		},
	})
	if len(processed) != 1 {
		t.Fatalf("expected capped processor to preserve unchanged block, got %d block(s)", len(processed))
	}
	if !structureAnyMapEqual(processed[0].nbt, blocks[0].nbt) {
		t.Fatal("expected nested NBT list payload to survive capped processor equality check")
	}
}

func TestProcessStructurePlacementAppliesWorldgenFarmProcessor(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	template := singleBlockStructureTemplate("minecraft:wheat")
	placement := structureTemplatePlacement{
		processors: compileStructureProcessors(g.worldgen, gen.ProcessorListRef{Name: "farm_plains"}),
	}

	foundReplacement := false
	for x := 0; x < 32 && !foundReplacement; x++ {
		for z := 0; z < 32 && !foundReplacement; z++ {
			blocks := g.processStructureTemplatePlacement(c, x>>4, z>>4, cube.Pos{x, 1, z}, structureRotationNone, structureMirrorNone, cube.Pos{}, false, template, placement)
			if len(blocks) != 1 {
				continue
			}
			switch blocks[0].state.Name {
			case "carrots", "potatoes", "beetroots":
				foundReplacement = true
			}
		}
	}
	if !foundReplacement {
		t.Fatal("expected farm_plains processor list to replace wheat at least once across sampled positions")
	}
}

func TestRotatePlacedStructureStateRotatesDirectionalProperties(t *testing.T) {
	state := gen.BlockState{
		Name: "oak_stairs",
		Properties: map[string]string{
			"facing": "north",
			"axis":   "x",
			"north":  "low",
			"east":   "none",
			"south":  "tall",
			"west":   "low",
			"shape":  "ascending_north",
		},
	}

	rotated := rotatePlacedStructureState(state, structureRotationClockwise90)
	if rotated.Properties["facing"] != "east" {
		t.Fatalf("expected facing to rotate north->east, got %q", rotated.Properties["facing"])
	}
	if rotated.Properties["axis"] != "z" {
		t.Fatalf("expected axis to rotate x->z, got %q", rotated.Properties["axis"])
	}
	if rotated.Properties["north"] != "low" || rotated.Properties["east"] != "low" || rotated.Properties["south"] != "none" || rotated.Properties["west"] != "tall" {
		t.Fatalf("expected cardinal connection properties to rotate, got %#v", rotated.Properties)
	}
	if rotated.Properties["shape"] != "ascending_east" {
		t.Fatalf("expected rail-like shape to rotate to ascending_east, got %q", rotated.Properties["shape"])
	}
}

func inlineProcessorList(t *testing.T, typ, raw string) gen.ProcessorListRef {
	t.Helper()

	return gen.ProcessorListRef{
		Inline: &gen.ProcessorListDef{
			Processors: []gen.StructureProcessorDef{{
				Type: typ,
				Raw:  json.RawMessage(raw),
			}},
		},
	}
}

func singleBlockStructureTemplate(name string) gen.StructureTemplate {
	return gen.StructureTemplate{
		Size: [3]int{1, 1, 1},
		Palette: []gen.StructureTemplateBlockState{
			{Name: name},
		},
		Blocks: []gen.StructureTemplateBlock{
			{Pos: [3]int{0, 0, 0}, State: 0},
		},
	}
}
