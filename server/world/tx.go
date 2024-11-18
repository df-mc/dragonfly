package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"iter"
	"time"
)

type Tx struct{ w *World }

// Range returns the lower and upper bounds of the World that the Tx is
// operating on.
func (tx *Tx) Range() cube.Range {
	return tx.w.ra
}

func (tx *Tx) SetBlock(pos cube.Pos, b Block, opts *SetOpts) {
	tx.w.setBlock(pos, b, opts)
}

func (tx *Tx) Block(pos cube.Pos) Block {
	return tx.w.block(pos)
}

func (tx *Tx) Liquid(pos cube.Pos) (Liquid, bool) {
	return tx.w.liquid(pos)
}

func (tx *Tx) SetLiquid(pos cube.Pos, b Liquid) {
	tx.w.setLiquid(pos, b)
}

func (tx *Tx) BuildStructure(pos cube.Pos, s Structure) {
	tx.w.buildStructure(pos, s)
}

func (tx *Tx) ScheduleBlockUpdate(pos cube.Pos, delay time.Duration) {
	tx.w.scheduleBlockUpdate(pos, delay)
}

func (tx *Tx) HighestLightBlocker(x, z int) int {
	return tx.w.highestLightBlocker(x, z)
}

func (tx *Tx) HighestBlock(x, z int) int {
	return tx.w.highestBlock(x, z)
}

func (tx *Tx) Light(pos cube.Pos) uint8 {
	return tx.w.light(pos)
}

func (tx *Tx) Skylight(pos cube.Pos) uint8 {
	return tx.w.skyLight(pos)
}

func (tx *Tx) SetBiome(pos cube.Pos, b Biome) {
	tx.w.setBiome(pos, b)
}

func (tx *Tx) Biome(pos cube.Pos) Biome {
	return tx.w.biome(pos)
}

func (tx *Tx) Temperature(pos cube.Pos) float64 {
	return tx.w.temperature(pos)
}

func (tx *Tx) RainingAt(pos cube.Pos) bool {
	return tx.w.rainingAt(pos)
}

func (tx *Tx) SnowingAt(pos cube.Pos) bool {
	return tx.w.snowingAt(pos)
}

func (tx *Tx) ThunderingAt(pos cube.Pos) bool {
	return tx.w.thunderingAt(pos)
}

func (tx *Tx) AddParticle(pos mgl64.Vec3, p Particle) {
	tx.w.addParticle(pos, p)
}

func (tx *Tx) PlaySound(pos mgl64.Vec3, s Sound) {
	tx.w.playSound(pos, s)
}

func (tx *Tx) AddEntity(e *EntityHandle) Entity {
	return tx.w.addEntity(tx, e)
}

func (tx *Tx) RemoveEntity(e Entity) *EntityHandle {
	return tx.w.removeEntity(e, tx)
}

func (tx *Tx) EntitiesWithin(box cube.BBox) iter.Seq[Entity] {
	return tx.w.entitiesWithin(tx, box)
}

func (tx *Tx) Entities() iter.Seq[Entity] {
	return tx.w.allEntities(tx)
}

func (tx *Tx) Viewers(pos mgl64.Vec3) []Viewer {
	return tx.w.viewersOf(pos)
}

// World returns the World of the Tx. It panics if the transaction was already
// marked complete.
func (tx *Tx) World() *World {
	return tx.w
}
