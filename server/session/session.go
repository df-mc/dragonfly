package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/debug"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/player/hud"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Session handles incoming packets from connections and sends outgoing packets by providing a thin layer
// of abstraction over direct packets. A Session basically 'controls' an entity.
type Session struct {
	conf           Config
	once, connOnce sync.Once

	ent      *world.EntityHandle
	conn     Conn
	handlers map[uint32]packetHandler
	packets  chan packet.Packet

	currentScoreboard atomic.Pointer[string]
	currentLines      atomic.Pointer[[]string]

	chunkLoader                 *world.Loader
	chunkRadius, maxChunkRadius int32

	emoteChatMuted bool

	teleportPos atomic.Pointer[mgl64.Vec3]

	entityMutex sync.RWMutex
	// currentEntityRuntimeID holds the runtime ID assigned to the last entity. It is incremented for every
	// entity spawned to the session.
	currentEntityRuntimeID uint64
	// entityRuntimeIDs holds the runtime IDs of entities shown to the session.
	entityRuntimeIDs map[*world.EntityHandle]uint64
	entities         map[uint64]*world.EntityHandle
	hiddenEntities   map[uuid.UUID]struct{}

	// heldSlot is the slot in the inventory that the controllable is holding.
	heldSlot                     *uint32
	inv, offHand, enderChest, ui *inventory.Inventory
	armour                       *inventory.Armour

	// joinSkin is the first skin that the player joined with. It is sent on
	// spawn for the player list, but otherwise updated immediately when the
	// player is viewed.
	joinSkin skin.Skin

	breakingPos cube.Pos

	inTransaction, containerOpened atomic.Bool
	openedWindowID                 atomic.Uint32
	openedContainerID              atomic.Uint32
	openedWindow                   atomic.Pointer[inventory.Inventory]
	openedPos                      atomic.Pointer[cube.Pos]
	swingingArm                    atomic.Bool
	changingSlot                   atomic.Bool
	changingDimension              atomic.Bool
	moving                         bool

	recipes map[uint32]recipe.Recipe

	blobMu                sync.Mutex
	blobs                 map[uint64][]byte
	openChunkTransactions []map[uint64]struct{}
	invOpened             bool

	hudMu      sync.RWMutex
	hudUpdates map[hud.Element]bool
	hiddenHud  map[hud.Element]struct{}

	debugShapesMu     sync.RWMutex
	debugShapes       map[int]debug.Shape
	debugShapesAdd    chan debug.Shape
	debugShapesRemove chan int

	closeBackground chan struct{}
}

// Conn represents a connection that packets are read from and written to by a Session. In addition, it holds some
// information on the identity of the Session.
type Conn interface {
	io.Closer
	// IdentityData returns the login.IdentityData of a Conn. It contains the UUID, XUID and username of the connection.
	IdentityData() login.IdentityData
	// ClientData returns the login.ClientData of a Conn. This includes less sensitive data of the player like its skin,
	// language code and other non-essential information.
	ClientData() login.ClientData
	// ClientCacheEnabled specifies if the Conn has the client cache, used for caching chunks client-side, enabled or
	// not. Some platforms, like the Nintendo Switch, have this disabled at all times.
	ClientCacheEnabled() bool
	// ChunkRadius returns the chunk radius as requested by the client at the other end of the Conn.
	ChunkRadius() int
	// Latency returns the current latency measured over the Conn.
	Latency() time.Duration
	// Flush flushes the packets buffered by the Conn, sending all of them out immediately.
	Flush() error
	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr
	// ReadPacket reads a packet.Packet from the Conn. An error is returned if a deadline was set that was
	// exceeded or if the Conn was closed while awaiting a packet.
	ReadPacket() (pk packet.Packet, err error)
	// WritePacket writes a packet.Packet to the Conn. An error is returned if the Conn was closed before sending the
	// packet.
	WritePacket(pk packet.Packet) error
	// StartGameContext starts the game for the Conn with a context to cancel it.
	StartGameContext(ctx context.Context, data minecraft.GameData) error
}

// Nop represents a no-operation session. It does not do anything when sending a packet to it.
var Nop = &Session{}

