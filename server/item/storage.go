package item

// Storage represents an item that can store other items, such as a bundle or shulker box.
type Storage interface {
	// StorageInfo returns the information of the item related to storing other items inside it.
	StorageInfo() StorageInfo
}

// StorageInfo is a struct returned by items that implement Storage. It contains the information required for
// the client to allow storing other items inside the item.
type StorageInfo struct {
	// MaxSlots is the maximum number of slots the item may store items in.
	MaxSlots int
	// MaxWeightLimit is the maximum weight of items the item may store.
	MaxWeightLimit int
	// WeightInStorageItem is the weight of the item itself when stored inside another storage item.
	WeightInStorageItem int
	// NumViewableSlots is the number of slots that may be viewed when interacting with the item. If set to
	// zero, no bundle interaction component is sent.
	NumViewableSlots int
	// AllowNestedStorageItems is true if other storage items may be stored inside the item.
	AllowNestedStorageItems bool
	// AllowedItems is a list of item identifiers that may be stored inside the item.
	AllowedItems []string
	// BannedItems is a list of item identifiers that may not be stored inside the item.
	BannedItems []string
}
