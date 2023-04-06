package mcdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
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
	ldat data
	set  *world.Settings
}

// New creates a new provider reading and writing from/to files under the path
// passed using default options. If a world is present at the path, New will
// parse its data and initialise the world with it. If the data cannot be
// parsed, an error is returned.
func New(dir string) (*DB, error) {
	var conf Config
	return conf.New(dir)
}

// initDefaultLevelDat initialises a default level.dat file.
func (db *DB) initDefaultLevelDat() {
	db.ldat.Abilities.AttackMobs = true
	db.ldat.Abilities.AttackPlayers = true
	db.ldat.Abilities.Build = true
	db.ldat.Abilities.DoorsAndSwitches = true
	db.ldat.Abilities.FlySpeed = 0.05
	db.ldat.Abilities.Mine = true
	db.ldat.Abilities.OpenContainers = true
	db.ldat.Abilities.PlayerPermissionsLevel = 1
	db.ldat.Abilities.WalkSpeed = 0.1
	db.ldat.BaseGameVersion = "*"
	db.ldat.CommandBlockOutput = true
	db.ldat.CommandBlocksEnabled = true
	db.ldat.CommandsEnabled = true
	db.ldat.Difficulty = 2
	db.ldat.DoDayLightCycle = true
	db.ldat.DoEntityDrops = true
	db.ldat.DoFireTick = true
	db.ldat.DoInsomnia = true
	db.ldat.DoMobLoot = true
	db.ldat.DoMobSpawning = true
	db.ldat.DoTileDrops = true
	db.ldat.DoWeatherCycle = true
	db.ldat.DrowningDamage = true
	db.ldat.FallDamage = true
	db.ldat.FireDamage = true
	db.ldat.FreezeDamage = true
	db.ldat.FunctionCommandLimit = 10000
	db.ldat.GameType = 1
	db.ldat.Generator = 2
	db.ldat.HasBeenLoadedInCreative = true
	db.ldat.InventoryVersion = protocol.CurrentVersion
	db.ldat.LANBroadcast = true
	db.ldat.LANBroadcastIntent = true
	db.ldat.LastOpenedWithVersion = minimumCompatibleClientVersion
	db.ldat.LevelName = "World"
	db.ldat.LightningLevel = 1.0
	db.ldat.LimitedWorldDepth = 16
	db.ldat.LimitedWorldOriginY = math.MaxInt16
	db.ldat.LimitedWorldWidth = 16
	db.ldat.MaxCommandChainLength = math.MaxUint16
	db.ldat.MinimumCompatibleClientVersion = minimumCompatibleClientVersion
	db.ldat.MobGriefing = true
	db.ldat.MultiPlayerGame = true
	db.ldat.MultiPlayerGameIntent = true
	db.ldat.NaturalRegeneration = true
	db.ldat.NetherScale = 8
	db.ldat.NetworkVersion = protocol.CurrentProtocol
	db.ldat.PVP = true
	db.ldat.Platform = 2
	db.ldat.PlatformBroadcastIntent = 3
	db.ldat.RainLevel = 1.0
	db.ldat.RandomSeed = time.Now().Unix()
	db.ldat.RandomTickSpeed = 1
	db.ldat.RespawnBlocksExplode = true
	db.ldat.SendCommandFeedback = true
	db.ldat.ServerChunkTickRange = 6
	db.ldat.ShowBorderEffect = true
	db.ldat.ShowDeathMessages = true
	db.ldat.ShowTags = true
	db.ldat.SpawnMobs = true
	db.ldat.SpawnRadius = 5
	db.ldat.SpawnRadius = 5
	db.ldat.SpawnY = math.MaxInt16
	db.ldat.StorageVersion = 9
	db.ldat.TNTExplodes = true
	db.ldat.WorldVersion = 1
	db.ldat.XBLBroadcastIntent = 3
}

// Settings returns the world.Settings of the world loaded by the DB.
func (db *DB) Settings() *world.Settings {
	return db.set
}

// loadSettings loads the settings in the level.dat into a world.Settings struct and stores it, so that it can be
// returned through a call to Settings.
func (db *DB) loadSettings() {
	db.ldat.WorldStartCount += 1
	difficulty, _ := world.DifficultyByID(int(db.ldat.Difficulty))
	mode, _ := world.GameModeByID(int(db.ldat.GameType))
	db.set = &world.Settings{
		Name:            db.ldat.LevelName,
		Spawn:           cube.Pos{int(db.ldat.SpawnX), int(db.ldat.SpawnY), int(db.ldat.SpawnZ)},
		Time:            db.ldat.Time,
		TimeCycle:       db.ldat.DoDayLightCycle,
		RainTime:        int64(db.ldat.RainTime),
		Raining:         db.ldat.RainLevel > 0,
		ThunderTime:     int64(db.ldat.LightningTime),
		Thundering:      db.ldat.LightningLevel > 0,
		WeatherCycle:    db.ldat.DoWeatherCycle,
		CurrentTick:     db.ldat.CurrentTick,
		DefaultGameMode: mode,
		Difficulty:      difficulty,
		TickRange:       db.ldat.ServerChunkTickRange,
	}
}

