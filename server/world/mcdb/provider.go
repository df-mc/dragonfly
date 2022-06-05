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
	"github.com/df-mc/goleveldb/leveldb/opt"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
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
	dim world.Dimension
	dir string
	d   data
}

// chunkVersion is the current version of chunks.
const chunkVersion = 27

// New creates a new provider reading and writing from/to files under the path passed. If a world is present
// at the path, New will parse its data and initialise the world with it. If the data cannot be parsed, an
// error is returned.
// A compression type may be passed which will be used for the compression of new blocks written to the database. This
// will only influence the compression. Decompression of the database will happen based on IDs found in the compressed
// blocks.
func New(dir string, d world.Dimension, compression opt.Compression) (*Provider, error) {
	_ = os.MkdirAll(filepath.Join(dir, "db"), 0777)

	p := &Provider{dir: dir, dim: d}
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
	db, ok := cacheLoad(dir)
	if !ok {
		var err error
		if db, err = leveldb.OpenFile(filepath.Join(dir, "db"), &opt.Options{
			Compression: compression,
			BlockSize:   16 * opt.KiB,
		}); err != nil {
			return nil, fmt.Errorf("error opening leveldb database: %w", err)
		}
		cacheStore(dir, db)
	}

	p.db = db
	return p, nil
}

// initDefaultLevelDat initialises a default level.dat file.
func (p *Provider) initDefaultLevelDat() {
	p.d.DoDayLightCycle = true
	p.d.DoWeatherCycle = true
	p.d.BaseGameVersion = protocol.CurrentVersion
	p.d.NetworkVersion = protocol.CurrentProtocol
	p.d.LastOpenedWithVersion = minimumCompatibleClientVersion
	p.d.MinimumCompatibleClientVersion = minimumCompatibleClientVersion
	p.d.LevelName = "World"
	p.d.GameType = 1
	p.d.StorageVersion = 8
	p.d.Generator = 1
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
	p.d.Difficulty = 2
	p.d.DoWeatherCycle = true
	p.d.RainLevel = 1.0
	p.d.LightningLevel = 1.0
	p.d.ServerChunkTickRange = 6
	p.d.NetherScale = 8
}

// Settings returns the world.Settings of the world loaded by the Provider.
func (p *Provider) Settings(s *world.Settings) {
	s.Name = p.d.LevelName
	s.Spawn = cube.Pos{int(p.d.SpawnX), int(p.d.SpawnY), int(p.d.SpawnZ)}
	s.Time = p.d.Time
	s.TimeCycle = p.d.DoDayLightCycle
	s.WeatherCycle = p.d.DoWeatherCycle
	s.RainTime = int64(p.d.RainTime)
	s.Raining = p.d.RainLevel > 0
	s.ThunderTime = int64(p.d.LightningTime)
	s.Thundering = p.d.LightningLevel > 0
	s.CurrentTick = p.d.CurrentTick
	s.DefaultGameMode = p.loadDefaultGameMode()
	s.Difficulty = p.loadDifficulty()
	s.TickRange = p.d.ServerChunkTickRange
}

// SaveSettings saves the world.Settings passed to the level.dat.
func (p *Provider) SaveSettings(s *world.Settings) {
	p.d.LevelName = s.Name
	p.d.SpawnX, p.d.SpawnY, p.d.SpawnZ = int32(s.Spawn.X()), int32(s.Spawn.Y()), int32(s.Spawn.Z())
	p.d.Time = s.Time
	p.d.DoDayLightCycle = s.TimeCycle
	p.d.DoWeatherCycle = s.WeatherCycle
	p.d.RainTime, p.d.RainLevel = int32(s.RainTime), 0
	p.d.LightningTime, p.d.LightningLevel = int32(s.ThunderTime), 0
	if s.Raining {
		p.d.RainLevel = 1
	}
	if s.Thundering {
		p.d.LightningLevel = 1
	}
	p.d.CurrentTick = s.CurrentTick
	p.d.ServerChunkTickRange = s.TickRange
	p.saveDefaultGameMode(s.DefaultGameMode)
	p.saveDifficulty(s.Difficulty)
}

