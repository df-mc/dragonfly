package mcdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/dragonfly/server/world/mcdb/leveldat"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"golang.org/x/exp/maps"
	"os"
	"path/filepath"
	"time"
)

// DB implements a world provider for the Minecraft world format, which
// is based on a leveldb database.
type DB struct {
	conf Config
	ldb  *leveldb.DB
	dir  string
	ldat *leveldat.Data
	set  *world.Settings
}

// Open creates a new provider reading and writing from/to files under the path
// passed using default options. If a world is present at the path, Open will
// parse its data and initialise the world with it. If the data cannot be
// parsed, an error is returned.
func Open(dir string) (*DB, error) {
	var conf Config
	return conf.Open(dir)
}

// Settings returns the world.Settings of the world loaded by the DB.
func (db *DB) Settings() *world.Settings {
	return db.set
}

// SaveSettings saves the world.Settings passed to the level.dat.
func (db *DB) SaveSettings(s *world.Settings) {
	db.ldat.PutSettings(s)
}

// playerData holds the fields that indicate where player data is stored for a player with a specific UUID.
type playerData struct {
	UUID         string `nbt:"MsaId"`
	ServerID     string `nbt:"ServerId"`
	SelfSignedID string `nbt:"SelfSignedId"`
}

// LoadPlayerSpawnPosition loads the players spawn position stored in the level.dat from their UUID.
func (db *DB) LoadPlayerSpawnPosition(id uuid.UUID) (pos cube.Pos, exists bool, err error) {
	serverData, _, exists, err := db.loadPlayerData(id)
	if !exists || err != nil {
		return cube.Pos{}, exists, err
	}
	x, y, z := serverData["SpawnX"], serverData["SpawnY"], serverData["SpawnZ"]
	if x == nil || y == nil || z == nil {
		return cube.Pos{}, true, fmt.Errorf("error reading spawn fields from server data for player %v", id)
	}
	return cube.Pos{int(x.(int32)), int(y.(int32)), int(z.(int32))}, true, nil
}

// loadPlayerData loads the data stored in a LevelDB database for a specific UUID.
func (db *DB) loadPlayerData(id uuid.UUID) (serverData map[string]interface{}, key string, exists bool, err error) {
	data, err := db.ldb.Get([]byte("player_"+id.String()), nil)
	if err == leveldb.ErrNotFound {
		return nil, "", false, nil
	} else if err != nil {
		return nil, "", true, fmt.Errorf("error reading player data for uuid %v: %w", id, err)
	}

	var d playerData
	if err := nbt.UnmarshalEncoding(data, &d, nbt.LittleEndian); err != nil {
		return nil, "", true, fmt.Errorf("error decoding player data for uuid %v: %w", id, err)
	}
	if d.UUID != id.String() || d.ServerID == "" {
		return nil, d.ServerID, true, fmt.Errorf("invalid player data for uuid %v: %v", id, d)
	}
	serverDB, err := db.ldb.Get([]byte(d.ServerID), nil)
	if err != nil {
		return nil, d.ServerID, true, fmt.Errorf("error reading server data for player %v (%v): %w", id, d.ServerID, err)
	}

	if err := nbt.UnmarshalEncoding(serverDB, &serverData, nbt.LittleEndian); err != nil {
		return nil, d.ServerID, true, fmt.Errorf("error decoding server data for player %v", id)
	}
	return serverData, d.ServerID, true, nil
}

// SavePlayerSpawnPosition saves the player spawn position passed to the levelDB database.
func (db *DB) SavePlayerSpawnPosition(id uuid.UUID, pos cube.Pos) error {
	_, err := db.ldb.Get([]byte("player_"+id.String()), nil)
	d := make(map[string]interface{})
	k := "player_server_" + id.String()

	if errors.Is(err, leveldb.ErrNotFound) {
		data, err := nbt.MarshalEncoding(playerData{
			UUID:     id.String(),
			ServerID: k,
		}, nbt.LittleEndian)
		if err != nil {
			panic(err)
		}
		if err := db.ldb.Put([]byte("player_"+id.String()), data, nil); err != nil {
			return fmt.Errorf("error writing player data for id %v: %w", id, err)
		}
	} else {
		if d, k, _, err = db.loadPlayerData(id); err != nil {
			return err
		}
	}
	d["SpawnX"] = int32(pos.X())
	d["SpawnY"] = int32(pos.Y())
	d["SpawnZ"] = int32(pos.Z())

	data, err := nbt.MarshalEncoding(d, nbt.LittleEndian)
	if err != nil {
		panic(err)
	}
	if err = db.ldb.Put([]byte(k), data, nil); err != nil {
		return fmt.Errorf("error writing server data for player %v: %w", id, err)
	}
	return nil
}

