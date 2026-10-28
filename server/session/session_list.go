package session

import (
	"image/color"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

var sessions = new(sessionList)

type sessionList struct {
	mu sync.Mutex
	s  []*Session
}

func (l *sessionList) Add(s *Session) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, other := range l.s {
		// Show all sessions to the new session and the new session to all
		// existing sessions.
		l.sendSessionTo(s, other)
		l.sendSessionTo(other, s)
	}
	// Show the new session to itself.
	l.sendSessionTo(s, s)
	l.s = append(l.s, s)
}

func (l *sessionList) Remove(s *Session, entity world.Entity) {
	l.mu.Lock()
	removedFrom := slices.Clone(l.s)
	for _, other := range l.s {
		l.unsendSessionFrom(s, other)
	}
	l.s = sliceutil.DeleteVal(l.s, s)
	l.mu.Unlock()

	if entity == nil {
		return
	}
	for _, other := range removedFrom {
		if other.viewLayer != nil {
			other.viewLayer.Remove(entity)
		}
	}
}

func (l *sessionList) Lookup(id uuid.UUID) (*Session, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if index := slices.IndexFunc(l.s, func(session *Session) bool {
		return session.ent.UUID() == id
	}); index != -1 {
		return l.s[index], true
	}
	return nil, false
}

func (l *sessionList) sendSessionTo(s, to *Session) {
	runtimeID := uint64(selfEntityRuntimeID)

	to.entityMutex.Lock()
	if s != to {
		to.currentEntityRuntimeID += 1
		runtimeID = to.currentEntityRuntimeID
	}
	to.entityRuntimeIDs[s.ent] = runtimeID
	to.entities[runtimeID] = s.ent
	to.entityMutex.Unlock()

	to.writePacket(&packet.PlayerList{
		Entries: []protocol.PlayerListEntry{{
			ActionType:     protocol.PlayerListActionAdd,
			UUID:           s.ent.UUID(),
			EntityUniqueID: int64(runtimeID),
			Username:       s.conn.IdentityData().DisplayName,
			XUID:           s.conn.IdentityData().XUID,
			BuildPlatform:  int32(protocol.DeviceUnknown),
			Skin:           skinToProtocol(s.joinSkin),
		}},
	})
}

func (l *sessionList) unsendSessionFrom(s, from *Session) {
	from.entityMutex.Lock()
	delete(from.entities, from.entityRuntimeIDs[s.ent])
	delete(from.entityRuntimeIDs, s.ent)
	from.entityMutex.Unlock()

	from.writePacket(&packet.PlayerList{
		Entries: []protocol.PlayerListEntry{{
			ActionType: protocol.PlayerListActionRemove,
			UUID:       s.ent.UUID(),
		}},
	})
}

