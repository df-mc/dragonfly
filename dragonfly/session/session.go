package session

import (
	"bytes"
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/chat"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
	"time"
)

// Session handles incoming packets from connections and sends outgoing packets by providing a thin layer
// of abstraction over direct packets. A Session basically 'controls' an entity.
type Session struct {
	log *logrus.Logger

	c                  Controllable
	world              *world.World
	controllableClosed atomic.Value
	conn               *minecraft.Conn

	cmdOrigin     protocol.CommandOrigin
	scoreboardObj atomic.Value

	chunkBuf       *bytes.Buffer
	chunkLoader    atomic.Value
	chunkRadius    int32
	maxChunkRadius int32

	// currentEntityRuntimeID holds the runtime ID assigned to the last entity. It is incremented for every
	// entity spawned to the session.
	currentEntityRuntimeID uint64
	entityMutex            sync.RWMutex
	// entityRuntimeIDs holds a list of all runtime IDs of entities spawned to the session.
	entityRuntimeIDs map[world.Entity]uint64

	// onStop is called when the session is stopped. The controllable passed is the controllable that the
	// session controls.
	onStop func(controllable Controllable)
}

// Nop represents a no-operation session. It does not do anything when sending a packet to it.
var Nop = &Session{}

// session is a slice of all open sessions. It is protected by the sessionMutex, which must be locked whenever
// accessing the value.
var sessions []*Session
var sessionMutex sync.Mutex

// New returns a new session using a controllable entity. The session will control this entity using the
// packets that it receives.
// New takes the connection from which to accept packets. It will start handling these packets after a call to
// Session.Start().
func New(c Controllable, conn *minecraft.Conn, w *world.World, maxChunkRadius int, log *logrus.Logger) *Session {
	s := &Session{
		c:              c,
		conn:           conn,
		log:            log,
		chunkBuf:       bytes.NewBuffer(make([]byte, 0, 4096)),
		world:          w,
		chunkRadius:    int32(maxChunkRadius / 2),
		maxChunkRadius: int32(maxChunkRadius),
		entityRuntimeIDs: map[world.Entity]uint64{
			// We initialise the runtime ID of the controllable of the session. It will always have runtime ID
			// 1, because we treat entity runtime IDs as session-local.
			c: 1,
		},
		currentEntityRuntimeID: 1,
	}
	s.chunkLoader.Store(world.NewLoader(maxChunkRadius/2, w, s))
	s.scoreboardObj.Store("")
	s.controllableClosed.Store(false)
	return s
}

// Start makes the session start handling incoming packets from the client and initialises the controllable of
// the session in the world.
// The function passed will be called when the session stops running.
func (s *Session) Start(onStop func(controllable Controllable)) {
	s.onStop = onStop

	go s.handlePackets()
	s.SendAvailableCommands()

	s.initPlayerList()
	s.world.AddEntity(s.c)

	yellow := text.Yellow()
	chat.Global.Println(yellow(s.conn.IdentityData().DisplayName, "has joined the game"))
}

// Close closes the session, which in turn closes the controllable and the connection that the session
// manages.
func (s *Session) Close() error {
	_ = s.c.Close()
	_ = s.conn.Close()
	_ = s.chunkLoader.Load().(*world.Loader).Close()
	s.world.RemoveEntity(s.c)

	yellow := text.Yellow()
	chat.Global.Println(yellow(s.conn.IdentityData().DisplayName, "has left the game"))

	// This should always be called last due to the timing of the removal of entity runtime IDs.
	s.closePlayerList()

	if s.onStop != nil {
		s.onStop(s.c)
	}
	return nil
}

// handlePackets continuously handles incoming packets from the connection. It processes them accordingly.
// Once the connection is closed, handlePackets will return.
func (s *Session) handlePackets() {
	c := make(chan struct{})
	defer func() {
		c <- struct{}{}
		_ = s.Close()
	}()
	go s.sendChunks(c)
	for {
		pk, err := s.conn.ReadPacket()
		if err != nil {
			return
		}
		if s.controllableClosed.Load().(bool) {
			// The controllable closed itself, so we need to stop handling packets and close the session.
			return
		}
		if err := s.handlePacket(pk); err != nil {
			// An error occurred during the handling of a packet. Print the error and stop handling any more
			// packets.
			s.log.Errorf("error processing packet from %v: %v\n", s.conn.RemoteAddr(), err)
			return
		}
	}
}