// LoadColumn reads a world.Column from the DB at a position and dimension in
// the DB. If no column at that position exists, errors.Is(err,
// leveldb.ErrNotFound) equals true.
func (db *DB) LoadColumn(pos world.ChunkPos, dim world.Dimension) (*world.Column, error) {
	k := dbKey{pos: pos, dim: dim}
	col, err := db.column(k)
	if err != nil {
		return nil, fmt.Errorf("load column %v (%v): %w", pos, dim, err)
	}
	return col, nil
}

const chunkVersion = 40

func (db *DB) column(k dbKey) (*world.Column, error) {
	var cdata chunk.SerialisedData
	col := new(world.Column)

	ver, err := db.version(k)
	if err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}
	if ver != chunkVersion {
		db.conf.Log.Debugf("column %v (%v): unsupported chunk version %v, trying to load anyway", k.pos, k.dim, ver)
	}
	cdata.Biomes, err = db.biomes(k)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		// Some chunks still use 2D chunk data and might not have this field, in
		// which case we can just move on.
		return nil, fmt.Errorf("read biomes: %w", err)
	}
	cdata.SubChunks, err = db.subChunks(k)
	if err != nil {
		return nil, fmt.Errorf("read sub chunks: %w", err)
	}
	col.Chunk, err = chunk.DiskDecode(cdata, k.dim.Range())
	if err != nil {
		return nil, fmt.Errorf("decode chunk data: %w", err)
	}
	col.Entities, err = db.entities(k)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		// Not all chunks need to have entities, so an ErrNotFound is fine here.
		return nil, fmt.Errorf("read entities: %w", err)
	}
	col.BlockEntities, err = db.blockEntities(k, col.Chunk)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		// Same as with entities, an ErrNotFound is fine here.
		return nil, fmt.Errorf("read block entities: %w", err)
	}
	return col, nil
}

func (db *DB) version(k dbKey) (byte, error) {
	p, err := db.ldb.Get(k.Sum(keyVersion), nil)
	switch err {
	default:
		return 0, err
	case leveldb.ErrNotFound:
		// Although the version at `keyVersion` may not be found, there is
		// another `keyVersionOld` where the version may be found.
		if p, err = db.ldb.Get(k.Sum(keyVersionOld), nil); err != nil {
			return 0, err
		}
		fallthrough
	case nil:
		if n := len(p); n != 1 {
			return 0, fmt.Errorf("expected 1 version byte, found %v", n)
		}
		return p[0], nil
	}
}

func (db *DB) biomes(k dbKey) ([]byte, error) {
	biomes, err := db.ldb.Get(k.Sum(key3DData), nil)
	if err != nil {
		return nil, err
	}
	// The first 512 bytes is a heightmap (16*16 int16s), the biomes follow. We
	// calculate a heightmap on startup so the heightmap is discarded.
	if n := len(biomes); n <= 512 {
		return nil, fmt.Errorf("expected at least 513 bytes for 3D data, got %v", n)
	}
	return biomes[512:], nil
}

func (db *DB) subChunks(k dbKey) ([][]byte, error) {
	r := k.dim.Range()
	sub := make([][]byte, (r.Height()>>4)+1)

	var err error
	for i := range sub {
		y := uint8(i + (r[0] >> 4))
		sub[i], err = db.ldb.Get(k.Sum(keySubChunkData, y), nil)
		if err == leveldb.ErrNotFound {
			// No sub chunk present at this Y level. We skip this one and move
			// to the next, which might still be present.
			continue
		} else if err != nil {
			return nil, fmt.Errorf("sub chunk %v: %w", int8(i), err)
		}
	}
	return sub, nil
}

