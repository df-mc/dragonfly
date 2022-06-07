package item

// EchoShard is an item found in ancient cities which can be used to craft recovery compasses.
type EchoShard struct{}

// EncodeItem ...
func (EchoShard) EncodeItem() (name string, meta int16) {
	return "minecraft:echo_shard", 0
}
