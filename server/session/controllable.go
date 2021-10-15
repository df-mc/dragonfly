package session

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

// Controllable represents an entity that may be controlled by a Session. Generally, a Controllable is
// implemented in the form of a Player.
// Methods in Controllable will be added as Session needs them in order to handle packets.
type Controllable interface {
	world.Entity
	item.User
	form.Submitter
	cmd.Source

	Move(deltaPos mgl64.Vec3, deltaYaw, deltaPitch float64)
	Speed() float64

	Chat(msg ...interface{})
	ExecuteCommand(commandLine string)
	GameMode() world.GameMode
	SetGameMode(mode world.GameMode)
	Effects() []effect.Effect

	UseItem()
	ReleaseItem()
	UseItemOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3)
	UseItemOnEntity(e world.Entity)
	BreakBlock(pos cube.Pos)
	PickBlock(pos cube.Pos)
	AttackEntity(e world.Entity)
	Drop(s item.Stack) (n int)
	SwingArm()
	PunchAir()

	Respawn()

	StartSneaking()
	Sneaking() bool
	StopSneaking()
	StartSprinting()
	Sprinting() bool
	StopSprinting()
	StartSwimming()
	Swimming() bool
	StopSwimming()
	StartFlying()
	Flying() bool
	StopFlying()

	StartBreaking(pos cube.Pos, face cube.Face)
	ContinueBreaking(face cube.Face)
	FinishBreaking()
	AbortBreaking()

	Exhaust(points float64)

	EditSign(pos cube.Pos, text string) error

	EnderChestInventory() *inventory.Inventory

	// UUID returns the UUID of the controllable. It must be unique for all controllable entities present in
	// the server.
	UUID() uuid.UUID
	// XUID returns the XBOX Live User ID of the controllable. Every controllable must have one of these if
	// they are authenticated via XBOX Live, as they must be connected to an XBOX Live account.
	XUID() string
	// Skin returns the skin of the controllable. Each controllable must have a skin, as it defines how the
	// entity looks in the world.
	Skin() skin.Skin
	SetSkin(skin.Skin)
}
