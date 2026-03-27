package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/pprof"
	"time"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func main() {
	var (
		seed   = flag.Int64("seed", 0, "world seed")
		radius = flag.Int("radius", 8, "chunk loader radius")
		x      = flag.Int("x", 8192, "benchmark block x")
		z      = flag.Int("z", 8192, "benchmark block z")
		y      = flag.Int("y", 96, "benchmark block y")
		loadN  = flag.Int("load", 0, "chunks to load; 0 means all chunks in radius")
		cpu    = flag.String("cpuprofile", "", "optional CPU profile output path")
	)
	flag.Parse()

	if *cpu != "" {
		f, err := os.Create(*cpu)
		if err != nil {
			fmt.Fprintln(os.Stderr, "worldgen bench failed:", err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			_ = f.Close()
			fmt.Fprintln(os.Stderr, "worldgen bench failed:", err)
			os.Exit(1)
		}
		defer func() {
			pprof.StopCPUProfile()
			_ = f.Close()
		}()
	}

	userConf := server.DefaultConfig()
	userConf.World.SaveData = false
	userConf.Players.SaveData = false
	userConf.World.Seed = *seed

	conf, err := userConf.Config(slog.Default())
	if err != nil {
		fmt.Fprintln(os.Stderr, "worldgen bench failed:", err)
		os.Exit(1)
	}
	conf.Listeners = nil

	startupStart := time.Now()
	srv := conf.New()
	w := srv.World()
	startupDuration := time.Since(startupStart)
	loader := world.NewLoader(*radius, w, world.NopViewer{})

	totalChunks := *loadN
	if totalChunks <= 0 {
		totalChunks = chunksInRadius(*radius)
	}

	var (
		loadDuration  time.Duration
		totalDuration time.Duration
		verifyErr     error
	)

	target := mgl64.Vec3{float64(*x), float64(*y), float64(*z)}
	targetChunk := world.ChunkPos{int32(*x >> 4), int32(*z >> 4)}
	loadStart := time.Now()
	deadline := loadStart.Add(30 * time.Second)
	for {
		<-w.Exec(func(tx *world.Tx) {
			loader.Move(tx, target)
			loader.Load(tx, totalChunks)
			if _, ok := loader.Chunk(targetChunk); ok {
				// Touch a block to ensure the column was materialized.
				_ = tx.Block(cube.Pos{*x, *y, *z})
			}
		})
		if _, ok := loader.Chunk(targetChunk); ok {
			break
		}
		if time.Now().After(deadline) {
			verifyErr = fmt.Errorf("target chunk %v was not loaded before timeout", targetChunk)
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	loadDuration = time.Since(loadStart)
	totalDuration = startupDuration + loadDuration

	<-w.Exec(func(tx *world.Tx) {
		loader.Close(tx)
	})

	if verifyErr != nil {
		fmt.Fprintln(os.Stderr, "worldgen bench failed:", verifyErr)
		os.Exit(1)
	}

	chunksPerSecond := float64(totalChunks)
	if loadDuration > 0 {
		chunksPerSecond /= loadDuration.Seconds()
	}
	fmt.Printf(
		"seed=%d radius=%d chunks=%d pos=(%d,%d,%d) startup_duration=%s load_duration=%s total_duration=%s chunks_per_second=%.2f\n",
		*seed,
		*radius,
		totalChunks,
		*x,
		*y,
		*z,
		startupDuration,
		loadDuration,
		totalDuration,
		chunksPerSecond,
	)
}

func chunksInRadius(radius int) int {
	count := 0
	r := radius * radius
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			if x*x+z*z <= r {
				count++
			}
		}
	}
	return count
}