// SaveSettings saves the world.Settings passed to the level.dat.
func (db *DB) SaveSettings(s *world.Settings) {
	db.ldat.LevelName = s.Name
	db.ldat.SpawnX, db.ldat.SpawnY, db.ldat.SpawnZ = int32(s.Spawn.X()), int32(s.Spawn.Y()), int32(s.Spawn.Z())
	db.ldat.LimitedWorldOriginX, db.ldat.LimitedWorldOriginY, db.ldat.LimitedWorldOriginZ = db.ldat.SpawnX, db.ldat.SpawnY, db.ldat.SpawnZ
	db.ldat.Time = s.Time
	db.ldat.DoDayLightCycle = s.TimeCycle
	db.ldat.DoWeatherCycle = s.WeatherCycle
	db.ldat.RainTime, db.ldat.RainLevel = int32(s.RainTime), 0
	db.ldat.LightningTime, db.ldat.LightningLevel = int32(s.ThunderTime), 0
	if s.Raining {
		db.ldat.RainLevel = 1
	}
	if s.Thundering {
		db.ldat.LightningLevel = 1
	}
	db.ldat.CurrentTick = s.CurrentTick
	db.ldat.ServerChunkTickRange = s.TickRange
	mode, _ := world.GameModeID(s.DefaultGameMode)
	db.ldat.GameType = int32(mode)
	difficulty, _ := world.DifficultyID(s.Difficulty)
	db.ldat.Difficulty = int32(difficulty)
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

// LoadChunk loads a chunk at the position passed from the leveldb database. If it doesn't exist, exists is
// false. If an error is returned, exists is always assumed to be true.
func (db *DB) LoadChunk(position world.ChunkPos, dim world.Dimension) (c *chunk.Chunk, exists bool, err error) {
	data := chunk.SerialisedData{}
	key := db.index(position, dim)

	// This key is where the version of a chunk resides. The chunk version has changed many times, without any
	// actual substantial changes, so we don't check this.
	_, err = db.ldb.Get(append(key, keyVersion), nil)
	if err == leveldb.ErrNotFound {
		// The new key was not found, so we try the old key.
		if _, err = db.ldb.Get(append(key, keyVersionOld), nil); err != nil {
			return nil, false, nil
		}
	} else if err != nil {
		return nil, true, fmt.Errorf("error reading version: %w", err)
	}

	data.Biomes, err = db.ldb.Get(append(key, key3DData), nil)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, true, fmt.Errorf("error reading 3D data: %w", err)
	}
	if len(data.Biomes) > 512 {
		// Strip the heightmap from the biomes.
		data.Biomes = data.Biomes[512:]
	}

	data.BlockNBT, err = db.ldb.Get(append(key, keyBlockEntities), nil)
	// Block entities aren't present when there aren't any, so it's okay if we can't find the key.
	if err != nil && err != leveldb.ErrNotFound {
		return nil, true, fmt.Errorf("error reading block entities: %w", err)
	}
	data.SubChunks = make([][]byte, (dim.Range().Height()>>4)+1)
	for i := range data.SubChunks {
		data.SubChunks[i], err = db.ldb.Get(append(key, keySubChunkData, uint8(i+(dim.Range()[0]>>4))), nil)
		if err == leveldb.ErrNotFound {
			// No sub chunk present at this Y level. We skip this one and move to the next, which might still
			// be present.
			continue
		} else if err != nil {
			return nil, true, fmt.Errorf("error reading sub chunk data %v: %w", i, err)
		}
	}
	c, err = chunk.DiskDecode(data, dim.Range())
	return c, true, err
}

// SaveChunk saves a chunk at the position passed to the leveldb database. Its version is written as the
// version in the chunkVersion constant.
func (db *DB) SaveChunk(position world.ChunkPos, c *chunk.Chunk, dim world.Dimension) error {
	data := chunk.Encode(c, chunk.DiskEncoding)

	key := db.index(position, dim)
	_ = db.ldb.Put(append(key, keyVersion), []byte{chunkVersion}, nil)
	// Write the heightmap by just writing 512 empty bytes.
	_ = db.ldb.Put(append(key, key3DData), append(make([]byte, 512), data.Biomes...), nil)

	finalisation := make([]byte, 4)
	binary.LittleEndian.PutUint32(finalisation, 2)
	_ = db.ldb.Put(append(key, keyFinalisation), finalisation, nil)

	for i, sub := range data.SubChunks {
		_ = db.ldb.Put(append(key, keySubChunkData, byte(i+(c.Range()[0]>>4))), sub, nil)
	}
	return nil
}