// selfEntityRuntimeID is the entity runtime (or unique) ID of the controllable that the session holds.
const selfEntityRuntimeID = 1

// errSelfRuntimeID is an error returned during packet handling for fields that refer to the player itself and
// must therefore always be 1.
var errSelfRuntimeID = errors.New("invalid entity runtime ID: runtime ID for self must always be 1")

type Config struct {
	Log *slog.Logger

	MaxChunkRadius int

	EmoteChatMuted bool

	JoinMessage, QuitMessage chat.Translation

	HandleStop func(*world.Tx, Controllable)
}

func (conf Config) New(conn Conn) *Session {
	r := conn.ChunkRadius()
	if r > conf.MaxChunkRadius {
		r = conf.MaxChunkRadius
		_ = conn.WritePacket(&packet.ChunkRadiusUpdated{ChunkRadius: int32(r)})
	}
	if conf.Log == nil {
		conf.Log = slog.Default()
	}
	conf.Log = conf.Log.With("name", conn.IdentityData().DisplayName, "uuid", conn.IdentityData().Identity, "raddr", conn.RemoteAddr().String())

	s := &Session{}
	*s = Session{
		openChunkTransactions:  make([]map[uint64]struct{}, 0, 8),
		closeBackground:        make(chan struct{}),
		handlers:               map[uint32]packetHandler{},
		packets:                make(chan packet.Packet, 256),
		entityRuntimeIDs:       map[*world.EntityHandle]uint64{},
		entities:               map[uint64]*world.EntityHandle{},
		hiddenEntities:         map[uuid.UUID]struct{}{},
		blobs:                  map[uint64][]byte{},
		chunkRadius:            int32(r),
		maxChunkRadius:         int32(conf.MaxChunkRadius),
		emoteChatMuted:         conf.EmoteChatMuted,
		conn:                   conn,
		currentEntityRuntimeID: 1,
		heldSlot:               new(uint32),
		recipes:                make(map[uint32]recipe.Recipe),
		conf:                   conf,
		hudUpdates:             make(map[hud.Element]bool),
		hiddenHud:              make(map[hud.Element]struct{}),
		debugShapes:            make(map[int]debug.Shape),
		debugShapesAdd:         make(chan debug.Shape, 256),
		debugShapesRemove:      make(chan int, 256),
	}
	s.openedWindow.Store(inventory.New(1, nil))
	s.openedPos.Store(&cube.Pos{})

	var scoreboardName string
	var scoreboardLines []string
	s.currentScoreboard.Store(&scoreboardName)
	s.currentLines.Store(&scoreboardLines)

	s.registerHandlers()
	s.sendBiomes()
	groups, items := creativeContent()
	s.writePacket(&packet.CreativeContent{Groups: groups, Items: items})
	s.sendRecipes()
	s.sendArmourTrimData()
	s.SendSpeed(0.1)
	go func() {
		for {
			select {
			case <-s.closeBackground:
				return
			case pk := <-s.packets:
				_ = conn.WritePacket(pk)
			}
		}
	}()
	return s
}

// SetHandle sets the world.EntityHandle of the Session and attaches a skin to
// other players on join.
func (s *Session) SetHandle(handle *world.EntityHandle, skin skin.Skin) {
	s.ent = handle
	s.entityRuntimeIDs[handle] = selfEntityRuntimeID
	s.entities[selfEntityRuntimeID] = handle

	s.joinSkin = skin
	sessions.Add(s)
}