// sendChunks continuously sends chunks to the player, until a value is sent to the closeChan passed.
func (s *Session) sendChunks(closeChan <-chan struct{}) {
	t := time.NewTicker(time.Second / 20)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if err := s.chunkLoader.Load().(*world.Loader).Load(4); err != nil {
				s.log.Errorf("error loading chunk: %v", err)
				continue
			}
		case <-closeChan:
			return
		}
	}
}

// handlePacket handles an incoming packet, processing it accordingly. If the packet had invalid data or was
// otherwise not valid in its context, an error is returned.
func (s *Session) handlePacket(pk packet.Packet) error {
	switch pk := pk.(type) {
	case *packet.Text:
		return s.handleText(pk)
	case *packet.CommandRequest:
		return s.handleCommandRequest(pk)
	case *packet.MovePlayer:
		return s.handleMovePlayer(pk)
	case *packet.RequestChunkRadius:
		return s.handleRequestChunkRadius(pk)
	default:
		s.log.Debugf("unhandled packet %T%v from %v\n", pk, fmt.Sprintf("%+v", pk)[1:], s.conn.RemoteAddr())
	}
	return nil
}

// writePacket writes a packet to the session's connection if it is not Nop.
func (s *Session) writePacket(pk packet.Packet) {
	if s == Nop {
		return
	}
	_ = s.conn.WritePacket(pk)
}

// initPlayerList initialises the player list of the session and sends the session itself to all other
// sessions currently open.
func (s *Session) initPlayerList() {
	sessionMutex.Lock()
	sessions = append(sessions, s)
	for _, session := range sessions {
		// Add the player of the session to all sessions currently open, and add the players of all sessions
		// currently open to the player list of the new session.
		session.addToPlayerList(s)
		s.addToPlayerList(session)
	}
	sessionMutex.Unlock()
}

// closePlayerList closes the player list of the session and removes the session from the player list of all
// other sessions.
func (s *Session) closePlayerList() {
	sessionMutex.Lock()
	for i, session := range sessions {
		if session == s {
			// Remove the session from the slice.
			sessions = append(sessions[:i], sessions[i+1:]...)
			continue
		}
		// Remove the player of the session from the player list of all other sessions.
		session.removeFromPlayerList(s)
	}
	sessionMutex.Unlock()
}

// addToPlayerList adds the player of a session to the player list of this session. It will be shown in the
// in-game pause menu screen.
func (s *Session) addToPlayerList(session *Session) {
	c := session.c

	s.entityMutex.Lock()
	runtimeID := atomic.AddUint64(&s.currentEntityRuntimeID, 1)
	s.entityRuntimeIDs[c] = runtimeID
	s.entityMutex.Unlock()

	s.writePacket(&packet.PlayerList{
		ActionType: packet.PlayerListActionAdd,
		Entries: []protocol.PlayerListEntry{{
			UUID:             c.UUID(),
			EntityUniqueID:   int64(runtimeID),
			Username:         c.Name(),
			SkinID:           c.Skin().ID,
			SkinData:         c.Skin().Pix,
			CapeData:         c.Skin().Cape.Pix,
			SkinGeometryName: c.Skin().ModelName,
			SkinGeometry:     c.Skin().Model,
			XUID:             c.XUID(),
		}},
	})
}

// removeFromPlayerList removes the player of a session from the player list of this session. It will no
// longer be shown in the in-game pause menu screen.
func (s *Session) removeFromPlayerList(session *Session) {
	c := session.c

	s.entityMutex.Lock()
	delete(s.entityRuntimeIDs, c)
	s.entityMutex.Unlock()

	s.writePacket(&packet.PlayerList{
		ActionType: packet.PlayerListActionRemove,
		Entries: []protocol.PlayerListEntry{{
			UUID: c.UUID(),
		}},
	})
}
