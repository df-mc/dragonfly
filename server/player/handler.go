package player

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/healing"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"net"
	"time"
)

// Handler handles events that are called by a player. Implementations of Handler may be used to listen to
// specific events such as when a player chats or moves.
type Handler interface {
	// HandleMove handles the movement of a player. ctx.Cancel() may be called to cancel the movement event.
	// The new position, yaw and pitch are passed.
	HandleMove(ctx *event.Context, newPos mgl64.Vec3, newYaw, newPitch float64)
	// HandleJump handles the player jumping.
	HandleJump()
	// HandleTeleport handles the teleportation of a player. ctx.Cancel() may be called to cancel it.
	HandleTeleport(ctx *event.Context, pos mgl64.Vec3)
	// HandleChangeWorld handles when the player is added to a new world. before may be nil.
	HandleChangeWorld(before, after *world.World)
	// HandleToggleSprint handles when the player starts or stops sprinting.
	// After is true if the player is sprinting after toggling (changing their sprinting state).
	HandleToggleSprint(ctx *event.Context, after bool)
	// HandleToggleSneak handles when the player starts or stops sneaking.
	// After is true if the player is sneaking after toggling (changing their sneaking state).
	HandleToggleSneak(ctx *event.Context, after bool)
	// HandleChat handles a message sent in the chat by a player. ctx.Cancel() may be called to cancel the
	// message being sent in chat.
	// The message may be changed by assigning to *message.
	HandleChat(ctx *event.Context, message *string)
	// HandleFoodLoss handles the food bar of a player depleting naturally, for example because the player was
	// sprinting and jumping. ctx.Cancel() may be called to cancel the food points being lost.
	HandleFoodLoss(ctx *event.Context, from, to int)
	// HandleHeal handles the player being healed by a healing source. ctx.Cancel() may be called to cancel
	// the healing.
	// The health added may be changed by assigning to *health.
	HandleHeal(ctx *event.Context, health *float64, src healing.Source)
	// HandleHurt handles the player being hurt by any damage source. ctx.Cancel() may be called to cancel the
	// damage being dealt to the player.
	// The damage dealt to the player may be changed by assigning to *damage.
	HandleHurt(ctx *event.Context, damage *float64, attackImmunity *time.Duration, src damage.Source)
	// HandleDeath handles the player dying to a particular damage cause.
	HandleDeath(src damage.Source)
	// HandleRespawn handles the respawning of the player in the world. The spawn position passed may be
	// changed by assigning to *pos. The world.World in which the Player is respawned may be modifying by assigning to
	// *w. This world may be the world the Player died in, but it might also point to a different world (the overworld)
	// if the Player died in the nether or end.
	HandleRespawn(pos *mgl64.Vec3, w **world.World)
	// HandleSkinChange handles the player changing their skin. ctx.Cancel() may be called to cancel the skin
	// change.
	HandleSkinChange(ctx *event.Context, skin *skin.Skin)
	// HandleStartBreak handles the player starting to break a block at the position passed. ctx.Cancel() may
	// be called to stop the player from breaking the block completely.
	HandleStartBreak(ctx *event.Context, pos cube.Pos)
	// HandleBlockBreak handles a block that is being broken by a player. ctx.Cancel() may be called to cancel
	// the block being broken. A pointer to a slice of the block's drops is passed, and may be altered
	// to change what items will actually be dropped.
	HandleBlockBreak(ctx *event.Context, pos cube.Pos, drops *[]item.Stack)
	// HandleBlockPlace handles the player placing a specific block at a position in its world. ctx.Cancel()
	// may be called to cancel the block being placed.
	HandleBlockPlace(ctx *event.Context, pos cube.Pos, b world.Block)
	// HandleBlockPick handles the player picking a specific block at a position in its world. ctx.Cancel()
	// may be called to cancel the block being picked.
	HandleBlockPick(ctx *event.Context, pos cube.Pos, b world.Block)
	// HandleItemUse handles the player using an item in the air. It is called for each item, although most
	// will not actually do anything. Items such as snowballs may be thrown if HandleItemUse does not cancel
	// the context using ctx.Cancel(). It is not called if the player is holding no item.
	HandleItemUse(ctx *event.Context)
	// HandleItemUseOnBlock handles the player using the item held in its main hand on a block at the block
	// position passed. The face of the block clicked is also passed, along with the relative click position.
	// The click position has X, Y and Z values which are all in the range 0.0-1.0. It is also called if the
	// player is holding no item.
	HandleItemUseOnBlock(ctx *event.Context, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3)
	// HandleItemUseOnEntity handles the player using the item held in its main hand on an entity passed to
	// the method.
	// HandleItemUseOnEntity is always called when a player uses an item on an entity, regardless of whether
	// the item actually does anything when used on an entity. It is also called if the player is holding no
	// item.
	HandleItemUseOnEntity(ctx *event.Context, e world.Entity)
	// HandleItemConsume handles the player consuming an item. This is called whenever a consumable such as
	// food is consumed.
	HandleItemConsume(ctx *event.Context, item item.Stack)
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
	HandleAttackEntity(ctx *event.Context, e world.Entity, force, height *float64, critical *bool)
	// HandleExperienceGain handles the player gaining experience. ctx.Cancel() may be called to cancel
	// the gain.
	// The amount is also provided which can be modified.
	HandleExperienceGain(ctx *event.Context, amount *int)
	// HandlePunchAir handles the player punching air.
	HandlePunchAir(ctx *event.Context)
	// HandleSignEdit handles the player editing a sign. It is called for every keystroke while editing a sign and
	// has both the old text passed and the text after the edit. This typically only has a change of one character.
	HandleSignEdit(ctx *event.Context, oldText, newText string)
	// HandleItemDamage handles the event wherein the item either held by the player or as armour takes
	// damage through usage.
	// The type of the item may be checked to determine whether it was armour or a tool used. The damage to
	// the item is passed.
	HandleItemDamage(ctx *event.Context, i item.Stack, damage int)
	// HandleItemPickup handles the player picking up an item from the ground. The item stack laying on the
	// ground is passed. ctx.Cancel() may be called to prevent the player from picking up the item.
	HandleItemPickup(ctx *event.Context, i item.Stack)
	// HandleItemDrop handles the player dropping an item on the ground. The dropped item entity is passed.
	// ctx.Cancel() may be called to prevent the player from dropping the entity.Item passed on the ground.
	// e.Item() may be called to obtain the item stack dropped.
	HandleItemDrop(ctx *event.Context, e *entity.Item)
	// HandleTransfer handles a player being transferred to another server. ctx.Cancel() may be called to
	// cancel the transfer.
	HandleTransfer(ctx *event.Context, addr *net.UDPAddr)
	// HandleCommandExecution handles the command execution of a player, who wrote a command in the chat.
	// ctx.Cancel() may be called to cancel the command execution.
	HandleCommandExecution(ctx *event.Context, command cmd.Command, args []string)
	// HandleQuit handles the closing of a player. It is always called when the player is disconnected,
	// regardless of the reason.
	HandleQuit()
}

