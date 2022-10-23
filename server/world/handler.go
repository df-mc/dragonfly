package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/go-gl/mathgl/mgl64"
)

// Handler handles events that are called by a world. Implementations of Handler may be used to listen to
// specific events such as when an entity is added to the world.
type Handler interface {
	// HandleLiquidFlow handles the flowing of a liquid from one block position from into another block
	// position into. The liquid that will replace the block is also passed. This replaced block might
	// also be a Liquid. The Liquid's depth and falling state can be checked to see if the resulting
	// liquid is a new source block (in the case of water).
	HandleLiquidFlow(ctx *event.Context, from, into cube.Pos, liquid Liquid, replaced Block)
	// HandleLiquidDecay handles the decaying of a Liquid block at a position. Liquid decaying happens
	// when there is no Liquid that can serve as the source block neighbouring it. The state of the
	// Liquid before and after the decaying is passed. The Liquid after is nil if the liquid is
	// completely removed as a result of the decay.
	HandleLiquidDecay(ctx *event.Context, pos cube.Pos, before, after Liquid)
	// HandleLiquidHarden handles the hardening of a liquid at hardenedPos. The liquid that was hardened,
	// liquidHardened, and the liquid that caused it to harden, otherLiquid, are passed. The block created
	// as a result is also passed.
	HandleLiquidHarden(ctx *event.Context, hardenedPos cube.Pos, liquidHardened, otherLiquid, newBlock Block)
	// HandleSound handles a Sound being played in the World at a specific position. ctx.Cancel() may be called
	// to stop the Sound from playing to viewers of the position.
	HandleSound(ctx *event.Context, s Sound, pos mgl64.Vec3)
	// HandleFireSpread handles when a fire block spreads from one block to another block. When this event handler gets
	// called, both the position of the original fire will be passed, and the position where it will spread to after the
	// event. The age of the fire may also be altered by changing the underlying value of the newFireAge pointer, which
	// decides how long the fire will stay before burning out.
	HandleFireSpread(ctx *event.Context, from, to cube.Pos)
	// HandleBlockBurn handles a block at a cube.Pos being burnt by fire. This event may be called for blocks such as
	// wood, that can be broken by fire. HandleBlockBurn is often succeeded by HandleFireSpread, when fire spreads to
	// the position of the original block and the event.Context is not cancelled in HandleBlockBurn.
	HandleBlockBurn(ctx *event.Context, pos cube.Pos)
	// HandleEntitySpawn handles an entity being spawned into a World through a call to World.AddEntity.
	HandleEntitySpawn(e Entity)
	// HandleEntityDespawn handles an entity being despawned from a World through a call to World.RemoveEntity.
	HandleEntityDespawn(e Entity)
	// HandleClose handles the World being closed. HandleClose may be used as a moment to finish code running on other
	// goroutines that operates on the World specifically. HandleClose is called directly before the World stops
	// ticking and before any chunks are saved to disk.
	HandleClose()
}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = (*NopHandler)(nil)

// NopHandler implements the Handler interface but does not execute any code when an event is called. The
// default Handler of worlds is set to NopHandler.
// Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

func (NopHandler) HandleLiquidFlow(*event.Context, cube.Pos, cube.Pos, Liquid, Block) {}
func (NopHandler) HandleLiquidDecay(*event.Context, cube.Pos, Liquid, Liquid)         {}
func (NopHandler) HandleLiquidHarden(*event.Context, cube.Pos, Block, Block, Block)   {}
func (NopHandler) HandleSound(*event.Context, Sound, mgl64.Vec3)                      {}
func (NopHandler) HandleFireSpread(*event.Context, cube.Pos, cube.Pos)                {}
func (NopHandler) HandleBlockBurn(*event.Context, cube.Pos)                           {}
func (NopHandler) HandleEntitySpawn(Entity)                                           {}
func (NopHandler) HandleEntityDespawn(Entity)                                         {}
func (NopHandler) HandleClose()                                                       {}
