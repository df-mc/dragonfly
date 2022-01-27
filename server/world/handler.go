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
	// position into. The liquid that will replace the block replaced is also passed.
	HandleLiquidFlow(ctx *event.Context, from, into cube.Pos, liquid, replaced Block)
	// HandleLiquidHarden handles the hardening of a liquid at hardenedPos. The liquid that was hardened,
	// liquidHardened, and the liquid that caused it to harden, otherLiquid, are passed. The block created
	// as a result is also passed.
	HandleLiquidHarden(ctx *event.Context, hardenedPos cube.Pos, liquidHardened, otherLiquid, newBlock Block)
	// HandleSound handles a Sound being played in the World at a specific position. ctx.Cancel() may be called
	// to stop the Sound from playing to viewers of the position.
	HandleSound(ctx *event.Context, s Sound, pos mgl64.Vec3)
	// HandleEntitySpawn handles an entity being spawned into the world.
	HandleEntitySpawn(ctx *event.Context, e Entity)
	// HandleEntityDespawn handles an entity being despawned from the world.
	HandleEntityDespawn(ctx *event.Context, e Entity)
}

// NopHandler implements the Handler interface but does not execute any code when an event is called. The
// default Handler of worlds is set to NopHandler.
// Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = (*NopHandler)(nil)

// HandleLiquidFlow ...
func (NopHandler) HandleLiquidFlow(*event.Context, cube.Pos, cube.Pos, Block, Block) {}

// HandleLiquidHarden ...
func (NopHandler) HandleLiquidHarden(*event.Context, cube.Pos, Block, Block, Block) {}

// HandleSound ...
func (NopHandler) HandleSound(*event.Context, Sound, mgl64.Vec3) {}

// HandleEntitySpawn ...
func (NopHandler) HandleEntitySpawn(*event.Context,Entity)  {}

// HandleEntityDespawn ...
func (NopHandler) HandleEntityDespawn(*event.Context,Entity)  {}
