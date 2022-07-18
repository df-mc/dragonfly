package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Wall is a block similar to fences that prevents players from jumping over and is thinner than the usual block. It is
// available for many blocks and all types connect together as if they were the same type.
type Wall struct {
	// Block is the block to use for the type of wall.
	Block world.Block
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
func (w Wall) EncodeItem() (string, int16) {
	name, meta := encodeWallBlock(w.Block)
	if meta == 0 {
		return "minecraft:" + name + "_wall", 0
	}
	return "minecraft:cobblestone_wall", meta
}

// EncodeBlock ...
func (w Wall) EncodeBlock() (string, map[string]any) {
	properties := map[string]any{
		"wall_connection_type_north": w.NorthConnection.String(),
		"wall_connection_type_east":  w.EastConnection.String(),
		"wall_connection_type_south": w.SouthConnection.String(),
		"wall_connection_type_west":  w.WestConnection.String(),
		"wall_post_bit":              boolByte(w.Post),
	}
	name, meta := encodeWallBlock(w.Block)
	if meta > 0 || name == "cobblestone" {
		properties["wall_block_type"] = name
		name = "cobblestone"
	}
	return "minecraft:" + name + "_wall", properties
}

// Model ...
func (w Wall) Model() world.BlockModel {
	return model.Wall{
		NorthConnection: w.NorthConnection.String(),
		EastConnection:  w.EastConnection.String(),
		SouthConnection: w.SouthConnection.String(),
		WestConnection:  w.WestConnection.String(),
		Post:            w.Post,
	}
}

// BreakInfo ...
func (w Wall) BreakInfo() BreakInfo {
	hardness := 2.0
	name, _ := encodeWallBlock(w.Block)
	if name == "cobbled_deepslate" || name == "deepslate_brick" || name == "deepslate_tile" || name == "polished_deepslate" {
		hardness = 3.5
	}
	return newBreakInfo(hardness, pickaxeHarvestable, pickaxeEffective, oneOf(w))
}

// NeighbourUpdateTick ...
func (w Wall) NeighbourUpdateTick(pos, _ cube.Pos, wo *world.World) {
	w, updated := w.calculateState(wo, pos)
	if updated {
		wo.SetBlock(pos, w, nil)
	}
}

// UseOnBlock ...
func (w Wall) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, wo *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(wo, pos, face, w)
	if !used {
		return
	}
	w, _ = w.calculateState(wo, pos)
	place(wo, pos, w, user, ctx)
	return placed(ctx)
}

// CanDisplace ...
func (Wall) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (Wall) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// ConnectionType returns the connection type of the wall in the given direction.
func (w Wall) ConnectionType(direction cube.Direction) WallConnectionType {
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
func (w Wall) WithConnectionType(direction cube.Direction, connection WallConnectionType) Wall {
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
func (w Wall) calculateState(wo *world.World, pos cube.Pos) (Wall, bool) {
	var updated bool
	abovePos := pos.Add(cube.Pos{0, 1, 0})
	above := wo.Block(abovePos)
	for _, face := range cube.HorizontalFaces() {
		sidePos := pos.Side(face)
		side := wo.Block(sidePos)
		var connectionType WallConnectionType
		if side.Model().FaceSolid(sidePos, face.Opposite(), wo) {
			// If the wall is connected to the side, it has the possibility of having a tall connection. This is
			//calculated by checking for any overlapping blocks in the area of the connection.
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
	case Torch:
		post = above.Facing == cube.FaceDown
	case Wall:
		post = above.Post
	default:
		post = true
	}
	if !post {
		// If a wall has two connections that are in different axis then it becomes a post regardless of the above block.
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

// allWalls returns a list of all wall types.
func allWalls() (walls []world.Block) {
	for _, block := range WallBlocks() {
		for _, north := range WallConnectionTypes() {
			for _, east := range WallConnectionTypes() {
				for _, south := range WallConnectionTypes() {
					for _, west := range WallConnectionTypes() {
						walls = append(walls, Wall{
							Block:           block,
							NorthConnection: north,
							EastConnection:  east,
							SouthConnection: south,
							WestConnection:  west,
							Post:            false,
						})
						walls = append(walls, Wall{
							Block:           block,
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
