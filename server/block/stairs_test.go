package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type stairsTestSource map[cube.Pos]world.Block

func (s stairsTestSource) Block(pos cube.Pos) world.Block {
	if b, ok := s[pos]; ok {
		return b
	}
	return Air{}
}

// TestStairsCorner checks the corner calculated for every layout of neighbouring stairs, for each facing.
func TestStairsCorner(t *testing.T) {
	oak, stone := Planks{Wood: OakWood()}, Cobblestone{}
	type neighbour struct {
		// side is the side of the stairs the neighbour is on, relative to their facing: 0 front, 1 right, 2 back, 3 left.
		side int
		// facing is the facing of the neighbour relative to the facing of the stairs, in right turns.
		facing     int
		upsideDown bool
		block      world.Block
	}
	tests := []struct {
		name       string
		neighbours []neighbour
		want       StairsCorner
	}{
		{"none", nil, NoStairsCorner()},
		{"outer left", []neighbour{{side: 0, facing: 3}}, OuterLeftStairsCorner()},
		{"outer right", []neighbour{{side: 0, facing: 1}}, OuterRightStairsCorner()},
		{"inner left", []neighbour{{side: 2, facing: 3}}, InnerLeftStairsCorner()},
		{"inner right", []neighbour{{side: 2, facing: 1}}, InnerRightStairsCorner()},
		{"outer left continued right", []neighbour{{side: 0, facing: 3}, {side: 1}}, NoStairsCorner()},
		{"outer right continued left", []neighbour{{side: 0, facing: 1}, {side: 3}}, NoStairsCorner()},
		{"inner left continued left", []neighbour{{side: 2, facing: 3}, {side: 3}}, NoStairsCorner()},
		{"inner right continued right", []neighbour{{side: 2, facing: 1}, {side: 1}}, NoStairsCorner()},
		{"outer left not continued by left side", []neighbour{{side: 0, facing: 3}, {side: 3}}, OuterLeftStairsCorner()},
		{"outer right not continued by right side", []neighbour{{side: 0, facing: 1}, {side: 1}}, OuterRightStairsCorner()},
		{"inner left not continued by right side", []neighbour{{side: 2, facing: 3}, {side: 1}}, InnerLeftStairsCorner()},
		{"inner right not continued by left side", []neighbour{{side: 2, facing: 1}, {side: 3}}, InnerRightStairsCorner()},
		{"continued outer falls through to inner", []neighbour{{side: 0, facing: 3}, {side: 1}, {side: 2, facing: 3}}, InnerLeftStairsCorner()},
		{"continued by other block type", []neighbour{{side: 0, facing: 3}, {side: 1, block: stone}}, OuterLeftStairsCorner()},
		{"continued by other half", []neighbour{{side: 0, facing: 1}, {side: 3, upsideDown: true}}, OuterRightStairsCorner()},
		{"continued by other facing", []neighbour{{side: 0, facing: 1}, {side: 3, facing: 2}}, OuterRightStairsCorner()},
		{"corner with other block type", []neighbour{{side: 0, facing: 1, block: stone}}, OuterRightStairsCorner()},
		{"no corner with other half", []neighbour{{side: 0, facing: 1, upsideDown: true}}, NoStairsCorner()},
		{"no corner with parallel stairs", []neighbour{{side: 0, facing: 2}, {side: 2}}, NoStairsCorner()},
	}
	for _, tt := range tests {
		for _, facing := range cube.Directions() {
			s := Stairs{Block: oak, Facing: facing}
			src := stairsTestSource{}
			for _, n := range tt.neighbours {
				b := n.block
				if b == nil {
					b = oak
				}
				src[cube.Pos{}.Side(rotateRight(facing, n.side).Face())] = Stairs{Block: b, Facing: rotateRight(facing, n.facing), UpsideDown: n.upsideDown}
			}
			if got := s.calculateCorner(src, cube.Pos{}); got != tt.want {
				t.Errorf("%v facing %v: got corner %v, want %v", tt.name, facing, got, tt.want)
			}
		}
	}
}

func rotateRight(d cube.Direction, n int) cube.Direction {
	for range n {
		d = d.RotateRight()
	}
	return d
}
