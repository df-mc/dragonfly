package leveldat

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
	"time"
)

// Data holds a collection of data that specify a range of Settings of the
// world. These Settings usually alter the way that players interact with the
// world. The data held here is usually saved in a level.dat file of the world.
// Data may be used in LevelDat.Unmarshal to collect the data of the level.dat.
type Data struct {
	BaseGameVersion                string `nbt:"baseGameVersion"`
	BiomeOverride                  string
	ConfirmedPlatformLockedContent bool
	CenterMapsToOrigin             bool
	CheatsEnabled                  bool  `nbt:"cheatsEnabled"`
	DaylightCycle                  int32 `nbt:"daylightCycle"`
	Difficulty                     int32
	EduOffer                       int32 `nbt:"eduOffer"`
	FlatWorldLayers                string
	ForceGameType                  bool
	GameType                       int32
	Generator                      int32
	InventoryVersion               string
	LANBroadcast                   bool
	LANBroadcastIntent             bool
	LastPlayed                     int64
	LevelName                      string
	LimitedWorldOriginX            int32
	LimitedWorldOriginY            int32
	LimitedWorldOriginZ            int32
	LimitedWorldDepth              int32 `nbt:"limitedWorldDepth"`
	LimitedWorldWidth              int32 `nbt:"limitedWorldWidth"`
	LocatorBar                     bool  `nbt:"locatorbar"`
	MinimumCompatibleClientVersion []int32
	MultiPlayerGame                bool `nbt:"MultiplayerGame"`
	MultiPlayerGameIntent          bool `nbt:"MultiplayerGameIntent"`
	NetherScale                    int32
	NetworkVersion                 int32
	Platform                       int32
	PlatformBroadcastIntent        int32
	RandomSeed                     int64
	ShowTags                       bool `nbt:"showtags"`
	SingleUseWorld                 bool `nbt:"isSingleUseWorld"`
	SpawnX, SpawnY, SpawnZ         int32
	SpawnV1Villagers               bool
	StorageVersion                 int32
	Time                           int64
	XBLBroadcast                   bool
	XBLBroadcastIntent             int32
	XBLBroadcastMode               int32
	Abilities                      struct {
		AttackMobs             bool    `nbt:"attackmobs"`
		AttackPlayers          bool    `nbt:"attackplayers"`
		Build                  bool    `nbt:"build"`
		Mine                   bool    `nbt:"mine"`
		DoorsAndSwitches       bool    `nbt:"doorsandswitches"`
		FlySpeed               float32 `nbt:"flySpeed"`
		Flying                 bool    `nbt:"flying"`
		InstantBuild           bool    `nbt:"instabuild"`
		Invulnerable           bool    `nbt:"invulnerable"`
		Lightning              bool    `nbt:"lightning"`
		MayFly                 bool    `nbt:"mayfly"`
		OP                     bool    `nbt:"op"`
		OpenContainers         bool    `nbt:"opencontainers"`
		PermissionsLevel       int32   `nbt:"permissionsLevel"`
		PlayerPermissionsLevel int32   `nbt:"playerPermissionsLevel"`
		Teleport               bool    `nbt:"teleport"`
		WalkSpeed              float32 `nbt:"walkSpeed"`
		VerticalFlySpeed       float32 `nbt:"verticalFlySpeed"`
	} `nbt:"abilities"`
	BonusChestEnabled              bool           `nbt:"bonusChestEnabled"`
	BonusChestSpawned              bool           `nbt:"bonusChestSpawned"`
	CommandBlockOutput             bool           `nbt:"commandblockoutput"`
	CommandBlocksEnabled           bool           `nbt:"commandblocksenabled"`
	CommandsEnabled                bool           `nbt:"commandsEnabled"`
	CurrentTick                    int64          `nbt:"currentTick"`
	DoDayLightCycle                bool           `nbt:"dodaylightcycle"`
	DoEntityDrops                  bool           `nbt:"doentitydrops"`
	DoFireTick                     bool           `nbt:"dofiretick"`
	DoImmediateRespawn             bool           `nbt:"doimmediaterespawn"`
	DoInsomnia                     bool           `nbt:"doinsomnia"`
	DoMobLoot                      bool           `nbt:"domobloot"`
	DoMobSpawning                  bool           `nbt:"domobspawning"`
	DoTileDrops                    bool           `nbt:"dotiledrops"`
	DoWeatherCycle                 bool           `nbt:"doweathercycle"`
	DrowningDamage                 bool           `nbt:"drowningdamage"`
	EduLevel                       bool           `nbt:"eduLevel"`
	EducationFeaturesEnabled       bool           `nbt:"educationFeaturesEnabled"`
	ExperimentalGamePlay           bool           `nbt:"experimentalgameplay"`
	FallDamage                     bool           `nbt:"falldamage"`
	FireDamage                     bool           `nbt:"firedamage"`
	FunctionCommandLimit           int32          `nbt:"functioncommandlimit"`
	HasBeenLoadedInCreative        bool           `nbt:"hasBeenLoadedInCreative"`
	HasLockedBehaviourPack         bool           `nbt:"hasLockedBehaviorPack"`
	HasLockedResourcePack          bool           `nbt:"hasLockedResourcePack"`
	ImmutableWorld                 bool           `nbt:"immutableWorld"`
	IsCreatedInEditor              bool           `nbt:"isCreatedInEditor"`
	IsExportedFromEditor           bool           `nbt:"isExportedFromEditor"`
	IsFromLockedTemplate           bool           `nbt:"isFromLockedTemplate"`
	IsFromWorldTemplate            bool           `nbt:"isFromWorldTemplate"`
	IsWorldTemplateOptionLocked    bool           `nbt:"isWorldTemplateOptionLocked"`
	KeepInventory                  bool           `nbt:"keepinventory"`
	LastOpenedWithVersion          []int32        `nbt:"lastOpenedWithVersion"`
	LightningLevel                 float32        `nbt:"lightningLevel"`
	LightningTime                  int32          `nbt:"lightningTime"`
	MaxCommandChainLength          int32          `nbt:"maxcommandchainlength"`
	MobGriefing                    bool           `nbt:"mobgriefing"`
	NaturalRegeneration            bool           `nbt:"naturalregeneration"`
	PRID                           string         `nbt:"prid"`
	PVP                            bool           `nbt:"pvp"`
	RainLevel                      float32        `nbt:"rainLevel"`
	RainTime                       int32          `nbt:"rainTime"`
	RandomTickSpeed                int32          `nbt:"randomtickspeed"`
	RequiresCopiedPackRemovalCheck bool           `nbt:"requiresCopiedPackRemovalCheck"`
	SendCommandFeedback            bool           `nbt:"sendcommandfeedback"`
	ServerChunkTickRange           int32          `nbt:"serverChunkTickRange"`
	ShowCoordinates                bool           `nbt:"showcoordinates"`
	ShowDeathMessages              bool           `nbt:"showdeathmessages"`
	SpawnMobs                      bool           `nbt:"spawnMobs"`
	SpawnRadius                    int32          `nbt:"spawnradius"`
	StartWithMapEnabled            bool           `nbt:"startWithMapEnabled"`
	TexturePacksRequired           bool           `nbt:"texturePacksRequired"`
	TNTExplodes                    bool           `nbt:"tntexplodes"`
	UseMSAGamerTagsOnly            bool           `nbt:"useMsaGamertagsOnly"`
	WorldStartCount                int64          `nbt:"worldStartCount"`
	Experiments                    map[string]any `nbt:"experiments"`
	FreezeDamage                   bool           `nbt:"freezedamage"`
	WorldPolicies                  map[string]any `nbt:"world_policies"`
	WorldVersion                   int32          `nbt:"WorldVersion"`
	RespawnBlocksExplode           bool           `nbt:"respawnblocksexplode"`
	ShowBorderEffect               bool           `nbt:"showbordereffect"`
	PermissionsLevel               int32          `nbt:"permissionsLevel"`
	PlayerPermissionsLevel         int32          `nbt:"playerPermissionsLevel"`
	IsRandomSeedAllowed            bool           `nbt:"isRandomSeedAllowed"`
	DoLimitedCrafting              bool           `nbt:"dolimitedcrafting"`
	EditorWorldType                int32          `nbt:"editorWorldType"`
	PlayersSleepingPercentage      int32          `nbt:"playerssleepingpercentage"`
	RecipesUnlock                  bool           `nbt:"recipesunlock"`
	NaturalGeneration              bool           `nbt:"naturalgeneration"`
	ProjectilesCanBreakBlocks      bool           `nbt:"projectilescanbreakblocks"`
	ShowRecipeMessages             bool           `nbt:"showrecipemessages"`
	IsHardcore                     bool           `nbt:"IsHardcore"`
	ShowDaysPlayed                 bool           `nbt:"showdaysplayed"`
	TNTExplosionDropDecay          bool           `nbt:"tntexplosiondropdecay"`
	HasUncompleteWorldFileOnDisk   bool           `nbt:"HasUncompleteWorldFileOnDisk"`
	PlayerHasDied                  bool           `nbt:"PlayerHasDied"`
}

