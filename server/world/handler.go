package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/go-gl/mathgl/mgl64"
)

type Context = event.Context[*Tx]

// Handler handles events that are called by a world. Implementations of
// Handler may be used to listen to specific events such as when an Entity is
// added to the world.
type Handler interface {
	// HandleLiquidFlow handles the flowing of a liquid from one block position
	// from into another block position into. The liquid that will replace the
	// block is also passed. This replaced block might also be a Liquid. The
	// Liquid's depth and falling state can be checked to see if the resulting
	// liquid is a new source block (in the case of water).
	HandleLiquidFlow(ctx *Context, from, into cube.Pos, liquid Liquid, replaced Block)
	// HandleLiquidDecay handles the decaying of a Liquid block at a position.
	// Liquid decaying happens when there is no Liquid that can serve as the
	// source block neighbouring it. The state of the Liquid before and after
	// the decaying is passed. The Liquid after is nil if the liquid is
	// completely removed as a result of the decay.
	HandleLiquidDecay(ctx *Context, pos cube.Pos, before, after Liquid)
	// HandleLiquidHarden handles the hardening of a liquid at hardenedPos. The
	// liquid that was hardened, liquidHardened, and the liquid that caused it
	// to harden, otherLiquid, are passed. The block created as a result is also
	// passed.
	HandleLiquidHarden(ctx *Context, hardenedPos cube.Pos, liquidHardened, otherLiquid, newBlock Block)
	// HandleSound handles a Sound being played in the World at a specific
	// position. ctx.Cancel() may be called to stop the Sound from playing to
	// viewers of the position.
	HandleSound(ctx *Context, s Sound, pos mgl64.Vec3)
	// HandleFireSpread handles when a fire block spreads from one block to
	// another block. When this event handler gets called, both the position of
	// the original fire will be passed, and the position where it will spread
	// to after the event. The age of the fire may also be altered by changing
	// the underlying value of the newFireAge pointer, which decides how long
	// the fire will stay before burning out.
	HandleFireSpread(ctx *Context, from, to cube.Pos)
	// HandleBlockBurn handles a block at a cube.Pos being burnt by fire. This
	// event may be called for blocks such as wood, that can be broken by fire.
	// HandleBlockBurn is often succeeded by HandleFireSpread, when fire spreads
	// to the position of the original block and the Context is not cancelled in
	// HandleBlockBurn.
	HandleBlockBurn(ctx *Context, pos cube.Pos)
	// HandleCropTrample handles an Entity trampling a crop.
	HandleCropTrample(ctx *Context, pos cube.Pos)
	// HandleLeavesDecay handles the decaying of a Leaves block at a position.
	// Leaves decaying happens when there is no wood block neighbouring it.
	// ctx.Cancel() may be called to prevent leaves from decaying.
	HandleLeavesDecay(ctx *Context, pos cube.Pos)
	// HandleEntitySpawn handles an Entity being spawned into a World through a
	// call to Tx.AddEntity.
	HandleEntitySpawn(tx *Tx, e Entity)
	// HandleEntityDespawn handles an Entity being despawned from a World
	// through a call to Tx.RemoveEntity.
	HandleEntityDespawn(tx *Tx, e Entity)
	// HandleExplosion handles an explosion in the world. ctx.Cancel() may be called
	// to cancel the explosion.
	// The affected entities, affected blocks, item drop chance, and whether the
	// explosion spawns fire may be altered.
	HandleExplosion(ctx *Context, position mgl64.Vec3, entities *[]Entity, blocks *[]cube.Pos, itemDropChance *float64, spawnFire *bool)
	// HandleClose handles the World being closed. HandleClose may be used as a
	// moment to finish code running on other goroutines that operates on the
	// World specifically. HandleClose is called directly before the World stops
	// ticking and before any chunks are saved to disk.
	HandleClose(tx *Tx)
}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = (*NopHandler)(nil)

// NopHandler implements the Handler interface but does not execute any code
// when an event is called. The default Handler of worlds is set to NopHandler.
// Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

func (NopHandler) HandleLiquidFlow(*Context, cube.Pos, cube.Pos, Liquid, Block)                  {}
func (NopHandler) HandleLiquidDecay(*Context, cube.Pos, Liquid, Liquid)                          {}
func (NopHandler) HandleLiquidHarden(*Context, cube.Pos, Block, Block, Block)                    {}
func (NopHandler) HandleSound(*Context, Sound, mgl64.Vec3)                                       {}
func (NopHandler) HandleFireSpread(*Context, cube.Pos, cube.Pos)                                 {}
func (NopHandler) HandleBlockBurn(*Context, cube.Pos)                                            {}
func (NopHandler) HandleCropTrample(*Context, cube.Pos)                                          {}
func (NopHandler) HandleLeavesDecay(*Context, cube.Pos)                                          {}
func (NopHandler) HandleEntitySpawn(*Tx, Entity)                                                 {}
func (NopHandler) HandleEntityDespawn(*Tx, Entity)                                               {}
func (NopHandler) HandleExplosion(*Context, mgl64.Vec3, *[]Entity, *[]cube.Pos, *float64, *bool) {}
func (NopHandler) HandleClose(*Tx)                                                               {}
