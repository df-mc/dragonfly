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

const (
	redstoneSourceStone = iota
	redstoneSourcePolishedBlackstone
	redstoneSourceOak
	redstoneSourceSpruce
	redstoneSourceBirch
	redstoneSourceJungle
	redstoneSourceAcacia
	redstoneSourceDarkOak
	redstoneSourceMangrove
	redstoneSourceCherry
	redstoneSourceBamboo
	redstoneSourceCrimson
	redstoneSourceWarped
	redstoneSourcePaleOak
	redstoneSourceLightWeighted
	redstoneSourceHeavyWeighted
)

// Lever is a switch that emits redstone power while active.
type Lever struct {
	empty
	transparent
	sourceWaterDisplacer

	// Facing is the face the lever is attached to.
	Facing cube.Face
	// Axis is the horizontal axis used by floor and ceiling levers.
	//blockhash:lever_axis
	Axis cube.Axis
	// Powered is true if the lever is switched on.
	Powered bool
}

// UseOnBlock places a lever attached to the clicked face.
func (l Lever) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, l)
	if !used || !redstoneAttachmentSupported(tx, pos, face) {
		return false
	}
	l.Facing = face
	if user != nil && (face == cube.FaceUp || face == cube.FaceDown) {
		l.Axis = user.Rotation().Direction().Face().Axis()
	}
	place(tx, pos, l, user, ctx)
	return placed(ctx)
}

// Activate toggles the lever.
func (l Lever) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	l.Powered = !l.Powered
	tx.SetBlock(pos, l, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.Click{})
	return true
}

// NeighbourUpdateTick breaks the lever if its supporting block is removed.
func (l Lever) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !redstoneAttachmentSupported(tx, pos, l.Facing) {
		breakBlock(l, pos, tx)
	}
}

// RedstonePower returns maximum power while the lever is active.
func (l Lever) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	if l.Powered {
		return 15
	}
	return 0
}

// RedstoneStrongPower strongly powers the block the lever is attached to.
func (l Lever) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if l.Powered && face == l.Facing.Opposite() {
		return 15
	}
	return 0
}

// BreakInfo ...
func (l Lever) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, nothingEffective, oneOf(l))
}

// SideClosed ...
func (Lever) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// EncodeItem ...
func (Lever) EncodeItem() (name string, meta int16) {
	return "minecraft:lever", 0
}

// EncodeBlock ...
func (l Lever) EncodeBlock() (string, map[string]any) {
	return "minecraft:lever", map[string]any{
		"lever_direction": leverDirection(l.Facing, l.Axis),
		"open_bit":        boolByte(l.Powered),
	}
}

func allLevers() (levers []world.Block) {
	for _, face := range cube.HorizontalFaces() {
		levers = append(levers, Lever{Facing: face}, Lever{Facing: face, Powered: true})
	}
	for _, face := range []cube.Face{cube.FaceDown, cube.FaceUp} {
		for _, axis := range []cube.Axis{cube.X, cube.Z} {
			levers = append(levers, Lever{Facing: face, Axis: axis}, Lever{Facing: face, Axis: axis, Powered: true})
		}
	}
	return
}

func leverDirection(face cube.Face, axis cube.Axis) string {
	switch face {
	case cube.FaceDown:
		if axis == cube.Z {
			return "down_north_south"
		}
		return "down_east_west"
	case cube.FaceUp:
		if axis == cube.Z {
			return "up_north_south"
		}
		return "up_east_west"
	case cube.FaceNorth:
		return "north"
	case cube.FaceSouth:
		return "south"
	case cube.FaceWest:
		return "west"
	case cube.FaceEast:
		return "east"
	default:
		return "down_east_west"
	}
}

func leverAxisHash(l Lever) uint64 {
	if (l.Facing == cube.FaceDown || l.Facing == cube.FaceUp) && l.Axis == cube.Z {
		return 1
	}
	return 0
}

// Button is a pressable redstone power source.
type Button struct {
	empty
	transparent
	sourceWaterDisplacer

	// Type identifies the button material.
	Type int
	// Facing is the face the button is attached to.
	Facing cube.Face
	// Pressed is true while the button emits power.
	Pressed bool
}

// UseOnBlock places a button attached to the clicked face.
func (b Button) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, b)
	if !used || !redstoneAttachmentSupported(tx, pos, face) {
		return false
	}
	b.Facing = face
	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// Activate presses the button and schedules release.
func (b Button) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	if b.Pressed {
		return true
	}
	b.Pressed = true
	tx.SetBlock(pos, b, nil)
	tx.ScheduleBlockUpdate(pos, b, b.pressDuration())
	tx.PlaySound(pos.Vec3Centre(), sound.Click{})
	return true
}

