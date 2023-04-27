package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/google/uuid"
	"io"
)

// Provider represents a value that may provide world data to a World value. It usually does the reading and
// writing of the world data so that the World may use it.
type Provider interface {
	io.Closer
	// Settings loads the settings for a World and returns them.
	Settings() *Settings
	// SaveSettings saves the settings of a World.
	SaveSettings(*Settings)

	// LoadPlayerSpawnPosition loads the player spawn point if found, otherwise an error will be returned.
	LoadPlayerSpawnPosition(uuid uuid.UUID) (pos cube.Pos, exists bool, err error)
	// SavePlayerSpawnPosition saves the player spawn point. In vanilla, this can be done with beds in the overworld
	// or respawn anchors in the nether.
	SavePlayerSpawnPosition(uuid uuid.UUID, pos cube.Pos) error

	LoadColumn(pos ChunkPos, dim Dimension) (*Column, error)
	// SaveChunk saves a chunk at a specific position in the provider. If writing was not successful, an error
	// is returned.
	SaveChunk(position ChunkPos, c *chunk.Chunk, dim Dimension) error
	// SaveEntities saves a list of entities in a chunk position. If writing is not successful, an error is
	// returned.
	SaveEntities(position ChunkPos, entities []Entity, dim Dimension) error
	// SaveBlockNBT saves block NBT, or block entities, to a specific chunk position. If the NBT cannot be
	// stored, SaveBlockNBT returns a non-nil error.
	SaveBlockNBT(position ChunkPos, data []map[string]any, dim Dimension) error
}

// Compile time check to make sure NopProvider implements Provider.
var _ Provider = (*NopProvider)(nil)

// NopProvider implements a Provider that does not perform any disk I/O. It generates values on the run and
// dynamically, instead of reading and writing data, and otherwise returns empty values. A Settings struct can be passed
// to initialize a world with specific settings. Since Settings is a pointer, using the same NopProvider for multiple
// worlds means those worlds will share the same settings.
type NopProvider struct {
	Set *Settings
}

func (n NopProvider) Settings() *Settings {
	if n.Set == nil {
		return defaultSettings()
	}
	return n.Set
}
func (NopProvider) SaveSettings(*Settings)                                   {}
func (NopProvider) LoadColumn(ChunkPos, Dimension) (*Column, error)          { return nil, leveldb.ErrNotFound }
func (NopProvider) SaveEntities(ChunkPos, []Entity, Dimension) error         { return nil }
func (NopProvider) SaveBlockNBT(ChunkPos, []map[string]any, Dimension) error { return nil }
func (NopProvider) SaveChunk(ChunkPos, *chunk.Chunk, Dimension) error        { return nil }
func (NopProvider) LoadPlayerSpawnPosition(uuid.UUID) (cube.Pos, bool, error) {
	return cube.Pos{}, false, nil
}
func (NopProvider) SavePlayerSpawnPosition(uuid.UUID, cube.Pos) error { return nil }
func (NopProvider) Close() error                                      { return nil }