// NopHandler implements the Handler interface but does not execute any code when an event is called. The
// default Handler of players is set to NopHandler.
// Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = NopHandler{}

func (NopHandler) HandleItemDrop(*event.Context, *entity.Item)                                {}
func (NopHandler) HandleMove(*event.Context, mgl64.Vec3, float64, float64)                    {}
func (NopHandler) HandleJump()                                                                {}
func (NopHandler) HandleTeleport(*event.Context, mgl64.Vec3)                                  {}
func (NopHandler) HandleChangeWorld(*world.World, *world.World)                               {}
func (NopHandler) HandleToggleSprint(*event.Context, bool)                                    {}
func (NopHandler) HandleToggleSneak(*event.Context, bool)                                     {}
func (NopHandler) HandleCommandExecution(*event.Context, cmd.Command, []string)               {}
func (NopHandler) HandleTransfer(*event.Context, *net.UDPAddr)                                {}
func (NopHandler) HandleChat(*event.Context, *string)                                         {}
func (NopHandler) HandleSkinChange(*event.Context, *skin.Skin)                                {}
func (NopHandler) HandleStartBreak(*event.Context, cube.Pos)                                  {}
func (NopHandler) HandleBlockBreak(*event.Context, cube.Pos, *[]item.Stack)                   {}
func (NopHandler) HandleBlockPlace(*event.Context, cube.Pos, world.Block)                     {}
func (NopHandler) HandleBlockPick(*event.Context, cube.Pos, world.Block)                      {}
func (NopHandler) HandleSignEdit(*event.Context, string, string)                              {}
func (NopHandler) HandleItemPickup(*event.Context, item.Stack)                                {}
func (NopHandler) HandleItemUse(*event.Context)                                               {}
func (NopHandler) HandleItemUseOnBlock(*event.Context, cube.Pos, cube.Face, mgl64.Vec3)       {}
func (NopHandler) HandleItemUseOnEntity(*event.Context, world.Entity)                         {}
func (NopHandler) HandleItemConsume(*event.Context, item.Stack)                               {}
func (NopHandler) HandleItemDamage(*event.Context, item.Stack, int)                           {}
func (NopHandler) HandleAttackEntity(*event.Context, world.Entity, *float64, *float64, *bool) {}
func (NopHandler) HandleExperienceGain(*event.Context, *int)                                  {}
func (NopHandler) HandlePunchAir(*event.Context)                                              {}
func (NopHandler) HandleHurt(*event.Context, *float64, *time.Duration, damage.Source)         {}
func (NopHandler) HandleHeal(*event.Context, *float64, healing.Source)                        {}
func (NopHandler) HandleFoodLoss(*event.Context, int, int)                                    {}
func (NopHandler) HandleDeath(damage.Source)                                                  {}
func (NopHandler) HandleRespawn(*mgl64.Vec3, **world.World)                                   {}
func (NopHandler) HandleQuit()                                                                {}
