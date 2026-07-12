package session

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestHandleInputFlagsSneakDownWhileFlying(t *testing.T) {
	c := &inputFlagControllable{flying: true}
	flags := protocol.NewBitset(128)
	flags.Set(packet.InputFlagSneakDown)

	(PlayerAuthInputHandler{}).handleInputFlags(flags, nil, c)
	if !c.sneaking {
		t.Fatal("SneakDown while flying did not start sneaking")
	}
}

type inputFlagControllable struct {
	Controllable
	sneaking bool
	flying   bool
}

func (c *inputFlagControllable) StartSneaking() { c.sneaking = true }
func (c *inputFlagControllable) Sneaking() bool { return c.sneaking }
func (c *inputFlagControllable) StopSneaking()  { c.sneaking = false }
func (c *inputFlagControllable) Flying() bool   { return c.flying }
