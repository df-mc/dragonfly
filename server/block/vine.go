package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// Vines are climbable non-solid vegetation blocks that grow on walls.
type Vines struct {
	replaceable
	transparent
	empty
	sourceWaterDisplacer

	// NorthDirection is true if the vines are attached towards north.
	NorthDirection bool
	// EastDirection is true if the vines are attached towards east.
	EastDirection bool
	// SouthDirection is true if the vines are attached towards south.
	SouthDirection bool
	// WestDirection is true if the vines are attached towards west.
	WestDirection bool
}

// CompostChance ...
func (v Vines) CompostChance() float64 {
	return 0.5
}

// SideClosed ...
func (v Vines) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// HasLiquidDrops ...
func (v Vines) HasLiquidDrops() bool {
	return false
}

// FlammabilityInfo ...
func (Vines) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(15, 100, true)
}

// BreakInfo ...
func (v Vines) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, func(t item.Tool) bool {
		return t.ToolType() == item.TypeShears || t.ToolType() == item.TypeAxe
	}, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if t.ToolType() == item.TypeShears {
			return []item.Stack{item.NewStack(v, 1)}
		}
		return nil
	})
}

// EntityInside ...
func (Vines) EntityInside(_ cube.Pos, _ *world.Tx, e world.Entity) {
	if fallEntity, ok := e.(fallDistanceEntity); ok {
		fallEntity.ResetFallDistance()
	}
}

// SetAttachment sets an attachment on the given cube.Direction.
func (v Vines) SetAttachment(direction cube.Direction, attached bool) Vines {
	switch direction {
	case cube.North:
		v.NorthDirection = attached
		return v
	case cube.East:
		v.EastDirection = attached
		return v
	case cube.South:
		v.SouthDirection = attached
		return v
	case cube.West:
		v.WestDirection = attached
		return v
	}
	panic("should never happen")
}

// Attachment returns the attachment of the vines at the given direction.
func (v Vines) Attachment(direction cube.Direction) bool {
	switch direction {
	case cube.North:
		return v.NorthDirection
	case cube.East:
		return v.EastDirection
	case cube.South:
		return v.SouthDirection
	case cube.West:
		return v.WestDirection
	}
	panic("should never happen")
}

// Attachments returns all attachments of the vines.
func (v Vines) Attachments() (attachments []cube.Direction) {
	for _, d := range cube.Directions() {
		if v.Attachment(d) {
			attachments = append(attachments, d)
		}
	}
	return
}

// UseOnBlock ...
func (v Vines) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	if _, ok := tx.Block(pos).Model().(model.Solid); !ok || face.Axis() == cube.Y {
		return false
	}
	pos, face, used := firstReplaceable(tx, pos, face, v)
	if !used {
		return false
	}
	//noinspection GoAssignmentToReceiver
	v = v.SetAttachment(face.Direction().Opposite(), true)

	place(tx, pos, v, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (v Vines) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	above, updated := tx.Block(pos.Side(cube.FaceUp)), false
	for _, d := range v.Attachments() {
		if _, ok := tx.Block(pos.Side(d.Face())).Model().(model.Solid); !ok {
			if o, ok := above.(Vines); !ok || ok && !o.Attachment(d) {
				//noinspection GoAssignmentToReceiver
				v = v.SetAttachment(d, false)
				updated = true
			}
		}
	}
	if !updated {
		return
	}
	if _, ok := above.Model().(model.Solid); !ok && len(v.Attachments()) == 0 {
		tx.SetBlock(pos, nil, nil)
		return
	}
	tx.SetBlock(pos, v, nil)
}