// LoadPlayerSpawnPosition loads the players spawn position stored in the level.dat from their UUID.
func (p *Provider) LoadPlayerSpawnPosition(uuid uuid.UUID) (pos mgl64.Vec3, exists bool, err error) {
	data, err := p.db.Get(append([]byte("player_"), uuid[:]...), nil)
	if err != nil {
		return mgl64.Vec3{}, false, err
	}
	var playerData map[string]string

	if err := nbt.UnmarshalEncoding(data, &playerData, nbt.LittleEndian); err != nil {
		return mgl64.Vec3{}, false, err
	}
	if playerData["MsaID"] != uuid.String() || playerData["ServerId"] == "" {
		return mgl64.Vec3{}, false, fmt.Errorf("load player spawn: invalid data for player uuid: %v", uuid.String())
	}
	serverDB, err := p.db.Get([]byte(playerData["ServerId"]), nil)
	if err != nil {
		return mgl64.Vec3{}, false, err
	}

	var serverData map[string]any
	if err := nbt.UnmarshalEncoding(serverDB, &serverData, nbt.LittleEndian); err != nil {
		return mgl64.Vec3{}, false, err
	}

	x, y, z := serverData["SpawnX"], serverData["SpawnY"], serverData["SpawnZ"]
	if x == nil || y == nil || z == nil {
		return mgl64.Vec3{}, false, nil
	}
	return mgl64.Vec3{x.(float64), y.(float64), z.(float64)}, true, nil
}

// SavePlayerSpawnPosition saves the player spawn position passed to the level.dat.
func (p Provider) SavePlayerSpawnPosition(uuid uuid.UUID, pos mgl64.Vec3) error {
	data, err := p.db.Get(append([]byte("player_"), uuid[:]...), nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		data, err = nbt.MarshalEncoding(map[string]string{
			"MsaID":    uuid.String(),
			"ServerID": "player_server_" + uuid.String(),
		}, nbt.LittleEndian)
		if err != nil {
			return err
		}
		if err := p.db.Put(append([]byte("player_"), uuid[:]...), data, nil); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	var playerData map[string]string

	if err := nbt.UnmarshalEncoding(data, &playerData, nbt.LittleEndian); err != nil {
		return err
	}
	if playerData["MsaID"] != uuid.String() || playerData["ServerId"] == "" {
		return fmt.Errorf("save player spawn: invalid data for player: %v", uuid.String())
	}

	serverDB, err := p.db.Get([]byte(playerData["ServerId"]), nil)
	if err != nil {
		return err
	}

	var serverData map[string]any
	if err := nbt.UnmarshalEncoding(serverDB, &serverData, nbt.LittleEndian); err != nil {
		return err
	}
	serverData["SpawnX"] = pos.X()
	serverData["SpawnY"] = pos.Y()
	serverData["SpawnZ"] = pos.Z()
	coords, err := nbt.MarshalEncoding(serverData, nbt.LittleEndian)
	if err != nil {
		return err
	}
	if err := p.db.Put([]byte(playerData["ServerId"]), coords, nil); err != nil {
		return err
	}
	return nil
}

// LoadChunk loads a chunk at the position passed from the leveldb database. If it doesn't exist, exists is
// false. If an error is returned, exists is always assumed to be true.
func (p *Provider) LoadChunk(position world.ChunkPos) (c *chunk.Chunk, exists bool, err error) {
	data := chunk.SerialisedData{}
	key := p.index(position)

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

	data.Biomes, err = p.db.Get(append(key, key3DData), nil)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, false, fmt.Errorf("error reading 3D data: %w", err)
	}
	if len(data.Biomes) > 512 {
		// Strip the heightmap from the biomes.
		data.Biomes = data.Biomes[512:]
	}

	data.BlockNBT, err = p.db.Get(append(key, keyBlockEntities), nil)
	// Block entities aren't present when there aren't any, so it's okay if we can't find the key.
	if err != nil && err != leveldb.ErrNotFound {
		return nil, true, fmt.Errorf("error reading block entities: %w", err)
	}
	data.SubChunks = make([][]byte, (p.dim.Range().Height()>>4)+1)
	for i := range data.SubChunks {
		data.SubChunks[i], err = p.db.Get(append(key, keySubChunkData, uint8(i+(p.dim.Range()[0]>>4))), nil)
		if err == leveldb.ErrNotFound {
			// No sub chunk present at this Y level. We skip this one and move to the next, which might still
			// be present.
			continue
		} else if err != nil {
			return nil, true, fmt.Errorf("error reading sub chunk data %v: %w", i, err)
		}
	}
	c, err = chunk.DiskDecode(data, p.dim.Range())
	return c, true, err
}

// SaveChunk saves a chunk at the position passed to the leveldb database. Its version is written as the
// version in the chunkVersion constant.
func (p *Provider) SaveChunk(position world.ChunkPos, c *chunk.Chunk) error {
	data := chunk.Encode(c, chunk.DiskEncoding)

	key := p.index(position)
	_ = p.db.Put(append(key, keyVersion), []byte{chunkVersion}, nil)
	// Write the heightmap by just writing 512 empty bytes.
	_ = p.db.Put(append(key, key3DData), append(make([]byte, 512), data.Biomes...), nil)

	finalisation := make([]byte, 4)
	binary.LittleEndian.PutUint32(finalisation, 2)
	_ = p.db.Put(append(key, keyFinalisation), finalisation, nil)

	for i, sub := range data.SubChunks {
		_ = p.db.Put(append(key, keySubChunkData, byte(i+(c.Range()[0]>>4))), sub, nil)
	}
	return nil
}

