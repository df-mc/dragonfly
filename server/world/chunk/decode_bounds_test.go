package chunk

import (
	"bytes"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

func TestNetworkDecode_RejectsInvalidCount(t *testing.T) {
	r := cube.Range{0, 127}
	count := len(New(testBlockRegistry{}, r).Sub()) + 1
	payload := bytes.Repeat([]byte{8, 0}, count)

	if _, err := NetworkDecode(testBlockRegistry{}, payload, count, r); err == nil {
		t.Fatal("expected an error for a sub-chunk count larger than the chunk range")
	}
}

func TestNetworkDecode_RejectsOutOfRangeVersion9Index(t *testing.T) {
	tests := []struct {
		name  string
		index byte
	}{
		{name: "above range", index: 8},
		{name: "below range", index: 0xff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := []byte{9, 0, tt.index}
			if _, err := NetworkDecode(testBlockRegistry{}, payload, 1, cube.Range{0, 127}); err == nil {
				t.Fatal("expected an error for an encoded sub-chunk index outside the chunk range")
			}
		})
	}
}