// NeighbourUpdateTick breaks the button if its supporting block is removed.
func (b Button) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !redstoneAttachmentSupported(tx, pos, b.Facing) {
		breakBlock(b, pos, tx)
	}
}

// ScheduledTick releases a pressed button.
func (b Button) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !b.Pressed {
		return
	}
	b.Pressed = false
	tx.SetBlock(pos, b, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.Click{})
}

// RedstonePower returns maximum power while the button is pressed.
func (b Button) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	if b.Pressed {
		return 15
	}
	return 0
}

// RedstoneStrongPower strongly powers the block the button is attached to.
func (b Button) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if b.Pressed && face == b.Facing.Opposite() {
		return 15
	}
	return 0
}

// BreakInfo ...
func (b Button) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, oneOf(b))
}

// SideClosed ...
func (Button) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// EncodeItem ...
func (b Button) EncodeItem() (name string, meta int16) {
	return sourceName(b.Type, "button"), 0
}

// EncodeBlock ...
func (b Button) EncodeBlock() (string, map[string]any) {
	return sourceName(b.Type, "button"), map[string]any{"button_pressed_bit": boolByte(b.Pressed), "facing_direction": int32(b.Facing)}
}

func (b Button) pressDuration() time.Duration {
	if b.Type >= redstoneSourceOak && b.Type <= redstoneSourcePaleOak {
		return time.Second * 3 / 2
	}
	return time.Second
}

func allButtons() (buttons []world.Block) {
	for _, typ := range redstoneSourceTypes() {
		for _, face := range cube.Faces() {
			buttons = append(buttons, Button{Type: typ, Facing: face}, Button{Type: typ, Facing: face, Pressed: true})
		}
	}
	return
}