// Spawn makes the Controllable passed spawn in the world.World.
// The function passed will be called when the session stops running.
func (s *Session) Spawn(c Controllable, tx *world.Tx) {
	s.SendHealth(c.Health(), c.MaxHealth(), c.Absorption())
	s.SendExperience(c.ExperienceLevel(), c.ExperienceProgress())
	s.SendFood(c.Food(), 0, 0)

	pos := c.Position()
	s.chunkLoader = world.NewLoader(int(s.chunkRadius), tx.World(), s)
	s.chunkLoader.Move(tx, pos)
	s.writePacket(&packet.NetworkChunkPublisherUpdate{
		Position: protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])},
		Radius:   uint32(s.chunkRadius) << 4,
	})

	s.sendAvailableEntities(tx.World())

	c.SetGameMode(c.GameMode())
	for _, e := range c.Effects() {
		s.SendEffect(e)
	}
	s.ViewEntityState(c)

	s.sendInv(s.inv, protocol.WindowIDInventory)
	s.sendInv(s.ui, protocol.WindowIDUI)
	s.sendInv(s.offHand, protocol.WindowIDOffHand)
	s.sendInv(s.armour.Inventory(), protocol.WindowIDArmour)

	chat.Global.Subscribe(c)
	if !s.conf.JoinMessage.Zero() {
		chat.Global.Writet(s.conf.JoinMessage, s.conn.IdentityData().DisplayName)
	}

	go s.background()
	go s.handlePackets()
}

// Close closes the session, which in turn closes the controllable and the connection that the session
// manages. Close ensures the method only runs code on the first call.
func (s *Session) Close(tx *world.Tx, c Controllable) {
	s.once.Do(func() {
		s.close(tx, c)
	})
}

// close closes the session, which in turn closes the controllable and the connection that the session
// manages.
func (s *Session) close(tx *world.Tx, c Controllable) {
	c.MoveItemsToInventory()
	s.closeCurrentContainer(tx)

	s.conf.HandleStop(tx, c)

	// Clear the inventories so that they no longer hold references to the connection.
	_ = s.inv.Close()
	_ = s.offHand.Close()
	_ = s.armour.Close()

	s.chunkLoader.Close(tx)

	if !s.conf.QuitMessage.Zero() {
		chat.Global.Writet(s.conf.QuitMessage, s.conn.IdentityData().DisplayName)
	}
	chat.Global.Unsubscribe(c)

	// Note: Be aware of where RemoveEntity is called. This must not be done too
	// early.
	tx.RemoveEntity(c)
	_ = s.ent.Close()

	// This should always be called last due to the timing of the removal of
	// entity runtime IDs.
	sessions.Remove(s)
	s.entityMutex.Lock()
	clear(s.entityRuntimeIDs)
	clear(s.entities)
	s.entityMutex.Unlock()
}

// CloseConnection closes the underlying connection of the session so that the session ends up being closed
// eventually.
func (s *Session) CloseConnection() {
	s.connOnce.Do(func() {
		_ = s.conn.Close()
		close(s.closeBackground)
	})
}

// Addr returns the net.Addr of the client.
func (s *Session) Addr() net.Addr {
	return s.conn.RemoteAddr()
}

// Latency returns the latency of the connection.
func (s *Session) Latency() time.Duration {
	return s.conn.Latency()
}

// ClientData returns the login.ClientData of the underlying *minecraft.Conn.
func (s *Session) ClientData() login.ClientData {
	return s.conn.ClientData()
}

// handlePackets continuously handles incoming packets from the connection. It processes them accordingly.
// Once the connection is closed, handlePackets will return.
func (s *Session) handlePackets() {
	defer func() {
		// First close the Controllable. This might lead to a world change
		// (player might be dead while disconnecting, in which case it will
		// respawn first).
		s.ent.ExecWorld(func(tx *world.Tx, e world.Entity) {
			_ = e.(Controllable).Close()
		})
		// Because the player might no longer be in the same world after
		// closing, we create a new transaction
		s.ent.ExecWorld(func(tx *world.Tx, e world.Entity) {
			s.Close(tx, e.(Controllable))
		})
	}()
	for {
		pk, err := s.conn.ReadPacket()
		if err != nil {
			return
		}
		s.ent.ExecWorld(func(tx *world.Tx, e world.Entity) {
			err = s.handlePacket(pk, tx, e.(Controllable))
		})
		if err != nil {
			s.conf.Log.Debug("process packet: " + err.Error())
			return
		}
	}
}

