package world_internal

// LiquidRemovable is an array indexed by the runtime IDs of blocks. If present in this array, a block with that
// runtime ID is removable by liquid and either will or will not drop items, depending on the bool in the map,
// when flown into by a liquid.
var LiquidRemovable []bool
