package vanilla

import (
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/block/cube"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

type structureMirror uint8

const (
	structureMirrorNone structureMirror = iota
	structureMirrorLeftRight
	structureMirrorFrontBack
)

type plannedStructureBlock struct {
	worldPos cube.Pos
	state    gen.BlockState
}

func structureTemplateWorldPos(reference cube.Pos, pos [3]int, rotation structureRotation, mirror structureMirror, pivot cube.Pos) cube.Pos {
	transformed := transformStructurePos(pos, rotation, mirror, pivot)
	return cube.Pos{
		reference[0] + transformed[0],
		reference[1] + transformed[1],
		reference[2] + transformed[2],
	}
}

func transformStructurePos(pos [3]int, rotation structureRotation, mirror structureMirror, pivot cube.Pos) [3]int {
	x := pos[0]
	y := pos[1]
	z := pos[2]
	wasMirrored := true

	switch mirror {
	case structureMirrorLeftRight:
		z = -z
	case structureMirrorFrontBack:
		x = -x
	default:
		wasMirrored = false
	}

	pivotX := pivot[0]
	pivotZ := pivot[2]
	switch rotation {
	case structureRotationCounterclockwise90:
		return [3]int{pivotX - pivotZ + z, y, pivotX + pivotZ - x}
	case structureRotationClockwise90:
		return [3]int{pivotX + pivotZ - z, y, pivotZ - pivotX + x}
	case structureRotationClockwise180:
		return [3]int{pivotX + pivotX - x, y, pivotZ + pivotZ - z}
	default:
		if wasMirrored {
			return [3]int{x, y, z}
		}
		return pos
	}
}

func structureTemplateWorldBox(template gen.StructureTemplate, reference cube.Pos, rotation structureRotation, mirror structureMirror, pivot cube.Pos) structureBox {
	box := emptyStructureBox()
	if len(template.Blocks) == 0 {
		corners := [][3]int{
			{0, 0, 0},
			{max(template.Size[0]-1, 0), 0, 0},
			{0, max(template.Size[1]-1, 0), 0},
			{0, 0, max(template.Size[2]-1, 0)},
			{max(template.Size[0]-1, 0), max(template.Size[1]-1, 0), max(template.Size[2]-1, 0)},
		}
		for _, corner := range corners {
			pos := structureTemplateWorldPos(reference, corner, rotation, mirror, pivot)
			box = unionStructureBoxes(box, structureBox{minX: pos[0], minY: pos[1], minZ: pos[2], maxX: pos[0], maxY: pos[1], maxZ: pos[2]})
		}
		return box
	}
	for _, blockInfo := range template.Blocks {
		pos := structureTemplateWorldPos(reference, blockInfo.Pos, rotation, mirror, pivot)
		box = unionStructureBoxes(box, structureBox{minX: pos[0], minY: pos[1], minZ: pos[2], maxX: pos[0], maxY: pos[1], maxZ: pos[2]})
	}
	return box
}

func applyPlacedStructureStateTransform(state gen.BlockState, mirror structureMirror, rotation structureRotation) gen.BlockState {
	out := mirrorPlacedStructureState(state, mirror)
	return rotatePlacedStructureState(out, rotation)
}

func mirrorPlacedStructureState(state gen.BlockState, mirror structureMirror) gen.BlockState {
	if mirror == structureMirrorNone || len(state.Properties) == 0 {
		return cloneBlockState(state)
	}
	out := cloneBlockState(state)
	props := out.Properties

	switch mirror {
	case structureMirrorFrontBack:
		swapPropertyValues(props, "east", "west")
	case structureMirrorLeftRight:
		swapPropertyValues(props, "north", "south")
	}

	for _, key := range []string{"facing", "direction", "horizontal_facing"} {
		if value, ok := props[key]; ok {
			props[key] = mirrorHorizontalDirectionName(value, mirror)
		}
	}
	if value, ok := props["rotation"]; ok {
		props["rotation"] = mirrorStructureRotationProperty(value, mirror)
	}
	if value, ok := props["shape"]; ok {
		props["shape"] = mirrorShapeProperty(value, mirror)
	}
	if value, ok := props["orientation"]; ok {
		props["orientation"] = mirrorOrientationProperty(value, mirror)
	}

	return out
}

func swapPropertyValues(properties map[string]string, a, b string) {
	if _, ok := properties[a]; !ok {
		return
	}
	if _, ok := properties[b]; !ok {
		return
	}
	properties[a], properties[b] = properties[b], properties[a]
}

func mirrorHorizontalDirectionName(value string, mirror structureMirror) string {
	switch mirror {
	case structureMirrorFrontBack:
		switch value {
		case "east":
			return "west"
		case "west":
			return "east"
		}
	case structureMirrorLeftRight:
		switch value {
		case "north":
			return "south"
		case "south":
			return "north"
		}
	}
	return value
}

func mirrorStructureRotationProperty(value string, mirror structureMirror) string {
	n, err := strconv.Atoi(value)
	if err != nil {
		return value
	}
	n = n % 16
	switch mirror {
	case structureMirrorFrontBack:
		n = (16 - n) % 16
	case structureMirrorLeftRight:
		n = (8 - n + 16) % 16
	}
	return strconv.Itoa(n)
}

func mirrorShapeProperty(value string, mirror structureMirror) string {
	switch value {
	case "ascending_east":
		if mirror == structureMirrorFrontBack {
			return "ascending_west"
		}
	case "ascending_west":
		if mirror == structureMirrorFrontBack {
			return "ascending_east"
		}
	case "ascending_north":
		if mirror == structureMirrorLeftRight {
			return "ascending_south"
		}
	case "ascending_south":
		if mirror == structureMirrorLeftRight {
			return "ascending_north"
		}
	case "south_east":
		if mirror == structureMirrorFrontBack {
			return "south_west"
		}
		if mirror == structureMirrorLeftRight {
			return "north_east"
		}
	case "south_west":
		if mirror == structureMirrorFrontBack {
			return "south_east"
		}
		if mirror == structureMirrorLeftRight {
			return "north_west"
		}
	case "north_east":
		if mirror == structureMirrorFrontBack {
			return "north_west"
		}
		if mirror == structureMirrorLeftRight {
			return "south_east"
		}
	case "north_west":
		if mirror == structureMirrorFrontBack {
			return "north_east"
		}
		if mirror == structureMirrorLeftRight {
			return "south_west"
		}
	}
	return value
}

func mirrorOrientationProperty(value string, mirror structureMirror) string {
	parts := strings.SplitN(value, "_", 2)
	if len(parts) != 2 {
		return value
	}
	return mirrorHorizontalDirectionName(parts[0], mirror) + "_" + parts[1]
}
