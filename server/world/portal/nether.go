package portal

import (
	"math"
	"math/rand/v2"
	"slices"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// Nether contains information about a Nether portal structure.
type Nether struct {
	w, h      int
	framed    bool
	axis      cube.Axis
	tx        *world.Tx
	spawnPos  cube.Pos
	positions []cube.Pos
}

const (
	minimumNetherPortalWidth, maximumNetherPortalWidth   = 2, 21
	minimumNetherPortalHeight, maximumNetherPortalHeight = 3, 21
	minimumArea                                          = minimumNetherPortalWidth * minimumNetherPortalHeight
)

// NetherPortalFromPos returns Nether portal information from a given position in the frame.
func NetherPortalFromPos(tx *world.Tx, pos cube.Pos) (Nether, bool) {
	if tx.World().Dimension() == world.End {
		return Nether{}, false
	}

	axis, positions, width, height, completed, ok := multiAxisScan(pos, tx, []string{"minecraft:air", "minecraft:fire"})
	if !ok {
		axis, positions, width, height, completed, ok = multiAxisScan(pos, tx, []string{"minecraft:portal"})
	}
	return Nether{
		w:         width,
		h:         height,
		spawnPos:  pos,
		positions: positions,
		framed:    completed,
		axis:      axis,
		tx:        tx,
	}, ok
}

// FindOrCreateNetherPortal finds or creates a Nether portal at the given position.
func FindOrCreateNetherPortal(tx *world.Tx, pos cube.Pos, radius int) (Nether, bool) {
	n, ok := FindNetherPortal(tx, pos, radius)
	if ok {
		return n, true
	}
	return CreateNetherPortal(tx, pos)
}

type portalBlock interface {
	Portal() world.Dimension
}

type frameBlock interface {
	Frame(dimension world.Dimension) bool
}

// FindNetherPortal searches a provided radius for a Nether portal.
func FindNetherPortal(tx *world.Tx, pos cube.Pos, radius int) (Nether, bool) {
	if tx.World().Dimension() == world.End {
		return Nether{}, false
	}

	for _, candidate := range findNetherPortalCandidates(tx, pos, radius) {
		preloadPortalSearchNeighbourhood(tx, candidate.pos)
		if found, ok := NetherPortalFromPos(tx, candidate.pos); ok {
			return found, true
		}
	}
	return Nether{}, false
}

// CreateNetherPortal creates a Nether portal at the given position.
func CreateNetherPortal(tx *world.Tx, pos cube.Pos) (Nether, bool) {
	if tx.World().Dimension() == world.End {
		return Nether{}, false
	}

	resultPos, random, distance, a, r := pos, rand.IntN(4), -1.0, 0, tx.Range()
	searchValidArea := func(directions int, valid func(pos cube.Pos, riv int, coEff1, coEff2 int) bool) {
		for tempX := pos.X() - 16; tempX <= pos.X()+16; tempX++ {
			offsetX := float64(tempX-pos.X()) + 0.5
			for tempZ := pos.Z() - 16; tempZ <= pos.Z()+16; tempZ++ {
				offsetZ := float64(tempZ-pos.Z()) + 0.5
				for tempY := r.Max() - 1; tempY >= r.Min(); tempY-- {
					entryPos := cube.Pos{tempX, tempY, tempZ}
					if tx.Block(entryPos) != air() {
						continue
					}

					for tempY > r.Min() && tx.Block(entryPos.Side(cube.FaceDown)) == air() {
						tempY--
						entryPos[1]--
					}

					for riv := random; riv < random+directions; riv++ {
						coEff1 := riv % 2
						coEff2 := 1 - coEff1

						if !valid(entryPos, riv, coEff1, coEff2) {
							break
						}

						offsetY := float64(tempY-pos.Y()) + 0.5
						newDist := offsetX*offsetX + offsetY*offsetY + offsetZ*offsetZ
						if distance < 0 || newDist < distance {
							distance = newDist
							a = riv % directions
							resultPos = cube.Pos{tempX, tempY, tempZ}
						}
					}
				}
			}
		}
	}

	searchValidArea(4, func(pos cube.Pos, riv int, coEff1, coEff2 int) bool {
		if riv%4 >= 2 {
			coEff1 = -coEff1
			coEff2 = -coEff2
		}

		for safeSpace1 := 0; safeSpace1 < 3; safeSpace1++ {
			for safeSpace2 := -1; safeSpace2 < 3; safeSpace2++ {
				for height := -1; height < 4; height++ {
					b := tx.Block(cube.Pos{
						pos.X() + safeSpace2*coEff1 + safeSpace1*coEff2,
						pos.Y() + height,
						pos.Z() + safeSpace2*coEff2 - safeSpace1*coEff1,
					})
					_, solid := b.Model().(model.Solid)
					if (height < 0 && !solid) || (height >= 0 && b != air()) {
						return false
					}
				}
			}
		}
		return true
	})

	if distance < 0 {
		searchValidArea(2, func(pos cube.Pos, riv int, coEff1, coEff2 int) bool {
			for safeSpace := 0; safeSpace < 3; safeSpace++ {
				for height := -1; height < 4; height++ {
					b := tx.Block(cube.Pos{
						pos.X() + safeSpace*coEff1,
						pos.Y() + height,
						pos.Z() + safeSpace*coEff2,
					})
					_, solid := b.Model().(model.Solid)
					if (height < 0 && !solid) || (height >= 0 && b != air()) {
						return false
					}
				}
			}
			return true
		})
	}

	coEff1 := a % 2
	coEff2 := 1 - coEff1
	if a%4 >= 2 {
		coEff1 = -coEff1
		coEff2 = -coEff2
	}

	axis := cube.X
	if coEff1 == 0 {
		axis = cube.Z
	}

	if distance < 0 {
		resultPos[1] = int(math.Min(math.Max(float64(resultPos[1]), 70), float64(r.Max()-10)))
		for safeBeforeAfter := -1; safeBeforeAfter <= 1; safeBeforeAfter++ {
			for safeWidth := 0; safeWidth < 2; safeWidth++ {
				for height := -1; height < 3; height++ {
					entryPos := cube.Pos{
						resultPos.X() + safeWidth*coEff1 + safeBeforeAfter*coEff2,
						resultPos.Y() + height,
						resultPos.Z() + safeWidth*coEff2 - safeBeforeAfter*coEff1,
					}

					tx.SetBlock(entryPos, nil, nil)
					if height < 0 {
						tx.SetBlock(entryPos, obsidian(), nil)
					}
				}
			}
		}
	}

	var positions []cube.Pos
	for width := -1; width < 3; width++ {
		for height := -1; height < 4; height++ {
			entryPos := cube.Pos{
				resultPos.X() + width*coEff1,
				resultPos.Y() + height,
				resultPos.Z() + width*coEff2,
			}

			if width == -1 || width == 2 || height == -1 || height == 3 {
				tx.SetBlock(entryPos, obsidian(), nil)
				continue
			}
			positions = append(positions, entryPos)
			tx.SetBlock(entryPos, portal(axis), nil)
		}
	}

	return Nether{
		w:         minimumNetherPortalWidth,
		h:         minimumNetherPortalHeight,
		framed:    true,
		spawnPos:  resultPos,
		positions: positions,
		axis:      axis,
		tx:        tx,
	}, true
}

// Bounds returns the inner portal width and height.
func (n Nether) Bounds() (int, int) {
	return n.w, n.h
}

// Axis returns the portal plane axis.
func (n Nether) Axis() cube.Axis {
	return n.axis
}

// Activate fills the framed portal with active portal blocks.
func (n Nether) Activate() {
	for _, pos := range n.Positions() {
		n.tx.SetBlock(pos, portal(n.axis), nil)
	}
}

type netherPortalCandidate struct {
	pos    cube.Pos
	distSq float64
}

func findNetherPortalCandidates(tx *world.Tx, pos cube.Pos, radius int) []netherPortalCandidate {
	portalX, portalZ := world.BlockRuntimeID(portal(cube.X)), world.BlockRuntimeID(portal(cube.Z))
	minChunkX, maxChunkX := floorDiv(pos.X()-radius, 16), floorDiv(pos.X()+radius, 16)
	minChunkZ, maxChunkZ := floorDiv(pos.Z()-radius, 16), floorDiv(pos.Z()+radius, 16)

	candidates := make([]netherPortalCandidate, 0, 8)
	for chunkX := minChunkX; chunkX <= maxChunkX; chunkX++ {
		for chunkZ := minChunkZ; chunkZ <= maxChunkZ; chunkZ++ {
			col, ok, err := tx.LoadExistingChunk(world.ChunkPos{int32(chunkX), int32(chunkZ)})
			if err != nil || !ok {
				continue
			}
			candidates = append(candidates, scanNetherPortalCandidatesInColumn(col, pos, radius, chunkX, chunkZ, portalX, portalZ)...)
		}
	}
	slices.SortFunc(candidates, func(a, b netherPortalCandidate) int {
		switch {
		case a.distSq < b.distSq:
			return -1
		case a.distSq > b.distSq:
			return 1
		default:
			return 0
		}
	})
	return candidates
}

func scanNetherPortalCandidatesInColumn(col *world.Column, center cube.Pos, radius, chunkX, chunkZ int, portalX, portalZ uint32) []netherPortalCandidate {
	baseX, baseZ := chunkX<<4, chunkZ<<4
	minX, maxX := max(baseX, center.X()-radius), min(baseX+15, center.X()+radius)
	minZ, maxZ := max(baseZ, center.Z()-radius), min(baseZ+15, center.Z()+radius)
	if minX > maxX || minZ > maxZ {
		return nil
	}

	candidates := make([]netherPortalCandidate, 0, 4)
	for subIndex := len(col.Sub()) - 1; subIndex >= 0; subIndex-- {
		sub := col.Sub()[subIndex]
		if sub.Empty() {
			continue
		}
		baseY := int(col.SubY(int16(subIndex)))
		for yOffset := 15; yOffset >= 0; yOffset-- {
			y := baseY + yOffset
			for x := minX; x <= maxX; x++ {
				localX := uint8(x - baseX)
				for z := minZ; z <= maxZ; z++ {
					rid := sub.Block(byte(localX), byte(yOffset), byte(z-baseZ), 0)
					if rid != portalX && rid != portalZ {
						continue
					}
					dx, dy, dz := float64(x-center.X()), float64(y-center.Y()), float64(z-center.Z())
					candidates = append(candidates, netherPortalCandidate{
						pos:    cube.Pos{x, y, z},
						distSq: dx*dx + dy*dy + dz*dz,
					})
				}
			}
		}
	}
	return candidates
}

func preloadPortalSearchNeighbourhood(tx *world.Tx, pos cube.Pos) {
	chunkX, chunkZ := floorDiv(pos.X(), 16), floorDiv(pos.Z(), 16)
	for neighbourX := chunkX - 1; neighbourX <= chunkX+1; neighbourX++ {
		for neighbourZ := chunkZ - 1; neighbourZ <= chunkZ+1; neighbourZ++ {
			_, _, _ = tx.LoadExistingChunk(world.ChunkPos{int32(neighbourX), int32(neighbourZ)})
		}
	}
}

func floorDiv(x, y int) int {
	if y == 0 {
		panic("division by zero")
	}
	q := x / y
	r := x % y
	if r != 0 && ((r < 0) != (y < 0)) {
		q--
	}
	return q
}

// Deactivate removes all active portal blocks from the frame.
func (n Nether) Deactivate() {
	for _, pos := range n.Positions() {
		n.tx.SetBlock(pos, nil, nil)
	}
}

// Framed reports if the portal scan found a complete obsidian frame.
func (n Nether) Framed() bool {
	return n.framed
}

// Activated reports if all scanned inner positions currently contain active portal blocks.
func (n Nether) Activated() bool {
	for _, pos := range n.Positions() {
		if n.tx.Block(pos) != portal(n.axis) {
			return false
		}
	}
	return true
}

// Spawn returns the base spawn position associated with the portal.
func (n Nether) Spawn() cube.Pos {
	return n.spawnPos
}

// Positions returns all inner portal positions.
func (n Nether) Positions() []cube.Pos {
	return n.positions
}
