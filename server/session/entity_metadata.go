package session

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// entityMetadata represents a map that holds metadata associated with an entity. The data held in the map
// depends on the entity and varies on a per-entity basis.
type entityMetadata map[uint32]any

// parseEntityMetadata returns an entity metadata object with default values. It is equivalent to setting
// all properties to their default values and disabling all flags.
func (s *Session) parseEntityMetadata(e world.Entity) entityMetadata {
	bb := e.BBox()
	m := entityMetadata{
		dataKeyBoundingBoxWidth:  float32(bb.Width()),
		dataKeyBoundingBoxHeight: float32(bb.Height()),
		dataKeyPotionColour:      int32(0),
		dataKeyPotionAmbient:     byte(0),
		dataKeyColour:            byte(0),
		dataKeyFlags:             int64(0),
		dataKeyFlagsExtended:     int64(0),
	}

	m.setFlag(dataKeyFlags, dataFlagAffectedByGravity)
	m.setFlag(dataKeyFlags, dataFlagCanClimb)
	if sn, ok := e.(sneaker); ok && sn.Sneaking() {
		m.setFlag(dataKeyFlags, dataFlagSneaking)
		if b, ok := e.(blocker); ok {
			if _, ok = b.Blocking(); ok {
				m.setFlag(dataKeyFlagsExtended, dataFlagBlocking)
			}
		}
	}
	if sp, ok := e.(sprinter); ok && sp.Sprinting() {
		m.setFlag(dataKeyFlags, dataFlagSprinting)
	}
	if sw, ok := e.(swimmer); ok && sw.Swimming() {
		m.setFlag(dataKeyFlags, dataFlagSwimming)
	}
	if gl, ok := e.(glider); ok && gl.Gliding() {
		m.setFlag(dataKeyFlags, dataFlagGliding)
	}
	if b, ok := e.(breather); ok {
		m[dataKeyAir] = int16(b.AirSupply().Milliseconds() / 50)
		m[dataKeyMaxAir] = int16(b.MaxAirSupply().Milliseconds() / 50)
		if b.Breathing() {
			m.setFlag(dataKeyFlags, dataFlagBreathing)
		}
	}
	if i, ok := e.(invisible); ok && i.Invisible() {
		m.setFlag(dataKeyFlags, dataFlagInvisible)
	}
	if i, ok := e.(immobile); ok && i.Immobile() {
		m.setFlag(dataKeyFlags, dataFlagNoAI)
	}
	if o, ok := e.(onFire); ok && o.OnFireDuration() > 0 {
		m.setFlag(dataKeyFlags, dataFlagOnFire)
	}
	if u, ok := e.(using); ok && u.UsingItem() {
		m.setFlag(dataKeyFlags, dataFlagUsingItem)
	}
	if c, ok := e.(arrow); ok && c.Critical() {
		m.setFlag(dataKeyFlags, dataFlagCritical)
	}
	if g, ok := e.(gameMode); ok && g.GameMode().HasCollision() {
		m.setFlag(dataKeyFlags, dataFlagHasCollision)
	}
	if o, ok := e.(orb); ok {
		m[dataKeyExperienceValue] = int32(o.Experience())
	}
	if f, ok := e.(firework); ok {
		m[dataKeyFireworkItem] = nbtconv.WriteItem(item.NewStack(f.Firework(), 1), false)
		if o, ok := e.(owned); ok && f.Attached() {
			m[dataKeyCustomDisplay] = int64(s.entityRuntimeID(o.Owner()))
		}
	} else if o, ok := e.(owned); ok {
		m[dataKeyOwnerRuntimeID] = int64(s.entityRuntimeID(o.Owner()))
	}
	if sc, ok := e.(scaled); ok {
		m[dataKeyScale] = float32(sc.Scale())
	}
	if t, ok := e.(tnt); ok {
		m[dataKeyFuseLength] = int32(t.Fuse().Milliseconds() / 50)
		m.setFlag(dataKeyFlags, dataFlagIgnited)
	}
	if n, ok := e.(named); ok {
		m[dataKeyNameTag] = n.NameTag()
		m[dataKeyAlwaysShowNameTag] = uint8(1)
		m.setFlag(dataKeyFlags, dataFlagAlwaysShowNameTag)
		m.setFlag(dataKeyFlags, dataFlagCanShowNameTag)
	}
	if sc, ok := e.(scoreTag); ok {
		m[dataKeyScoreTag] = sc.ScoreTag()
	}
	if c, ok := e.(areaEffectCloud); ok {
		radius, radiusOnUse, radiusGrowth := c.Radius()
		colour, am := effect.ResultingColour(c.Effects())
		m[dataKeyAreaEffectCloudDuration] = int32(c.Duration().Milliseconds() / 50)
		m[dataKeyAreaEffectCloudRadius] = float32(radius)
		m[dataKeyAreaEffectCloudRadiusChangeOnPickup] = float32(radiusOnUse)
		m[dataKeyAreaEffectCloudRadiusPerTick] = float32(radiusGrowth)
		m[dataKeyPotionColour] = nbtconv.Int32FromRGBA(colour)
		if am {
			m[dataKeyPotionAmbient] = byte(1)
		} else {
			m[dataKeyPotionAmbient] = byte(0)
		}
	}
	if p, ok := e.(splash); ok {
		m[dataKeyPotionAuxValue] = int16(p.Type().Uint8())
	}
	if g, ok := e.(glint); ok && g.Glint() {
		m.setFlag(dataKeyFlags, dataFlagEnchanted)
	}
	if l, ok := e.(lingers); ok && l.Lingers() {
		m.setFlag(dataKeyFlags, dataFlagLinger)
	}
	if t, ok := e.(tipped); ok {
		if tip := t.Tip().Uint8(); tip > 4 {
			m[dataKeyCustomDisplay] = tip + 1
		}
	}
	if eff, ok := e.(effectBearer); ok && len(eff.Effects()) > 0 {
		visibleEffects := make([]effect.Effect, 0, len(eff.Effects()))
		for _, ef := range eff.Effects() {
			if !ef.ParticlesHidden() {
				visibleEffects = append(visibleEffects, ef)
			}
		}
		if len(visibleEffects) > 0 {
			colour, am := effect.ResultingColour(visibleEffects)
			m[dataKeyPotionColour] = nbtconv.Int32FromRGBA(colour)
			if am {
				m[dataKeyPotionAmbient] = byte(1)
			} else {
				m[dataKeyPotionAmbient] = byte(0)
			}
		}
	}
	return m
}

