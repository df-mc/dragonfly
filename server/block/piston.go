package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// Piston is a block capable of pushing blocks, players, and mobs when given a redstone pulse.
type Piston struct {
	solid
	transparent
	sourceWaterDisplacer

	// Facing represents the direction the piston is facing.
	Facing cube.Face
	// Sticky is true if the piston is sticky, false if not.
	Sticky bool

	// AttachedBlocks ...
	AttachedBlocks []cube.Pos
	// BreakBlocks ...
	BreakBlocks []cube.Pos

	// Progress is how far the block has been moved. It can either be 0.0, 0.5, or 1.0.
	Progress float64
	// LastProgress ...
	LastProgress float64

	// State ...
	State int
	// NewState ...
	NewState int
}

// PistonImmovable represents a block that cannot be moved by a piston.
type PistonImmovable interface {
	// PistonImmovable returns whether the block is immovable.
	PistonImmovable() bool
}

// PistonBreakable represents a block that can be broken by a piston.
type PistonBreakable interface {
	// PistonBreakable returns whether the block can be broken by a piston.
	PistonBreakable() bool
}

// PistonUpdater represents a block that can be updated through a piston movement.
type PistonUpdater interface {
	// PistonUpdate is called when a piston moves the block.
	PistonUpdate(pos cube.Pos, tx *world.Tx)
}

// BreakInfo ...
func (p Piston) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, pickaxeEffective, oneOf(Piston{Sticky: p.Sticky}))
}

// SideClosed ...
func (Piston) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeItem ...
func (p Piston) EncodeItem() (name string, meta int16) {
	if p.Sticky {
		return "minecraft:sticky_piston", 0
	}
	return "minecraft:piston", 0
}

// EncodeBlock ...
func (p Piston) EncodeBlock() (string, map[string]any) {
	if p.Sticky {
		return "minecraft:sticky_piston", map[string]any{"facing_direction": int32(p.Facing)}
	}
	return "minecraft:piston", map[string]any{"facing_direction": int32(p.Facing)}
}

// UseOnBlock ...
func (p Piston) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, p)
	if !used {
		return false
	}
	p.Facing = calculateFace(user, pos, false)

	place(tx, pos, p, user, ctx)
	if placed(ctx) {
		p.RedstoneUpdate(pos, tx)
		return true
	}
	return false
}

// EncodeNBT ...
func (p Piston) EncodeNBT() map[string]any {
	attachedBlocks := make([]int32, 0, len(p.AttachedBlocks)*3)
	for _, pos := range p.AttachedBlocks {
		attachedBlocks = append(attachedBlocks, int32(pos[0]), int32(pos[1]), int32(pos[2]))
	}
	breakBlocks := make([]int32, 0, len(p.BreakBlocks)*3)
	for _, pos := range p.BreakBlocks {
		breakBlocks = append(breakBlocks, int32(pos[0]), int32(pos[1]), int32(pos[2]))
	}
	return map[string]any{
		"AttachedBlocks": attachedBlocks,
		"BreakBlocks":    breakBlocks,

		"LastProgress": float32(p.LastProgress),
		"Progress":     float32(p.Progress),

		"NewState": uint8(p.NewState),
		"State":    uint8(p.State),

		"Sticky": p.Sticky,

		"id": "PistonArm",
	}
}

// DecodeNBT ...
func (p Piston) DecodeNBT(m map[string]any) any {
	if attached := nbtconv.Slice(m, "AttachedBlocks"); attached != nil {
		p.AttachedBlocks = make([]cube.Pos, 0, len(attached)/3)
		for i := 0; i < len(attached); i += 3 {
			p.AttachedBlocks = append(p.AttachedBlocks, cube.Pos{
				int(attached[i].(int32)),
				int(attached[i+1].(int32)),
				int(attached[i+2].(int32)),
			})
		}
	}
	if breakBlocks := nbtconv.Slice(m, "BreakBlocks"); breakBlocks != nil {
		p.BreakBlocks = make([]cube.Pos, 0, len(breakBlocks)/3)
		for i := 0; i < len(breakBlocks); i += 3 {
			p.BreakBlocks = append(p.BreakBlocks, cube.Pos{
				int(breakBlocks[i].(int32)),
				int(breakBlocks[i+1].(int32)),
				int(breakBlocks[i+2].(int32)),
			})
		}
	}
	p.LastProgress = float64(nbtconv.Float32(m, "LastProgress"))
	p.Progress = float64(nbtconv.Float32(m, "Progress"))
	p.NewState = int(nbtconv.Uint8(m, "NewState"))
	p.State = int(nbtconv.Uint8(m, "State"))
	p.Sticky = nbtconv.Bool(m, "Sticky")
	return p
}

// RedstoneUpdate ...
func (p Piston) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
	tx.ScheduleBlockUpdate(pos, p, time.Millisecond*50)
}

// ScheduledTick ...
func (p Piston) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if receivedRedstonePower(pos, tx, p.armFace()) {
		if !p.push(pos, tx) {
			return
		}
	} else if !p.pull(pos, tx) {
		return
	}
	tx.ScheduleBlockUpdate(pos, p, time.Millisecond*50)
}

