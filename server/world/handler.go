package world

// Handler handles events that are called by a world. Implementations of Handler may be used to listen to
// specific events such as when an entity is added to the world.
type Handler interface {
	// HandleLiquidFlow handles the flowing of a liquid from one block position from into another block
	// position into. The liquid that will replace the block is also passed. This replaced block might
	// also be a Liquid. The Liquid's depth and falling state can be checked to see if the resulting
	// liquid is a new source block (in the case of water).
	HandleLiquidFlow(EventLiquidFlow)
	// HandleLiquidDecay handles the decaying of a Liquid block at a position. Liquid decaying happens
	// when there is no Liquid that can serve as the source block neighbouring it. The state of the
	// Liquid before and after the decaying is passed. The Liquid after is nil if the liquid is
	// completely removed as a result of the decay.
	HandleLiquidDecay(EventLiquidDecay)
	// HandleLiquidHarden handles the hardening of a liquid at hardenedPos. The liquid that was hardened,
	// liquidHardened, and the liquid that caused it to harden, otherLiquid, are passed. The block created
	// as a result is also passed.
	HandleLiquidHarden(EventLiquidHarden)
	// HandleSound handles a Sound being played in the World at a specific position. ctx.Cancel() may be called
	// to stop the Sound from playing to viewers of the position.
	HandleSound(EventSound)
	// HandleFireSpread handles when a fire block spreads from one block to another block. When this event handler gets
	// called, both the position of the original fire will be passed, and the position where it will spread to after the
	// event. The age of the fire may also be altered by changing the underlying value of the newFireAge pointer, which
	// decides how long the fire will stay before burning out.
	HandleFireSpread(EventFireSpread)
	// HandleBlockBurn handles a block at a cube.Pos being burnt by fire. This event may be called for blocks such as
	// wood, that can be broken by fire. HandleBlockBurn is often succeeded by HandleFireSpread, when fire spreads to
	// the position of the original block and the event.Context is not cancelled in HandleBlockBurn.
	HandleBlockBurn(EventBlockBurn)
	// HandleEntitySpawn handles an entity being spawned into a World through a call to World.AddEntity.
	HandleEntitySpawn(EventEntitySpawn)
	// HandleEntityDespawn handles an entity being despawned from a World through a call to World.RemoveEntity.
	HandleEntityDespawn(EventEntityDespawn)
	// HandleClose handles the World being closed. HandleClose may be used as a moment to finish code running on other
	// goroutines that operates on the World specifically. HandleClose is called directly before the World stops
	// ticking and before any chunks are saved to disk.
	HandleClose(EventClose)
}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = (*NopHandler)(nil)

// NopHandler implements the Handler interface but does not execute any code when an event is called. The
// default Handler of worlds is set to NopHandler.
// Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

func (NopHandler) HandleLiquidFlow(EventLiquidFlow)       {}
func (NopHandler) HandleLiquidDecay(EventLiquidDecay)     {}
func (NopHandler) HandleLiquidHarden(EventLiquidHarden)   {}
func (NopHandler) HandleSound(EventSound)                 {}
func (NopHandler) HandleFireSpread(EventFireSpread)       {}
func (NopHandler) HandleBlockBurn(EventBlockBurn)         {}
func (NopHandler) HandleEntitySpawn(EventEntitySpawn)     {}
func (NopHandler) HandleEntityDespawn(EventEntityDespawn) {}
func (NopHandler) HandleClose(EventClose)                 {}
