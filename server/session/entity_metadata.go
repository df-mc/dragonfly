package session

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
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
	}

	m.setFlag(dataKeyFlags, dataFlagAffectedByGravity)
	m.setFlag(dataKeyFlags, dataFlagCanClimb)
	if sn, ok := e.(sneaker); ok && sn.Sneaking() {
		m.setFlag(dataKeyFlags, dataFlagSneaking)
	}
	if sp, ok := e.(sprinter); ok && sp.Sprinting() {
		m.setFlag(dataKeyFlags, dataFlagSprinting)
	}
	if sw, ok := e.(swimmer); ok && sw.Swimming() {
		m.setFlag(dataKeyFlags, dataFlagSwimming)
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
	if sc, ok := e.(scaled); ok {
		m[dataKeyScale] = float32(sc.Scale())
	}
	if o, ok := e.(owned); ok {
		m[dataKeyOwnerRuntimeID] = int64(s.entityRuntimeID(o.Owner()))
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
	if sp, ok := e.(splash); ok {
		pot := sp.Type()
		m[dataKeyPotionAuxValue] = int16(pot.Uint8())
		if len(pot.Effects()) > 0 {
			m.setFlag(dataKeyFlags, dataFlagEnchanted)
		}
	}
	if t, ok := e.(tipped); ok {
		if tip := t.Tip().Uint8(); tip > 4 {
			m[dataKeyCustomDisplay] = tip + 1
		}
	}
	if eff, ok := e.(effectBearer); ok && len(eff.Effects()) > 0 {
		colour, am := effect.ResultingColour(eff.Effects())
		if (colour != color.RGBA{}) {
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
	if v, ok := m[key]; !ok {
		m[key] = int64(0) ^ (1 << uint64(index))
	} else {
		m[key] = v.(int64) ^ (1 << uint64(index))
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
	dataKeyExperienceValue   = 15
	dataKeyCustomDisplay     = 18
	dataKeyPotionAuxValue    = 36
	dataKeyScale             = 38
	dataKeyMaxAir            = 42
	dataKeyBoundingBoxWidth  = 53
	dataKeyBoundingBoxHeight = 54
	dataKeyAlwaysShowNameTag = 81
	dataKeyScoreTag          = 84
)

//noinspection GoUnusedConst
const (
	dataFlagOnFire = iota
	dataFlagSneaking
	dataFlagRiding
	dataFlagSprinting
	dataFlagUsingItem
	dataFlagInvisible
	dataFlagCritical          = 13
	dataFlagCanShowNameTag    = 14
	dataFlagAlwaysShowNameTag = 15
	dataFlagNoAI              = 16
	dataFlagCanClimb          = 19
	dataFlagBreathing         = 35
	dataFlagHasCollision      = 47
	dataFlagAffectedByGravity = 48
	dataFlagEnchanted         = 51
	dataFlagSwimming          = 56
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

type gameMode interface {
	GameMode() world.GameMode
}
