package session

// entityMetadata represents a map that holds metadata associated with an entity. The data held in the map
// depends on the entity and varies on a per-entity basis.
type entityMetadata map[uint32]interface{}

// defaultEntityMetadata returns an entity metadata object with default values. It is equivalent to setting
// all properties to their default values and disabling all flags.
func defaultEntityMetadata() entityMetadata {
	m := entityMetadata{}
	m.setFlag(dataKeyFlags, dataFlagAffectedByGravity)
	return m
}

// setFlag sets a flag with a specific index in the int64 stored in the entity metadata map to the value
// passed. It is typically used for entity metadata flags.
func (m entityMetadata) setFlag(key uint32, index uint8) {
	if v, ok := m[key]; !ok {
		m[key] = int64(0) ^ (1 << uint64(index))
	} else {
		m[key] = v.(int64) ^ (1 << uint64(index))
	}
}

const (
	dataKeyFlags = iota
	dataKeyHealth
	dataKeyVariant
	dataKeyColour
	dataKeyNameTag
	dataKeyOwnerRuntimeID
	dataKeyTargetRuntimeID
	dataKeyAir
)

const (
	dataFlagOnFire = iota
	dataFlagSneaking
	dataFlagRiding
	dataFlagSprinting
	dataFlagAction
	dataFlagInvisible
	dataFlagBreathing         = 35
	dataFlagAffectedByGravity = 48
)
