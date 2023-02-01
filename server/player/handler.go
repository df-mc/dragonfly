package player

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
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
	HandleMove(EventMove)
	// HandleJump handles the player jumping.
	HandleJump(EventJump)
	// HandleTeleport handles the teleportation of a player. ctx.Cancel() may be called to cancel it.
	HandleTeleport(EventTeleport)
	// HandleChangeWorld handles when the player is added to a new world. before may be nil.
	HandleChangeWorld(EventChangeWorld)
	// HandleToggleSprint handles when the player starts or stops sprinting.
	// After is true if the player is sprinting after toggling (changing their sprinting state).
	HandleToggleSprint(EventToggleSprint)
	// HandleToggleSneak handles when the player starts or stops sneaking.
	// After is true if the player is sneaking after toggling (changing their sneaking state).
	HandleToggleSneak(EventToggleSneak)
	// HandleChat handles a message sent in the chat by a player. ctx.Cancel() may be called to cancel the
	// message being sent in chat.
	// The message may be changed by assigning to *message.
	HandleChat(EventChat)
	// HandleFoodLoss handles the food bar of a player depleting naturally, for example because the player was
	// sprinting and jumping. ctx.Cancel() may be called to cancel the food points being lost.
	HandleFoodLoss(EventFoodLoss)
	// HandleHeal handles the player being healed by a healing source. ctx.Cancel() may be called to cancel
	// the healing.
	// The health added may be changed by assigning to *health.
	HandleHeal(EventHeal)
	// HandleHurt handles the player being hurt by any damage source. ctx.Cancel() may be called to cancel the
	// damage being dealt to the player.
	// The damage dealt to the player may be changed by assigning to *damage.
	HandleHurt(EventHurt)
	// HandleDeath handles the player dying to a particular damage cause.
	HandleDeath(EventDeath)
	// HandleRespawn handles the respawning of the player in the world. The spawn position passed may be
	// changed by assigning to *pos. The world.World in which the Player is respawned may be modifying by assigning to
	// *w. This world may be the world the Player died in, but it might also point to a different world (the overworld)
	// if the Player died in the nether or end.
	HandleRespawn(EventRespawn)
	// HandleSkinChange handles the player changing their skin. ctx.Cancel() may be called to cancel the skin
	// change.
	HandleSkinChange(EventSkinChange)
	// HandleStartBreak handles the player starting to break a block at the position passed. ctx.Cancel() may
	// be called to stop the player from breaking the block completely.
	HandleStartBreak(EventStartBreak)
	// HandleBlockBreak handles a block that is being broken by a player. ctx.Cancel() may be called to cancel
	// the block being broken. A pointer to a slice of the block's drops is passed, and may be altered
	// to change what items will actually be dropped.
	HandleBlockBreak(EventBlockBreak)
	// HandleBlockPlace handles the player placing a specific block at a position in its world. ctx.Cancel()
	// may be called to cancel the block being placed.
	HandleBlockPlace(EventBlockPlace)
	// HandleBlockPick handles the player picking a specific block at a position in its world. ctx.Cancel()
	// may be called to cancel the block being picked.
	HandleBlockPick(EventBlockPick)
	// HandleItemUse handles the player using an item in the air. It is called for each item, although most
	// will not actually do anything. Items such as snowballs may be thrown if HandleItemUse does not cancel
	// the context using ctx.Cancel(). It is not called if the player is holding no item.
	HandleItemUse(EventItemUse)
	// HandleItemUseOnBlock handles the player using the item held in its main hand on a block at the block
	// position passed. The face of the block clicked is also passed, along with the relative click position.
	// The click position has X, Y and Z values which are all in the range 0.0-1.0. It is also called if the
	// player is holding no item.
	HandleItemUseOnBlock(EventItemUseOnBlock)
	// HandleItemUseOnEntity handles the player using the item held in its main hand on an entity passed to
	// the method.
	// HandleItemUseOnEntity is always called when a player uses an item on an entity, regardless of whether
	// the item actually does anything when used on an entity. It is also called if the player is holding no
	// item.
	HandleItemUseOnEntity(EventItemUseOnEntity)
	// HandleItemConsume handles the player consuming an item. This is called whenever a consumable such as
	// food is consumed.
	HandleItemConsume(EventItemConsume)
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
	HandleAttackEntity(EventAttackEntity)
	// HandleExperienceGain handles the player gaining experience. ctx.Cancel() may be called to cancel
	// the gain.
	// The amount is also provided which can be modified.
	HandleExperienceGain(EventExperienceGain)
	// HandlePunchAir handles the player punching air.
	HandlePunchAir(EventPunchAir)
	// HandleSignEdit handles the player editing a sign. It is called for every keystroke while editing a sign and
	// has both the old text passed and the text after the edit. This typically only has a change of one character.
	HandleSignEdit(EventSignEdit)
	// HandleItemDamage handles the event wherein the item either held by the player or as armour takes
	// damage through usage.
	// The type of the item may be checked to determine whether it was armour or a tool used. The damage to
	// the item is passed.
	HandleItemDamage(EventItemDamage)
	// HandleItemPickup handles the player picking up an item from the ground. The item stack laying on the
	// ground is passed. ctx.Cancel() may be called to prevent the player from picking up the item.
	HandleItemPickup(EventItemPickup)
	// HandleItemDrop handles the player dropping an item on the ground. The dropped item entity is passed.
	// ctx.Cancel() may be called to prevent the player from dropping the entity.Item passed on the ground.
	// e.Item() may be called to obtain the item stack dropped.
	HandleItemDrop(EventItemDrop)
	// HandleTransfer handles a player being transferred to another server. ctx.Cancel() may be called to
	// cancel the transfer.
	HandleTransfer(EventTransfer)
	// HandleCommandExecution handles the command execution of a player, who wrote a command in the chat.
	// ctx.Cancel() may be called to cancel the command execution.
	HandleCommandExecution(EventCommandExecution)
	// HandleQuit handles the closing of a player. It is always called when the player is disconnected,
	// regardless of the reason.
	HandleQuit(EventQuit)
}

