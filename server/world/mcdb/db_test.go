package mcdb

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
)

func TestPlayerSpawnDimensionPersists(t *testing.T) {
	id := uuid.New()
	want := world.PlayerSpawn{Pos: cube.Pos{17, 71, -9}, Dim: world.Nether}
	db, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open DB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.SavePlayerSpawn(id, want); err != nil {
		t.Fatalf("SavePlayerSpawn(): %v", err)
	}
	got, exists, err := db.LoadPlayerSpawn(id)
	if err != nil {
		t.Fatalf("LoadPlayerSpawn(): %v", err)
	}
	if !exists || got != want {
		t.Fatalf("LoadPlayerSpawn() = %v, %t, want %v, true", got, exists, want)
	}
	data, _, _, err := db.loadPlayerData(id)
	if err != nil {
		t.Fatalf("loadPlayerData(): %v", err)
	}
	for _, key := range []string{"SpawnBlockPositionX", "SpawnBlockPositionY", "SpawnBlockPositionZ", "SpawnDimension", "SpawnX", "SpawnY", "SpawnZ"} {
		if _, ok := data[key].(int32); !ok {
			t.Fatalf("saved field %q = %T, want int32", key, data[key])
		}
	}
}

func TestPlayerSpawnFromDataCompatibility(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name       string
		data       map[string]any
		want       world.PlayerSpawn
		wantExists bool
	}{
		{
			name: "dragonfly legacy",
			data: map[string]any{"SpawnX": int32(4), "SpawnY": int32(5), "SpawnZ": int32(6)},
			want: world.PlayerSpawn{Pos: cube.Pos{4, 5, 6}, Dim: world.Overworld}, wantExists: true,
		},
		{
			name: "vanilla legacy",
			data: map[string]any{"BedPositionX": int32(7), "BedPositionY": int32(8), "BedPositionZ": int32(9)},
			want: world.PlayerSpawn{Pos: cube.Pos{7, 8, 9}, Dim: world.Overworld}, wantExists: true,
		},
		{
			name: "invalid dimension",
			data: map[string]any{
				"SpawnBlockPositionX": int32(1), "SpawnBlockPositionY": int32(2), "SpawnBlockPositionZ": int32(3),
				"SpawnDimension": int32(3),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, exists, err := playerSpawnFromData(test.data, id)
			if err != nil {
				t.Fatalf("playerSpawnFromData(): %v", err)
			}
			if got != test.want || exists != test.wantExists {
				t.Fatalf("playerSpawnFromData() = %v, %t, want %v, %t", got, exists, test.want, test.wantExists)
			}
		})
	}
}
