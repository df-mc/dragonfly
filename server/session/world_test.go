package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type offsetTestConfig struct{}

func (offsetTestConfig) Apply(*world.EntityData) {}

type offsetTestType struct {
	name   string
	offset float64
}

func (t offsetTestType) Open(*world.Tx, *world.EntityHandle, *world.EntityData) world.Entity {
	return nil
}

func (t offsetTestType) EncodeEntity() string { return t.name }
func (offsetTestType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.3, 0, -0.3, 0.3, 1.8, 0.3)
}
func (offsetTestType) DecodeNBT(map[string]any, *world.EntityData) {}
func (offsetTestType) EncodeNBT(*world.EntityData) map[string]any  { return nil }
func (t offsetTestType) NetworkOffset() float64                    { return t.offset }

type offsetTestEntity struct {
	handle   *world.EntityHandle
	sleeping bool
	sneaking bool
	swimming bool
	crawling bool
	gliding  bool
}

func (e *offsetTestEntity) Close() error               { return nil }
func (e *offsetTestEntity) H() *world.EntityHandle     { return e.handle }
func (e *offsetTestEntity) Position() mgl64.Vec3       { return mgl64.Vec3{} }
func (e *offsetTestEntity) Rotation() cube.Rotation    { return cube.Rotation{} }
func (e *offsetTestEntity) Sleeping() (cube.Pos, bool) { return cube.Pos{}, e.sleeping }
func (e *offsetTestEntity) Sneaking() bool             { return e.sneaking }
func (e *offsetTestEntity) Swimming() bool             { return e.swimming }
func (e *offsetTestEntity) Crawling() bool             { return e.crawling }
func (e *offsetTestEntity) Gliding() bool              { return e.gliding }

func TestEntityOffsetPlayerPoses(t *testing.T) {
	tests := []struct {
		name   string
		entity offsetTestEntity
		want   float64
	}{
		{name: "standing", want: 1.62001},
		{name: "sleeping", entity: offsetTestEntity{sleeping: true}, want: 0.2},
		{name: "swimming", entity: offsetTestEntity{swimming: true}, want: 0.4},
		{name: "crawling", entity: offsetTestEntity{crawling: true}, want: 0.4},
		{name: "gliding", entity: offsetTestEntity{gliding: true}, want: 0.4},
		{name: "sneaking", entity: offsetTestEntity{sneaking: true}, want: 1.27001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.entity.handle = world.EntitySpawnOpts{}.New(offsetTestType{name: "minecraft:player", offset: 1.62001}, offsetTestConfig{})
			if got := entityOffset(&tt.entity)[1]; got != tt.want {
				t.Fatalf("entity offset = %v, want %v", got, tt.want)
			}
		})
	}
}
