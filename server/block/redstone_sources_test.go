package block

import (
	"fmt"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestLeverPower(t *testing.T) {
	if power := (Lever{}).RedstonePower(cube.Pos{}, nil, cube.FaceUp); power != 0 {
		t.Fatalf("unpowered lever power = %d, want 0", power)
	}
	if power := (Lever{Powered: true}).RedstonePower(cube.Pos{}, nil, cube.FaceUp); power != 15 {
		t.Fatalf("powered lever power = %d, want 15", power)
	}
}

func TestLeverEncodeBlock(t *testing.T) {
	tests := []struct {
		name string
		l    Lever
		want string
	}{
		{name: "wall", l: Lever{Facing: cube.FaceEast}, want: "east"},
		{name: "floor east west", l: Lever{Facing: cube.FaceUp, Axis: cube.X}, want: "up_east_west"},
		{name: "floor north south", l: Lever{Facing: cube.FaceUp, Axis: cube.Z}, want: "up_north_south"},
		{name: "ceiling east west", l: Lever{Facing: cube.FaceDown, Axis: cube.X}, want: "down_east_west"},
		{name: "ceiling north south", l: Lever{Facing: cube.FaceDown, Axis: cube.Z}, want: "down_north_south"},
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

func TestRedstoneAttachableSupport(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{0, 1, 0}
		if redstoneAttachmentSupported(tx, pos, cube.FaceUp) {
			err = fmt.Errorf("lever without support was supported")
			return
		}
		tx.SetBlock(pos.Side(cube.FaceDown), Lever{}, nil)
		if redstoneAttachmentSupported(tx, pos, cube.FaceUp) {
			err = fmt.Errorf("lever on lever support was supported")
			return
		}
		tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, nil)
		if !redstoneAttachmentSupported(tx, pos, cube.FaceUp) {
			err = fmt.Errorf("lever on solid support was not supported")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedstoneFloorComponentSupport(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{0, 1, 0}
		if redstoneFloorComponentSupported(tx, pos) {
			err = fmt.Errorf("floor component without support was supported")
			return
		}
		tx.SetBlock(pos.Side(cube.FaceDown), Button{Facing: cube.FaceUp}, nil)
		if redstoneFloorComponentSupported(tx, pos) {
			err = fmt.Errorf("floor component on button support was supported")
			return
		}
		tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, nil)
		if !redstoneFloorComponentSupported(tx, pos) {
			err = fmt.Errorf("floor component on solid support was not supported")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestButtonPowerAndDuration(t *testing.T) {
	stone := Button{Type: redstoneSourceStone}
	if power := stone.RedstonePower(cube.Pos{}, nil, cube.FaceUp); power != 0 {
		t.Fatalf("unpressed button power = %d, want 0", power)
	}
	stone.Pressed = true
	if power := stone.RedstonePower(cube.Pos{}, nil, cube.FaceUp); power != 15 {
		t.Fatalf("pressed button power = %d, want 15", power)
	}
	if stone.pressDuration() >= (Button{Type: redstoneSourceOak}).pressDuration() {
		t.Fatal("stone button duration should be shorter than wooden button duration")
	}
}

func TestPressurePlatePower(t *testing.T) {
	plate := PressurePlate{Type: redstoneSourceStone, Power: 15}
	if power := plate.RedstonePower(cube.Pos{}, nil, cube.FaceUp); power != 15 {
		t.Fatalf("stone pressure plate power = %d, want 15", power)
	}
	if power := (PressurePlate{Type: redstoneSourceLightWeighted}).stepPower(); power != 1 {
		t.Fatalf("weighted pressure plate step power = %d, want first analog level 1", power)
	}
	if power := (PressurePlate{Type: redstoneSourceLightWeighted}).weightedPower(16); power != 15 {
		t.Fatalf("light weighted pressure plate count power = %d, want 15", power)
	}
	if power := (PressurePlate{Type: redstoneSourceHeavyWeighted}).weightedPower(11); power != 2 {
		t.Fatalf("heavy weighted pressure plate count power = %d, want 2", power)
	}
	if power := (PressurePlate{Type: redstoneSourceHeavyWeighted}).weightedPower(141); power != 15 {
		t.Fatalf("heavy weighted pressure plate max power = %d, want 15", power)
	}
}

func TestPressurePlateItemActivation(t *testing.T) {
	itemEntity := fakeItemEntity{}
	if power := (PressurePlate{Type: redstoneSourceStone}).entityPower(itemEntity); power != 0 {
		t.Fatalf("stone pressure plate item power = %d, want 0", power)
	}
	if power := (PressurePlate{Type: redstoneSourceOak}).entityPower(itemEntity); power != 15 {
		t.Fatalf("wood pressure plate item power = %d, want 15", power)
	}
	if power := (PressurePlate{Type: redstoneSourceLightWeighted}).entityPower(itemEntity); power != 1 {
		t.Fatalf("light weighted pressure plate item power = %d, want 1", power)
	}
	if power := (PressurePlate{Type: redstoneSourceHeavyWeighted}).entityPower(fakeSnowballEntity{}); power != 0 {
		t.Fatalf("heavy weighted pressure plate snowball power = %d, want 0", power)
	}
	if power := (PressurePlate{Type: redstoneSourceStone}).entityPower(fakeLivingEntity{health: 20}); power != 15 {
		t.Fatalf("stone pressure plate living entity power = %d, want 15", power)
	}
	if power := (PressurePlate{Type: redstoneSourceStone}).entityPower(fakeLivingEntity{}); power != 0 {
		t.Fatalf("stone pressure plate dead living entity power = %d, want 0", power)
	}
}

func TestPressurePlateDetectsEntityBoundingBoxOnEdge(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{0, 1, 0}
		tx.AddEntity(world.EntitySpawnOpts{Position: mgl64.Vec3{1.2, 1.0625, 0.5}}.New(pressurePlateTestEntityType{name: "minecraft:player"}, pressurePlateTestEntityConfig{}))

		if power := (PressurePlate{Type: redstoneSourceStone}).detectPower(pos, tx); power != 15 {
			err = fmt.Errorf("edge-overlapping pressure plate power = %d, want 15", power)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWeightedPressurePlateCountsWorldEntities(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		pos := cube.Pos{0, 1, 0}
		for i := 0; i < 11; i++ {
			tx.AddEntity(world.EntitySpawnOpts{Position: mgl64.Vec3{0.5 + float64(i%3)*0.01, 1.0625, 0.5}}.New(pressurePlateTestEntityType{name: "minecraft:item"}, pressurePlateTestEntityConfig{}))
		}
		tx.AddEntity(world.EntitySpawnOpts{Position: mgl64.Vec3{0.5, 1.0625, 0.5}}.New(pressurePlateTestEntityType{name: "minecraft:snowball"}, pressurePlateTestEntityConfig{}))

		if power := (PressurePlate{Type: redstoneSourceLightWeighted}).detectPower(pos, tx); power != 11 {
			err = fmt.Errorf("light weighted plate power = %d, want 11", power)
			return
		}
		if power := (PressurePlate{Type: redstoneSourceHeavyWeighted}).detectPower(pos, tx); power != 2 {
			err = fmt.Errorf("heavy weighted plate power = %d, want 2", power)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedstoneSourceNames(t *testing.T) {
	tests := map[int]string{
		redstoneSourceStone:              "minecraft:stone_button",
		redstoneSourcePolishedBlackstone: "minecraft:polished_blackstone_button",
		redstoneSourceOak:                "minecraft:wooden_button",
		redstoneSourceMangrove:           "minecraft:mangrove_button",
		redstoneSourcePaleOak:            "minecraft:pale_oak_button",
	}
	for typ, want := range tests {
		if got, _ := (Button{Type: typ}).EncodeItem(); got != want {
			t.Fatalf("button type %d encodes to %q, want %q", typ, got, want)
		}
	}
	if got, _ := (PressurePlate{Type: redstoneSourceLightWeighted}).EncodeItem(); got != "minecraft:light_weighted_pressure_plate" {
		t.Fatalf("light weighted pressure plate encodes to %q", got)
	}
	if got, _ := (PressurePlate{Type: redstoneSourceHeavyWeighted}).EncodeItem(); got != "minecraft:heavy_weighted_pressure_plate" {
		t.Fatalf("heavy weighted pressure plate encodes to %q", got)
	}
}

type fakeItemEntity struct{}

func (fakeItemEntity) Close() error            { return nil }
func (fakeItemEntity) H() *world.EntityHandle  { return nil }
func (fakeItemEntity) Position() mgl64.Vec3    { return mgl64.Vec3{} }
func (fakeItemEntity) Rotation() cube.Rotation { return cube.Rotation{} }
func (fakeItemEntity) Item() item.Stack        { return item.Stack{} }

type fakeSnowballEntity struct{ fakeItemEntity }

func (fakeSnowballEntity) H() *world.EntityHandle {
	return world.EntitySpawnOpts{}.New(pressurePlateTestEntityType{name: "minecraft:snowball"}, pressurePlateTestEntityConfig{})
}

type fakeLivingEntity struct {
	fakeItemEntity
	health float64
}

func (e fakeLivingEntity) Health() float64 { return e.health }
func (e fakeLivingEntity) Dead() bool      { return e.health <= 0 }

type pressurePlateTestEntityConfig struct{}

func (pressurePlateTestEntityConfig) Apply(*world.EntityData) {}

type pressurePlateTestEntityType struct {
	name string
}

func (pressurePlateTestEntityType) Open(_ *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return pressurePlateTestEntity{handle: handle, data: data}
}

func (t pressurePlateTestEntityType) EncodeEntity() string {
	if t.name != "" {
		return t.name
	}
	return "minecraft:test_entity"
}
func (pressurePlateTestEntityType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.3, 0, -0.3, 0.3, 1.8, 0.3)
}
func (pressurePlateTestEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (pressurePlateTestEntityType) EncodeNBT(*world.EntityData) map[string]any  { return nil }

type pressurePlateTestEntity struct {
	handle *world.EntityHandle
	data   *world.EntityData
}

func (e pressurePlateTestEntity) Close() error            { return nil }
func (e pressurePlateTestEntity) H() *world.EntityHandle  { return e.handle }
func (e pressurePlateTestEntity) Position() mgl64.Vec3    { return e.data.Pos }
func (e pressurePlateTestEntity) Rotation() cube.Rotation { return e.data.Rot }

func TestRedstoneSourceHashesIncludeMaterial(t *testing.T) {
	_, stoneButton := (Button{Type: redstoneSourceStone, Facing: cube.FaceUp}).Hash()
	_, oakButton := (Button{Type: redstoneSourceOak, Facing: cube.FaceUp}).Hash()
	if stoneButton == oakButton {
		t.Fatal("stone and oak buttons produced the same block hash")
	}

	_, stonePlate := (PressurePlate{Type: redstoneSourceStone}).Hash()
	_, oakPlate := (PressurePlate{Type: redstoneSourceOak}).Hash()
	if stonePlate == oakPlate {
		t.Fatal("stone and oak pressure plates produced the same block hash")
	}
}