// loadDefaultGameMode returns the default game mode stored in the level.dat.
func (p *Provider) loadDefaultGameMode() world.GameMode {
	switch p.d.GameType {
	default:
		return world.GameModeSurvival
	case 1:
		return world.GameModeCreative
	case 2:
		return world.GameModeAdventure
	case 3:
		return world.GameModeSpectator
	}
}

// saveDefaultGameMode changes the default game mode in the level.dat.
func (p *Provider) saveDefaultGameMode(mode world.GameMode) {
	switch mode {
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

// loadDifficulty loads the difficulty stored in the level.dat.
func (p *Provider) loadDifficulty() world.Difficulty {
	switch p.d.Difficulty {
	default:
		return world.DifficultyNormal
	case 0:
		return world.DifficultyPeaceful
	case 1:
		return world.DifficultyEasy
	case 3:
		return world.DifficultyHard
	}
}

// saveDifficulty saves the difficulty passed to the level.dat.
func (p *Provider) saveDifficulty(d world.Difficulty) {
	switch d {
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
func (p *Provider) LoadEntities(pos world.ChunkPos) ([]world.SaveableEntity, error) {
	data, err := p.db.Get(append(p.index(pos), keyEntities), nil)
	if err != leveldb.ErrNotFound && err != nil {
		return nil, err
	}
	var a []world.SaveableEntity

	buf := bytes.NewBuffer(data)
	dec := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian)

	for buf.Len() != 0 {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("error decoding block NBT: %w", err)
		}
		id, ok := m["identifier"]
		if !ok {
			return nil, fmt.Errorf("entity has no ID but data (%v)", m)
		}
		name, _ := id.(string)
		e, ok := world.EntityByName(name)
		if !ok {
			// Entity was not registered: This can only be expected sometimes, so the best we can do is to just
			// ignore this and proceed.
			continue
		}
		if v := e.DecodeNBT(m); v != nil {
			a = append(a, v.(world.SaveableEntity))
		}
	}
	return a, nil
}

// SaveEntities saves all entities to the chunk position passed.
func (p *Provider) SaveEntities(pos world.ChunkPos, entities []world.SaveableEntity) error {
	if len(entities) == 0 {
		return p.db.Delete(append(p.index(pos), keyEntities), nil)
	}

	buf := bytes.NewBuffer(nil)
	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	for _, e := range entities {
		x := e.EncodeNBT()
		x["identifier"] = e.EncodeEntity()
		if err := enc.Encode(x); err != nil {
			return fmt.Errorf("save entities: error encoding NBT: %w", err)
		}
	}
	return p.db.Put(append(p.index(pos), keyEntities), buf.Bytes(), nil)
}

// LoadBlockNBT loads all block entities from the chunk position passed.
func (p *Provider) LoadBlockNBT(position world.ChunkPos) ([]map[string]any, error) {
	data, err := p.db.Get(append(p.index(position), keyBlockEntities), nil)
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
func (p *Provider) SaveBlockNBT(position world.ChunkPos, data []map[string]any) error {
	if len(data) == 0 {
		return p.db.Delete(append(p.index(position), keyBlockEntities), nil)
	}
	buf := bytes.NewBuffer(nil)
	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	for _, d := range data {
		if err := enc.Encode(d); err != nil {
			return fmt.Errorf("error encoding block NBT: %w", err)
		}
	}
	return p.db.Put(append(p.index(position), keyBlockEntities), buf.Bytes(), nil)
}

// Close closes the provider, saving any file that might need to be saved, such as the level.dat.
func (p *Provider) Close() error {
	p.d.LastPlayed = time.Now().Unix()
	if cacheDelete(p.dir) != 0 {
		// The same provider is still alive elsewhere. Don't store the data to the level.dat and levelname.txt just yet.
		return nil
	}
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

// index returns a byte buffer holding the written index of the chunk position passed. If the dimension passed to New
// is not world.Overworld, the length of the index returned is 12. It is 8 otherwise.
func (p *Provider) index(position world.ChunkPos) []byte {
	x, z, dim := uint32(position[0]), uint32(position[1]), uint32(p.dim.EncodeDimension())
	b := make([]byte, 12)

	binary.LittleEndian.PutUint32(b, x)
	binary.LittleEndian.PutUint32(b[4:], z)
	if dim == 0 {
		return b[:8]
	}
	binary.LittleEndian.PutUint32(b[8:], dim)
	return b
}
