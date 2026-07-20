package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// PressurePlate is a non-solid block that emits redstone power while entities
// stand on it. Weighted variants emit an analog power level based on the
// number of entities on the plate.
type PressurePlate struct {
	empty
	transparent
	sourceWaterDisplacer

	// Type is the material the pressure plate is made of.
	Type PressurePlateType
	// Power is the current redstone signal emitted by the plate.
	Power int
}

// Model ...
func (PressurePlate) Model() world.BlockModel {
	return model.Empty{}
}

// UseOnBlock places the pressure plate on a solid surface.
func (p PressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, p)
	if !used || !redstoneFloorComponentSupported(tx, pos) {
		return false
	}
	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// EntityInside powers the plate when an entity enters its activation area.
func (p PressurePlate) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if p.entityPower(e) == 0 || !pressurePlateEntityIntersects(e, pressurePlateActivationBox(pos)) {
		return
	}
	if p.Power > 0 {
		// The plate is already active. Its scheduled tick keeps the (weighted)
		// level current, so a stepping entity only needs to defer the release.
		tx.ScheduleBlockUpdate(pos, p, p.releaseDelay())
		return
	}
	power := p.stepPower()
	if p.Type.Weighted() {
		power = max(power, p.detectPower(pos, tx))
	}
	p.Power = power
	tx.SetBlock(pos, p, nil)
	tx.ScheduleBlockUpdate(pos, p, p.releaseDelay())
	tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOn{})
}

// NeighbourUpdateTick breaks the pressure plate if its supporting block is removed.
func (p PressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !redstoneFloorComponentSupported(tx, pos) {
		breakBlock(p, pos, tx)
	}
}

// ScheduledTick releases the plate if no entity keeps it pressed.
func (p PressurePlate) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	power := p.detectPower(pos, tx)
	if power > 0 {
		if p.Power != power {
			p.Power = power
			tx.SetBlock(pos, p, nil)
		}
		tx.ScheduleBlockUpdate(pos, p, p.releaseDelay())
		return
	}
	if p.Power == 0 {
		return
	}
	p.Power = 0
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOff{})
}

// RedstonePower returns the plate's analog power level.
func (p PressurePlate) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	return p.Power
}

// RedstoneStrongPower strongly powers the block below the pressure plate.
func (p PressurePlate) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if face == cube.FaceDown {
		return p.Power
	}
	return 0
}

// BreakInfo ...
func (p PressurePlate) BreakInfo() BreakInfo {
	effective := pickaxeEffective
	if p.Type.Wood() {
		effective = axeEffective
	}
	return newBreakInfo(0.5, alwaysHarvestable, effective, oneOf(PressurePlate{Type: p.Type}))
}

// SideClosed ...
func (PressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// FuelInfo ...
func (p PressurePlate) FuelInfo() item.FuelInfo {
	if p.Type.Flammable() {
		return newFuelInfo(time.Second * 15)
	}
	return item.FuelInfo{}
}

// EncodeItem ...
func (p PressurePlate) EncodeItem() (name string, meta int16) {
	return "minecraft:" + p.Type.String(), 0
}

// EncodeBlock ...
func (p PressurePlate) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + p.Type.String(), map[string]any{"redstone_signal": int32(max(0, min(p.Power, 15)))}
}

// stepPower is the power a single detected entity contributes: the first
// analog level for weighted plates and full power otherwise.
func (p PressurePlate) stepPower() int {
	if p.Type.Weighted() {
		return 1
	}
	return 15
}

func (p PressurePlate) entityPower(e world.Entity) int {
	if !p.detectsEntity(e) {
		return 0
	}
	return p.stepPower()
}

// detectsEntity reports whether an entity activates the plate. Stone-like
// plates only react to living entities, players and armour stands; wooden and
// weighted plates react to any entity.
func (p PressurePlate) detectsEntity(e world.Entity) bool {
	if !p.Type.Wood() && !p.Type.Weighted() {
		return pressurePlateStoneEntity(e)
	}
	return true
}

// detectPower scans the entities intersecting the plate's activation box and
// returns the power level they produce.
func (p PressurePlate) detectPower(pos cube.Pos, tx *world.Tx) int {
	box := pressurePlateActivationBox(pos)
	entities := 0
	for e := range tx.EntitiesWithin(box.Grow(1)) {
		if p.entityPower(e) == 0 || !pressurePlateEntityIntersects(e, box) {
			continue
		}
		if !p.Type.Weighted() {
			return 15
		}
		entities++
		if entities >= p.weightedMaxEntities() {
			return 15
		}
	}
	if p.Type.Weighted() {
		return p.weightedPower(entities)
	}
	return 0
}

// weightedPower converts an entity count to the analog power of a weighted
// plate: one level per entity for light plates and per ten entities, rounded
// up, for heavy plates.
func (p PressurePlate) weightedPower(entities int) int {
	if entities <= 0 {
		return 0
	}
	if p.Type == HeavyWeightedPressurePlate() {
		return min(15, (entities+9)/10)
	}
	return min(15, entities)
}

// weightedMaxEntities is the entity count at which a weighted plate reaches
// full power, so scanning may stop early.
func (p PressurePlate) weightedMaxEntities() int {
	if p.Type == HeavyWeightedPressurePlate() {
		return 141
	}
	return 15
}

// releaseDelay is the delay before the plate re-checks its entities: 0.5
// seconds for weighted plates and 1 second otherwise.
func (p PressurePlate) releaseDelay() time.Duration {
	if p.Type.Weighted() {
		return time.Second / 2
	}
	return time.Second
}

type pressurePlateLivingEntity interface {
	Health() float64
	Dead() bool
}

func pressurePlateStoneEntity(e world.Entity) bool {
	if living, ok := e.(pressurePlateLivingEntity); ok {
		return living.Health() > 0 && !living.Dead()
	}
	return pressurePlateEntityName(e) == "minecraft:player" || pressurePlateEntityName(e) == "minecraft:armor_stand"
}

func pressurePlateEntityName(e world.Entity) string {
	h := e.H()
	if h == nil || h.Type() == nil {
		return ""
	}
	return h.Type().EncodeEntity()
}

// pressurePlateActivationBox is the box entities must intersect to press the
// plate at a position.
func pressurePlateActivationBox(pos cube.Pos) cube.BBox {
	return cube.Box(float64(pos[0])+0.125, float64(pos[1]), float64(pos[2])+0.125, float64(pos[0])+0.875, float64(pos[1])+0.25, float64(pos[2])+0.875)
}

func pressurePlateEntityIntersects(e world.Entity, box cube.BBox) bool {
	h := e.H()
	if h == nil || h.Type() == nil {
		return false
	}
	return h.Type().BBox(e).Translate(e.Position()).IntersectsWith(box)
}

// allPressurePlates ...
func allPressurePlates() (plates []world.Block) {
	for _, t := range PressurePlateTypes() {
		for power := 0; power <= 15; power++ {
			plates = append(plates, PressurePlate{Type: t, Power: power})
		}
	}
	return
}
