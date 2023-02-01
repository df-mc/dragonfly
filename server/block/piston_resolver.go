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

	sticky bool
	push   bool

	attachedBlocks []cube.Pos
	breakBlocks    []cube.Pos

	checked map[cube.Pos]struct{}

	success bool
}

// newPistonResolver ...
func newPistonResolver(w *world.World, pos cube.Pos, sticky, push bool) *pistonResolver {
	return &pistonResolver{
		w:   w,
		pos: pos,

		sticky: sticky,
		push:   push,

		checked: make(map[cube.Pos]struct{}),
	}
}

// resolve ...
func (r *pistonResolver) resolve() {
	piston, ok := r.w.Block(r.pos).(Piston)
	if !ok {
		return
	}
	face := piston.armFace()
	if r.push {
		if r.calculateBlocks(r.pos.Side(face), face, face) {
			r.success = true
		}
	} else {
		if r.sticky {
			r.calculateBlocks(r.pos.Side(face).Side(face), face, face.Opposite())
		}
		r.success = true
	}
	sort.SliceStable(r.attachedBlocks, func(i, j int) bool {
		posOne := r.attachedBlocks[i]
		posTwo := r.attachedBlocks[j]

		push := 1
		if !r.push {
			push = -1
		}

		positive := 1
		if !face.Positive() {
			positive = -1
		}

		direction := push * positive
		switch face.Axis() {
		case cube.Y:
			return (posTwo.Y()-posOne.Y())*direction > 0
		case cube.Z:
			return (posTwo.Z()-posOne.Z())*direction > 0
		case cube.X:
			return (posTwo.X()-posOne.X())*direction > 0
		}
		panic("should never happen")
	})
}

// calculateBlocks ...
func (r *pistonResolver) calculateBlocks(pos cube.Pos, face cube.Face, breakFace cube.Face) bool {
	if _, ok := r.checked[pos]; ok {
		return true
	}
	r.checked[pos] = struct{}{}

	block := r.w.Block(pos)
	if _, ok := block.(Air); ok {
		return true
	}
	if !r.canMove(pos, block) {
		if face == breakFace {
			r.breakBlocks = nil
			r.attachedBlocks = nil
			return false
		}
		return true
	}
	if r.canBreak(block) {
		if face == breakFace {
			r.breakBlocks = append(r.breakBlocks, pos)
		}
		return true
	}
	if pos.Side(breakFace).OutOfBounds(r.w.Range()) {
		r.breakBlocks = nil
		r.attachedBlocks = nil
		return false
	}

	r.attachedBlocks = append(r.attachedBlocks, pos)
	if len(r.attachedBlocks) >= 13 {
		r.breakBlocks = nil
		r.attachedBlocks = nil
		return false
	}
	return r.calculateBlocks(pos.Side(breakFace), breakFace, breakFace)
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
