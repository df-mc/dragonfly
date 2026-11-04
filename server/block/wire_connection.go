package block

// WireConnection represents the way redstone wire connects to one of its
// horizontal sides. It is calculated by the server whenever the wire or the
// blocks around it change.
type WireConnection struct {
	wireConnection
}

// NoWireConnection returns the connection of redstone wire to a side that it
// does not connect to.
func NoWireConnection() WireConnection {
	return WireConnection{0}
}

// SideWireConnection returns the connection of redstone wire to a side that
// it runs to along the ground.
func SideWireConnection() WireConnection {
	return WireConnection{1}
}

// UpWireConnection returns the connection of redstone wire to a side that it
// runs up, climbing the block on that side to reach wire on top of it.
func UpWireConnection() WireConnection {
	return WireConnection{2}
}

type wireConnection uint8

// Uint8 returns the wire connection as a uint8.
func (w wireConnection) Uint8() uint8 {
	return uint8(w)
}

// String ...
func (w wireConnection) String() string {
	switch w {
	case 0:
		return "none"
	case 1:
		return "side"
	case 2:
		return "up"
	}
	panic("should never happen")
}

// WireConnections holds the way redstone wire connects to each of its
// horizontal sides.
type WireConnections struct {
	north, east, south, west WireConnection
}

// Uint8 returns the WireConnections as a uint8.
func (c WireConnections) Uint8() uint8 {
	return c.north.Uint8() + c.east.Uint8()*3 + c.south.Uint8()*9 + c.west.Uint8()*27
}

// properties returns the block state properties of the WireConnections.
func (c WireConnections) properties() map[string]any {
	return map[string]any{
		"redstone_north": c.north.String(),
		"redstone_east":  c.east.String(),
		"redstone_south": c.south.String(),
		"redstone_west":  c.west.String(),
	}
}

// allWireConnections returns a list of all combinations of WireConnections.
func allWireConnections() (all []WireConnections) {
	for i := range 81 {
		all = append(all, WireConnections{
			north: WireConnection{wireConnection(i % 3)},
			east:  WireConnection{wireConnection(i / 3 % 3)},
			south: WireConnection{wireConnection(i / 9 % 3)},
			west:  WireConnection{wireConnection(i / 27 % 3)},
		})
	}
	return
}
