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
	dir string
	d   data
	set *world.Settings
	log Logger
}

// chunkVersion is the current version of chunks.
const chunkVersion = 40

// Logger is a logger implementation that may be passed to the Log field of Config. World will send errors and debug
// messages to this Logger when appropriate.
type Logger interface {
	Errorf(format string, a ...any)
	Debugf(format string, a ...any)
}

// New creates a new provider reading and writing from/to files under the path passed. If a world is present
// at the path, New will parse its data and initialise the world with it. If the data cannot be parsed, an
// error is returned.
// A compression type may be passed which will be used for the compression of new blocks written to the database. This
// will only influence the compression. Decompression of the database will happen based on IDs found in the compressed
// blocks.
func New(log Logger, dir string, compression opt.Compression) (*Provider, error) {
	_ = os.MkdirAll(filepath.Join(dir, "db"), 0777)

	p := &Provider{log: log, dir: dir}
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
	p.loadSettings()
	db, err := leveldb.OpenFile(filepath.Join(dir, "db"), &opt.Options{
		Compression: compression,
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
	p.d.Abilities.AttackMobs = true
	p.d.Abilities.AttackPlayers = true
	p.d.Abilities.Build = true
	p.d.Abilities.DoorsAndSwitches = true
	p.d.Abilities.FlySpeed = 0.05
	p.d.Abilities.Mine = true
	p.d.Abilities.OpenContainers = true
	p.d.Abilities.PlayerPermissionsLevel = 1
	p.d.Abilities.WalkSpeed = 0.1
	p.d.BaseGameVersion = "*"
	p.d.CommandBlockOutput = true
	p.d.CommandBlocksEnabled = true
	p.d.CommandsEnabled = true
	p.d.Difficulty = 2
	p.d.DoDayLightCycle = true
	p.d.DoEntityDrops = true
	p.d.DoFireTick = true
	p.d.DoInsomnia = true
	p.d.DoMobLoot = true
	p.d.DoMobSpawning = true
	p.d.DoTileDrops = true
	p.d.DoWeatherCycle = true
	p.d.DrowningDamage = true
	p.d.FallDamage = true
	p.d.FireDamage = true
	p.d.FreezeDamage = true
	p.d.FunctionCommandLimit = 10000
	p.d.GameType = 1
	p.d.Generator = 1
	p.d.HasBeenLoadedInCreative = true
	p.d.InventoryVersion = protocol.CurrentVersion
	p.d.LANBroadcast = true
	p.d.LANBroadcastIntent = true
	p.d.LastOpenedWithVersion = minimumCompatibleClientVersion
	p.d.LevelName = "My World"
	p.d.LightningLevel = 1.0
	p.d.LimitedWorldDepth = 16
	p.d.LimitedWorldOriginY = math.MaxInt16
	p.d.LimitedWorldWidth = 16
	p.d.MaxCommandChainLength = math.MaxUint16
	p.d.MinimumCompatibleClientVersion = minimumCompatibleClientVersion
	p.d.MobGriefing = true
	p.d.MultiPlayerGame = true
	p.d.MultiPlayerGameIntent = true
	p.d.NaturalRegeneration = true
	p.d.NetherScale = 8
	p.d.NetworkVersion = protocol.CurrentProtocol
	p.d.PVP = true
	p.d.Platform = 2
	p.d.PlatformBroadcastIntent = 3
	p.d.RainLevel = 1.0
	p.d.RandomSeed = time.Now().Unix()
	p.d.RandomTickSpeed = 1
	p.d.RespawnBlocksExplode = true
	p.d.SendCommandFeedback = true
	p.d.ServerChunkTickRange = 6
	p.d.ShowBorderEffect = true
	p.d.ShowDeathMessages = true
	p.d.ShowTags = true
	p.d.SpawnMobs = true
	p.d.SpawnRadius = 5
	p.d.SpawnRadius = 5
	p.d.SpawnY = math.MaxInt16
	p.d.StorageVersion = 9
	p.d.TNTExplodes = true
	p.d.WorldStartCount = 1
	p.d.WorldVersion = 1
	p.d.XBLBroadcastIntent = 3
}

// Settings returns the world.Settings of the world loaded by the Provider.
func (p *Provider) Settings() *world.Settings {
	return p.set
}

// loadSettings loads the settings in the level.dat into a world.Settings struct and stores it, so that it can be
// returned through a call to Settings.
func (p *Provider) loadSettings() {
	p.set = &world.Settings{
		Name:            p.d.LevelName,
		Spawn:           cube.Pos{int(p.d.SpawnX), int(p.d.SpawnY), int(p.d.SpawnZ)},
		Time:            p.d.Time,
		TimeCycle:       p.d.DoDayLightCycle,
		RainTime:        int64(p.d.RainTime),
		Raining:         p.d.RainLevel > 0,
		ThunderTime:     int64(p.d.LightningTime),
		Thundering:      p.d.LightningLevel > 0,
		WeatherCycle:    p.d.DoWeatherCycle,
		CurrentTick:     p.d.CurrentTick,
		DefaultGameMode: p.loadDefaultGameMode(),
		Difficulty:      p.loadDifficulty(),
		TickRange:       p.d.ServerChunkTickRange,
	}
}

// SaveSettings saves the world.Settings passed to the level.dat.
func (p *Provider) SaveSettings(s *world.Settings) {
	p.d.LevelName = s.Name
	p.d.SpawnX, p.d.SpawnY, p.d.SpawnZ = int32(s.Spawn.X()), int32(s.Spawn.Y()), int32(s.Spawn.Z())
	p.d.LimitedWorldOriginX, p.d.LimitedWorldOriginY, p.d.LimitedWorldOriginZ = p.d.SpawnX, p.d.SpawnY, p.d.SpawnZ
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

// playerData holds the fields that indicate where player data is stored for a player with a specific UUID.
type playerData struct {
	UUID         string `nbt:"MsaId"`
	ServerID     string `nbt:"ServerId"`
	SelfSignedID string `nbt:"SelfSignedId"`
}

// LoadPlayerSpawnPosition loads the players spawn position stored in the level.dat from their UUID.
func (p *Provider) LoadPlayerSpawnPosition(id uuid.UUID) (pos cube.Pos, exists bool, err error) {
	serverData, _, exists, err := p.loadPlayerData(id)
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
func (p *Provider) loadPlayerData(id uuid.UUID) (serverData map[string]interface{}, key string, exists bool, err error) {
	data, err := p.db.Get([]byte("player_"+id.String()), nil)
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
	serverDB, err := p.db.Get([]byte(d.ServerID), nil)
	if err != nil {
		return nil, d.ServerID, true, fmt.Errorf("error reading server data for player %v (%v): %w", id, d.ServerID, err)
	}

	if err := nbt.UnmarshalEncoding(serverDB, &serverData, nbt.LittleEndian); err != nil {
		return nil, d.ServerID, true, fmt.Errorf("error decoding server data for player %v", id)
	}
	return serverData, d.ServerID, true, nil
}

// SavePlayerSpawnPosition saves the player spawn position passed to the levelDB database.
func (p *Provider) SavePlayerSpawnPosition(id uuid.UUID, pos cube.Pos) error {
	_, err := p.db.Get([]byte("player_"+id.String()), nil)
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
		if err := p.db.Put([]byte("player_"+id.String()), data, nil); err != nil {
			return fmt.Errorf("error writing player data for id %v: %w", id, err)
		}
	} else {
		if d, k, _, err = p.loadPlayerData(id); err != nil {
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
	if err = p.db.Put([]byte(k), data, nil); err != nil {
		return fmt.Errorf("error writing server data for player %v: %w", id, err)
	}
	return nil
}

// LoadChunk loads a chunk at the position passed from the leveldb database. If it doesn't exist, exists is
// false. If an error is returned, exists is always assumed to be true.
func (p *Provider) LoadChunk(position world.ChunkPos, dim world.Dimension) (c *chunk.Chunk, exists bool, err error) {
	data := chunk.SerialisedData{}
	key := p.index(position, dim)

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
	data.SubChunks = make([][]byte, (dim.Range().Height()>>4)+1)
	for i := range data.SubChunks {
		data.SubChunks[i], err = p.db.Get(append(key, keySubChunkData, uint8(i+(dim.Range()[0]>>4))), nil)
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
func (p *Provider) SaveChunk(position world.ChunkPos, c *chunk.Chunk, dim world.Dimension) error {
	data := chunk.Encode(c, chunk.DiskEncoding)

	key := p.index(position, dim)
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
func (p *Provider) LoadEntities(pos world.ChunkPos, dim world.Dimension) ([]world.SaveableEntity, error) {
	data, err := p.db.Get(append(p.index(pos, dim), keyEntities), nil)
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
			p.log.Errorf("load entities: failed loading %v: entity had data but no identifier (%v)", pos, m)
			continue
		}
		name, _ := id.(string)
		e, ok := world.EntityByName(name)
		if !ok {
			p.log.Errorf("load entities: failed loading %v: entity %s was not registered (%v)", pos, name, m)
			continue
		}
		if s, ok := e.(world.SaveableEntity); ok {
			if v := s.DecodeNBT(m); v != nil {
				a = append(a, v.(world.SaveableEntity))
			}
		}
	}
	return a, nil
}

// SaveEntities saves all entities to the chunk position passed.
func (p *Provider) SaveEntities(pos world.ChunkPos, entities []world.SaveableEntity, dim world.Dimension) error {
	if len(entities) == 0 {
		return p.db.Delete(append(p.index(pos, dim), keyEntities), nil)
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
	return p.db.Put(append(p.index(pos, dim), keyEntities), buf.Bytes(), nil)
}

// LoadBlockNBT loads all block entities from the chunk position passed.
func (p *Provider) LoadBlockNBT(position world.ChunkPos, dim world.Dimension) ([]map[string]any, error) {
	data, err := p.db.Get(append(p.index(position, dim), keyBlockEntities), nil)
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
func (p *Provider) SaveBlockNBT(position world.ChunkPos, data []map[string]any, dim world.Dimension) error {
	if len(data) == 0 {
		return p.db.Delete(append(p.index(position, dim), keyBlockEntities), nil)
	}
	buf := bytes.NewBuffer(nil)
	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	for _, d := range data {
		if err := enc.Encode(d); err != nil {
			return fmt.Errorf("error encoding block NBT: %w", err)
		}
	}
	return p.db.Put(append(p.index(position, dim), keyBlockEntities), buf.Bytes(), nil)
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

// index returns a byte buffer holding the written index of the chunk position passed. If the dimension passed to New
// is not world.Overworld, the length of the index returned is 12. It is 8 otherwise.
func (p *Provider) index(position world.ChunkPos, d world.Dimension) []byte {
	x, z, dim := uint32(position[0]), uint32(position[1]), uint32(d.EncodeDimension())
	b := make([]byte, 12)

	binary.LittleEndian.PutUint32(b, x)
	binary.LittleEndian.PutUint32(b[4:], z)
	if dim == 0 {
		return b[:8]
	}
	binary.LittleEndian.PutUint32(b[8:], dim)
	return b
}