// PressurePlate emits redstone power while stepped on.
type PressurePlate struct {
	empty
	transparent
	sourceWaterDisplacer

	// Type identifies the pressure plate material.
	Type int
	// Power is the current redstone signal emitted by the plate.
	Power int
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

// Model ...
func (PressurePlate) Model() world.BlockModel {
	return model.Carpet{}
}

// EntityStepOn powers the plate when an entity stands on it.
func (p PressurePlate) EntityStepOn(pos cube.Pos, tx *world.Tx, e world.Entity) {
	power := p.entityPower(e)
	if power == 0 {
		return
	}
	if p.weighted() {
		power = max(power, p.detectPower(pos, tx))
	}
	if p.Power == power {
		tx.ScheduleBlockUpdate(pos, p, p.releaseDelay())
		return
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

// ScheduledTick releases the plate if nothing refreshes it.
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
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, oneOf(p))
}

// SideClosed ...
func (PressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// EncodeItem ...
func (p PressurePlate) EncodeItem() (name string, meta int16) {
	return pressurePlateSourceName(p.Type), 0
}

// EncodeBlock ...
func (p PressurePlate) EncodeBlock() (string, map[string]any) {
	return pressurePlateSourceName(p.Type), map[string]any{"redstone_signal": int32(max(0, min(p.Power, 15)))}
}

func (p PressurePlate) stepPower() int {
	if p.Type == redstoneSourceLightWeighted || p.Type == redstoneSourceHeavyWeighted {
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

func (p PressurePlate) detectsEntity(e world.Entity) bool {
	if p.ignoresEntity(e) {
		return false
	}
	if p.stoneLike() {
		return pressurePlateStoneEntity(e)
	}
	return true
}

func (p PressurePlate) detectPower(pos cube.Pos, tx *world.Tx) int {
	box := pressurePlateActivationBox(pos)
	entities := 0
	for e := range tx.EntitiesWithin(box.Grow(1)) {
		if p.entityPower(e) == 0 || !pressurePlateEntityIntersects(e, box) {
			continue
		}
		if !p.weighted() {
			return 15
		}
		entities++
		if entities >= p.weightedMaxEntities() {
			return 15
		}
	}
	if p.weighted() {
		return p.weightedPower(entities)
	}
	return 0
}

func (p PressurePlate) stoneLike() bool {
	return p.Type == redstoneSourceStone || p.Type == redstoneSourcePolishedBlackstone
}

func (p PressurePlate) weighted() bool {
	return p.Type == redstoneSourceLightWeighted || p.Type == redstoneSourceHeavyWeighted
}

func (p PressurePlate) weightedPower(entities int) int {
	if entities <= 0 {
		return 0
	}
	if p.Type == redstoneSourceHeavyWeighted {
		return min(15, (entities+9)/10)
	}
	return min(15, entities)
}

func (p PressurePlate) weightedMaxEntities() int {
	if p.Type == redstoneSourceHeavyWeighted {
		return 141
	}
	return 15
}

func (p PressurePlate) releaseDelay() time.Duration {
	return time.Second
}

func (p PressurePlate) ignoresEntity(e world.Entity) bool {
	return pressurePlateEntityName(e) == "minecraft:snowball"
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

func pressurePlateActivationBox(pos cube.Pos) cube.BBox {
	const inset = 1.0 / 16.0
	return cube.Box(float64(pos[0])+inset, float64(pos[1]), float64(pos[2])+inset, float64(pos[0]+1)-inset, float64(pos[1])+0.25, float64(pos[2]+1)-inset)
}

func pressurePlateEntityIntersects(e world.Entity, box cube.BBox) bool {
	h := e.H()
	if h == nil || h.Type() == nil {
		return false
	}
	return h.Type().BBox(e).Translate(e.Position()).IntersectsWith(box)
}

func allPressurePlates() (plates []world.Block) {
	for _, typ := range append(redstoneSourceTypes(), redstoneSourceLightWeighted, redstoneSourceHeavyWeighted) {
		for i := 0; i <= 15; i++ {
			plates = append(plates, PressurePlate{Type: typ, Power: i})
		}
	}
	return
}

func redstoneSourceTypes() []int {
	types := []int{
		redstoneSourceStone,
		redstoneSourcePolishedBlackstone,
		redstoneSourceOak,
		redstoneSourceSpruce,
		redstoneSourceBirch,
		redstoneSourceJungle,
		redstoneSourceAcacia,
		redstoneSourceDarkOak,
		redstoneSourceMangrove,
		redstoneSourceCherry,
		redstoneSourceBamboo,
		redstoneSourceCrimson,
		redstoneSourceWarped,
		redstoneSourcePaleOak,
	}
	return types
}

func pressurePlateSourceName(typ int) string {
	if typ == redstoneSourceLightWeighted {
		return "minecraft:light_weighted_pressure_plate"
	}
	if typ == redstoneSourceHeavyWeighted {
		return "minecraft:heavy_weighted_pressure_plate"
	}
	return sourceName(typ, "pressure_plate")
}

func sourceName(typ int, suffix string) string {
	switch typ {
	case redstoneSourceStone:
		return "minecraft:stone_" + suffix
	case redstoneSourcePolishedBlackstone:
		return "minecraft:polished_blackstone_" + suffix
	case redstoneSourceOak:
		if suffix == "button" {
			return "minecraft:wooden_button"
		}
		return "minecraft:wooden_pressure_plate"
	case redstoneSourceSpruce:
		return "minecraft:spruce_" + suffix
	case redstoneSourceBirch:
		return "minecraft:birch_" + suffix
	case redstoneSourceJungle:
		return "minecraft:jungle_" + suffix
	case redstoneSourceAcacia:
		return "minecraft:acacia_" + suffix
	case redstoneSourceDarkOak:
		return "minecraft:dark_oak_" + suffix
	case redstoneSourceMangrove:
		return "minecraft:mangrove_" + suffix
	case redstoneSourceCherry:
		return "minecraft:cherry_" + suffix
	case redstoneSourceBamboo:
		return "minecraft:bamboo_" + suffix
	case redstoneSourceCrimson:
		return "minecraft:crimson_" + suffix
	case redstoneSourceWarped:
		return "minecraft:warped_" + suffix
	case redstoneSourcePaleOak:
		return "minecraft:pale_oak_" + suffix
	default:
		return "minecraft:stone_" + suffix
	}
}

func (b Button) Model() world.BlockModel {
	return model.Empty{}
}

func (p PressurePlate) FuelInfo() item.FuelInfo {
	if p.Type >= redstoneSourceOak && p.Type <= redstoneSourcePaleOak {
		return newFuelInfo(time.Second * 15)
	}
	return item.FuelInfo{}
}

func (b Button) FuelInfo() item.FuelInfo {
	if b.Type >= redstoneSourceOak && b.Type <= redstoneSourcePaleOak {
		return newFuelInfo(time.Second * 5)
	}
	return item.FuelInfo{}
}

func redstoneAttachmentSupported(tx *world.Tx, pos cube.Pos, face cube.Face) bool {
	support := pos.Side(face.Opposite())
	if support.OutOfBounds(tx.Range()) {
		return false
	}
	return tx.Block(support).Model().FaceSolid(support, face, tx)
}

func redstoneFloorComponentSupported(tx *world.Tx, pos cube.Pos) bool {
	support := pos.Side(cube.FaceDown)
	if support.OutOfBounds(tx.Range()) {
		return false
	}
	return tx.Block(support).Model().FaceSolid(support, cube.FaceUp, tx)
}
