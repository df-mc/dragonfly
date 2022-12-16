package mcdb

// data holds a collection of data that specify a range of settings of the world. These settings usually
// alter the way that players interact with the world.
// The data held here is usually saved in a level.dat file of the world.
// noinspection SpellCheckingInspection
type data struct {
	BaseGameVersion                string `nbt:"baseGameVersion"`
	BiomeOverride                  string
	ConfirmedPlatformLockedContent bool
	CenterMapsToOrigin             bool
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
}