func (db *DB) entities(k dbKey) ([]world.Entity, error) {
	data, err := db.ldb.Get(k.Sum(keyEntities), nil)
	if err != nil {
		return nil, err
	}
	var entities []world.Entity

	buf := bytes.NewBuffer(data)
	dec := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian)

	var m map[string]any
	for buf.Len() != 0 {
		maps.Clear(m)
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("decode nbt: %w", err)
		}
		id, ok := m["identifier"]
		if !ok {
			db.conf.Log.Errorf("missing identifier field in %v", m)
			continue
		}
		name, _ := id.(string)
		t, ok := db.conf.Entities.Lookup(name)
		if !ok {
			db.conf.Log.Errorf("entity %v was not registered (%v)", name, m)
			continue
		}
		if s, ok := t.(world.SaveableEntityType); ok {
			if v := s.DecodeNBT(m); v != nil {
				entities = append(entities, v)
			}
		}
	}
	return entities, nil
}

func (db *DB) blockEntities(k dbKey, c *chunk.Chunk) (map[cube.Pos]world.Block, error) {
	blockEntities := make(map[cube.Pos]world.Block)

	data, err := db.ldb.Get(k.Sum(keyBlockEntities), nil)
	if err != nil {
		return blockEntities, err
	}

	buf := bytes.NewBuffer(data)
	dec := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian)

	var m map[string]any
	for buf.Len() != 0 {
		maps.Clear(m)
		if err := dec.Decode(&m); err != nil {
			return blockEntities, fmt.Errorf("decode nbt: %w", err)
		}
		pos := blockPosFromNBT(m)

		id := c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0)
		b, ok := world.BlockByRuntimeID(id)
		if !ok {
			db.conf.Log.Errorf("no block registered with runtime id %v", id)
			continue
		}
		nbter, ok := b.(world.NBTer)
		if !ok {
			db.conf.Log.Errorf("block %#v has nbt but does not implement world.nbter", b)
			continue
		}
		blockEntities[pos] = nbter.DecodeNBT(m).(world.Block)
	}
	return blockEntities, nil
}

// StoreColumn stores a world.Column at a position and dimension in the DB. An
// error is returned if storing was unsuccessful.
func (db *DB) StoreColumn(pos world.ChunkPos, dim world.Dimension, col *world.Column) error {
	k := dbKey{pos: pos, dim: dim}
	if err := db.storeColumn(k, col); err != nil {
		return fmt.Errorf("store column %v (%v): %w", pos, dim, err)
	}
	return nil
}

func (db *DB) storeColumn(k dbKey, col *world.Column) error {
	data := chunk.Encode(col.Chunk, chunk.DiskEncoding)
	n := 5 + len(data.SubChunks)
	batch := leveldb.MakeBatch(n)

	db.storeVersion(batch, k, chunkVersion)
	db.storeBiomes(batch, k, data.Biomes)
	db.storeSubChunks(batch, k, data.SubChunks, col.Chunk.Range())
	db.storeFinalisation(batch, k, finalisationPopulated)
	db.storeEntities(batch, k, col.Entities)
	db.storeBlockEntities(batch, k, col.BlockEntities)

	return db.ldb.Write(batch, nil)
}

func (db *DB) storeVersion(batch *leveldb.Batch, k dbKey, ver uint8) {
	batch.Put(k.Sum(keyVersion), []byte{ver})
}

var emptyHeightmap = make([]byte, 512)

func (db *DB) storeBiomes(batch *leveldb.Batch, k dbKey, biomes []byte) {
	batch.Put(k.Sum(key3DData), append(emptyHeightmap, biomes...))
}

func (db *DB) storeSubChunks(batch *leveldb.Batch, k dbKey, subChunks [][]byte, r cube.Range) {
	for i, sub := range subChunks {
		batch.Put(k.Sum(keySubChunkData, byte(i+(r[0]>>4))), sub)
	}
}