// LoadEntities loads all entities from the chunk position passed.
func (db *DB) LoadEntities(pos world.ChunkPos, dim world.Dimension, reg world.EntityRegistry) ([]world.Entity, error) {
	data, err := db.ldb.Get(append(db.index(pos, dim), keyEntities), nil)
	if err != leveldb.ErrNotFound && err != nil {
		return nil, err
	}
	var a []world.Entity

	buf := bytes.NewBuffer(data)
	dec := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian)

	for buf.Len() != 0 {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("error decoding block NBT: %w", err)
		}
		id, ok := m["identifier"]
		if !ok {
			db.conf.Log.Errorf("load entities: failed loading %v: entity had data but no identifier (%v)", pos, m)
			continue
		}
		name, _ := id.(string)
		t, ok := reg.Lookup(name)
		if !ok {
			db.conf.Log.Errorf("load entities: failed loading %v: entity %s was not registered (%v)", pos, name, m)
			continue
		}
		if s, ok := t.(world.SaveableEntityType); ok {
			if v := s.DecodeNBT(m); v != nil {
				a = append(a, v)
			}
		}
	}
	return a, nil
}

// SaveEntities saves all entities to the chunk position passed.
func (db *DB) SaveEntities(pos world.ChunkPos, entities []world.Entity, dim world.Dimension) error {
	if len(entities) == 0 {
		return db.ldb.Delete(append(db.index(pos, dim), keyEntities), nil)
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
			return fmt.Errorf("save entities: error encoding NBT: %w", err)
		}
	}
	return db.ldb.Put(append(db.index(pos, dim), keyEntities), buf.Bytes(), nil)
}

// LoadBlockNBT loads all block entities from the chunk position passed.
func (db *DB) LoadBlockNBT(position world.ChunkPos, dim world.Dimension) ([]map[string]any, error) {
	data, err := db.ldb.Get(append(db.index(position, dim), keyBlockEntities), nil)
	if err != leveldb.ErrNotFound && err != nil {
		return nil, err
	}
	var a []map[string]any

	buf := bytes.NewBuffer(data)
	dec := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian)

	for buf.Len() != 0 {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("error decoding block NBT: %w", err)
		}
		a = append(a, m)
	}
	return a, nil
}

// SaveBlockNBT saves all block NBT data to the chunk position passed.
func (db *DB) SaveBlockNBT(position world.ChunkPos, data []map[string]any, dim world.Dimension) error {
	if len(data) == 0 {
		return db.ldb.Delete(append(db.index(position, dim), keyBlockEntities), nil)
	}
	buf := bytes.NewBuffer(nil)
	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	for _, d := range data {
		if err := enc.Encode(d); err != nil {
			return fmt.Errorf("error encoding block NBT: %w", err)
		}
	}
	return db.ldb.Put(append(db.index(position, dim), keyBlockEntities), buf.Bytes(), nil)
}

// NewChunkIterator returns a ChunkIterator that may be used to iterate over all
// position/chunk pairs in a database.
// An IteratorRange r may be passed to specify limits in terms of what chunks
// should be read. r may be set to nil to read all chunks from the DB.
func (db *DB) NewChunkIterator(r *IteratorRange) *ChunkIterator {
	if r == nil {
		r = &IteratorRange{}
	}
	return newChunkIterator(db, r)
}

// Close closes the provider, saving any file that might need to be saved, such as the level.dat.
func (db *DB) Close() error {
	db.ldat.LastPlayed = time.Now().Unix()

	buf := bytes.NewBuffer(nil)
	if err := db.ldat.marshal(buf); err != nil {
		return fmt.Errorf("encode level.dat: %w", err)
	}
	if err := os.WriteFile(filepath.Join(db.dir, "level.dat"), buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing levelname.txt: %w", err)
	}
	if err := os.WriteFile(filepath.Join(db.dir, "levelname.txt"), []byte(db.ldat.LevelName), 0644); err != nil {
		return fmt.Errorf("error writing levelname.txt: %w", err)
	}
	return db.ldb.Close()
}

// index returns a byte buffer holding the written index of the chunk position passed. If the dimension passed to New
// is not world.Overworld, the length of the index returned is 12. It is 8 otherwise.
func (db *DB) index(position world.ChunkPos, d world.Dimension) []byte {
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
