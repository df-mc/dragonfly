package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"sort"
)

// pistonResolver ...
type pistonResolver struct {
	w   *world.World
	pos cube.Pos

	attachedPositions []cube.Pos
	breakPositions    []cube.Pos

	history map[cube.Pos]struct{}

	success bool
}

// pistonResolve ...
func pistonResolve(w *world.World, pos cube.Pos, piston Piston, push bool) *pistonResolver {
	r := &pistonResolver{
		w:   w,
		pos: pos,

		history: make(map[cube.Pos]struct{}),
	}

	face := piston.armFace()
	if push {
		if r.calculateBlocks(r.pos.Side(face), face, face) {
			r.success = true
		}
	} else {
		if piston.Sticky {
			r.calculateBlocks(r.pos.Side(face).Side(face), face, face.Opposite())
		}
		r.success = true
	}
	sort.SliceStable(r.attachedPositions, func(i, j int) bool {
		posOne := r.attachedPositions[i]
		posTwo := r.attachedPositions[j]

		pushI := 1
		if !push {
			pushI = -1
		}

		positive := 1
		if !face.Positive() {
			positive = -1
		}

		offset := posOne.Sub(posTwo)
		direction := pushI * positive
		switch face.Axis() {
		case cube.Y:
			return offset.Y()*direction > 0
		case cube.Z:
			return offset.Z()*direction > 0
		case cube.X:
			return offset.X()*direction > 0
		}
		panic("should never happen")
	})
	return r
}

// calculateBlocks ...
func (r *pistonResolver) calculateBlocks(pos cube.Pos, face cube.Face, breakFace cube.Face) bool {
	if pos.Side(breakFace).OutOfBounds(r.w.Range()) {
		r.breakPositions = nil
		r.attachedPositions = nil
		return false
	}
	if _, ok := r.history[pos]; ok {
		return true
	}
	r.history[pos] = struct{}{}

	block := r.w.Block(pos)
	if _, ok := block.(Air); ok {
		return true
	}
	if !r.canMove(pos, block) {
		if face == breakFace {
			r.breakPositions = nil
			r.attachedPositions = nil
			return false
		}
		return true
	}
	if r.canBreak(block) {
		if face == breakFace {
			r.breakPositions = append(r.breakPositions, pos)
		}
		return true
	}

	if _, ok := block.(GlazedTerracotta); ok && face != breakFace {
		// Glazed terracotta can't be pushed, but can be pulled.
		return true
	}

	r.attachedPositions = append(r.attachedPositions, pos)
	if len(r.attachedPositions) >= 13 {
		r.breakPositions = nil
		r.attachedPositions = nil
		return false
	}

	if _, ok := block.(Slime); ok {
		for _, otherFace := range cube.Faces() {
			if face.Opposite() == otherFace {
				continue
			}
			if !r.calculateBlocks(pos.Side(otherFace), otherFace, breakFace) {
				return false
			}
		}
	} else if !r.calculateBlocks(pos.Side(breakFace), breakFace, breakFace) {
		return false
	}
	return true
}

// canMove ...
func (r *pistonResolver) canMove(pos cube.Pos, block world.Block) bool {
	if p, ok := block.(Piston); ok {
		if r.pos == pos {
			return false
		}
		return p.State == 0
	}
	p, ok := block.(PistonImmovable)
	return !ok || !p.PistonImmovable()
}

// canBreak ...
func (r *pistonResolver) canBreak(block world.Block) bool {
	if l, ok := block.(LiquidRemovable); ok && l.HasLiquidDrops() {
		return true
	}
	p, ok := block.(PistonBreakable)
	return ok && p.PistonBreakable()
}