// skinToProtocol converts a skin to its protocol representation. Persona pieces, tints and animation data are
// carried through unchanged, so receiving clients can build persona skins.
func skinToProtocol(s skin.Skin) protocol.Skin {
	animations := make([]protocol.SkinAnimation, 0, len(s.Animations))
	for _, animation := range s.Animations {
		protocolAnim := protocol.SkinAnimation{
			ImageWidth:  uint32(animation.Bounds().Max.X),
			ImageHeight: uint32(animation.Bounds().Max.Y),
			ImageData:   animation.Pix,
			FrameCount:  float32(animation.FrameCount),
		}
		switch animation.Type() {
		case skin.AnimationHead:
			protocolAnim.AnimationType = protocol.SkinAnimationHead
		case skin.AnimationBody32x32:
			protocolAnim.AnimationType = protocol.SkinAnimationBody32x32
		case skin.AnimationBody128x128:
			protocolAnim.AnimationType = protocol.SkinAnimationBody128x128
		}
		protocolAnim.ExpressionType = uint32(animation.AnimationExpression)
		animations = append(animations, protocolAnim)
	}

	pieces := make([]protocol.PersonaPiece, 0, len(s.PersonaPieces))
	for _, piece := range s.PersonaPieces {
		packID, err := uuid.Parse(piece.PackID)
		if err != nil {
			// A malformed pack ID is not worth dropping the whole piece over: the client keys the piece off
			// PieceID, and a nil pack ID is what it receives for a piece with no pack anyway.
			packID = uuid.Nil
		}
		pieces = append(pieces, protocol.PersonaPiece{
			PieceID:   piece.PieceID,
			PieceType: personaPieceType(piece.PieceType),
			PackID:    packID,
			Default:   piece.Default,
			ProductID: piece.ProductID,
		})
	}

	tints := make([]protocol.PersonaPieceTintColour, 0, len(s.PieceTintColours))
	for _, tint := range s.PieceTintColours {
		t := protocol.PersonaPieceTintColour{PieceType: tint.PieceType}
		for i, colour := range tint.Colours {
			t.Colours[i] = parseARGB(colour)
		}
		tints = append(tints, t)
	}

	fullID := s.FullID
	if fullID == "" {
		fullID = uuid.New().String()
	}
	skinID := s.SkinID
	if skinID == "" {
		skinID = fullID
	}
	model := s.Model
	if len(model) == 0 {
		model = []byte("{}")
	}
	geometryVersion := s.GeometryVersion
	if geometryVersion == "" {
		geometryVersion = protocol.CurrentVersion
	}
	return protocol.Skin{
		PlayFabID:                s.PlayFabID,
		SkinID:                   skinID,
		SkinResourcePatch:        s.ModelConfig.Encode(),
		SkinImageWidth:           uint32(s.Bounds().Max.X),
		SkinImageHeight:          uint32(s.Bounds().Max.Y),
		SkinData:                 s.Pix,
		CapeImageWidth:           uint32(s.Cape.Bounds().Max.X),
		CapeImageHeight:          uint32(s.Cape.Bounds().Max.Y),
		CapeData:                 s.Cape.Pix,
		SkinGeometry:             model,
		AnimationData:            []byte(s.AnimationData),
		ArmSize:                  armSizeToProtocol(s.ArmSize),
		SkinColour:               parseARGB(s.SkinColour),
		PremiumSkin:              s.Premium,
		PersonaSkin:              s.Persona,
		PersonaCapeOnClassicSkin: s.CapeOnClassic,
		PrimaryUser:              s.PrimaryUser,
		PersonaPieces:            pieces,
		PieceTintColours:         tints,
		CapeID:                   s.CapeID,
		FullID:                   fullID,
		Animations:               animations,
		Trusted:                  true,
		// OverrideAppearance always true: this tells the receiving client to use the skin data sent
		// here rather than falling back to any appearance it might otherwise guess for the player.
		OverrideAppearance:        true,
		GeometryDataEngineVersion: []byte(geometryVersion),
	}
}

// parseARGB reads a colour written as hex with a leading '#', as both login.ClientData.SkinColour and the
// persona piece tints use. Anything unparsable becomes a fully transparent zero colour, which is what an
// unused tint slot ("#0") means anyway.
func parseARGB(s string) color.RGBA {
	v, err := strconv.ParseUint(strings.TrimPrefix(s, "#"), 16, 32)
	if err != nil {
		return color.RGBA{}
	}
	return color.RGBA{A: uint8(v >> 24), R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v)}
}

// armSizeToProtocol maps the real client's ArmSize string (login.ClientData.ArmSize, "wide" or
// "slim") to the protocol.ArmSize* constant a re-broadcast skin packet needs. Defaults to
// ArmSizeWide (the more common/vanilla-default value) for anything else, including an empty
// string from an older client that never sent one, rather than silently defaulting to slim.
func armSizeToProtocol(armSize string) uint8 {
	if armSize == "slim" {
		return protocol.ArmSizeSlim
	}
	return protocol.ArmSizeWide
}
