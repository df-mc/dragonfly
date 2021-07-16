package mcdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/df-mc/goleveldb/leveldb/opt"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"time"
)

// Provider implements a world provider for the Minecraft world format, which is based on a leveldb database.
type Provider struct {
	db  *leveldb.DB
	dir string
	d   data
}

// chunkVersion is the current version of chunks.
const chunkVersion = 19

// New creates a new provider reading and writing files to files under the path passed. If a world is present
// at the path, New will parse its data and initialise the world with it. If the data cannot be parsed, an
// error is returned.
func New(dir string) (*Provider, error) {
	_ = os.MkdirAll(filepath.Join(dir, "db"), 0777)

	p := &Provider{dir: dir}
	if _, err := os.Stat(filepath.Join(dir, "level.dat")); os.IsNotExist(err) {
		// A level.dat was not currently present for the world.
		p.initDefaultLevelDat()
	} else {
		f, err := ioutil.ReadFile(filepath.Join(dir, "level.dat"))
		if err != nil {
			return nil, fmt.Errorf("error opening level.dat file: %w", err)
		}
		// The first 8 bytes are a useless header (version and length): We don't need it.
		if len(f) < 8 {
			// The file did not have enough content, meaning it is corrupted. We return an error.
			return nil, fmt.Errorf("level.dat exists but has no data")
		}
		if err := nbt.UnmarshalEncoding(f[8:], &p.d, nbt.LittleEndian); err != nil {
			return nil, fmt.Errorf("error decoding level.dat NBT: %w", err)
		}
		p.d.WorldStartCount++
	}
	db, err := leveldb.OpenFile(filepath.Join(dir, "db"), &opt.Options{
		Compression: opt.FlateCompression,
		BlockSize:   16 * opt.KiB,
	})
	if err != nil {
		return nil, fmt.Errorf("error opening leveldb database: %w", err)
	}
	p.db = db
	return p, nil
}

// initDefaultLevelDat initialises a default level.dat file.
func (p *Provider) initDefaultLevelDat() {
	p.d.DoDayLightCycle = true
	p.d.BaseGameVersion = protocol.CurrentVersion
	p.d.LevelName = "World"
	p.d.GameType = 1
	p.d.StorageVersion = 8
	p.d.Generator = 1
	p.d.NetworkVersion = protocol.CurrentProtocol
	p.d.Abilities.WalkSpeed = 0.1
	p.d.PVP = true
	p.d.WorldStartCount = 1
	p.d.RandomTickSpeed = 1
	p.d.FallDamage = true
	p.d.FireDamage = true
	p.d.DrowningDamage = true
	p.d.CommandsEnabled = true
	p.d.MultiPlayerGame = true
	p.d.SpawnY = math.MaxInt32
}

// Settings returns the world.Settings of the world loaded by the Provider.
func (p *Provider) Settings() world.Settings {
	return world.Settings{
		Name:            p.d.LevelName,
		Spawn:           cube.Pos{int(p.d.SpawnX), int(p.d.SpawnY), int(p.d.SpawnZ)},
		Time:            p.d.Time,
		TimeCycle:       p.d.DoDayLightCycle,
		CurrentTick:     p.d.CurrentTick,
		DefaultGameMode: p.LoadDefaultGameMode(),
		Difficulty:      p.LoadDifficulty(),
	}
}

// SaveSettings saves the world.Settings passed to the level.dat.
func (p *Provider) SaveSettings(s world.Settings) {
	p.d.LevelName = s.Name
	p.d.SpawnX, p.d.SpawnY, p.d.SpawnZ = int32(s.Spawn.X()), int32(s.Spawn.Y()), int32(s.Spawn.Z())
	p.d.Time = s.Time
	p.d.DoDayLightCycle = s.TimeCycle
	p.d.CurrentTick = s.CurrentTick
	p.SaveDefaultGameMode(s.DefaultGameMode)
	p.SaveDifficulty(s.Difficulty)
}

