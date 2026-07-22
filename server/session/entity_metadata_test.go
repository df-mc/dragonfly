package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestNameTagMetadata(t *testing.T) {
	tests := []struct {
		name       string
		entity     any
		wantValue  uint8
		wantShow   bool
		wantAlways bool
	}{
		{
			name:       "name tag only defaults to always visible",
			entity:     testNameTag{name: "Dragonfly"},
			wantValue:  1,
			wantShow:   true,
			wantAlways: true,
		},
		{
			name:      "always visible may be disabled",
			entity:    testConfigurableNameTag{name: "Dragonfly"},
			wantValue: 0,
			wantShow:  true,
		},
		{
			name:   "empty name tag remains hidden",
			entity: testConfigurableNameTag{name: "", alwaysShow: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := protocol.NewEntityMetadata()
			new(Session).addSpecificMetadata(tt.entity, metadata)

			if got := metadata[protocol.EntityDataKeyAlwaysShowNameTag]; got != tt.wantValue {
				t.Fatalf("always-show name tag value: got %v, want %v", got, tt.wantValue)
			}
			if got := metadata.Flag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagShowName); got != tt.wantShow {
				t.Errorf("show-name flag: got %v, want %v", got, tt.wantShow)
			}
			if got := metadata.Flag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagAlwaysShowName); got != tt.wantAlways {
				t.Errorf("always-show-name flag: got %v, want %v", got, tt.wantAlways)
			}
		})
	}
}

func TestViewAlwaysShowNameTagMetadata(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	w.Do(func(tx *world.Tx) {
		e := tx.AddEntity(entity.NewText("Dragonfly", mgl64.Vec3{}))
		viewLayer := world.NewViewLayer(nil)
		s := &Session{viewLayer: viewLayer}

		viewLayer.ViewAlwaysShowNameTag(e, false)
		assertAlwaysShowNameTag(t, s.entityMetadata(e), false)

		viewLayer.ViewPublicAlwaysShowNameTag(e)
		assertAlwaysShowNameTag(t, s.entityMetadata(e), true)
	})
}

func assertAlwaysShowNameTag(t *testing.T, metadata protocol.EntityMetadata, want bool) {
	t.Helper()
	if got := metadata[protocol.EntityDataKeyAlwaysShowNameTag] == uint8(1); got != want {
		t.Errorf("always-show name tag value: got %v, want %v", got, want)
	}
	if got := metadata.Flag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagAlwaysShowName); got != want {
		t.Errorf("always-show-name flag: got %v, want %v", got, want)
	}
}

type testNameTag struct {
	name string
}

func (t testNameTag) NameTag() string {
	return t.name
}

type testConfigurableNameTag struct {
	name       string
	alwaysShow bool
}

func (t testConfigurableNameTag) NameTag() string {
	return t.name
}

func (t testConfigurableNameTag) AlwaysShowNameTag() bool {
	return t.alwaysShow
}
