package item

// RepairMaterials represents a durable item that can be repaired by other items and provides the data
// required for the client to display the repair materials of the item.
type RepairMaterials interface {
	// RepairMaterials returns the repair materials of the item.
	RepairMaterials() []RepairItem
}

// RepairItem is a struct returned by items that implement RepairMaterials. It contains the information
// required for the client to display a single repair material of the item.
type RepairItem struct {
	// Item is the identifier of the item used to repair the item.
	Item string
	// Tag is the tag used to match items that can repair the item.
	Tag string
	// RepairAmount is the expression used to determine the amount of durability restored when repairing the
	// item.
	RepairAmount string
}