// setFlag sets a flag with a specific index in the int64 stored in the entity metadata map to the value
// passed. It is typically used for entity metadata flags.
func (m entityMetadata) setFlag(key uint32, index uint8) {
	actualIndex := index % 64
	if v, ok := m[key]; ok {
		m[key] = v.(int64) ^ (1 << uint64(actualIndex))
	}
}

//noinspection GoUnusedConst
const (
	dataKeyFlags = iota
	dataKeyHealth
	dataKeyVariant
	dataKeyColour
	dataKeyNameTag
	dataKeyOwnerRuntimeID
	dataKeyTargetRuntimeID
	dataKeyAir
	dataKeyPotionColour
	dataKeyPotionAmbient
	dataKeyExperienceValue                     = 15
	dataKeyFireworkItem                        = 16
	dataKeyCustomDisplay                       = 18
	dataKeyPotionAuxValue                      = 36
	dataKeyScale                               = 38
	dataKeyMaxAir                              = 42
	dataKeyBoundingBoxWidth                    = 53
	dataKeyBoundingBoxHeight                   = 54
	dataKeyFuseLength                          = 55
	dataKeyAreaEffectCloudRadius               = 61
	dataKeyAreaEffectCloudParticleID           = 63
	dataKeyAlwaysShowNameTag                   = 81
	dataKeyScoreTag                            = 84
	dataKeyFlagsExtended                       = 92
	dataKeyAreaEffectCloudDuration             = 95
	dataKeyAreaEffectCloudSpawnTime            = 96
	dataKeyAreaEffectCloudRadiusPerTick        = 97
	dataKeyAreaEffectCloudRadiusChangeOnPickup = 98
	dataKeyAreaEffectCloudPickupCount          = 99
)

//noinspection GoUnusedConst
const (
	dataFlagOnFire = iota
	dataFlagSneaking
	dataFlagRiding
	dataFlagSprinting
	dataFlagUsingItem
	dataFlagInvisible
	dataFlagIgnited            = 10
	dataFlagCritical           = 13
	dataFlagCanShowNameTag     = 14
	dataFlagAlwaysShowNameTag  = 15
	dataFlagNoAI               = 16
	dataFlagCanClimb           = 19
	dataFlagGliding            = 32
	dataFlagBreathing          = 35
	dataFlagLinger             = 46
	dataFlagHasCollision       = 47
	dataFlagAffectedByGravity  = 48
	dataFlagEnchanted          = 51
	dataFlagSwimming           = 56
	dataFlagBlocking           = 71
	dataFlagTransitionBlocking = 72
)

type sneaker interface {
	Sneaking() bool
}

type sprinter interface {
	Sprinting() bool
}

type swimmer interface {
	Swimming() bool
}

type glider interface {
	Gliding() bool
}

type blocker interface {
	Blocking() (bool, bool)
}

type breather interface {
	Breathing() bool
	AirSupply() time.Duration
	MaxAirSupply() time.Duration
}

type immobile interface {
	Immobile() bool
}

type invisible interface {
	Invisible() bool
}

type scaled interface {
	Scale() float64
}

type owned interface {
	Owner() world.Entity
}

type named interface {
	NameTag() string
}

type scoreTag interface {
	ScoreTag() string
}

type splash interface {
	Type() potion.Potion
}

type glint interface {
	Glint() bool
}

type lingers interface {
	Lingers() bool
}

type areaEffectCloud interface {
	effectBearer
	Duration() time.Duration
	Radius() (radius, radiusOnUse, radiusGrowth float64)
}

type onFire interface {
	OnFireDuration() time.Duration
}

type effectBearer interface {
	Effects() []effect.Effect
}

type tipped interface {
	Tip() potion.Potion
}

type using interface {
	UsingItem() bool
}

type arrow interface {
	Critical() bool
}

type orb interface {
	Experience() int
}

type firework interface {
	Firework() item.Firework
	Attached() bool
}

type gameMode interface {
	GameMode() world.GameMode
}

type tnt interface {
	Fuse() time.Duration
}
