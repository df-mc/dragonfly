// Package block exports implementations of the Block interface found in the server/world package. The blocks
// implemented in this package are automatically registered in the server/world package using the world.RegisterBlock
// and world.RegisterItem functions through an init function, so that methods in server/world that return blocks and
// items are able to return them by the respective types implemented here.
package block
