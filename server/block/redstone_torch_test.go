package block

import (
	"fmt"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

func TestRedstoneTorchPowerAndEncode(t *testing.T) {
	torch := RedstoneTorch{Facing: cube.FaceDown, Lit: true}
	if power := torch.RedstonePower(cube.Pos{}, nil, cube.FaceUp); power != 15 {
		t.Fatalf("lit redstone torch output = %d, want 15", power)
	}
	if power := torch.RedstonePower(cube.Pos{}, nil, cube.FaceDown); power != 0 {
		t.Fatalf("redstone torch attached-face output = %d, want 0", power)
	}
	if power := (RedstoneTorch{Facing: cube.FaceDown}).RedstonePower(cube.Pos{}, nil, cube.FaceUp); power != 0 {
		t.Fatalf("unlit redstone torch output = %d, want 0", power)
	}
	if light := torch.LightEmissionLevel(); light != 7 {
		t.Fatalf("lit redstone torch light = %d, want 7", light)
	}

	name, props := torch.EncodeBlock()
	if name != "minecraft:redstone_torch" {
		t.Fatalf("redstone torch block name = %q, want minecraft:redstone_torch", name)
	}
	if facing := props["torch_facing_direction"]; facing != "top" {
		t.Fatalf("torch_facing_direction = %v, want top", facing)
	}
	if _, ok := world.BlockByName(name, props); !ok {
		t.Fatalf("BlockByName(%s, %#v) was not found", name, props)
	}

	name, props = (RedstoneTorch{Facing: cube.FaceNorth}).EncodeBlock()
	if name != "minecraft:unlit_redstone_torch" {
		t.Fatalf("unlit redstone torch block name = %q, want minecraft:unlit_redstone_torch", name)
	}
	if facing := props["torch_facing_direction"]; facing != "north" {
		t.Fatalf("unlit torch_facing_direction = %v, want north", facing)
	}
	if count := len(allRedstoneTorches()); count != len(cube.Faces())*2 {
		t.Fatalf("allRedstoneTorches returned %d states, want %d", count, len(cube.Faces())*2)
	}
}

func TestRedstoneTorchInverseScheduledTick(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		support := cube.Pos{0, 1, 0}
		pos := support.Side(cube.FaceUp)
		tx.SetBlock(support, RedstoneBlock{}, nil)
		torch := RedstoneTorch{Facing: cube.FaceDown, Lit: true}
		tx.SetBlock(pos, torch, nil)
		torch.ScheduledTick(pos, tx, nil)
		after, ok := tx.Block(pos).(RedstoneTorch)
		if !ok {
			err = fmt.Errorf("redstone torch missing after scheduled tick")
			return
		}
		if after.Lit {
			err = fmt.Errorf("redstone torch stayed lit while attached block was powered")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedstoneTorchRegisteredStates(t *testing.T) {
	if b, ok := world.BlockByName("minecraft:redstone_torch", map[string]any{
		"torch_facing_direction": "west",
	}); !ok {
		t.Fatal("lit redstone torch state was not registered")
	} else if torch, ok := b.(RedstoneTorch); !ok || !torch.Lit || torch.Facing != cube.FaceWest {
		t.Fatalf("registered lit redstone torch = %#v, want west lit RedstoneTorch", b)
	}

	if b, ok := world.BlockByName("minecraft:unlit_redstone_torch", map[string]any{
		"torch_facing_direction": "top",
	}); !ok {
		t.Fatal("unlit redstone torch state was not registered")
	} else if torch, ok := b.(RedstoneTorch); !ok || torch.Lit || torch.Facing != cube.FaceDown {
		t.Fatalf("registered unlit redstone torch = %#v, want standing unlit RedstoneTorch", b)
	}
}
