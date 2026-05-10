package world

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

var redstoneGraphIDBenchmarkResult uint64

func BenchmarkRedstoneGraphID(b *testing.B) {
	nodes := redstoneBenchmarkNodes(256)
	edges := redstoneBenchmarkEdges(len(nodes))

	b.ReportAllocs()
	b.SetBytes(int64(len(nodes)*32 + len(edges)*24))
	b.ResetTimer()

	var result uint64
	for i := 0; i < b.N; i++ {
		result = redstoneGraphID(nodes, edges)
	}
	redstoneGraphIDBenchmarkResult = result
}

func redstoneBenchmarkNodes(n int) []redstoneNode {
	nodes := make([]redstoneNode, n)
	for i := range nodes {
		nodes[i] = redstoneNode{
			pos:    cube.Pos{i%16 - 8, i/16 - 8, (i * 7) % 16},
			power:  i % 16,
			source: i%3 == 0,
			sink:   i%5 == 0,
		}
	}
	return nodes
}

func redstoneBenchmarkEdges(nodes int) []redstoneEdge {
	edges := make([]redstoneEdge, 0, nodes*2)
	for i := 0; i < nodes; i++ {
		if i+1 < nodes {
			edges = append(edges, redstoneEdge{from: i, to: i + 1, weight: i%3 + 1})
		}
		if i+16 < nodes {
			edges = append(edges, redstoneEdge{from: i, to: i + 16, weight: i%5 + 1})
		}
	}
	return edges
}
