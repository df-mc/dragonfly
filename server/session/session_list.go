package session

import (
	"slices"
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
	s.entityMutex.RLock()
	name := s.name
	s.entityMutex.RUnlock()

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
		ActionType: packet.PlayerListActionAdd,
		Entries:    []protocol.PlayerListEntry{playerListEntry(s, runtimeID, name, s.joinSkin)},
	})
}

// UpdateName changes the player list name of a session and resends its entry to
// every session so the new name is shown on the pause screen. The current skin
// is resent with it to avoid reverting a skin changed during the session.
func (l *sessionList) UpdateName(s *Session, name string, sk skin.Skin) {
	l.mu.Lock()
	defer l.mu.Unlock()

	s.entityMutex.Lock()
	s.name = name
	s.entityMutex.Unlock()

	for _, to := range l.s {
		to.entityMutex.RLock()
		runtimeID, ok := to.entityRuntimeIDs[s.ent]
		to.entityMutex.RUnlock()
		if !ok {
			continue
		}
		to.writePacket(&packet.PlayerList{
			ActionType: packet.PlayerListActionAdd,
			Entries:    []protocol.PlayerListEntry{playerListEntry(s, runtimeID, name, sk)},
		})
	}
}

// playerListEntry builds the player list entry of a session, with the name and
// skin passed, as shown to a recipient that holds the runtime ID passed.
func playerListEntry(s *Session, runtimeID uint64, name string, sk skin.Skin) protocol.PlayerListEntry {
	return protocol.PlayerListEntry{
		UUID:           s.ent.UUID(),
		EntityUniqueID: int64(runtimeID),
		Username:       name,
		XUID:           s.conn.IdentityData().XUID,
		Skin:           skinToProtocol(sk),
	}
}

func (l *sessionList) unsendSessionFrom(s, from *Session) {
	from.entityMutex.Lock()
	delete(from.entities, from.entityRuntimeIDs[s.ent])
	delete(from.entityRuntimeIDs, s.ent)
	from.entityMutex.Unlock()

	from.writePacket(&packet.PlayerList{
		ActionType: packet.PlayerListActionRemove,
		Entries:    []protocol.PlayerListEntry{{UUID: s.ent.UUID()}},
	})
}

// skinToProtocol converts a skin to its protocol representation.
func skinToProtocol(s skin.Skin) protocol.Skin {
	var animations []protocol.SkinAnimation
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

	fullID := s.FullID
	if fullID == "" {
		fullID = uuid.New().String()
	}
	return protocol.Skin{
		PlayFabID:                 s.PlayFabID,
		SkinID:                    uuid.New().String(),
		SkinResourcePatch:         s.ModelConfig.Encode(),
		SkinImageWidth:            uint32(s.Bounds().Max.X),
		SkinImageHeight:           uint32(s.Bounds().Max.Y),
		SkinData:                  s.Pix,
		CapeImageWidth:            uint32(s.Cape.Bounds().Max.X),
		CapeImageHeight:           uint32(s.Cape.Bounds().Max.Y),
		CapeData:                  s.Cape.Pix,
		SkinGeometry:              s.Model,
		PersonaSkin:               s.Persona,
		CapeID:                    uuid.New().String(),
		FullID:                    fullID,
		Animations:                animations,
		Trusted:                   true,
		OverrideAppearance:        true,
		GeometryDataEngineVersion: []byte(protocol.CurrentVersion),
	}
}
