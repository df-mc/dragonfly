package mcdb

//lint:file-ignore U1000 Unused unexported constants are present for future code using these.

// Keys on a per-sub chunk basis. These are prefixed by the chunk coordinates and subchunk ID.
const (
	keySubChunkData = '/' // 2f
)

// Keys on a per-chunk basis. These are prefixed by only the chunk coordinates.
const (
	// keyVersion holds a single byte of data with the version of the chunk.
	keyVersion = ',' // 2c
	// keyVersionOld was replaced by keyVersion. It is still used by vanilla to check compatibility, but vanilla no
	// longer writes this tag.
	keyVersionOld = 'v' // 76
	// keyBlockEntities holds n amount of NBT compound tags appended to each other (not a TAG_List, just appended). The
	// compound tags contain the position of the block entities.
	keyBlockEntities = '1' // 31
	// keyEntities holds n amount of NBT compound tags appended to each other (not a TAG_List, just appended). The
	// compound tags contain the position of the entities.
	keyEntities = '2' // 32
	// keyFinalisation contains a single LE int32 that indicates the state of generation of the chunk. If 0, the chunk
	// needs to be ticked. If 1, the chunk needs to be populated and if 2 (which is the state generally found in world
	// saves from vanilla), the chunk is fully finalised.
	keyFinalisation = '6' // 36
	// key3DData holds 3-dimensional biomes for the entire chunk.
	key3DData = '+' // 2b
	// key2DData is no longer used in worlds with world height change. It was replaced by key3DData in newer worlds
	// which has 3-dimensional biomes.
	key2DData = '-' // 2d
	// keyChecksum holds a list of checksums of some sort. It's not clear of what data this checksum is composed or what
	// these checksums are used for.
	keyChecksums = ';' // 3b
)

// Keys on a per-world basis. These are found only once in a leveldb world save.
const (
	keyAutonomousEntities = "AutonomousEntities"
	keyOverworld          = "Overworld"
	keyMobEvents          = "mobevents"
	keyBiomeData          = "BiomeData"
	keyScoreboard         = "scoreboard"
	keyLocalPlayer        = "~local_player"
)