// background performs background tasks of the Session. This includes chunk sending and automatic command updating.
// background returns when the Session's connection is closed using CloseConnection.
func (s *Session) background() {
	var (
		r          map[string]map[int]cmd.Runnable
		enums      map[string]cmd.Enum
		enumValues map[string][]string
		ok         bool
		i          int
	)

	s.ent.ExecWorld(func(tx *world.Tx, e world.Entity) {
		co := e.(Controllable)
		r = s.sendAvailableCommands(co)
		enums, enumValues = s.enums(co)
	})

	t := time.NewTicker(time.Second / 20)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			s.ent.ExecWorld(func(tx *world.Tx, e world.Entity) {
				c := e.(Controllable)

				if i++; i%20 == 0 {
					// Enum resending happens relatively often and frequent updates are more important than with full
					// command changes. Those are generally only related to permission changes, which doesn't happen often.
					s.resendEnums(enums, enumValues, c)
				}
				if i%100 == 0 {
					// Try to resend commands only every 5 seconds.
					if r, ok = s.resendCommands(r, c); ok {
						enums, enumValues = s.enums(c)
					}
				}
				s.sendChunks(tx, c)
			})
		case <-s.closeBackground:
			return
		}
	}
}

// sendChunks sends the next up to 4 chunks to the connection. What chunks are loaded depends on the connection of
// the chunk loader and the chunks that were previously loaded.
func (s *Session) sendChunks(tx *world.Tx, c Controllable) {
	if w := tx.World(); s.chunkLoader.World() != w && w != nil {
		s.handleWorldSwitch(w, tx, c)
	}
	pos := c.Position()
	s.chunkLoader.Move(tx, pos)
	s.writePacket(&packet.NetworkChunkPublisherUpdate{
		Position: protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])},
		Radius:   uint32(s.chunkRadius) << 4,
	})

	s.blobMu.Lock()
	const maxChunkTransactions = 8
	toLoad := maxChunkTransactions - len(s.openChunkTransactions)
	s.blobMu.Unlock()
	if toLoad > 4 {
		toLoad = 4
	}
	s.chunkLoader.Load(tx, toLoad)
}

// handleWorldSwitch handles the player of the Session switching worlds.
func (s *Session) handleWorldSwitch(w *world.World, tx *world.Tx, c Controllable) {
	if s.conn.ClientCacheEnabled() {
		s.blobMu.Lock()
		s.blobs = map[uint64][]byte{}
		s.openChunkTransactions = nil
		s.blobMu.Unlock()
	}

	dim, _ := world.DimensionID(w.Dimension())
	same := w.Dimension() == s.chunkLoader.World().Dimension()
	if !same {
		s.changeDimension(int32(dim), false, c)
	}
	s.ViewEntityTeleport(c, c.Position())
	s.chunkLoader.ChangeWorld(tx, w)
}

// changeDimension changes the dimension of the client. If silent is set to true, the portal noise will be stopped
// immediately.
func (s *Session) changeDimension(dim int32, silent bool, c Controllable) {
	s.changingDimension.Store(true)
	h := s.handlers[packet.IDServerBoundLoadingScreen].(*ServerBoundLoadingScreenHandler)
	id := h.currentID.Add(1)
	h.expectedID.Store(id)

	s.writePacket(&packet.ChangeDimension{
		Dimension:       dim,
		Position:        vec64To32(c.Position().Add(entityOffset(c))),
		LoadingScreenID: protocol.Option(id),
	})
	s.writePacket(&packet.StopSound{StopAll: silent})
	s.writePacket(&packet.PlayStatus{Status: packet.PlayStatusPlayerSpawn})

	// As of v1.19.50, the dimension ack that is meant to be sent by the client is now sent by the server. The client
	// still sends the ack, but after the server has sent it. Thanks to Mojang for another groundbreaking change.
	s.writePacket(&packet.PlayerAction{
		EntityRuntimeID: selfEntityRuntimeID,
		ActionType:      protocol.PlayerActionDimensionChangeDone,
	})
}

// ChangingDimension returns whether the session is currently changing dimension or not.
func (s *Session) ChangingDimension() bool {
	return s.changingDimension.Load()
}