// NopHandler implements the Handler interface but does not execute any code when an event is called. The
// default Handler of players is set to NopHandler.
// Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = NopHandler{}

func (NopHandler) HandleAttackEntity(EventAttackEntity)         {}
func (NopHandler) HandleBlockBreak(EventBlockBreak)             {}
func (NopHandler) HandleBlockPick(EventBlockPick)               {}
func (NopHandler) HandleBlockPlace(EventBlockPlace)             {}
func (NopHandler) HandleChangeWorld(EventChangeWorld)           {}
func (NopHandler) HandleChat(EventChat)                         {}
func (NopHandler) HandleCommandExecution(EventCommandExecution) {}
func (NopHandler) HandleDeath(EventDeath)                       {}
func (NopHandler) HandleExperienceGain(EventExperienceGain)     {}
func (NopHandler) HandleFoodLoss(EventFoodLoss)                 {}
func (NopHandler) HandleHeal(EventHeal)                         {}
func (NopHandler) HandleHurt(EventHurt)                         {}
func (NopHandler) HandleItemConsume(EventItemConsume)           {}
func (NopHandler) HandleItemDamage(EventItemDamage)             {}
func (NopHandler) HandleItemDrop(EventItemDrop)                 {}
func (NopHandler) HandleItemPickup(EventItemPickup)             {}
func (NopHandler) HandleItemUse(EventItemUse)                   {}
func (NopHandler) HandleItemUseOnBlock(EventItemUseOnBlock)     {}
func (NopHandler) HandleItemUseOnEntity(EventItemUseOnEntity)   {}
func (NopHandler) HandleJump(EventJump)                         {}
func (NopHandler) HandleMove(EventMove)                         {}
func (NopHandler) HandlePunchAir(EventPunchAir)                 {}
func (NopHandler) HandleQuit(EventQuit)                         {}
func (NopHandler) HandleRespawn(EventRespawn)                   {}
func (NopHandler) HandleSignEdit(EventSignEdit)                 {}
func (NopHandler) HandleSkinChange(EventSkinChange)             {}
func (NopHandler) HandleStartBreak(EventStartBreak)             {}
func (NopHandler) HandleTeleport(EventTeleport)                 {}
func (NopHandler) HandleToggleSneak(EventToggleSneak)           {}
func (NopHandler) HandleToggleSprint(EventToggleSprint)         {}
func (NopHandler) HandleTransfer(EventTransfer)                 {}
