package item

// Record represents an item that can play music when placed in a jukebox.
type Record interface {
	// RecordInfo returns the record information of the item.
	RecordInfo() RecordInfo
}

// RecordInfo is a struct returned by items that implement Record. It contains the music playback
// configuration of the item.
type RecordInfo struct {
	// ComparatorSignal is the signal strength for comparator blocks, from 1 to 13.
	ComparatorSignal int
	// Duration is the duration of the sound event in seconds.
	Duration float64
	// SoundEvent is the sound event played by the record.
	SoundEvent string
}
