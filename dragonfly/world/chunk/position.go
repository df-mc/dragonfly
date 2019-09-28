package chunk

// Position holds the position of a chunk. The type is provided as a utility struct for keeping track of a
// chunk's position. Chunks do not themselves keep track of that.
type Position struct {
	X, Z int32
}
