package leveldat

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func TestDataUnmarshalEditorWorldTags(t *testing.T) {
	payload, err := nbt.MarshalEncoding(map[string]any{
		"LevelName":                              "World",
		"allowAnonymousBlockDropsInEditorWorlds": true,
		"playerwaypoints":                        int32(2),
		"serverEditorConnectionPolicy":           int32(1),
	}, nbt.LittleEndian)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	var got Data
	if err = nbt.UnmarshalEncoding(payload, &got, nbt.LittleEndian); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if !got.AllowAnonymousBlockDropsInEditorWorlds {
		t.Fatal("expected allowAnonymousBlockDropsInEditorWorlds=true")
	}
	if got.PlayerWaypoints != 2 {
		t.Fatalf("playerwaypoints: got %d, want 2", got.PlayerWaypoints)
	}
	if got.ServerEditorConnectionPolicy != 1 {
		t.Fatalf("serverEditorConnectionPolicy: got %d, want 1", got.ServerEditorConnectionPolicy)
	}
}
