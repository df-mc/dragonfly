package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// StoneWall is a block similar to fences that prevents players from jumping over and is thinner than the usual block.
type StoneWall struct {
	// Type is the type of stone of the wall.
	Type StoneWallType
	// NorthConnection is the type of connection in the north direction of the post.
	NorthConnection WallConnectionType
	// EastConnection is the type of connection in the east direction of the post.
	EastConnection WallConnectionType
	// SouthConnection is the type of connection in the south direction of the post.
	SouthConnection WallConnectionType
	// WestConnection is the type of connection in the west direction of the post.
	WestConnection WallConnectionType
	// Post is if the wall is extended to the full height of a block or not.
	Post bool
}

// EncodeItem ...
func (w StoneWall) EncodeItem() (name string, meta int16) {
	return "minecraft:cobblestone_wall", int16(w.Type.Uint8())
}

// EncodeBlock ...
func (w StoneWall) EncodeBlock() (string, map[string]any) {
	return "minecraft:cobblestone_wall", map[string]any{
		"wall_block_type":            w.Type.String(),
		"wall_connection_type_north": w.NorthConnection.String(),
		"wall_connection_type_east":  w.EastConnection.String(),
		"wall_connection_type_south": w.SouthConnection.String(),
		"wall_connection_type_west":  w.WestConnection.String(),
		"wall_post_bit":              boolByte(w.Post),
	}
}

// Model ...
func (w StoneWall) Model() world.BlockModel {
	return model.Wall{
		NorthConnection: w.NorthConnection.String(),
		EastConnection:  w.EastConnection.String(),
		SouthConnection: w.SouthConnection.String(),
		WestConnection:  w.WestConnection.String(),
		Post:            w.Post,
	}
}

// BreakInfo ...
func (w StoneWall) BreakInfo() BreakInfo {
	hardness := 1.5
	switch w.Type {
	case SandstoneWall(), RedSandstoneWall():
		hardness = 0.8
	case BrickWall(), MossyCobblestoneWall(), RedNetherBrickWall():
		hardness = 2.0
	case EndBrickWall():
		hardness = 3.0
	}
	return newBreakInfo(hardness, pickaxeHarvestable, pickaxeEffective, oneOf(w))
}

// NeighbourUpdateTick ...
func (w StoneWall) NeighbourUpdateTick(pos, _ cube.Pos, wo *world.World) {
	w, updated := w.calculateState(wo, pos)
	if updated {
		wo.SetBlock(pos, w, nil)
	}
}

// UseOnBlock ...
func (w StoneWall) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, wo *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(wo, pos, face, w)
	if !used {
		return
	}
	w, _ = w.calculateState(wo, pos)
	place(wo, pos, w, user, ctx)
	return placed(ctx)
}

// CanDisplace ...
func (StoneWall) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (StoneWall) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// ConnectionType returns the connection type of the wall in the given direction.
func (w StoneWall) ConnectionType(direction cube.Direction) WallConnectionType {
	switch direction {
	case cube.North:
		return w.NorthConnection
	case cube.East:
		return w.EastConnection
	case cube.South:
		return w.SouthConnection
	case cube.West:
		return w.WestConnection
	}
	panic("unknown direction")
}

// WithConnectionType returns the wall with the given connection type in the given direction.
func (w StoneWall) WithConnectionType(direction cube.Direction, connection WallConnectionType) StoneWall {
	switch direction {
	case cube.North:
		w.NorthConnection = connection
	case cube.East:
		w.EastConnection = connection
	case cube.South:
		w.SouthConnection = connection
	case cube.West:
		w.WestConnection = connection
	}
	return w
}

// calculateState returns the wall with the correct state based on walls around it. If any of the connections have been
// updated then the wall and true are returned.
func (w StoneWall) calculateState(wo *world.World, pos cube.Pos) (StoneWall, bool) {
	var updated bool
	abovePos := pos.Add(cube.Pos{0, 1, 0})
	above := wo.Block(abovePos)
	for _, face := range cube.HorizontalFaces() {
		sidePos := pos.Side(face)
		side := wo.Block(sidePos)
		var connectionType WallConnectionType
		if side.Model().FaceSolid(sidePos, face.Opposite(), wo) {
			connectionType = ShortWallConnection()
			boxes := above.Model().BBox(abovePos, wo)
			for _, bb := range boxes {
				if bb.Min().Y() == 0 {
					xOverlap := bb.Min().X() < 0.75 && bb.Max().X() > 0.25
					zOverlap := bb.Min().Z() < 0.75 && bb.Max().Z() > 0.25
					var tall bool
					switch face {
					case cube.FaceNorth:
						tall = xOverlap && bb.Min().Z() < 0.25
					case cube.FaceEast:
						tall = bb.Max().X() > 0.75 && zOverlap
					case cube.FaceSouth:
						tall = xOverlap && bb.Max().Z() > 0.75
					case cube.FaceWest:
						tall = bb.Min().X() < 0.25 && zOverlap
					}
					if tall {
						connectionType = TallWallConnection()
						break
					}
				}
			}

		}
		if w.ConnectionType(face.Direction()) != connectionType {
			updated = true
			w = w.WithConnectionType(face.Direction(), connectionType)
		}
	}
	var connections int
	for _, face := range cube.HorizontalFaces() {
		if w.ConnectionType(face.Direction()) != NoWallConnection() {
			connections++
		}
	}
	var post bool
	switch above := above.(type) {
	case Air:
	case Lantern:
		post = !above.Hanging
	case Sign:
		post = !above.Attach.hanging
	case StoneWall:
		post = above.Post
	case Torch:
		post = above.Facing == cube.FaceDown
	case Wall:
		post = above.Post
	default:
		post = true
	}
	if !post {
		post = connections < 2
		if connections >= 2 {
			if w.NorthConnection != NoWallConnection() && w.SouthConnection != NoWallConnection() {
				post = w.EastConnection != NoWallConnection() || w.WestConnection != NoWallConnection()
			} else if w.EastConnection != NoWallConnection() && w.WestConnection != NoWallConnection() {
				post = w.NorthConnection != NoWallConnection() || w.SouthConnection != NoWallConnection()
			} else {
				post = true
			}
		}
	}
	if w.Post != post {
		updated = true
		w.Post = post
	}
	return w, updated
}

// allStoneWalls returns a list of all cobblestone wall types.
func allStoneWalls() (walls []world.Block) {
	for _, w := range StoneWallTypes() {
		for _, north := range WallConnectionTypes() {
			for _, east := range WallConnectionTypes() {
				for _, south := range WallConnectionTypes() {
					for _, west := range WallConnectionTypes() {
						walls = append(walls, StoneWall{Type: w,
							NorthConnection: north,
							EastConnection:  east,
							SouthConnection: south,
							WestConnection:  west,
							Post:            false,
						})
						walls = append(walls, StoneWall{Type: w,
							NorthConnection: north,
							EastConnection:  east,
							SouthConnection: south,
							WestConnection:  west,
							Post:            true,
						})
					}
				}
			}
		}
	}
	return
}