// armFace ...
func (p Piston) armFace() cube.Face {
	if p.Facing.Axis() == cube.Y {
		return p.Facing
	}
	return p.Facing.Opposite()
}

// push ...
func (p Piston) push(pos cube.Pos, tx *world.Tx) bool {
	if p.State == 0 {
		resolver := pistonResolve(tx, pos, p, true)
		if !resolver.success {
			return false
		}

		for _, breakPos := range resolver.breakPositions {
			p.BreakBlocks = append(p.BreakBlocks, breakPos)
			if b, ok := tx.Block(breakPos).(Breakable); ok {
				tx.SetBlock(breakPos, nil, nil)
				for _, drop := range b.BreakInfo().Drops(item.ToolNone{}, nil) {
					dropItem(tx, drop, breakPos.Vec3Centre())
				}
			}
		}

		face := p.armFace()
		for _, attachedPos := range resolver.attachedPositions {
			side := attachedPos.Side(face)
			p.AttachedBlocks = append(p.AttachedBlocks, attachedPos)

			tx.SetBlock(side, Moving{Piston: pos, Moving: tx.Block(attachedPos), Expanding: true}, nil)
			tx.SetBlock(attachedPos, nil, nil)
			updateAroundRedstone(attachedPos, tx)
		}

		p.State = 1
		tx.SetBlock(pos.Side(face), PistonArmCollision{Facing: p.Facing}, nil)
	} else if p.State == 1 {
		if p.Progress == 1 {
			p.State = 2
		}
		p.LastProgress = p.Progress

		if p.State == 1 {
			p.Progress += 0.5
			if p.Progress == 0.5 {
				tx.PlaySound(pos.Vec3Centre(), sound.PistonExtend{})
			}
		}

		if p.State == 2 {
			face := p.armFace()
			for _, attachedPos := range p.AttachedBlocks {
				side := attachedPos.Side(face)
				moving, ok := tx.Block(side).(Moving)
				if !ok {
					continue
				}

				tx.SetBlock(side, moving.Moving, nil)
				if u, ok := moving.Moving.(RedstoneUpdater); ok {
					u.RedstoneUpdate(side, tx)
				}
				if u, ok := moving.Moving.(PistonUpdater); ok {
					u.PistonUpdate(side, tx)
				}
				updateAroundRedstone(side, tx)
			}

			p.AttachedBlocks = nil
			p.BreakBlocks = nil
		}
	} else if p.State == 3 {
		return p.pull(pos, tx)
	} else {
		return false
	}

	p.NewState = p.State
	tx.SetBlock(pos, p, nil)
	return true
}

// pull ...
func (p Piston) pull(pos cube.Pos, tx *world.Tx) bool {
	if p.State == 2 {
		face := p.armFace()
		tx.SetBlock(pos.Side(face), nil, nil)

		resolver := pistonResolve(tx, pos, p, false)
		if !resolver.success {
			return false
		}

		for _, breakPos := range resolver.breakPositions {
			p.BreakBlocks = append(p.BreakBlocks, breakPos)
			if b, ok := tx.Block(breakPos).(Breakable); ok {
				tx.SetBlock(breakPos, nil, nil)
				for _, drop := range b.BreakInfo().Drops(item.ToolNone{}, nil) {
					dropItem(tx, drop, breakPos.Vec3Centre())
				}
			}
		}

		face = face.Opposite()
		for _, attachedPos := range resolver.attachedPositions {
			side := attachedPos.Side(face)
			p.AttachedBlocks = append(p.AttachedBlocks, attachedPos)

			tx.SetBlock(side, Moving{Piston: pos, Moving: tx.Block(attachedPos)}, nil)
			tx.SetBlock(attachedPos, nil, nil)
			updateAroundRedstone(attachedPos, tx)
		}

		p.State = 3
	} else if p.State == 3 {
		if p.Progress == 0 {
			p.State = 0
		}
		p.LastProgress = p.Progress

		if p.State == 3 {
			p.Progress -= 0.5
			if p.Progress == 0.5 {
				tx.PlaySound(pos.Vec3Centre(), sound.PistonRetract{})
			}
		}

		if p.State == 0 {
			face := p.armFace()
			for _, attachedPos := range p.AttachedBlocks {
				side := attachedPos.Side(face.Opposite())
				moving, ok := tx.Block(side).(Moving)
				if !ok {
					continue
				}

				tx.SetBlock(side, moving.Moving, nil)
				if r, ok := moving.Moving.(RedstoneUpdater); ok {
					r.RedstoneUpdate(side, tx)
				}
				if r, ok := moving.Moving.(PistonUpdater); ok {
					r.PistonUpdate(side, tx)
				}
				updateAroundRedstone(side, tx)
			}

			p.AttachedBlocks = nil
			p.BreakBlocks = nil
		}
	} else if p.State == 1 {
		return p.push(pos, tx)
	} else {
		return false
	}

	p.NewState = p.State
	tx.SetBlock(pos, p, nil)
	return true
}

// allPistons ...
func allPistons() (pistons []world.Block) {
	for _, f := range cube.Faces() {
		for _, s := range []bool{false, true} {
			pistons = append(pistons, Piston{Facing: f, Sticky: s})
		}
	}
	return
}
