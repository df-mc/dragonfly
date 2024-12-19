package block

// SkullType represents a mob variant of a skull.
type SkullType struct {
	skull
}

// SkeletonSkull returns the skull variant for skeletons.
func SkeletonSkull() SkullType {
	return SkullType{0}
}

// WitherSkeletonSkull returns the skull variant for wither skeletons.
func WitherSkeletonSkull() SkullType {
	return SkullType{1}
}

// ZombieHead returns the skull variant for zombies.
func ZombieHead() SkullType {
	return SkullType{2}
}

// PlayerHead returns the skull variant for players.
func PlayerHead() SkullType {
	return SkullType{3}
}

// CreeperHead returns the skull variant for creepers.
func CreeperHead() SkullType {
	return SkullType{4}
}

// DragonHead returns the skull variant for ender dragons.
func DragonHead() SkullType {
	return SkullType{5}
}

// PiglinHead returns the skull variant for piglins.
func PiglinHead() SkullType {
	return SkullType{6}
}

// SkullTypes returns all variants of skulls.
func SkullTypes() []SkullType {
	return []SkullType{SkeletonSkull(), WitherSkeletonSkull(), ZombieHead(), PlayerHead(), CreeperHead(), DragonHead(), PiglinHead()}
}

type skull uint8

// Uint8 ...
func (s skull) Uint8() uint8 {
	return uint8(s)
}

// Name ...
func (s skull) Name() string {
	switch s {
	case 0:
		return "Skeleton Skull"
	case 1:
		return "Wither Skeleton Skull"
	case 2:
		return "Zombie Head"
	case 3:
		return "Player Head"
	case 4:
		return "Creeper Head"
	case 5:
		return "Dragon Head"
	case 6:
		return "Piglin Head"
	}
	panic("unknown skull type")
}

// String ...
func (s skull) String() string {
	switch s {
	case 0:
		return "skeleton_skull"
	case 1:
		return "wither_skeleton_skull"
	case 2:
		return "zombie_head"
	case 3:
		return "player_head"
	case 4:
		return "creeper_head"
	case 5:
		return "dragon_head"
	case 6:
		return "piglin_head"
	}
	panic("unknown skull type")
}
