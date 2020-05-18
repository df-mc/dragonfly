package session

import (
	"bytes"
	"errors"
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/inventory"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/player/chat"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/player/form"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
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

	c        Controllable
	conn     *minecraft.Conn
	handlers map[uint32]packetHandler

	// onStop is called when the session is stopped. The controllable passed is the controllable that the
	// session controls.
	onStop func(controllable Controllable)

	scoreboardObj atomic.Value

	chunkBuf                    *bytes.Buffer
	chunkLoader                 *world.Loader
	chunkRadius, maxChunkRadius int32

	// currentEntityRuntimeID holds the runtime ID assigned to the last entity. It is incremented for every
	// entity spawned to the session.
	currentEntityRuntimeID uint64
	entityMutex            sync.RWMutex
	// entityRuntimeIDs holds a list of all runtime IDs of entities spawned to the session.
	entityRuntimeIDs map[world.Entity]uint64
	entities         map[uint64]world.Entity

	// heldSlot is the slot in the inventory that the controllable is holding.
	heldSlot         *uint32
	inv, offHand, ui *inventory.Inventory
	armour           *inventory.Armour

	inTransaction, containerOpened, openedWindowID uint32
	openedWindow, openedPos                        atomic.Value
}

// Nop represents a no-operation session. It does not do anything when sending a packet to it.
var Nop = &Session{}

// session is a slice of all open sessions. It is protected by the sessionMu, which must be locked whenever
// accessing the value.
var sessions []*Session
var sessionMu sync.Mutex

// selfEntityRuntimeID is the entity runtime (or unique) ID of the controllable that the session holds.
const selfEntityRuntimeID = 1

// ErrSelfRuntimeID is an error returned during packet handling for fields that refer to the player itself and
// must therefore always be 1.
var ErrSelfRuntimeID = errors.New("invalid entity runtime ID: runtime ID for self must always be 1")

// New returns a new session using a controllable entity. The session will control this entity using the
// packets that it receives.
// New takes the connection from which to accept packets. It will start handling these packets after a call to
// Session.Start().
func New(conn *minecraft.Conn, maxChunkRadius int, log *logrus.Logger) *Session {
	s := &Session{
		conn:                   conn,
		log:                    log,
		handlers:               make(map[uint32]packetHandler),
		chunkBuf:               bytes.NewBuffer(make([]byte, 0, 4096)),
		chunkRadius:            int32(maxChunkRadius / 2),
		maxChunkRadius:         int32(maxChunkRadius),
		entityRuntimeIDs:       map[world.Entity]uint64{},
		entities:               map[uint64]world.Entity{},
		currentEntityRuntimeID: 1,
		heldSlot:               new(uint32),
		ui:                     inventory.New(51, nil),
	}
	s.scoreboardObj.Store("")
	s.openedWindow.Store(inventory.New(1, nil))
	s.openedPos.Store(world.BlockPos{})

	s.registerHandlers()
	return s
}

// Start makes the session start handling incoming packets from the client and initialises the controllable of
// the session in the world.
// The function passed will be called when the session stops running.
func (s *Session) Start(c Controllable, w *world.World, onStop func(controllable Controllable)) {
	s.onStop = onStop
	s.c = c
	s.entityRuntimeIDs[c] = selfEntityRuntimeID
	s.entities[selfEntityRuntimeID] = c
	s.chunkLoader = world.NewLoader(int(s.chunkRadius), w, s)
	s.initPlayerList()

	w.AddEntity(s.c)
	s.c.SetGameMode(w.DefaultGameMode())
	s.SendAvailableCommands()
	s.SendSpeed(0.1)

	go s.handlePackets()

	yellow := text.Yellow()
	chat.Global.Println(yellow(s.conn.IdentityData().DisplayName, "has joined the game"))

	s.writePacket(&packet.InventoryContent{
		WindowID: protocol.WindowIDCreative,
		Content:  creativeItems(),
	})
}