// LoadChunk loads a chunk at the position passed from the leveldb database. If it doesn't exist, exists is
// false. If an error is returned, exists is always assumed to be true.
func (p *Provider) LoadChunk(position world.ChunkPos) (c *chunk.Chunk, exists bool, err error) {
	data := chunk.SerialisedData{}
	key := index(position)

	// This key is where the version of a chunk resides. The chunk version has changed many times, without any
	// actual substantial changes, so we don't check this.
	_, err = p.db.Get(append(key, keyVersion), nil)
	if err == leveldb.ErrNotFound {
		// The new key was not found, so we try the old key.
		if _, err = p.db.Get(append(key, keyVersionOld), nil); err != nil {
			return nil, false, nil
		}
	} else if err != nil {
		return nil, true, fmt.Errorf("error reading version: %w", err)
	}

	data.Data2D, err = p.db.Get(append(key, key2DData), nil)
	if err == leveldb.ErrNotFound {
		return nil, false, nil
	} else if err != nil {
		return nil, true, fmt.Errorf("error reading 2D data: %w", err)
	}

	data.BlockNBT, err = p.db.Get(append(key, keyBlockEntities), nil)
	// Block entities aren't present when there aren't any, so it's okay if we can't find the key.
	if err != nil && err != leveldb.ErrNotFound {
		return nil, true, fmt.Errorf("error reading block entities: %w", err)
	}

	for y := byte(0); y < 16; y++ {
		data.SubChunks[y], err = p.db.Get(append(key, keySubChunkData, y), nil)
		if err == leveldb.ErrNotFound {
			// No sub chunk present at this Y level. We skip this one and move to the next, which might still
			// be present.
			continue
		} else if err != nil {
			return nil, true, fmt.Errorf("error reading 2D sub chunk %v: %w", y, err)
		}
	}
	c, err = chunk.DiskDecode(data)
	return c, true, err
}

// SaveChunk saves a chunk at the position passed to the leveldb database. Its version is written as the
// version in the chunkVersion constant.
func (p *Provider) SaveChunk(position world.ChunkPos, c *chunk.Chunk) error {
	data := chunk.DiskEncode(c, false)

	key := index(position)
	_ = p.db.Put(append(key, keyVersion), []byte{chunkVersion}, nil)
	_ = p.db.Put(append(key, key2DData), data.Data2D, nil)

	finalisation := make([]byte, 4)
	binary.LittleEndian.PutUint32(finalisation, 2)
	_ = p.db.Put(append(key, keyFinalisation), finalisation, nil)

	if len(data.BlockNBT) != 0 {
		// We only write block NBT if there actually is any.
		_ = p.db.Put(append(key, keyBlockEntities), data.BlockNBT, nil)
	}
	for y, sub := range data.SubChunks {
		if len(sub) == 0 {
			// No sub chunk here: Delete it from the database and continue.
			_ = p.db.Delete(append(key, keySubChunkData, byte(y)), nil)
			continue
		}
		_ = p.db.Put(append(key, keySubChunkData, byte(y)), sub, nil)
	}
	return nil
}

// LoadDefaultGameMode returns the default game mode stored in the level.dat.
func (p *Provider) LoadDefaultGameMode() world.GameMode {
	switch p.d.GameType {
	default:
		return world.GameModeAdventure{}
	case 0:
		return world.GameModeSurvival{}
	case 1:
		return world.GameModeCreative{}
	case 2:
		return world.GameModeAdventure{}
	case 3:
		return world.GameModeSpectator{}
	}
}

// SaveDefaultGameMode changes the default game mode in the level.dat.
func (p *Provider) SaveDefaultGameMode(mode world.GameMode) {
	switch mode.(type) {
	case world.GameModeSurvival:
		p.d.GameType = 0
	case world.GameModeCreative:
		p.d.GameType = 1
	case world.GameModeAdventure:
		p.d.GameType = 2
	case world.GameModeSpectator:
		p.d.GameType = 3
	}
}