// RandomTick ...
func (v Vines) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if r.Float64() > 0.25 {
		// Vines have a 25% chance of spreading.
		return
	}

	face := cube.Face(r.Intn(len(cube.Faces())))
	selectedPos := pos.Side(face)
	if selectedPos.OutOfBounds(tx.Range()) {
		return
	}

	if face.Axis() != cube.Y && !v.Attachment(face.Direction()) {
		if !v.canSpread(tx, pos) {
			return
		}
		if _, ok := tx.Block(selectedPos).(Air); ok {
			rightRotatedFace := face.RotateRight()
			leftRotatedFace := face.RotateLeft()

			attachedOnRight := v.Attachment(rightRotatedFace.Direction())
			attachedOnLeft := v.Attachment(leftRotatedFace.Direction())

			rightSelectedPos := selectedPos.Side(rightRotatedFace)
			leftSelectedPos := selectedPos.Side(leftRotatedFace)

			if attachedOnRight && v.canSpreadTo(tx, rightSelectedPos) {
				tx.SetBlock(selectedPos, (Vines{}).SetAttachment(rightRotatedFace.Direction(), true), nil)
			} else if attachedOnLeft && v.canSpreadTo(tx, leftSelectedPos) {
				tx.SetBlock(selectedPos, (Vines{}).SetAttachment(leftRotatedFace.Direction(), true), nil)
			} else if _, ok = tx.Block(rightSelectedPos).(Air); ok && attachedOnRight && v.canSpreadTo(tx, pos.Side(rightRotatedFace)) {
				tx.SetBlock(rightSelectedPos, (Vines{}).SetAttachment(face.Opposite().Direction(), true), nil)
			} else if _, ok = tx.Block(leftSelectedPos).(Air); ok && attachedOnLeft && v.canSpreadTo(tx, pos.Side(leftRotatedFace)) {
				tx.SetBlock(leftSelectedPos, (Vines{}).SetAttachment(face.Opposite().Direction(), true), nil)
			}
		} else if v.canSpreadTo(tx, selectedPos) {
			tx.SetBlock(pos, v.SetAttachment(face.Direction(), true), nil)
		}
		return
	}

	_, air := tx.Block(selectedPos).(Air)
	newVines := tx.Block(selectedPos).(Vines)
	if face == cube.FaceUp {
		if air {
			if !v.canSpread(tx, pos) {
				return
			}
			for _, f := range cube.HorizontalFaces() {
				if r.Intn(2) == 0 && v.canSpreadTo(tx, selectedPos.Side(f)) {
					newVines = newVines.SetAttachment(f.Direction(), v.Attachment(f.Direction()))
				}
			}
			if len(newVines.Attachments()) > 0 {
				tx.SetBlock(selectedPos, newVines, nil)
			}
			return
		}
	}

	selectedPos = pos.Side(cube.FaceDown)
	if selectedPos.OutOfBounds(tx.Range()) {
		return
	}
	var changed bool
	for _, f := range cube.HorizontalFaces() {
		if r.Intn(2) == 0 && v.Attachment(f.Direction()) && !newVines.Attachment(f.Direction()) {
			newVines, changed = newVines.SetAttachment(f.Direction(), true), true
		}
	}
	if changed {
		tx.SetBlock(selectedPos, newVines, nil)
	}
}

// EncodeItem ...
func (Vines) EncodeItem() (name string, meta int16) {
	return "minecraft:vine", 0
}

// EncodeBlock ...
func (v Vines) EncodeBlock() (string, map[string]any) {
	var bits int
	for i, ok := range []bool{v.SouthDirection, v.WestDirection, v.NorthDirection, v.EastDirection} {
		if ok {
			bits |= 1 << i
		}
	}
	return "minecraft:vine", map[string]any{"vine_direction_bits": int32(bits)}
}

// canSpreadTo returns true if the vines can spread onto the given position.
func (Vines) canSpreadTo(tx *world.Tx, pos cube.Pos) bool {
	_, ok := tx.Block(pos).Model().(model.Solid)
	return ok
}

// canSpread returns true if the vines can spread from the given position.
func (v Vines) canSpread(tx *world.Tx, pos cube.Pos) bool {
	var count int
	for x := -4; x <= 4; x++ {
		for z := -4; z <= 4; z++ {
			for y := -1; y <= 1; y++ {
				if _, ok := tx.Block(pos.Add(cube.Pos{x, y, z})).(Vines); ok {
					count++
					if count >= 5 {
						return false
					}
				}
			}
		}
	}
	return true
}

// allVines ...
func allVines() (b []world.Block) {
	for _, north := range []bool{true, false} {
		for _, east := range []bool{true, false} {
			for _, south := range []bool{true, false} {
				for _, west := range []bool{true, false} {
					b = append(b, Vines{
						NorthDirection: north,
						EastDirection:  east,
						SouthDirection: south,
						WestDirection:  west,
					})
				}
			}
		}
	}
	return
}