func (db *DB) storeFinalisation(batch *leveldb.Batch, k dbKey, finalisation uint32) {
	p := make([]byte, 4)
	binary.LittleEndian.PutUint32(p, finalisation)
	batch.Put(k.Sum(keyFinalisation), p)
}

func (db *DB) storeEntities(batch *leveldb.Batch, k dbKey, entities []world.Entity) {
	if len(entities) == 0 {
		batch.Delete(k.Sum(keyEntities))
		return
	}

	buf := bytes.NewBuffer(nil)
	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	for _, e := range entities {
		t, ok := e.Type().(world.SaveableEntityType)
		if !ok {
			continue
		}
		x := t.EncodeNBT(e)
		x["identifier"] = t.EncodeEntity()
		if err := enc.Encode(x); err != nil {
			db.conf.Log.Errorf("store entities: error encoding NBT: %w", err)
		}
	}
	batch.Put(k.Sum(keyEntities), buf.Bytes())
}

func (db *DB) storeBlockEntities(batch *leveldb.Batch, k dbKey, blockEntities map[cube.Pos]world.Block) {
	if len(blockEntities) == 0 {
		batch.Delete(k.Sum(keyBlockEntities))
		return
	}

	buf := bytes.NewBuffer(nil)
	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	for pos, b := range blockEntities {
		n, ok := b.(world.NBTer)
		if !ok {
			continue
		}
		data := n.EncodeNBT()
		data["x"], data["y"], data["z"] = int32(pos[0]), int32(pos[1]), int32(pos[2])
		if err := enc.Encode(data); err != nil {
			db.conf.Log.Errorf("store block entities: error encoding NBT: %w", err)
		}
	}
	batch.Put(k.Sum(keyBlockEntities), buf.Bytes())
}

// NewColumnIterator returns a ColumnIterator that may be used to iterate over all
// position/chunk pairs in a database.
// An IteratorRange r may be passed to specify limits in terms of what chunks
// should be read. r may be set to nil to read all chunks from the DB.
func (db *DB) NewColumnIterator(r *IteratorRange) *ColumnIterator {
	if r == nil {
		r = &IteratorRange{}
	}
	return newColumnIterator(db, r)
}

// Close closes the provider, saving any file that might need to be saved, such as the level.dat.
func (db *DB) Close() error {
	db.ldat.LastPlayed = time.Now().Unix()

	var ldat leveldat.LevelDat
	if err := ldat.Marshal(*db.ldat); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	if err := ldat.WriteFile(filepath.Join(db.dir, "level.dat")); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	if err := os.WriteFile(filepath.Join(db.dir, "levelname.txt"), []byte(db.ldat.LevelName), 0644); err != nil {
		return fmt.Errorf("close: write levelname.txt: %w", err)
	}
	return db.ldb.Close()
}

// dbKey holds a position and dimension.
type dbKey struct {
	pos world.ChunkPos
	dim world.Dimension
}

// Sum converts k to its []byte representation and appends p.
func (k dbKey) Sum(p ...byte) []byte {
	return append(index(k.pos, k.dim), p...)
}

// index returns a byte buffer holding the written index of the chunk position passed. If the dimension passed
// is not world.Overworld, the length of the index returned is 12. It is 8 otherwise.
func index(position world.ChunkPos, d world.Dimension) []byte {
	dim, _ := world.DimensionID(d)
	x, z := uint32(position[0]), uint32(position[1])
	b := make([]byte, 12)

	binary.LittleEndian.PutUint32(b, x)
	binary.LittleEndian.PutUint32(b[4:], z)
	if dim == 0 {
		return b[:8]
	}
	binary.LittleEndian.PutUint32(b[8:], uint32(dim))
	return b
}

// blockPosFromNBT returns a position from the X, Y and Z components stored in the NBT data map passed. The
// map is assumed to have an 'x', 'y' and 'z' key.
func blockPosFromNBT(data map[string]any) cube.Pos {
	x, _ := data["x"].(int32)
	y, _ := data["y"].(int32)
	z, _ := data["z"].(int32)
	return cube.Pos{int(x), int(y), int(z)}
}
