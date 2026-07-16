package world

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestAddCustomBlocksRejectsExcessStates(t *testing.T) {
	properties := make([]any, 17)
	for i := range properties {
		properties[i] = map[string]any{
			"name": fmt.Sprintf("test:property_%d", i),
			"enum": []any{false, true},
		}
	}
	_, err := NewCustomBlockRegistry([]protocol.BlockEntry{{
		Name:       "test:block",
		Properties: map[string]any{"properties": properties},
	}})
	if err == nil || !strings.Contains(err.Error(), "exceed limit") {
		t.Fatalf("NewCustomBlockRegistry() error = %v, want state limit error", err)
	}
}
