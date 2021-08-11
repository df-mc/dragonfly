package entity_internal

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
)

//go:linkname world_performThunder github.com/df-mc/dragonfly/server/world.performThunder
//noinspection ALL
var world_performThunder func(w *world.World, pos world.ChunkPos)

func init() {
	world_performThunder = func(w *world.World, pos world.ChunkPos) {
		LCG := int32(w.GetUpdateLGC() >> 2)

		chunkX, chunkZ := pos.X(), pos.Z()
		vec := adjustPosToNearbyEntities(w, mgl64.Vec3{float64(chunkX + (LCG & 0xf)), 0, float64(chunkZ + (LCG >> 8 & 0xf))})

		blockType := w.Block(cube.Pos{int(math.Floor(vec.X())) & 0xf, int(math.Floor(vec.Y())), int(math.Floor(vec.Z())) & 0xf})

		_, tallGrass := blockType.(block.TallGrass)
		_, flowingWater := blockType.(block.Water)
		if !tallGrass && !flowingWater {
			vec = vec.Add(mgl64.Vec3{0, 1, 0})
		}

		w.AddEntity(entity.NewLightning(vec))
	}
}

func canBlockSeeSky(w *world.World, pos mgl64.Vec3) bool {
	return w.HighestBlock(int(pos.X()), int(pos.Z())) < int(pos.Y())
}

func adjustPosToNearbyEntities(w *world.World, pos mgl64.Vec3) mgl64.Vec3 {
	pos = mgl64.Vec3{pos.X(), float64(w.HighestBlock(int(math.Floor(pos.X())), int(math.Floor(pos.Z())))), pos.Z()}
	aabb := physics.NewAABB(pos, mgl64.Vec3{pos.X(), 255, pos.Z()}).Extend(mgl64.Vec3{3, 3, 3})
	var list []mgl64.Vec3

	for _, e := range w.CollidingEntities(aabb) {
		if l, ok := e.(entity.Living); ok && l.Health() <= 0 && canBlockSeeSky(w, l.Position()) {
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

//go:linkname block_setBlocksOnFire github.com/df-mc/dragonfly/server/entity.setBlocksOnFire
//noinspection ALL
var block_setBlocksOnFire func(w *world.World, lPos mgl64.Vec3)

func init() {
	block_setBlocksOnFire = func(w *world.World, lPos mgl64.Vec3) {
		_, isNormal := w.Difficulty().(world.DifficultyNormal)
		_, isHard := w.Difficulty().(world.DifficultyHard)
		if isNormal || isHard { // difficulty >= 2
			bPos := cube.PosFromVec3(lPos)
			b := w.Block(bPos)
			_, isAir := b.(block.Air)
			_, isTallGrass := b.(block.TallGrass)
			if isAir || isTallGrass {
				below := w.Block(bPos.Side(cube.FaceDown))
				if below.Model().FaceSolid(bPos, cube.FaceUp, w) || block.NeighboursFlammable(bPos, w) {
					w.PlaceBlock(bPos, block.Fire{})
				}
			}
		}
	}
}
