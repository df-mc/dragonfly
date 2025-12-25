package session

import (
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// parseEntityMetadata returns an entity metadata object with default values. It is equivalent to setting
// all properties to their default values and disabling all flags.
func (s *Session) parseEntityMetadata(e world.Entity) protocol.EntityMetadata {
	bb := e.H().Type().BBox(e)
	m := protocol.NewEntityMetadata()

	m[protocol.EntityDataKeyWidth] = float32(bb.Width())
	m[protocol.EntityDataKeyHeight] = float32(bb.Height())
	m[protocol.EntityDataKeyEffectColor] = int32(0)
	m[protocol.EntityDataKeyEffectAmbience] = byte(0)
	m[protocol.EntityDataKeyColorIndex] = byte(0)

	m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagHasGravity)
	m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagClimb)
	if g, ok := e.H().Type().(glint); ok && g.Glint() {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagEnchanted)
	}
	if e.H().Type() == entity.LingeringPotionType {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagLingering)
	}
	s.addSpecificMetadata(e, m)
	if ent, ok := e.(*entity.Ent); ok {
		s.addSpecificMetadata(ent.Behaviour(), m)
	}
	return m
}

func (s *Session) addSpecificMetadata(e any, m protocol.EntityMetadata) {
	if sn, ok := e.(sneaker); ok && sn.Sneaking() {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagSneaking)
	}
	if sp, ok := e.(sprinter); ok && sp.Sprinting() {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagSprinting)
	}
	if sw, ok := e.(swimmer); ok && sw.Swimming() {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagSwimming)
	}
	if cr, ok := e.(crawler); ok && cr.Crawling() {
		m.SetFlag(protocol.EntityDataKeyFlagsTwo, protocol.EntityDataFlagCrawling&63)
	}
	if gl, ok := e.(glider); ok && gl.Gliding() {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagGliding)
	}
	if b, ok := e.(breather); ok {
		m[protocol.EntityDataKeyAirSupply] = int16(b.AirSupply().Milliseconds() / 50)
		m[protocol.EntityDataKeyAirSupplyMax] = int16(b.MaxAirSupply().Milliseconds() / 50)
		if b.Breathing() {
			m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagBreathing)
		}
	}
	if i, ok := e.(invisible); ok && i.Invisible() {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagInvisible)
	}
	if i, ok := e.(immobile); ok && i.Immobile() {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagNoAI)
	}
	if o, ok := e.(onFire); ok && o.OnFireDuration() > 0 {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagOnFire)
	}
	if u, ok := e.(using); ok && u.UsingItem() {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagUsingItem)
	}
	if c, ok := e.(arrow); ok && c.Critical() {
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagCritical)
	}
	if g, ok := e.(gameMode); ok {
		if g.GameMode().HasCollision() {
			m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagHasCollision)
		}
	}
	if o, ok := e.(orb); ok {
		m[protocol.EntityDataKeyValue] = int32(o.Experience())
	}
	if f, ok := e.(firework); ok {
		m[protocol.EntityDataKeyDisplayTileRuntimeID] = nbtconv.WriteItem(item.NewStack(f.Firework(), 1), false)
		if o, ok := e.(owned); ok && f.Attached() && o.Owner() != nil {
			m[protocol.EntityDataKeyCustomDisplay] = int64(s.handleRuntimeID(o.Owner()))
		}
	} else if o, ok := e.(owned); ok && o.Owner() != nil {
		m[protocol.EntityDataKeyOwner] = int64(s.handleRuntimeID(o.Owner()))
	}
	if sc, ok := e.(scaled); ok {
		m[protocol.EntityDataKeyScale] = float32(sc.Scale())
	}
	if t, ok := e.(tnt); ok {
		m[protocol.EntityDataKeyFuseTime] = int32(t.Fuse().Milliseconds() / 50)
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagIgnited)
	}
	if n, ok := e.(named); ok {
		m[protocol.EntityDataKeyName] = n.NameTag()
		m[protocol.EntityDataKeyAlwaysShowNameTag] = uint8(1)
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagAlwaysShowName)
		m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagShowName)
	}
	if sc, ok := e.(scoreTag); ok {
		m[protocol.EntityDataKeyScore] = sc.ScoreTag()
	}
	if sl, ok := e.(sleeper); ok {
		if pos, ok := sl.Sleeping(); ok {
			m[protocol.EntityDataKeyBedPosition] = protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])}

			// For some reason there is no such flag in gophertunnel.
			m.SetFlag(protocol.EntityDataKeyPlayerFlags, 1)
		}
	}
	if c, ok := e.(areaEffectCloud); ok {
		m[protocol.EntityDataKeyDataRadius] = float32(c.Radius())

		// We purposely fill these in with invalid values to disable the client-sided shrinking of the cloud.
		m[protocol.EntityDataKeyDataDuration] = int32(math.MaxInt32)
		m[protocol.EntityDataKeyDataChangeOnPickup] = float32(math.SmallestNonzeroFloat32)
		m[protocol.EntityDataKeyDataChangeRate] = float32(math.SmallestNonzeroFloat32)

		colour, am := effect.ResultingColour(c.Effects())
		m[protocol.EntityDataKeyEffectColor] = nbtconv.Int32FromRGBA(colour)
		if am {
			m[protocol.EntityDataKeyEffectAmbience] = byte(1)
		} else {
			m[protocol.EntityDataKeyEffectAmbience] = byte(0)
		}
	}

	if l, ok := e.(living); ok && s.ent.UUID() == l.UUID() {
		deathPos, deathDimension, died := l.DeathPosition()
		if died {
			dim, _ := world.DimensionID(deathDimension)
			m[protocol.EntityDataKeyPlayerLastDeathPosition] = vec64To32(deathPos)
			m[protocol.EntityDataKeyPlayerLastDeathDimension] = int32(dim)
		}
		m[protocol.EntityDataKeyPlayerHasDied] = boolByte(died)
	}
	if p, ok := e.(splash); ok {
		m[protocol.EntityDataKeyAuxValueData] = int16(p.Potion().Uint8())
		if tip := p.Potion().Uint8(); tip > 4 {
			m[protocol.EntityDataKeyCustomDisplay] = tip + 1
		}
	}
	if eff, ok := e.(effectBearer); ok {
		var packedEffects int64

		for _, ef := range eff.Effects() {
			if !ef.ParticlesHidden() {
				id, found := effect.ID(ef.Type())
				if !found {
					continue
				}
				packedEffects = (packedEffects << 7) | int64(id<<1)
				if ef.Ambient() {
					packedEffects |= 1
				}
			}
		}
		m[protocol.EntityDataKeyVisibleMobEffects] = packedEffects
	}
	if v, ok := e.(variable); ok {
		m[protocol.EntityDataKeyVariant] = v.Variant()
	}
	if mv, ok := e.(markVariable); ok {
		m[protocol.EntityDataKeyMarkVariant] = mv.MarkVariant()
	}
}

type sneaker interface {
	Sneaking() bool
}

type sprinter interface {
	Sprinting() bool
}

type swimmer interface {
	Swimming() bool
}

type crawler interface {
	Crawling() bool
}

type glider interface {
	Gliding() bool
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
	Owner() *world.EntityHandle
}

type named interface {
	NameTag() string
}

type scoreTag interface {
	ScoreTag() string
}

type splash interface {
	Potion() potion.Potion
}

type glint interface {
	Glint() bool
}

type areaEffectCloud interface {
	effectBearer
	Radius() float64
}

type onFire interface {
	OnFireDuration() time.Duration
}

type effectBearer interface {
	Effects() []effect.Effect
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

type sleeper interface {
	Sleeping() (cube.Pos, bool)
}

type tnt interface {
	Fuse() time.Duration
}

type living interface {
	UUID() uuid.UUID
	DeathPosition() (mgl64.Vec3, world.Dimension, bool)
}

type variable interface {
	Variant() int32
}

type markVariable interface {
	MarkVariant() int32
}
