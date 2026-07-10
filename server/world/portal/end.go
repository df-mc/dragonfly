package portal

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// endSpawnX, endSpawnY and endSpawnZ are the centre of the End arrival platform.
const (
	endSpawnX = 100
	endSpawnY = 49
	endSpawnZ = 0
)

// EndSpawnPosition returns the Bedrock End platform arrival position: y=49 for players and y=50 for other entities.
func EndSpawnPosition(player bool) mgl64.Vec3 {
	y := endSpawnY + 1
	if player {
		y = endSpawnY
	}
	return mgl64.Vec3{float64(endSpawnX) + 0.5, float64(y), float64(endSpawnZ) + 0.5}
}

// GenerateEndSpawnPlatform builds the 5x5 obsidian arrival platform at (100, 48, 0) and clears the 5x5x3 air column
// above it. It runs on every travel into the End, matching vanilla's unconditional regeneration.
func GenerateEndSpawnPlatform(tx *world.Tx) {
	ob := obsidian()
	for dx := -2; dx <= 2; dx++ {
		for dz := -2; dz <= 2; dz++ {
			tx.SetBlock(cube.Pos{endSpawnX + dx, endSpawnY - 1, endSpawnZ + dz}, ob, nil)
			for dy := 0; dy < 3; dy++ {
				tx.SetBlock(cube.Pos{endSpawnX + dx, endSpawnY + dy, endSpawnZ + dz}, nil, nil)
			}
		}
	}
}

// endRingFrame is one of the twelve canonical ring positions and the Facing each frame must have.
type endRingFrame struct {
	pos    cube.Pos
	facing cube.Direction
}

// endFrameBlock is implemented by block.EndPortalFrame, which cannot be imported here directly.
type endFrameBlock interface {
	world.Block
	EndPortalFrameState() (eye bool, facing cube.Direction)
}

// ActivateEndPortal fills the 3x3 interior with end_portal blocks if a complete twelve-frame ring exists around the
// frame at the position passed. All twelve frames must hold an eye and face toward the centre, as in vanilla.
func ActivateEndPortal(tx *world.Tx, framePos cube.Pos) bool {
	f, ok := tx.Block(framePos).(endFrameBlock)
	if !ok {
		return false
	}
	_, facing := f.EndPortalFrameState()

	// The frame may be the left, middle or right of its side: walk inward twice, then try the three candidate centres.
	inward, tangent := facing.Face(), facing.RotateRight().Face()
	base := framePos.Side(inward).Side(inward)
	for _, center := range []cube.Pos{base.Side(tangent.Opposite()), base, base.Side(tangent)} {
		interior, ok := matchEndRing(tx, center)
		if !ok {
			continue
		}
		ep := endPortal()
		for _, pos := range interior {
			if tx.Block(pos) != ep {
				tx.SetBlock(pos, ep, nil)
			}
		}
		return true
	}
	return false
}

// matchEndRing returns the 3x3 interior positions if the twelve canonical ring positions around centre all hold
// matching frames.
func matchEndRing(tx *world.Tx, center cube.Pos) ([]cube.Pos, bool) {
	for _, want := range expectedEndRingFrames(center) {
		b, ok := tx.Block(want.pos).(endFrameBlock)
		if !ok {
			return nil, false
		}
		eye, facing := b.EndPortalFrameState()
		if !eye || facing != want.facing {
			return nil, false
		}
	}
	return endRingInterior(center), true
}

// expectedEndRingFrames returns the twelve (position, facing) pairs a complete ring around the centre must have, with
// every frame facing toward the centre.
func expectedEndRingFrames(center cube.Pos) []endRingFrame {
	frames := make([]endRingFrame, 0, 12)
	for _, side := range cube.Directions() {
		base := center.Side(side.Face()).Side(side.Face())
		tangent := side.RotateRight().Face()
		inward := side.Opposite()
		for _, pos := range []cube.Pos{base.Side(tangent.Opposite()), base, base.Side(tangent)} {
			frames = append(frames, endRingFrame{pos: pos, facing: inward})
		}
	}
	return frames
}

// endRingInterior returns the nine 3x3 interior positions on the y plane of centre.
func endRingInterior(center cube.Pos) []cube.Pos {
	out := make([]cube.Pos, 0, 9)
	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			out = append(out, center.Add(cube.Pos{dx, 0, dz}))
		}
	}
	return out
}

// endPortal returns the end_portal block.
func endPortal() world.Block {
	p, ok := world.BlockByName("minecraft:end_portal", nil)
	if !ok {
		panic("could not find end_portal block")
	}
	return p
}
