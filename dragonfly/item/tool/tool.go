package tool

// Tool represents an item that may be used as a tool.
type Tool interface {
	// ToolType returns the type of the tool. The blocks that can be mined with this tool depend on this
	// tool type.
	ToolType() Type
	// HarvestLevel returns the level that this tool is able to harvest. If a block has a harvest level above
	// this one, this tool won't be able to harvest it.
	HarvestLevel() int
	// BaseMiningEfficiency is the base efficiency of the tool, when it comes to mining blocks. This decides
	// the speed with which blocks can be mined.
	BaseMiningEfficiency() float64
}
