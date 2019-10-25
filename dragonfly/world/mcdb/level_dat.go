package mcdb

// data holds a collection of data that specify a range of settings of the world. These settings usually
// alter the way that players interact with the world.
// The data held here is usually saved in a level.dat file of the world.
type data struct {
	ConfirmedPlatformLockedContent uint8
	CenterMapsToOrigin             uint8
	Difficulty                     int32
	FlatWorldLayers                string
	ForceGameType                  uint8
	GameType                       int32
	Generator                      int32
	InventoryVersion               string
	LANBroadcast                   uint8
	LANBroadcastIntent             uint8
	LastPlayed                     int64
	LevelName                      string
	LimitedWorldOriginX            int32
	LimitedWorldOriginY            int32
	LimitedWorldOriginZ            int32
	MinimumCompatibleClientVersion []int32
	MultiPlayerGame                uint8 `nbt:"MultiplayerGame"`
	MultiPlayerGameIntent          uint8 `nbt:"MultiplayerGameIntent"`
	NetherScale                    int32
	NetworkVersion                 int32
	Platform                       int32
	PlatformBroadcastIntent        int32
	RandomSeed                     int64
	SpawnX, SpawnY, SpawnZ         int32
	SpawnV1Villagers               uint8
	StorageVersion                 int32
	Time                           int64
	XBLBroadcast                   uint8
	XBLBroadcastIntent             int32
	XBLBroadcastMode               int32
	Abilities                      struct {
		AttackMobs             uint8   `nbt:"attackmobs"`
		AttackPlayers          uint8   `nbt:"attackplayers"`
		Build                  uint8   `nbt:"build"`
		Mine                   uint8   `nbt:"mine"`
		DoorsAndSwitches       uint8   `nbt:"doorsandswitches"`
		FlySpeed               float32 `nbt:"flySpeed"`
		Flying                 uint8   `nbt:"flying"`
		InstantBuild           uint8   `nbt:"instabuild"`
		Invulnerable           uint8   `nbt:"invulnerable"`
		Lightning              uint8   `nbt:"lightning"`
		MayFly                 uint8   `nbt:"mayfly"`
		OP                     uint8   `nbt:"op"`
		OpenContainers         uint8   `nbt:"opencontainers"`
		PermissionsLevel       int32   `nbt:"permissionsLevel"`
		PlayerPermissionsLevel int32   `nbt:"playerPermissionsLevel"`
		Teleport               uint8   `nbt:"teleport"`
		WalkSpeed              float32 `nbt:"walkSpeed"`
	} `nbt:"abilities"`
	BonusChestEnabled              uint8   `nbt:"bonusChestEnabled"`
	BonusChestSpawned              uint8   `nbt:"bonusChestSpawned"`
	CommandBlockOutput             uint8   `nbt:"commandblockoutput"`
	CommandBlocksEnabled           uint8   `nbt:"commandblocksenabled"`
	CommandsEnabled                uint8   `nbt:"commandsEnabled"`
	CurrentTick                    int64   `nbt:"currentTick"`
	DoDayLightCycle                uint8   `nbt:"dodaylightcycle"`
	DoEntityDrops                  uint8   `nbt:"doentitydrops"`
	DoFireTick                     uint8   `nbt:"dofiretick"`
	DoImmediateRespawn             uint8   `nbt:"doimmediaterespawn"`
	DoInsomnia                     uint8   `nbt:"doinsomnia"`
	DoMobLoot                      uint8   `nbt:"domobloot"`
	DoMobSpawning                  uint8   `nbt:"domobspawning"`
	DoTileDrops                    uint8   `nbt:"dotiledrops"`
	DoWeatherCycle                 uint8   `nbt:"doweathercycle"`
	DrowningDamage                 uint8   `nbt:"drowningdamage"`
	EduLevel                       uint8   `nbt:"eduLevel"`
	EducationFeaturesEnabled       uint8   `nbt:"educationFeaturesEnabled"`
	ExperimentalGamePlay           uint8   `nbt:"experimentalgameplay"`
	FallDamage                     uint8   `nbt:"falldamage"`
	FireDamage                     uint8   `nbt:"firedamage"`
	FunctionCommandLimit           int32   `nbt:"functioncommandlimit"`
	HasBeenLoadedInCreative        uint8   `nbt:"hasBeenLoadedInCreative"`
	HasLockedBehaviourPack         uint8   `nbt:"hasLockedBehaviorPack"`
	HasLockedResourcePack          uint8   `nbt:"hasLockedResourcePack"`
	ImmutableWorld                 uint8   `nbt:"immutableWorld"`
	IsFromLockedTemplate           uint8   `nbt:"isFromLockedTemplate"`
	IsFromWorldTemplate            uint8   `nbt:"isFromWorldTemplate"`
	IsWorldTemplateOptionLocked    uint8   `nbt:"isWorldTemplateOptionLocked"`
	KeepInventory                  uint8   `nbt:"keepinventory"`
	LastOpenedWithVersion          []int32 `nbt:"lastOpenedWithVersion"`
	LightningLevel                 float32 `nbt:"lightningLevel"`
	LightningTime                  int32   `nbt:"lightningTime"`
	MaxCommandChainLength          int32   `nbt:"maxcommandchainlength"`
	MobGriefing                    uint8   `nbt:"mobgriefing"`
	NaturalRegeneration            uint8   `nbt:"naturalregeneration"`
	PRID                           string  `nbt:"prid"`
	PVP                            uint8   `nbt:"pvp"`
	RainLevel                      float32 `nbt:"rainLevel"`
	RainTime                       int32   `nbt:"rainTime"`
	RandomTickSpeed                int32   `nbt:"randomtickspeed"`
	RequiresCopiedPackRemovalCheck uint8   `nbt:"requiresCopiedPackRemovalCheck"`
	SendCommandFeedback            uint8   `nbt:"sendcommandfeedback"`
	ServerChunkTickRange           int32   `nbt:"serverChunkTickRange"`
	ShowCoordinates                uint8   `nbt:"showcoordinates"`
	ShowDeathMessages              uint8   `nbt:"showdeathmessages"`
	SpawnMobs                      uint8   `nbt:"spawnMobs"`
	SpawnRadius                    int32   `nbt:"spawnradius"`
	StartWithMapEnabled            uint8   `nbt:"startWithMapEnabled"`
	TexturePacksRequired           uint8   `nbt:"texturePacksRequired"`
	TNTExplodes                    uint8   `nbt:"tntexplodes"`
	UseMSAGamerTagsOnly            uint8   `nbt:"useMsaGamertagsOnly"`
	WorldStartCount                int64   `nbt:"worldStartCount"`
}