// FillDefault fills out d with all the default level.dat values.
func (d *Data) FillDefault() {
	d.Abilities.AttackMobs = true
	d.Abilities.AttackPlayers = true
	d.Abilities.Build = true
	d.Abilities.DoorsAndSwitches = true
	d.Abilities.FlySpeed = 0.05
	d.Abilities.Mine = true
	d.Abilities.OpenContainers = true
	d.Abilities.PlayerPermissionsLevel = 1
	d.Abilities.WalkSpeed = 0.1
	d.Abilities.VerticalFlySpeed = 1.0
	d.BaseGameVersion = "*"
	d.CommandBlockOutput = true
	d.CommandBlocksEnabled = true
	d.CommandsEnabled = true
	d.Difficulty = 2
	d.DoDayLightCycle = true
	d.DoEntityDrops = true
	d.DoFireTick = true
	d.DoInsomnia = true
	d.DoMobLoot = true
	d.DoMobSpawning = true
	d.DoTileDrops = true
	d.DoWeatherCycle = true
	d.DrowningDamage = true
	d.FallDamage = true
	d.FireDamage = true
	d.FreezeDamage = true
	d.FunctionCommandLimit = 10000
	d.GameType = 1
	d.Generator = 2
	d.HasBeenLoadedInCreative = true
	d.InventoryVersion = protocol.CurrentVersion
	d.LANBroadcast = true
	d.LANBroadcastIntent = true
	d.LastOpenedWithVersion = minimumCompatibleClientVersion
	d.LevelName = "World"
	d.LightningLevel = 1.0
	d.LimitedWorldDepth = 16
	d.LimitedWorldOriginY = math.MaxInt16
	d.LimitedWorldWidth = 16
	d.MaxCommandChainLength = math.MaxUint16
	d.MinimumCompatibleClientVersion = minimumCompatibleClientVersion
	d.MobGriefing = true
	d.MultiPlayerGame = true
	d.MultiPlayerGameIntent = true
	d.NaturalRegeneration = true
	d.NetherScale = 8
	d.NetworkVersion = protocol.CurrentProtocol
	d.PVP = true
	d.Platform = 2
	d.PlatformBroadcastIntent = 3
	d.RainLevel = 1.0
	d.RandomSeed = time.Now().Unix()
	d.RandomTickSpeed = 1
	d.RespawnBlocksExplode = true
	d.SendCommandFeedback = true
	d.ServerChunkTickRange = 6
	d.ShowBorderEffect = true
	d.ShowDeathMessages = true
	d.ShowTags = true
	d.SpawnMobs = true
	d.SpawnRadius = 5
	d.SpawnRadius = 5
	d.SpawnY = math.MaxInt16
	d.StorageVersion = 9
	d.TNTExplodes = true
	d.WorldVersion = 1
	d.XBLBroadcastIntent = 3
}

