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

func TestLeverPower(t *testing.T) {
	if power := (Lever{}).RedstonePower(cube.Pos{}, nil, cube.FaceUp); power != 0 {
		t.Fatalf("unpowered lever power = %d, want 0", power)
	}
	if power := (Lever{Powered: true}).RedstonePower(cube.Pos{}, nil, cube.FaceUp); power != 15 {
		t.Fatalf("powered lever power = %d, want 15", power)
	}
	if power := (Lever{Facing: cube.FaceWest, Powered: true}).RedstoneStrongPower(cube.Pos{}, nil, cube.FaceEast); power != 15 {
		t.Fatalf("attached-face lever strong power = %d, want 15", power)
	}
	if power := (Lever{Facing: cube.FaceWest, Powered: true}).RedstoneStrongPower(cube.Pos{}, nil, cube.FaceWest); power != 0 {
		t.Fatalf("opposite-face lever strong power = %d, want 0", power)
	}
}

func TestLeverEncodeBlock(t *testing.T) {
	tests := []struct {
		name string
		l    Lever
		want string
	}{
		{name: "wall", l: Lever{Facing: cube.FaceEast}, want: "east"},
		{name: "floor east west", l: Lever{Facing: cube.FaceUp, Direction: cube.West}, want: "up_east_west"},
		{name: "floor north south", l: Lever{Facing: cube.FaceUp, Direction: cube.North}, want: "up_north_south"},
		{name: "ceiling east west", l: Lever{Facing: cube.FaceDown, Direction: cube.West}, want: "down_east_west"},
		{name: "ceiling north south", l: Lever{Facing: cube.FaceDown, Direction: cube.North}, want: "down_north_south"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, props := test.l.EncodeBlock()
			if direction := props["lever_direction"]; direction != test.want {
				t.Fatalf("lever_direction = %v, want %s", direction, test.want)
			}
		})
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

func TestRedstoneWireDoesNotPowerDown(t *testing.T) {
	wire := RedstoneWire{Power: 15}
	if power := wire.RedstonePower(cube.Pos{}, nil, cube.FaceDown); power != 0 {
		t.Fatalf("wire downward power = %d, want 0", power)
	}
	if power := wire.RedstonePower(cube.Pos{}, nil, cube.FaceUp); power != 15 {
		t.Fatalf("wire upward power = %d, want 15", power)
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

func TestRedstoneTorchBurnout(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{0, 1, 0}
		tx.SetBlock(pos.Side(cube.FaceDown), RedstoneBlock{}, nil)
		torch := RedstoneTorch{Facing: cube.FaceDown, Lit: true}
		tx.SetBlock(pos, torch, nil)

		for i := 0; i < 7; i++ {
			if tx.RecordRedstoneTorchToggle(pos) {
				err = fmt.Errorf("torch burned out after %d toggles, want below threshold", i+1)
				return
			}
		}
		torch.ScheduledTick(pos, tx, nil)
		after, ok := tx.Block(pos).(RedstoneTorch)
		if !ok {
			err = fmt.Errorf("redstone torch missing after burnout tick")
			return
		}
		if after.Lit {
			err = fmt.Errorf("redstone torch stayed lit after burnout")
			return
		}
		burnedOut, recoverable := tx.RedstoneTorchBurnoutStatus(pos)
		if !burnedOut || recoverable {
			err = fmt.Errorf("burnout status = burnedOut %t recoverable %t, want true false", burnedOut, recoverable)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
