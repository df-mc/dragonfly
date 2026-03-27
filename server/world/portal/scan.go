package portal

import (
	"container/list"
	"strings"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type scanIteration struct {
	lastPos cube.Pos
	face    cube.Face
	first   bool
}

func multiAxisScan(framePos cube.Pos, tx *world.Tx, matchers []string) (cube.Axis, []cube.Pos, int, int, bool, bool) {
	positions, width, height, completed := scan(cube.Z, framePos, tx, matchers)
	positionsTwo, widthTwo, heightTwo, completedTwo := scan(cube.X, framePos, tx, matchers)
	if len(positions) < minimumArea && len(positionsTwo) >= minimumArea {
		return cube.X, positionsTwo, widthTwo, heightTwo, completedTwo, len(positionsTwo) > 0
	}
	return cube.Z, positions, width, height, completed, len(positions) > 0
}

func scan(axis cube.Axis, framePos cube.Pos, tx *world.Tx, matchers []string) ([]cube.Pos, int, int, bool) {
	var width, height int
	positionsMap := make(map[cube.Pos]bool)

	completed := true
	queue := list.New()
	queue.PushBack(scanIteration{lastPos: framePos, first: true})
	for queue.Len() > 0 {
		e := queue.Front()
		queue.Remove(e)

		iteration := e.Value.(scanIteration)
		pos := iteration.lastPos
		if !iteration.first {
			pos = pos.Side(iteration.face)
		}

		b := tx.Block(pos)
		if _, ok := positionsMap[pos]; !ok && satisfiesMatchers(b, matchers) {
			positionsMap[pos] = true

			if pos.X() == framePos.X() && pos.Z() == framePos.Z() && iteration.face < cube.FaceNorth {
				height++
			}
			if pos.Y() == framePos.Y() {
				width++
			}
			if width > maximumNetherPortalWidth || height > maximumNetherPortalHeight {
				return nil, 0, 0, false
			}

			if axis == cube.Z {
				queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceSouth})
				queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceNorth})
			} else if axis == cube.X {
				queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceWest})
				queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceEast})
			}
			queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceUp})
			queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceDown})
		} else if _, ok = positionsMap[pos]; !(ok || isPortalObsidian(b)) {
			completed = false
		}
	}

	area, expectedArea := len(positionsMap), width*height
	completed = completed && width >= minimumNetherPortalWidth && height >= minimumNetherPortalHeight && area == expectedArea

	positions := make([]cube.Pos, 0, expectedArea)
	for pos := range positionsMap {
		positions = append(positions, pos)
	}
	return positions, width, height, completed
}

func satisfiesMatchers(b world.Block, matchers []string) bool {
	name, _ := b.EncodeBlock()
	name = normalizeBlockName(name)
	for _, matcher := range matchers {
		if name == normalizeBlockName(matcher) {
			return true
		}
	}
	return false
}

func normalizeBlockName(name string) string {
	return strings.TrimPrefix(name, "minecraft:")
}

func air() world.Block {
	a, ok := world.BlockByName("minecraft:air", nil)
	if !ok {
		panic("could not find air block")
	}
	return a
}

func portal(axis cube.Axis) world.Block {
	p, ok := world.BlockByName("minecraft:portal", map[string]any{"portal_axis": axis.String()})
	if !ok {
		panic("could not find portal block")
	}
	return p
}

func obsidian() world.Block {
	o, ok := world.BlockByName("minecraft:obsidian", nil)
	if !ok {
		panic("could not find obsidian block")
	}
	return o
}

func isPortalObsidian(b world.Block) bool {
	name, _ := b.EncodeBlock()
	return normalizeBlockName(name) == "obsidian"
}
