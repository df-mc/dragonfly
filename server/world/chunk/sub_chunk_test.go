package chunk

import (
	"testing"
	"time"
)

// TestLayerAddressableRange verifies that Layer creates every layer up to the highest addressable index without
// looping, and that it refuses an index that cannot be encoded.
func TestLayerAddressableRange(t *testing.T) {
	tests := []struct {
		name         string
		layer        uint8
		wantStorages int
		wantPanic    bool
	}{
		{name: "block layer", layer: 0, wantStorages: 1},
		{name: "waterlogging layer", layer: 1, wantStorages: 2},
		{name: "highest addressable layer", layer: MaxLayers - 1, wantStorages: MaxLayers},
		{name: "layer past the encodable maximum", layer: MaxLayers, wantPanic: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, done, panicked := NewSubChunk(0), make(chan int, 1), make(chan struct{}, 1)
			go func() {
				defer func() {
					if recover() != nil {
						panicked <- struct{}{}
					}
				}()
				sub.Layer(tt.layer)
				done <- len(sub.storages)
			}()

			select {
			case got := <-done:
				if tt.wantPanic {
					t.Fatalf("Layer(%v) grew to %v storages, want a panic", tt.layer, got)
				}
				if got != tt.wantStorages {
					t.Fatalf("Layer(%v) grew to %v storages, want %v", tt.layer, got, tt.wantStorages)
				}
			case <-panicked:
				if !tt.wantPanic {
					t.Fatalf("Layer(%v) panicked, want %v storages", tt.layer, tt.wantStorages)
				}
			case <-time.After(time.Second * 5):
				t.Fatalf("Layer(%v) did not return: the layer index comparison wrapped and looped", tt.layer)
			}
		})
	}
}

// TestBlockReadsEveryLayer verifies that Block reads back a block written to the highest addressable layer, and
// reports a layer that was never created as air.
func TestBlockReadsEveryLayer(t *testing.T) {
	sub := NewSubChunk(0)
	sub.SetBlock(1, 2, 3, MaxLayers-1, 1)

	if got := sub.Block(1, 2, 3, MaxLayers-1); got != 1 {
		t.Fatalf("Block(layer %v) = %v, want 1", MaxLayers-1, got)
	}
	if got := sub.Block(1, 2, 3, MaxLayers-2); got != 0 {
		t.Fatalf("Block(layer %v) = %v, want 0 (air)", MaxLayers-2, got)
	}
}
