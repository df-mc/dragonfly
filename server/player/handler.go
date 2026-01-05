package player

import (
	"net"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Context = event.Context[*Player]

// Handler handles events that are called by a player. Implementations of Handler may be used to listen to
// specific events such as when a player chats or moves.
type Handler interface {
	// HandleMove handles the movement of a player. ctx.Cancel() may be called to cancel the movement event.
	// The new position, yaw and pitch are passed.
	HandleMove(ctx *Context, newPos mgl64.Vec3, newRot cube.Rotation)
	// HandleJump handles the player jumping.
	HandleJump(p *Player)
	// HandleTeleport handles the teleportation of a player. ctx.Cancel() may be called to cancel it.
	HandleTeleport(ctx *Context, pos mgl64.Vec3)
	// HandleChangeWorld handles when the player is added to a new world. before may be nil.
	HandleChangeWorld(p *Player, before, after *world.World)
	// HandleToggleSprint handles when the player starts or stops sprinting.
	// After is true if the player is sprinting after toggling (changing their sprinting state).
	HandleToggleSprint(ctx *Context, after bool)
	// HandleToggleSneak handles when the player starts or stops sneaking.
	// After is true if the player is sneaking after toggling (changing their sneaking state).
	HandleToggleSneak(ctx *Context, after bool)
	// HandleChat handles a message sent in the chat by a player. ctx.Cancel() may be called to cancel the
	// message being sent in chat.
	// The message may be changed by assigning to *message.
	HandleChat(ctx *Context, message *string)
	// HandleFoodLoss handles the food bar of a player depleting naturally, for example because the player was
	// sprinting and jumping. ctx.Cancel() may be called to cancel the food points being lost.
	HandleFoodLoss(ctx *Context, from int, to *int)
	// HandleHeal handles the player being healed by a healing source. ctx.Cancel() may be called to cancel
	// the healing.
	// The health added may be changed by assigning to *health.
	HandleHeal(ctx *Context, health *float64, src world.HealingSource)
	// HandleHurt handles the player being hurt by any damage source. ctx.Cancel() may be called to cancel the
	// damage being dealt to the player.
	// The damage dealt to the player may be changed by assigning to *damage.
	// *damage is the final damage dealt to the player. Immune is set to true
	// if the player was hurt during an immunity frame with higher damage than
	// the original cause of the immunity frame. In this case, the damage is
	// reduced but the player is still knocked back.
	HandleHurt(ctx *Context, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource)
	// HandleDeath handles the player dying to a particular damage cause.
	HandleDeath(p *Player, src world.DamageSource, keepInv *bool)
	// HandleRespawn handles the respawning of the player in the world. The spawn position passed may be
	// changed by assigning to *pos. The world.World in which the Player is respawned may be modifying by assigning to
	// *w. This world may be the world the Player died in, but it might also point to a different world (the overworld)
	// if the Player died in the nether or end.
	HandleRespawn(p *Player, pos *mgl64.Vec3, w **world.World)
	// HandleSkinChange handles the player changing their skin. ctx.Cancel() may be called to cancel the skin
	// change.
	HandleSkinChange(ctx *Context, skin *skin.Skin)
	// HandleFireExtinguish handles the player extinguishing a fire at a specific position. ctx.Cancel() may
	// be called to cancel the fire being extinguished.
	// cube.Pos can be used to see where was the fire extinguished, may be used to cancel this on specific positions.
	HandleFireExtinguish(ctx *Context, pos cube.Pos)
	// HandleStartBreak handles the player starting to break a block at the position passed. ctx.Cancel() may
	// be called to stop the player from breaking the block completely.
	HandleStartBreak(ctx *Context, pos cube.Pos)
	// HandleBlockBreak handles a block that is being broken by a player. ctx.Cancel() may be called to cancel
	// the block being broken. A pointer to a slice of the block's drops is passed, and may be altered
	// to change what items will actually be dropped.
	HandleBlockBreak(ctx *Context, pos cube.Pos, drops *[]item.Stack, xp *int)
	// HandleBlockPlace handles the player placing a specific block at a position in its world. ctx.Cancel()
	// may be called to cancel the block being placed.
	HandleBlockPlace(ctx *Context, pos cube.Pos, b world.Block)
	// HandleBlockPick handles the player picking a specific block at a position in its world. ctx.Cancel()
	// may be called to cancel the block being picked.
	HandleBlockPick(ctx *Context, pos cube.Pos, b world.Block)
	// HandleItemUse handles the player using an item in the air. It is called for each item, although most
	// will not actually do anything. Items such as snowballs may be thrown if HandleItemUse does not cancel
	// the context using ctx.Cancel(). It is not called if the player is holding no item.
	HandleItemUse(ctx *Context)
	// HandleItemUseOnBlock handles the player using the item held in its main hand on a block at the block
	// position passed. The face of the block clicked is also passed, along with the relative click position.
	// The click position has X, Y and Z values which are all in the range 0.0-1.0. It is also called if the
	// player is holding no item.
	HandleItemUseOnBlock(ctx *Context, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3)
	// HandleItemUseOnEntity handles the player using the item held in its main hand on an entity passed to
	// the method.
	// HandleItemUseOnEntity is always called when a player uses an item on an entity, regardless of whether
	// the item actually does anything when used on an entity. It is also called if the player is holding no
	// item.
	HandleItemUseOnEntity(ctx *Context, e world.Entity)
	// HandleItemRelease handles the player releasing an item after using it for
	// a particular duration. These include items such as bows.
	HandleItemRelease(ctx *Context, item item.Stack, dur time.Duration)
	// HandleItemConsume handles the player consuming an item. This is called whenever a consumable such as
	// food is consumed.
	HandleItemConsume(ctx *Context, item item.Stack)
	// HandleAttackEntity handles the player attacking an entity using the item held in its hand. ctx.Cancel()
	// may be called to cancel the attack, which will cancel damage dealt to the target and will stop the
	// entity from being knocked back.
	// The entity attacked may not be alive (implements entity.Living), in which case no damage will be dealt
	// and the target won't be knocked back.
	// The entity attacked may also be immune when this method is called, in which case no damage and knock-
	// back will be dealt.
	// The knock back force and height is also provided which can be modified.
	// The attack can be a critical attack, which would increase damage by a factor of 1.5 and
	// spawn critical hit particles around the target entity. These particles will not be displayed
	// if no damage is dealt.
	HandleAttackEntity(ctx *Context, e world.Entity, force, height *float64, critical *bool)
	// HandleExperienceGain handles the player gaining experience. ctx.Cancel() may be called to cancel
	// the gain.
	// The amount is also provided which can be modified.
	HandleExperienceGain(ctx *Context, amount *int)
	// HandlePunchAir handles the player punching air.
	HandlePunchAir(ctx *Context)
	// HandleSignEdit handles the player editing a sign. It is called for every keystroke while editing a sign and
	// has both the old text passed and the text after the edit. This typically only has a change of one character.
	HandleSignEdit(ctx *Context, pos cube.Pos, frontSide bool, oldText, newText string)
	// HandleSleep handles the player beginning the sleep action. ctx.Cancel() may be called to cancel the action.
	HandleSleep(ctx *Context, sendReminder *bool)
	// HandleLecternPageTurn handles the player turning a page in a lectern. ctx.Cancel() may be called to cancel the
	// page turn. The page number may be changed by assigning to *page.
	HandleLecternPageTurn(ctx *Context, pos cube.Pos, oldPage int, newPage *int)
	// HandleItemDamage handles the event wherein the item either held by the player or as armour takes
	// damage through usage.
	// The type of the item may be checked to determine whether it was armour or a tool used. The damage to
	// the item is passed.
	HandleItemDamage(ctx *Context, i item.Stack, damage *int)
	// HandleItemPickup handles the player picking up an item from the ground. The item stack laying on the
	// ground is passed. ctx.Cancel() may be called to prevent the player from picking up the item.
	HandleItemPickup(ctx *Context, i *item.Stack)
	// HandleHeldSlotChange handles the player changing the slot they are currently holding.
	HandleHeldSlotChange(ctx *Context, from, to int)
	// HandleItemDrop handles the player dropping an item on the ground. The dropped item entity is passed.
	// ctx.Cancel() may be called to prevent the player from dropping the entity.Item passed on the ground.
	// e.Item() may be called to obtain the item stack dropped.
	HandleItemDrop(ctx *Context, s item.Stack)
	// HandleTransfer handles a player being transferred to another server. ctx.Cancel() may be called to
	// cancel the transfer.
	HandleTransfer(ctx *Context, addr *net.UDPAddr)
	// HandleCommandExecution handles the command execution of a player, who wrote a command in the chat.
	// ctx.Cancel() may be called to cancel the command execution.
	HandleCommandExecution(ctx *Context, command cmd.Command, args []string)
	// HandleQuit handles the closing of a player. It is always called when the player is disconnected,
	// regardless of the reason.
	HandleQuit(p *Player)
	// HandleDiagnostics handles the latest diagnostics data that the player has sent to the server. This is
	// not sent by every client however, only those with the "Creator > Enable Client Diagnostics" setting
	// enabled.
	HandleDiagnostics(p *Player, d session.Diagnostics)
}

// NopHandler implements the Handler interface but does not execute any code when an event is called. The
// default Handler of players is set to NopHandler.
// Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = NopHandler{}

func (NopHandler) HandleItemDrop(*Context, item.Stack)                                     {}
func (NopHandler) HandleHeldSlotChange(*Context, int, int)                                 {}
func (NopHandler) HandleMove(*Context, mgl64.Vec3, cube.Rotation)                          {}
func (NopHandler) HandleJump(*Player)                                                      {}
func (NopHandler) HandleTeleport(*Context, mgl64.Vec3)                                     {}
func (NopHandler) HandleChangeWorld(*Player, *world.World, *world.World)                   {}
func (NopHandler) HandleToggleSprint(*Context, bool)                                       {}
func (NopHandler) HandleToggleSneak(*Context, bool)                                        {}
func (NopHandler) HandleCommandExecution(*Context, cmd.Command, []string)                  {}
func (NopHandler) HandleTransfer(*Context, *net.UDPAddr)                                   {}
func (NopHandler) HandleChat(*Context, *string)                                            {}
func (NopHandler) HandleSkinChange(*Context, *skin.Skin)                                   {}
func (NopHandler) HandleFireExtinguish(*Context, cube.Pos)                                 {}
func (NopHandler) HandleStartBreak(*Context, cube.Pos)                                     {}
func (NopHandler) HandleBlockBreak(*Context, cube.Pos, *[]item.Stack, *int)                {}
func (NopHandler) HandleBlockPlace(*Context, cube.Pos, world.Block)                        {}
func (NopHandler) HandleBlockPick(*Context, cube.Pos, world.Block)                         {}
func (NopHandler) HandleSignEdit(*Context, cube.Pos, bool, string, string)                 {}
func (NopHandler) HandleSleep(*Context, *bool)                                             {}
func (NopHandler) HandleLecternPageTurn(*Context, cube.Pos, int, *int)                     {}
func (NopHandler) HandleItemPickup(*Context, *item.Stack)                                  {}
func (NopHandler) HandleItemUse(*Context)                                                  {}
func (NopHandler) HandleItemUseOnBlock(*Context, cube.Pos, cube.Face, mgl64.Vec3)          {}
func (NopHandler) HandleItemUseOnEntity(*Context, world.Entity)                            {}
func (NopHandler) HandleItemRelease(ctx *Context, item item.Stack, dur time.Duration)      {}
func (NopHandler) HandleItemConsume(*Context, item.Stack)                                  {}
func (NopHandler) HandleItemDamage(*Context, item.Stack, *int)                             {}
func (NopHandler) HandleAttackEntity(*Context, world.Entity, *float64, *float64, *bool)    {}
func (NopHandler) HandleExperienceGain(*Context, *int)                                     {}
func (NopHandler) HandlePunchAir(*Context)                                                 {}
func (NopHandler) HandleHurt(*Context, *float64, bool, *time.Duration, world.DamageSource) {}
func (NopHandler) HandleHeal(*Context, *float64, world.HealingSource)                      {}
func (NopHandler) HandleFoodLoss(*Context, int, *int)                                      {}
func (NopHandler) HandleDeath(*Player, world.DamageSource, *bool)                          {}
func (NopHandler) HandleRespawn(*Player, *mgl64.Vec3, **world.World)                       {}
func (NopHandler) HandleQuit(*Player)                                                      {}
func (NopHandler) HandleDiagnostics(*Player, session.Diagnostics)                          {}
