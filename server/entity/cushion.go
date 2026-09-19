package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	cushionWidth, cushionHeight = 0.999, 0.249
	// cushionDismountHeight puts a rider getting off just above the top of the cushion.
	cushionDismountHeight = cushionHeight + 0.025
)

// cushionSeat is the seat offset sent to clients: The 0.1875 seat height plus the vanilla offset of a player rider.
var cushionSeat = mgl64.Vec3{0, 1.40751, 0}

// NewCushion creates a cushion entity with the colour passed.
func NewCushion(opts world.EntitySpawnOpts, colour item.Colour) *world.EntityHandle {
	return opts.New(CushionType, CushionConfig{Colour: colour})
}

// CushionConfig holds the configuration of a cushion entity.
type CushionConfig struct {
	// Colour is the colour of the cushion.
	Colour item.Colour
}

// Apply applies the cushion configuration to data.
func (conf CushionConfig) Apply(data *world.EntityData) {
	data.Data = &cushionBehaviour{colour: conf.Colour}
	data.AlwaysShowNameTag = false
}

// CushionType is a world.EntityType implementation for cushions.
var CushionType cushionType

type cushionType struct{}

func (cushionType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Cushion{Ent: Open(tx, handle, data)}
}

func (cushionType) EncodeEntity() string { return "minecraft:cushion" }

func (cushionType) BBox(world.Entity) cube.BBox {
	return cube.Box(-cushionWidth/2, 0, -cushionWidth/2, cushionWidth/2, cushionHeight, cushionWidth/2)
}

func (cushionType) DecodeNBT(m map[string]any, data *world.EntityData) {
	variant := int32(15)
	if v, ok := m["Variant"].(int32); ok {
		variant = v
	}
	CushionConfig{Colour: cushionColour(variant)}.Apply(data)
}

func (cushionType) EncodeNBT(data *world.EntityData) map[string]any {
	return map[string]any{"Variant": data.Data.(*cushionBehaviour).Variant()}
}

// cushionColour returns the colour of a cushion variant. Variants count down from white (15) to black (0).
func cushionColour(variant int32) item.Colour {
	return item.Colours()[15-uint8(variant)&15]
}

// Cushion is an entity that rests on top of a block and may be sat on by players.
type Cushion struct {
	*Ent
}

func (c *Cushion) behaviour() *cushionBehaviour {
	return c.Behaviour().(*cushionBehaviour)
}

// Colour returns the colour of the cushion.
func (c *Cushion) Colour() item.Colour {
	return c.behaviour().colour
}

// Break breaks the cushion, dropping it as an item.
func (c *Cushion) Break() {
	c.behaviour().destroy(c.Ent, c.tx, true)
}

// SeatPositions ...
func (c *Cushion) SeatPositions() []mgl64.Vec3 {
	return []mgl64.Vec3{cushionSeat}
}

// NextFreeSeatIndex ...
func (c *Cushion) NextFreeSeatIndex(mgl64.Vec3) (int, bool) {
	return 0, c.behaviour().rider == nil
}

// AcceptsRider refuses sneaking entities, which cannot sit down on a cushion.
func (c *Cushion) AcceptsRider(rider world.Entity) bool {
	s, ok := rider.(interface{ Sneaking() bool })
	return !ok || !s.Sneaking()
}

// ControllingRider returns the entity sitting on the cushion, which vanilla links as its controlling rider.
func (c *Cushion) ControllingRider() *world.EntityHandle {
	return c.behaviour().rider
}

// ControllingSeatIndex ...
func (c *Cushion) ControllingSeatIndex() int {
	return 0
}

// Riders ...
func (c *Cushion) Riders() []RiderSeat {
	if r := c.behaviour().rider; r != nil {
		return []RiderSeat{{Handle: r, SeatIndex: 0}}
	}
	return nil
}

// AddRider ...
func (c *Cushion) AddRider(rider *world.EntityHandle, seatIndex int) bool {
	b := c.behaviour()
	if seatIndex != 0 || b.rider != nil || b.removed {
		return false
	}
	b.rider = rider
	c.tx.PlaySound(c.Position(), sound.CushionSit{})
	return true
}

// RemoveRider ...
func (c *Cushion) RemoveRider(rider *world.EntityHandle) {
	if b := c.behaviour(); b.rider == rider {
		b.rider = nil
	}
}

// MoveInput does nothing, as a cushion cannot move.
func (c *Cushion) MoveInput(mgl64.Vec2, float32, float32) {}

// DismountPosition returns the top centre of the cushion.
func (c *Cushion) DismountPosition(world.Entity) mgl64.Vec3 {
	return c.Position().Add(mgl64.Vec3{0, cushionDismountHeight, 0})
}

// SeatRotation turns the rider 90 degrees to the left without locking its rotation.
func (c *Cushion) SeatRotation(int) SeatRotation {
	return SeatRotation{RotateBy: -90}
}

// HoldsRiders ...
func (c *Cushion) HoldsRiders() bool {
	return true
}

// InteractText returns the sit button text, shown only while the cushion is free and the user is not sneaking.
func (c *Cushion) InteractText(user world.Entity) string {
	if c.behaviour().rider != nil || !c.AcceptsRider(user) {
		return ""
	}
	return "action.interact.ride.cushion"
}

// Pick returns the cushion item of the same colour as the cushion.
func (c *Cushion) Pick() item.Stack {
	return item.NewStack(item.Cushion{Colour: c.Colour()}, 1)
}