// LoadDifficulty loads the difficulty stored in the level.dat.
func (p *Provider) LoadDifficulty() world.Difficulty {
	switch p.d.Difficulty {
	default:
		return world.DifficultyNormal{}
	case 0:
		return world.DifficultyPeaceful{}
	case 1:
		return world.DifficultyEasy{}
	case 3:
		return world.DifficultyHard{}
	}
}

// SaveDifficulty saves the difficulty passed to the level.dat.
func (p *Provider) SaveDifficulty(d world.Difficulty) {
	switch d.(type) {
	case world.DifficultyPeaceful:
		p.d.Difficulty = 0
	case world.DifficultyEasy:
		p.d.Difficulty = 1
	case world.DifficultyNormal:
		p.d.Difficulty = 2
	case world.DifficultyHard:
		p.d.Difficulty = 3
	}
}

// LoadEntities loads all entities from the chunk position passed.
func (p *Provider) LoadEntities(world.ChunkPos) ([]world.Entity, error) {
	// TODO: Implement entities.
	return nil, nil
}

// SaveEntities saves all entities to the chunk position passed.
func (p *Provider) SaveEntities(world.ChunkPos, []world.Entity) error {
	// TODO: Implement entities.
	return nil
}

// LoadBlockNBT loads all block entities from the chunk position passed.
func (p *Provider) LoadBlockNBT(position world.ChunkPos) ([]map[string]interface{}, error) {
	data, err := p.db.Get(append(index(position), keyBlockEntities), nil)
	if err != leveldb.ErrNotFound && err != nil {
		return nil, err
	}
	var a []map[string]interface{}

	buf := bytes.NewBuffer(data)
	dec := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian)

	for buf.Len() != 0 {
		var m map[string]interface{}
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("error decoding block NBT: %w", err)
		}
		a = append(a, m)
	}
	return a, nil
}

// SaveBlockNBT saves all block NBT data to the chunk position passed.
func (p *Provider) SaveBlockNBT(position world.ChunkPos, data []map[string]interface{}) error {
	if len(data) == 0 {
		return p.db.Delete(append(index(position), keyBlockEntities), nil)
	}
	buf := bytes.NewBuffer(nil)
	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	for _, d := range data {
		if err := enc.Encode(d); err != nil {
			return fmt.Errorf("error encoding block NBT: %w", err)
		}
	}
	return p.db.Put(append(index(position), keyBlockEntities), buf.Bytes(), nil)
}

// Close closes the provider, saving any file that might need to be saved, such as the level.dat.
func (p *Provider) Close() error {
	p.d.LastPlayed = time.Now().Unix()

	f, err := os.OpenFile(filepath.Join(p.dir, "level.dat"), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening level.dat file: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.LittleEndian, int32(3))
	nbtData, err := nbt.MarshalEncoding(p.d, nbt.LittleEndian)
	if err != nil {
		return fmt.Errorf("error encoding level.dat to NBT: %w", err)
	}
	_ = binary.Write(buf, binary.LittleEndian, int32(len(nbtData)))
	_, _ = buf.Write(nbtData)

	_, _ = f.Write(buf.Bytes())

	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing level.dat: %w", err)
	}
	//noinspection SpellCheckingInspection
	if err := ioutil.WriteFile(filepath.Join(p.dir, "levelname.txt"), []byte(p.d.LevelName), 0644); err != nil {
		return fmt.Errorf("error writing levelname.txt: %w", err)
	}
	return p.db.Close()
}

// index returns a byte buffer holding the written index of the chunk position passed.
func index(position world.ChunkPos) []byte {
	x, z := uint32(position[0]), uint32(position[1])
	return []byte{
		byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24),
		byte(z), byte(z >> 8), byte(z >> 16), byte(z >> 24),
	}
}
