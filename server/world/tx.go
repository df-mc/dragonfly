package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"iter"
	"time"
)

type Tx struct {
	w      *World
	closed bool
}

// Range returns the lower and upper bounds of the World that the Tx is
// operating on.
func (tx *Tx) Range() cube.Range {
	return tx.w.ra
}

func (tx *Tx) SetBlock(pos cube.Pos, b Block, opts *SetOpts) {
	tx.World().setBlock(pos, b, opts)
}

func (tx *Tx) Block(pos cube.Pos) Block {
	return tx.World().block(pos)
}

func (tx *Tx) Liquid(pos cube.Pos) (Liquid, bool) {
	return tx.World().liquid(pos)
}

func (tx *Tx) SetLiquid(pos cube.Pos, b Liquid) {
	tx.World().setLiquid(pos, b)
}

func (tx *Tx) BuildStructure(pos cube.Pos, s Structure) {
	tx.World().buildStructure(pos, s)
}

func (tx *Tx) ScheduleBlockUpdate(pos cube.Pos, delay time.Duration) {
	tx.World().scheduleBlockUpdate(pos, delay)
}

func (tx *Tx) HighestLightBlocker(x, z int) int {
	return tx.World().highestLightBlocker(x, z)
}

func (tx *Tx) HighestBlock(x, z int) int {
	return tx.World().highestBlock(x, z)
}

func (tx *Tx) Light(pos cube.Pos) uint8 {
	return tx.World().light(pos)
}

func (tx *Tx) Skylight(pos cube.Pos) uint8 {
	return tx.World().skyLight(pos)
}

func (tx *Tx) SetBiome(pos cube.Pos, b Biome) {
	tx.World().setBiome(pos, b)
}

func (tx *Tx) Biome(pos cube.Pos) Biome {
	return tx.World().biome(pos)
}

func (tx *Tx) Temperature(pos cube.Pos) float64 {
	return tx.World().temperature(pos)
}

func (tx *Tx) RainingAt(pos cube.Pos) bool {
	return tx.World().rainingAt(pos)
}

func (tx *Tx) SnowingAt(pos cube.Pos) bool {
	return tx.World().snowingAt(pos)
}

func (tx *Tx) ThunderingAt(pos cube.Pos) bool {
	return tx.World().thunderingAt(pos)
}

func (tx *Tx) AddParticle(pos mgl64.Vec3, p Particle) {
	tx.World().addParticle(pos, p)
}

func (tx *Tx) PlaySound(pos mgl64.Vec3, s Sound) {
	tx.World().playSound(tx, pos, s)
}

func (tx *Tx) AddEntity(e *EntityHandle) Entity {
	return tx.World().addEntity(tx, e)
}

func (tx *Tx) RemoveEntity(e Entity) *EntityHandle {
	return tx.World().removeEntity(e, tx)
}

func (tx *Tx) EntitiesWithin(box cube.BBox) iter.Seq[Entity] {
	return tx.World().entitiesWithin(tx, box)
}

func (tx *Tx) Entities() iter.Seq[Entity] {
	return tx.World().allEntities(tx)
}

func (tx *Tx) Players() iter.Seq[Entity] {
	return tx.World().allPlayers(tx)
}

func (tx *Tx) Viewers(pos mgl64.Vec3) []Viewer {
	return tx.World().viewersOf(pos)
}

// World returns the World of the Tx. It panics if the transaction was already
// marked complete.
func (tx *Tx) World() *World {
	if tx.closed {
		panic("world.Tx: use of transaction after transaction finishes is not permitted")
	}
	return tx.w
}

func (tx *Tx) close() {
	tx.closed = true
}