// Settings returns a world.Settings value based on the properties stored in d.
func (d *Data) Settings() *world.Settings {
	d.WorldStartCount += 1
	difficulty, _ := world.DifficultyByID(int(d.Difficulty))
	mode, _ := world.GameModeByID(int(d.GameType))
	return &world.Settings{
		Name:            d.LevelName,
		Spawn:           cube.Pos{int(d.SpawnX), int(d.SpawnY), int(d.SpawnZ)},
		Time:            d.Time,
		TimeCycle:       d.DoDayLightCycle,
		RainTime:        int64(d.RainTime),
		Raining:         d.RainLevel > 0,
		ThunderTime:     int64(d.LightningTime),
		Thundering:      d.LightningLevel > 0,
		WeatherCycle:    d.DoWeatherCycle,
		CurrentTick:     d.CurrentTick,
		DefaultGameMode: mode,
		Difficulty:      difficulty,
		TickRange:       d.ServerChunkTickRange,
	}
}

// PutSettings updates d with the Settings stored in s.
func (d *Data) PutSettings(s *world.Settings) {
	d.LevelName = s.Name
	d.SpawnX, d.SpawnY, d.SpawnZ = int32(s.Spawn.X()), int32(s.Spawn.Y()), int32(s.Spawn.Z())
	d.LimitedWorldOriginX, d.LimitedWorldOriginY, d.LimitedWorldOriginZ = d.SpawnX, d.SpawnY, d.SpawnZ
	d.Time = s.Time
	d.DoDayLightCycle = s.TimeCycle
	d.DoWeatherCycle = s.WeatherCycle
	d.RainTime, d.RainLevel = int32(s.RainTime), 0
	d.LightningTime, d.LightningLevel = int32(s.ThunderTime), 0
	if s.Raining {
		d.RainLevel = 1
	}
	if s.Thundering {
		d.LightningLevel = 1
	}
	d.CurrentTick = s.CurrentTick
	d.ServerChunkTickRange = s.TickRange
	mode, _ := world.GameModeID(s.DefaultGameMode)
	d.GameType = int32(mode)
	difficulty, _ := world.DifficultyID(s.Difficulty)
	d.Difficulty = int32(difficulty)
}
