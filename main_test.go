package main

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/go-gl/mathgl/mgl64"
	"testing"
)

func Benchmark(b *testing.B) {
	inputs := []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384}

	for _, v := range inputs {
		b.StopTimer()
		p := player.New("Test", skin.New(0, 0), mgl64.Vec3{0, 0, 0})

		for i := 0; i < v; i++ {
			p.AddHandler(player.NopHandler{})
		}

		evt := player.EventBlockBreak{Player: p}

		b.StartTimer()
		b.Run(fmt.Sprintf("%d handlers:", v), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				p.HandleBlockBreak(evt)
			}
		})
	}
}

/*
var p = player.New("Test", skin.Skin{}, mgl64.Vec3{0, 0, 0})

func TestMain(m *testing.M) {
    for i := 0; i < 1000000; i++ {
        p.AddHandler(player.NopHandler{})
    }
    fmt.Println("hi")
    TestSomethingElse()
}

func TestSomething(t *testing.T) { fmt.Println("hi") }

func TestSomethingElse(t *testing.T) { t.Fail() }

func BenchmarkMultiHandler(b *testing.B) {
    fmt.Println("hi")
    p.HandleBlockBreak(player.EventBlockBreak{})
}
*/
