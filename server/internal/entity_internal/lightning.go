package entity_internal

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	lcgRand "golang.org/x/exp/rand"
	"math"
	"math/rand"
)

//go:linkname world_performThunder github.com/df-mc/dragonfly/server/world.performThunder
//noinspection ALL
var world_performThunder func(w *world.World, pos world.ChunkPos, tr *lcgRand.Rand)

func init() {
	world_performThunder = func(w *world.World, pos world.ChunkPos, tr *lcgRand.Rand) {
		lcg := int32(tr.Uint32())

		chunkX, chunkZ := pos.X(), pos.Z()
		vec := adjustPosToNearbyEntities(w, mgl64.Vec3{float64(chunkX + (lcg & 0xf)), 0, float64(chunkZ + (lcg >> 8 & 0xf))})

		blockType := w.Block(cube.Pos{int(math.Floor(vec.X())) & 0xf, int(math.Floor(vec.Y())), int(math.Floor(vec.Z())) & 0xf})

		_, tallGrass := blockType.(block.TallGrass)
		_, flowingWater := blockType.(block.Water)
		if !tallGrass && !flowingWater {
			vec = vec.Add(mgl64.Vec3{0, 1, 0})
		}

		w.AddEntity(entity.NewLightning(vec))
	}
}

func highestBlock(w *world.World, pos mgl64.Vec3) bool {
	return w.HighestBlock(int(pos.X()), int(pos.Z())) < int(pos.Y())
}

func adjustPosToNearbyEntities(w *world.World, pos mgl64.Vec3) mgl64.Vec3 {
	pos = mgl64.Vec3{pos.X(), float64(w.HighestBlock(int(math.Floor(pos.X())), int(math.Floor(pos.Z())))), pos.Z()}
	aabb := physics.NewAABB(pos, mgl64.Vec3{pos.X(), 255, pos.Z()}).Extend(mgl64.Vec3{3, 3, 3})
	var list []mgl64.Vec3

	for _, e := range w.CollidingEntities(aabb) {
		if l, ok := e.(entity.Living); ok && l.Health() <= 0 && highestBlock(w, l.Position()) {
			list = append(list, l.Position())
		}
	}

	if len(list) > 0 {
		return list[rand.Intn(len(list))]
	} else {
		if pos.Y() == -1 {
			pos = pos.Add(mgl64.Vec3{0, 2, 0})
		}

		return pos
	}
}
