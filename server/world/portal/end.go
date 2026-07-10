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

// End contains information about a complete End portal ring. Values returned from this package are tied to the
// transaction that produced them and must not be retained after that transaction finishes.
type End struct {
	tx       *world.Tx
	interior []cube.Pos
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

// EndPortalFromPos returns End portal information from any frame on a complete twelve-frame ring. All twelve frames
// must hold an eye and face toward the centre, as in vanilla.
func EndPortalFromPos(tx *world.Tx, framePos cube.Pos) (End, bool) {
	f, ok := tx.Block(framePos).(endFrameBlock)
	if !ok {
		return End{}, false
	}
	_, facing := f.EndPortalFrameState()

	// The frame may be the left, middle or right of its side: walk inward twice, then try the three candidate centres.
	inward := facing.Face()
	tangent := tangentFace(facing)
	base := framePos.Side(inward).Side(inward)
	for k := -1; k <= 1; k++ {
		if e, ok := matchEndRing(tx, stepAlong(base, tangent, k)); ok {
			return e, true
		}
	}
	return End{}, false
}

// matchEndRing returns a complete End if the twelve canonical ring positions around centre all hold matching frames.
func matchEndRing(tx *world.Tx, center cube.Pos) (End, bool) {
	frames := expectedEndRingFrames(center)
	for _, want := range frames {
		b, ok := tx.Block(want.pos).(endFrameBlock)
		if !ok {
			return End{}, false
		}
		eye, gotFacing := b.EndPortalFrameState()
		if !eye || gotFacing != want.facing {
			return End{}, false
		}
	}
	return End{
		tx:       tx,
		interior: endRingInterior(center),
	}, true
}

// ActivateEndPortal fills the 3x3 interior with end_portal blocks if a complete twelve-frame ring exists around the
// frame at the position passed.
func ActivateEndPortal(tx *world.Tx, framePos cube.Pos) bool {
	p, ok := EndPortalFromPos(tx, framePos)
	if !ok {
		return false
	}
	p.activate()
	return true
}

// activate fills the 3x3 interior with end_portal blocks. Positions that already hold an end_portal are skipped.
func (e End) activate() {
	ep := endPortal()
	for _, pos := range e.interior {
		if e.tx.Block(pos) == ep {
			continue
		}
		e.tx.SetBlock(pos, ep, nil)
	}
}

// expectedEndRingFrames returns the twelve (position, facing) pairs a complete ring around the centre must have, with
// every frame facing toward the centre.
func expectedEndRingFrames(center cube.Pos) []endRingFrame {
	frames := make([]endRingFrame, 0, 12)
	for _, side := range cube.Directions() {
		base := center.Side(side.Face()).Side(side.Face())
		t := tangentFace(side)
		inward := side.Opposite()
		for i := -1; i <= 1; i++ {
			frames = append(frames, endRingFrame{pos: stepAlong(base, t, i), facing: inward})
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

// tangentFace returns a horizontal Face perpendicular to side, used to walk along the three frames on a ring side.
func tangentFace(side cube.Direction) cube.Face {
	return side.RotateRight().Face()
}

// stepAlong returns p offset by n steps along face. Negative n walks the opposite direction.
func stepAlong(p cube.Pos, face cube.Face, n int) cube.Pos {
	if n < 0 {
		face, n = face.Opposite(), -n
	}
	for range n {
		p = p.Side(face)
	}
	return p
}

// endPortal returns the end_portal block.
func endPortal() world.Block {
	p, ok := world.BlockByName("minecraft:end_portal", nil)
	if !ok {
		panic("could not find end_portal block")
	}
	return p
}