// handlePacket handles an incoming packet, processing it accordingly. If the packet had invalid data or was
// otherwise not valid in its context, an error is returned.
func (s *Session) handlePacket(pk packet.Packet, tx *world.Tx, c Controllable) (err error) {
	handler, ok := s.handlers[pk.ID()]
	if !ok {
		s.conf.Log.Debug("unhandled packet", "packet", fmt.Sprintf("%T", pk), "data", fmt.Sprintf("%+v", pk)[1:])
		return nil
	}
	if handler == nil {
		// A nil handler means it was explicitly unhandled.
		return nil
	}
	if err := handler.Handle(pk, s, tx, c); err != nil {
		return fmt.Errorf("%T: %w", pk, err)
	}
	return nil
}

// registerHandlers registers all packet handlers found in the packetHandler package.
func (s *Session) registerHandlers() {
	s.handlers = map[uint32]packetHandler{
		packet.IDActorEvent:                nil,
		packet.IDAdventureSettings:         nil, // Deprecated, the client still sends this though.
		packet.IDAnimate:                   nil,
		packet.IDAnvilDamage:               nil,
		packet.IDBlockActorData:            &BlockActorDataHandler{},
		packet.IDBlockPickRequest:          &BlockPickRequestHandler{},
		packet.IDBookEdit:                  &BookEditHandler{},
		packet.IDBossEvent:                 nil,
		packet.IDClientCacheBlobStatus:     &ClientCacheBlobStatusHandler{},
		packet.IDCommandRequest:            &CommandRequestHandler{},
		packet.IDContainerClose:            &ContainerCloseHandler{},
		packet.IDEmote:                     &EmoteHandler{},
		packet.IDEmoteList:                 nil,
		packet.IDFilterText:                nil,
		packet.IDInteract:                  &InteractHandler{},
		packet.IDInventoryTransaction:      &InventoryTransactionHandler{},
		packet.IDItemStackRequest:          &ItemStackRequestHandler{changes: map[byte]map[byte]changeInfo{}, responseChanges: map[int32]map[*inventory.Inventory]map[byte]responseChange{}},
		packet.IDLecternUpdate:             &LecternUpdateHandler{},
		packet.IDMobEquipment:              &MobEquipmentHandler{},
		packet.IDModalFormResponse:         &ModalFormResponseHandler{forms: make(map[uint32]form.Form)},
		packet.IDMovePlayer:                nil,
		packet.IDNPCRequest:                &NPCRequestHandler{},
		packet.IDPlayerAction:              &PlayerActionHandler{},
		packet.IDPlayerAuthInput:           &PlayerAuthInputHandler{},
		packet.IDPlayerSkin:                &PlayerSkinHandler{},
		packet.IDRequestAbility:            &RequestAbilityHandler{},
		packet.IDRequestChunkRadius:        &RequestChunkRadiusHandler{},
		packet.IDRespawn:                   &RespawnHandler{},
		packet.IDSetPlayerInventoryOptions: nil,
		packet.IDSubChunkRequest:           &SubChunkRequestHandler{},
		packet.IDText:                      &TextHandler{},
		packet.IDServerBoundLoadingScreen:  &ServerBoundLoadingScreenHandler{},
		packet.IDServerBoundDiagnostics:    &ServerBoundDiagnosticsHandler{},
	}
}

// writePacket writes a packet to the session's connection if it is not Nop.
func (s *Session) writePacket(pk packet.Packet) {
	if s == Nop {
		return
	}
	select {
	case s.packets <- pk:
	case <-s.closeBackground:
	}
}

// actorIdentifier represents the structure of an actor identifier sent over the network.
type actorIdentifier struct {
	// ID is a unique namespaced identifier for the entity.
	ID string `nbt:"id"`
}

// sendAvailableEntities sends all registered entities to the player.
func (s *Session) sendAvailableEntities(w *world.World) {
	var identifiers []actorIdentifier
	for _, t := range w.EntityRegistry().Types() {
		identifiers = append(identifiers, actorIdentifier{ID: t.EncodeEntity()})
	}
	serializedEntityData, err := nbt.Marshal(map[string]any{"idlist": identifiers})
	if err != nil {
		panic("should never happen")
	}
	s.writePacket(&packet.AvailableActorIdentifiers{SerialisedEntityIdentifiers: serializedEntityData})
}