// Close closes the session, which in turn closes the controllable and the connection that the session
// manages.
func (s *Session) Close() error {
	s.closeCurrentContainer()

	_ = s.conn.Close()
	_ = s.chunkLoader.Close()
	_ = s.c.Close()

	yellow := text.Yellow()
	chat.Global.Println(yellow(s.conn.IdentityData().DisplayName, "has left the game"))

	s.c.World().RemoveEntity(s.c)

	// This should always be called last due to the timing of the removal of entity runtime IDs.
	s.closePlayerList()

	s.entityMutex.Lock()
	s.entityRuntimeIDs = map[world.Entity]uint64{}
	s.entities = map[uint64]world.Entity{}
	s.entityMutex.Unlock()

	if s.onStop != nil {
		s.onStop(s.c)
		s.onStop = nil
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
		if err := s.handlePacket(pk); err != nil {
			// An error occurred during the handling of a packet. Print the error and stop handling any more
			// packets.
			s.log.Debugf("failed processing packet from %v (%v): %v\n", s.conn.RemoteAddr(), s.c.Name(), err)
			return
		}
	}
}

// sendChunks continuously sends chunks to the player, until a value is sent to the closeChan passed.
func (s *Session) sendChunks(stop <-chan struct{}) {
	t := time.NewTicker(time.Second / 20)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if err := s.chunkLoader.Load(4); err != nil {
				// The world was closed. This should generally never happen.
				s.log.Errorf("error loading chunk: %v", err)
				return
			}
		case <-stop:
			return
		}
	}
}

// handlePacket handles an incoming packet, processing it accordingly. If the packet had invalid data or was
// otherwise not valid in its context, an error is returned.
func (s *Session) handlePacket(pk packet.Packet) error {
	handler, ok := s.handlers[pk.ID()]
	if !ok {
		s.log.Debugf("unhandled packet %T%v from %v\n", pk, fmt.Sprintf("%+v", pk)[1:], s.conn.RemoteAddr())
		return nil
	}
	if handler == nil {
		// A nil handler means it was explicitly unhandled.
		return nil
	}
	if err := handler.Handle(pk, s); err != nil {
		return fmt.Errorf("%T: %w", pk, err)
	}
	return nil
}

// registerHandlers registers all packet handlers found in the packetHandler package.
func (s *Session) registerHandlers() {
	s.handlers = map[uint32]packetHandler{
		packet.IDActorFall:            nil,
		packet.IDAnimate:              nil,
		packet.IDBossEvent:            nil,
		packet.IDCommandRequest:       &CommandRequestHandler{},
		packet.IDContainerClose:       &ContainerCloseHandler{},
		packet.IDInventoryTransaction: &InventoryTransactionHandler{},
		packet.IDMobEquipment:         &MobEquipmentHandler{},
		packet.IDModalFormResponse:    &ModalFormResponseHandler{forms: make(map[uint32]form.Form), currentID: new(uint32)},
		packet.IDMovePlayer:           &MovePlayerHandler{},
		packet.IDPlayerAction:         &PlayerActionHandler{},
		packet.IDPlayerAuthInput:      &PlayerAuthInputHandler{},
		packet.IDRequestChunkRadius:   &RequestChunkRadiusHandler{},
		packet.IDRespawn:              &RespawnHandler{},
		packet.IDText:                 &TextHandler{},
		packet.IDLevelSoundEvent:      nil,
	}
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
	sessionMu.Lock()
	sessions = append(sessions, s)
	for _, session := range sessions {
		// AddStack the player of the session to all sessions currently open, and add the players of all sessions
		// currently open to the player list of the new session.
		session.addToPlayerList(s)
		s.addToPlayerList(session)
	}
	sessionMu.Unlock()
}

// closePlayerList closes the player list of the session and removes the session from the player list of all
// other sessions.
func (s *Session) closePlayerList() {
	sessionMu.Lock()
	n := make([]*Session, 0, len(sessions)-1)
	for _, session := range sessions {
		if session != s {
			n = append(n, session)
		}
		// Remove the player of the session from the player list of all other sessions.
		session.removeFromPlayerList(s)
	}
	sessions = n
	sessionMu.Unlock()
}
