package portal

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// endSpawnX, endSpawnY and endSpawnZ are the centre of the End arrival platform. The platform extends two blocks in
// each horizontal direction at endSpawnY, so the obsidian fills the 5x5 area at y=endSpawnY-1 and the player arrives
// at endSpawnY+0.
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
// above it. Run on every Overworld→End travel to match vanilla, which regenerates the platform unconditionally — any
// player builds in the area are wiped.
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

// endFrameBlock is the local interface implemented by block.EndPortalFrame. It avoids importing server/block from this
// package, which would create an import cycle.
type endFrameBlock interface {
	world.Block
	EndPortalFrameState() (eye bool, facing cube.Direction)
}

// EndPortalFromPos validates a complete twelve-frame End portal ring starting from any frame on the ring. Frames must
// face TOWARD the centre (vanilla Bedrock: cardinal_direction = opposite of the placing player's facing, so a player
// standing at the centre and placing outward yields inward-facing frames — the only valid configuration). The starting
// frame may be the left, middle, or right of its side, so each of the three candidate centres along the tangent is
// tried. Returns ok=true only if the ring is complete (all twelve ring positions hold an EndPortalFrame with Eye=true
// and the correct inward Facing).
func EndPortalFromPos(tx *world.Tx, framePos cube.Pos) (End, bool) {
	f, ok := tx.Block(framePos).(endFrameBlock)
	if !ok {
		return End{}, false
	}
	_, facing := f.EndPortalFrameState()

	// Frames face toward the centre, so walk in the Facing direction twice to reach the row of three candidate centres.
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

// ActivateEndPortal places end_portal blocks in the 3x3 interior if and only if a complete twelve-frame ring exists
// around `framePos`. If interior positions already hold end_portal blocks, they are left untouched (idempotent). The
// starting frame must be one of the twelve ring positions; its Facing is used to derive the ring centre.
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

// expectedEndRingFrames returns the twelve canonical (position, facing) pairs around the centre. Each frame sits on
// the side of the centre indicated by `side` and must face TOWARD the centre (Facing = side.Opposite()), matching the
// vanilla Bedrock requirement that frames be placed by a player standing at or near the centre and looking outward.
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

// tangentFace returns a unit horizontal Face perpendicular to side. Used to walk along the row of three frames on a
// side of the ring. Sign does not matter: the {-1, 0, +1} step set covers the same three positions either way.
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
