package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Wall struct {
	Type            WallType
	NorthConnection WallConnectionType
	EastConnection  WallConnectionType
	SouthConnection WallConnectionType
	WestConnection  WallConnectionType
	Post            bool
}

// EncodeItem ...
func (w Wall) EncodeItem() (name string, meta int16) {
	if w.Type.IsCobblestoneWall() {
		return "minecraft:cobblestone_wall", int16(w.Type.Uint8())
	}
	return "minecraft:" + w.Type.String() + "_wall", 0
}

// EncodeBlock ...
func (w Wall) EncodeBlock() (string, map[string]any) {
	name := "minecraft:" + w.Type.String() + "_wall"
	properties := map[string]any{
		"wall_connection_type_north": w.NorthConnection.String(),
		"wall_connection_type_east":  w.EastConnection.String(),
		"wall_connection_type_south": w.SouthConnection.String(),
		"wall_connection_type_west":  w.WestConnection.String(),
		"wall_post_bit":              boolByte(w.Post),
	}
	if w.Type.IsCobblestoneWall() {
		name = "minecraft:cobblestone_wall"
		properties["wall_block_type"] = w.Type.String()
	}
	return name, properties
}

// Model ...
func (w Wall) Model() world.BlockModel {
	//TODO implement me
	return model.Solid{}
}

// BreakInfo ...
func (w Wall) BreakInfo() BreakInfo {
	hardness := 1.5
	switch w.Type {
	case SandstoneWall(), RedSandstoneWall():
		hardness = 0.8
	case BrickWall(), MossyCobblestoneWall(), RedNetherBrickWall(), PolishedBlackstoneWall():
		hardness = 2.0
	case EndBrickWall():
		hardness = 3.0
	case PolishedDeepslateWall(), DeepslateBrickWall(), DeepslateTileWall(), CobbledDeepslateWall():
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
func (w Wall) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, wo *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(wo, pos, face, w)
	if !used {
		return
	}
	w, _ = w.calculateState(wo, pos)
	place(wo, pos, w, user, ctx)
	return placed(ctx)
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
			var tall bool
			switch above := above.(type) {
			case Wall:
				tall = above.ConnectionType(face.Direction()) != NoWallConnection()
			default:
				tall = above.Model().FaceSolid(abovePos, cube.FaceDown, wo)
			}
			connectionType = ShortWallConnection()
			if tall {
				connectionType = TallWallConnection()
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
	for _, w := range WallTypes() {
		for _, north := range WallConnectionTypes() {
			for _, east := range WallConnectionTypes() {
				for _, south := range WallConnectionTypes() {
					for _, west := range WallConnectionTypes() {
						walls = append(walls, Wall{Type: w,
							NorthConnection: north,
							EastConnection:  east,
							SouthConnection: south,
							WestConnection:  west,
							Post:            false,
						})
						walls = append(walls, Wall{Type: w,
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
