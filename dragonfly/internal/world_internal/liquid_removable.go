package world_internal

// LiquidRemovable is a map indexed by the runtime IDs of blocks. If present in this map, a block with that
// runtime ID is removable by liquid and either will or will not drop items, depending on the bool in the map,
// when flown into by a liquid.
var LiquidRemovable = map[uint32]bool{}
