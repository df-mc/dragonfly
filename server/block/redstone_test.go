package block

import (
	"fmt"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

func TestRedstoneBlockPower(t *testing.T) {
	for _, face := range cube.Faces() {
		if power := (RedstoneBlock{}).RedstonePower(cube.Pos{}, nil, face); power != 15 {
			t.Fatalf("RedstoneBlock power from %v = %d, want 15", face, power)
		}
	}
}

func TestRedstoneWirePowerUpdate(t *testing.T) {
	wire := RedstoneWire{Power: 3}
	after, changed := wire.RedstonePowerUpdate(cube.Pos{}, nil, 12)
	if !changed {
		t.Fatal("RedstoneWire update did not report a change")
	}
	if got := after.(RedstoneWire).Power; got != 12 {
		t.Fatalf("RedstoneWire power = %d, want 12", got)
	}

	after, changed = wire.RedstonePowerUpdate(cube.Pos{}, nil, 24)
	if !changed {
		t.Fatal("RedstoneWire clamped update did not report a change")
	}
	if got := after.(RedstoneWire).Power; got != 15 {
		t.Fatalf("RedstoneWire clamped power = %d, want 15", got)
	}

	_, changed = (RedstoneWire{Power: 15}).RedstonePowerUpdate(cube.Pos{}, nil, 15)
	if changed {
		t.Fatal("RedstoneWire unchanged power reported a change")
	}
}

func TestRedstoneWireRequiresSolidSupport(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{0, 1, 0}
		if redstoneWireSupported(tx, pos) {
			err = fmt.Errorf("wire without support was supported")
			return
		}
		tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, nil)
		if !redstoneWireSupported(tx, pos) {
			err = fmt.Errorf("wire on solid support was not supported")
			return
		}
		tx.SetBlock(pos.Side(cube.FaceDown), RedstoneWire{}, nil)
		if redstoneWireSupported(tx, pos) {
			err = fmt.Errorf("wire on top of wire was supported")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedstoneWireConnectsUpBlocks(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{0, 1, 0}
		up := cube.Pos{1, 2, 0}
		tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(pos, RedstoneWire{}, nil)
		tx.SetBlock(up.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(up, RedstoneWire{}, nil)

		for _, neighbour := range (RedstoneWire{}).RedstoneRelayerNeighbours(pos, tx) {
			if neighbour == up {
				return
			}
		}
		err = fmt.Errorf("wire neighbours did not include upward step")
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStrongPowerConductsThroughSolidBlocks(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		conductor := cube.Pos{0, 1, 0}
		target := conductor.Side(cube.FaceEast)
		tx.SetBlock(conductor, Stone{}, nil)
		tx.SetBlock(conductor.Side(cube.FaceWest), RedstoneBlock{}, nil)

		if power := tx.RedstonePower(target); power != 15 {
			err = fmt.Errorf("conducted strong power = %d, want 15", power)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedstoneLampPowerUpdate(t *testing.T) {
	after, changed := (RedstoneLamp{}).RedstonePowerUpdate(cube.Pos{}, nil, 15)
	if !changed {
		t.Fatal("RedstoneLamp update did not report a change")
	}
	if !after.(RedstoneLamp).Lit {
		t.Fatal("RedstoneLamp did not become lit")
	}
	if after.(RedstoneLamp).LightEmissionLevel() != 15 {
		t.Fatal("lit RedstoneLamp should emit light level 15")
	}

	after, changed = (RedstoneLamp{Lit: true}).RedstonePowerUpdate(cube.Pos{}, nil, 0)
	if !changed {
		t.Fatal("RedstoneLamp unpower update did not report a change")
	}
	if after.(RedstoneLamp).Lit {
		t.Fatal("RedstoneLamp did not turn off")
	}
}
