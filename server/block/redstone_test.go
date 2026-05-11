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
		if power := (RedstoneBlock{}).RedstoneStrongPower(cube.Pos{}, nil, face); power != 0 {
			t.Fatalf("RedstoneBlock strong power from %v = %d, want 0", face, power)
		}
	}

	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		source := cube.Pos{0, 1, 0}
		target := source.Side(cube.FaceEast)
		tx.SetBlock(source, RedstoneBlock{}, nil)
		if power := tx.RedstonePower(target); power != 15 {
			err = fmt.Errorf("RedstoneBlock weak power = %d, want 15", power)
			return
		}
		if power := tx.RedstoneDirectPower(target); power != 15 {
			err = fmt.Errorf("RedstoneBlock direct power = %d, want 15", power)
			return
		}
		if power := tx.RedstoneStrongPower(target); power != 0 {
			err = fmt.Errorf("RedstoneBlock strong power = %d, want 0", power)
		}
	})
	if err != nil {
		t.Fatal(err)
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

func TestRedstoneWireRelayerNeighboursFollowShape(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{0, 1, 0}
		source := pos.Side(cube.FaceWest)
		endLamp := pos.Side(cube.FaceEast)
		sideLamp := pos.Side(cube.FaceNorth)
		tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(source, RedstoneBlock{}, nil)
		tx.SetBlock(pos, RedstoneWire{}, nil)
		tx.SetBlock(endLamp, RedstoneLamp{}, nil)
		tx.SetBlock(sideLamp, RedstoneLamp{}, nil)

		neighbours := (RedstoneWire{}).RedstoneRelayerNeighbours(pos, tx)
		if !redstoneNeighbourTestContains(neighbours, source) {
			err = fmt.Errorf("wire neighbours %v did not include source connection %v", neighbours, source)
			return
		}
		if !redstoneNeighbourTestContains(neighbours, endLamp) {
			err = fmt.Errorf("wire neighbours %v did not include line-end lamp %v", neighbours, endLamp)
			return
		}
		if redstoneNeighbourTestContains(neighbours, sideLamp) {
			err = fmt.Errorf("wire neighbours %v included side lamp %v", neighbours, sideLamp)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedstoneWireRelayerNeighboursDoNotLoadAdjacentChunks(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{15, 1, 0}
		unloadedNeighbour := pos.Side(cube.FaceEast)
		tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(pos, RedstoneWire{}, nil)

		if _, ok := tx.BlockLoaded(unloadedNeighbour); ok {
			err = fmt.Errorf("adjacent chunk was loaded before relayer neighbour lookup")
			return
		}
		_ = (RedstoneWire{}).RedstoneRelayerNeighbours(pos, tx)
		if _, ok := tx.BlockLoaded(unloadedNeighbour); ok {
			err = fmt.Errorf("relayer neighbour lookup loaded adjacent chunk")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedstoneWirePowersEndLampButNotSideLamp(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		wire := cube.Pos{0, 1, 0}
		endLamp := wire.Side(cube.FaceEast)
		sideLamp := wire.Side(cube.FaceNorth)
		tx.SetBlock(wire.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(wire.Side(cube.FaceWest), RedstoneBlock{}, nil)
		tx.SetBlock(wire, RedstoneWire{Power: 1}, nil)
		tx.SetBlock(endLamp, RedstoneLamp{}, nil)
		tx.SetBlock(sideLamp, RedstoneLamp{}, nil)

		after, changed := (RedstoneLamp{}).RedstonePowerUpdate(endLamp, tx, tx.RedstonePower(endLamp))
		if !changed || !after.(RedstoneLamp).Lit {
			err = fmt.Errorf("strength-1 line dust did not light end lamp: changed=%t after=%#v", changed, after)
			return
		}
		after, changed = (RedstoneLamp{}).RedstonePowerUpdate(sideLamp, tx, tx.RedstonePower(sideLamp))
		if changed || after.(RedstoneLamp).Lit {
			err = fmt.Errorf("line dust powered side lamp: changed=%t after=%#v", changed, after)
		}
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
		source := cube.Pos{0, 1, 0}
		conductor := source.Side(cube.FaceEast)
		target := conductor.Side(cube.FaceEast)
		tx.SetBlock(source, Lever{Facing: cube.FaceWest, Powered: true}, nil)
		tx.SetBlock(conductor, Stone{}, nil)

		if power := tx.RedstonePower(target); power != 15 {
			err = fmt.Errorf("conducted strong power = %d, want 15", power)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStrongPowerDoesNotConductThroughTransparentFullBlocks(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		source := cube.Pos{0, 1, 0}
		conductor := source.Side(cube.FaceEast)
		target := conductor.Side(cube.FaceEast)
		tx.SetBlock(source, Lever{Facing: cube.FaceWest, Powered: true}, nil)
		tx.SetBlock(conductor, Glass{}, nil)

		if power := tx.RedstonePower(target); power != 0 {
			err = fmt.Errorf("power conducted through glass = %d, want 0", power)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedstoneBlockDoesNotPowerLampThroughSolidBlock(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		source := cube.Pos{0, 1, 0}
		conductor := source.Side(cube.FaceEast)
		lampPos := conductor.Side(cube.FaceEast)
		tx.SetBlock(source, RedstoneBlock{}, nil)
		tx.SetBlock(conductor, Stone{}, nil)
		tx.SetBlock(lampPos, RedstoneLamp{}, nil)

		power := tx.RedstonePower(lampPos)
		if power != 0 {
			err = fmt.Errorf("lamp power through opaque block = %d, want 0", power)
			return
		}
		after, changed := (RedstoneLamp{}).RedstonePowerUpdate(lampPos, tx, power)
		if changed || after.(RedstoneLamp).Lit {
			err = fmt.Errorf("lamp lit from redstone block through opaque block")
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

func TestRedstoneLampDelayedTurnOff(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{0, 1, 0}
		lamp := RedstoneLamp{Lit: true}
		tx.SetBlock(pos, lamp, nil)

		after, changed := lamp.RedstonePowerUpdate(pos, tx, 0)
		if changed {
			err = fmt.Errorf("RedstoneLamp reported immediate off change: %#v", after)
			return
		}
		if !tx.Block(pos).(RedstoneLamp).Lit {
			err = fmt.Errorf("RedstoneLamp turned off before scheduled tick")
			return
		}
		tx.Block(pos).(RedstoneLamp).ScheduledTick(pos, tx, nil)
		if tx.Block(pos).(RedstoneLamp).Lit {
			err = fmt.Errorf("RedstoneLamp stayed lit after delayed off tick")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedstoneLampRepowerCancelsTurnOff(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{0, 1, 0}
		lamp := RedstoneLamp{Lit: true}
		tx.SetBlock(pos, lamp, nil)
		_, changed := lamp.RedstonePowerUpdate(pos, tx, 0)
		if changed {
			err = fmt.Errorf("RedstoneLamp reported immediate off change")
			return
		}
		tx.SetBlock(pos.Side(cube.FaceWest), RedstoneBlock{}, nil)
		tx.Block(pos).(RedstoneLamp).ScheduledTick(pos, tx, nil)
		if !tx.Block(pos).(RedstoneLamp).Lit {
			err = fmt.Errorf("RedstoneLamp turned off after being repowered")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func redstoneNeighbourTestContains(neighbours []cube.Pos, pos cube.Pos) bool {
	for _, neighbour := range neighbours {
		if neighbour == pos {
			return true
		}
	}
	return false
}
