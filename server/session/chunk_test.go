package session

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestShouldObserveChunkVisibility(t *testing.T) {
	tests := []struct {
		name   string
		result byte
		want   bool
	}{
		{name: "success", result: protocol.SubChunkResultSuccess, want: true},
		{name: "success_all_air", result: protocol.SubChunkResultSuccessAllAir, want: false},
		{name: "chunk_not_found", result: protocol.SubChunkResultChunkNotFound, want: false},
		{name: "index_out_of_bounds", result: protocol.SubChunkResultIndexOutOfBounds, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldObserveChunkVisibility(tt.result); got != tt.want {
				t.Fatalf("shouldObserveChunkVisibility(%d) = %v, want %v", tt.result, got, tt.want)
			}
		})
	}
}
