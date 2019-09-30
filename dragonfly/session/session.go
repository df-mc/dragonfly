package session

import (
	"bytes"
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/chat"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/cmd"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sirupsen/logrus"
	"net"
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

	cmdOrigin protocol.CommandOrigin

	chunkBuf       *bytes.Buffer
	chunkLoader    atomic.Value
	chunkRadius    int32
	maxChunkRadius int32
}

// Nop represents a no-operation session. It does not do anything when sending a packet to it.
var Nop = &Session{}

// New returns a new session using a controllable entity. The session will control this entity using the
// packets that it receives.
// New takes the connection from which to accept packets. It will start handling these packets after a call to
// Session.Handle().
func New(c Controllable, conn *minecraft.Conn, w *world.World, maxChunkRadius int, log *logrus.Logger) *Session {
	s := &Session{
		c:              c,
		conn:           conn,
		log:            log,
		chunkBuf:       bytes.NewBuffer(make([]byte, 0, 4096)),
		world:          w,
		chunkRadius:    int32(maxChunkRadius / 2),
		maxChunkRadius: int32(maxChunkRadius),
	}
	s.chunkLoader.Store(world.NewLoader(maxChunkRadius/2, w))
	s.controllableClosed.Store(false)

	yellow := text.Yellow()
	chat.Global.Println(yellow(s.conn.IdentityData().DisplayName, "has joined the game"))

	return s
}

// Handle makes the session start handling incoming packets from the client.
func (s *Session) Handle() {
	go s.handlePackets()
	s.SendAvailableCommands()
}

// Close closes the session, which in turn closes the controllable and the connection that the session
// manages.
func (s *Session) Close() error {
	_ = s.c.Close()
	_ = s.conn.Close()

	yellow := text.Yellow()
	chat.Global.Println(yellow(s.conn.IdentityData().DisplayName, "has left the game"))
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
			if err := s.chunkLoader.Load().(*world.Loader).Load(4, s.SendChunk); err != nil {
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

// handleText ...
func (s *Session) handleText(pk *packet.Text) error {
	if pk.TextType != packet.TextTypeChat {
		return fmt.Errorf("text packet can only contain text type of type chat (%v) but got %v", packet.TextTypeChat, pk.TextType)
	}
	if pk.SourceName != s.conn.IdentityData().DisplayName {
		return fmt.Errorf("text packet source name must be equal to display name")
	}
	s.c.Chat(pk.Message)
	return nil
}

// handleCommandRequest ...
func (s *Session) handleCommandRequest(pk *packet.CommandRequest) error {
	if pk.Internal {
		return fmt.Errorf("command request packet must never have the internal field set to true")
	}
	s.cmdOrigin = pk.CommandOrigin
	s.c.ExecuteCommand(pk.CommandLine)
	return nil
}

// handleMovePlayer ...
func (s *Session) handleMovePlayer(pk *packet.MovePlayer) error {
	if pk.EntityRuntimeID != s.conn.GameData().EntityRuntimeID {
		return fmt.Errorf("incorrect entity runtime ID %v: runtime ID must be equal to %v", pk.EntityRuntimeID, s.conn.GameData().EntityRuntimeID)
	}
	// TODO: Make players move.
	s.chunkLoader.Load().(*world.Loader).Move(pk.Position)
	s.writePacket(&packet.NetworkChunkPublisherUpdate{
		Position: protocol.BlockPos{int32(pk.Position[0]), int32(pk.Position[1]), int32(pk.Position[2])},
		Radius:   uint32(s.chunkRadius * 16),
	})
	return nil
}

// handleRequestChunkRadius ...
func (s *Session) handleRequestChunkRadius(pk *packet.RequestChunkRadius) error {
	if pk.ChunkRadius > s.maxChunkRadius {
		pk.ChunkRadius = s.maxChunkRadius
	}
	s.chunkRadius = pk.ChunkRadius
	s.chunkLoader.Store(world.NewLoader(int(s.chunkRadius), s.world))

	s.writePacket(&packet.ChunkRadiusUpdated{ChunkRadius: s.chunkRadius})
	return nil
}

// SendMessage ...
func (s *Session) SendMessage(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeRaw,
		Message:  message,
	})
}

// SendTip ...
func (s *Session) SendTip(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypePopup,
		Message:  message,
	})
}

// SendAnnouncement ...
func (s *Session) SendAnnouncement(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeAnnouncement,
		Message:  message,
	})
}

// SendPopup ...
func (s *Session) SendPopup(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypePopup,
		Message:  message,
	})
}

// SendJukeBoxPopup ...
func (s *Session) SendJukeBoxPopup(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeJukeboxPopup,
		Message:  message,
	})
}

// SendScoreboard ...
func (s *Session) SendScoreboard(displayName string, objName string) {
	s.writePacket(&packet.SetDisplayObjective{
		DisplaySlot:   "sidebar",
		ObjectiveName: objName,
		DisplayName:   displayName,
		CriteriaName:  "dummy",
	})
}

// RemoveScoreboard ...
func (s *Session) RemoveScoreboard(objName string) {
	s.writePacket(&packet.RemoveObjective{
		ObjectiveName: objName,
	})
}

const tickLength = time.Second / 20

// SetTitleDurations ...
func (s *Session) SetTitleDurations(fadeInDuration, remainDuration, fadeOutDuration time.Duration) {
	s.writePacket(&packet.SetTitle{
		ActionType:      packet.TitleActionSetDurations,
		FadeInDuration:  int32(fadeInDuration / tickLength),
		RemainDuration:  int32(remainDuration / tickLength),
		FadeOutDuration: int32(fadeOutDuration / tickLength),
	})
}

// SendTitle ...
func (s *Session) SendTitle(text string) {
	s.writePacket(&packet.SetTitle{ActionType: packet.TitleActionSetTitle, Text: text})
}

// SendSubtitle ...
func (s *Session) SendSubtitle(text string) {
	s.writePacket(&packet.SetTitle{ActionType: packet.TitleActionSetSubtitle, Text: text})
}

// SendActionbarMessage ...
func (s *Session) SendActionBarMessage(text string) {
	s.writePacket(&packet.SetTitle{ActionType: packet.TitleActionSetActionBar, Text: text})
}

// SendNetherDimension sends the player to the nether dimension
func (s *Session) SendNetherDimension() {
	s.writePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionNether,
		Position:  mgl32.Vec3{},
		Respawn:   false,
	})
}

// SendEndDimension sends the player to the end dimension
func (s *Session) SendEndDimension() {
	s.writePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionEnd,
		Position:  mgl32.Vec3{},
		Respawn:   false,
	})
}

// SendNetherDimension sends the player to the overworld dimension
func (s *Session) SendOverworldDimension() {
	s.writePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionOverworld,
		Position:  mgl32.Vec3{},
		Respawn:   false,
	})
}

// SendChunk sends a chunk to the player at the chunk X and Y passed.
func (s *Session) SendChunk(pos world.ChunkPos, c *chunk.Chunk) {
	data := chunk.NetworkEncode(c)

	count := 16
	for y := 15; y >= 0; y-- {
		if data.SubChunks[y] == nil {
			count--
			continue
		}
		break
	}
	for y := 0; y < count; y++ {
		if data.SubChunks[y] == nil {
			_ = s.chunkBuf.WriteByte(chunk.SubChunkVersion)
			// We write zero here, meaning the sub chunk has no block storages: The sub chunk is completely
			// empty.
			_ = s.chunkBuf.WriteByte(0)
			continue
		}
		_, _ = s.chunkBuf.Write(data.SubChunks[y])
	}
	_, _ = s.chunkBuf.Write(data.Data2D)
	_, _ = s.chunkBuf.Write(data.BlockNBT)

	s.writePacket(&packet.LevelChunk{
		ChunkX:        pos[0],
		ChunkZ:        pos[1],
		SubChunkCount: uint32(count),
		RawPayload:    append([]byte(nil), s.chunkBuf.Bytes()...),
	})
	s.chunkBuf.Reset()
}

// Disconnect disconnects the client and ultimately closes the session. If the message passed is non-empty,
// it will be shown to the client.
func (s *Session) Disconnect(message string) {
	s.writePacket(&packet.Disconnect{
		HideDisconnectionScreen: message == "",
		Message:                 message,
	})
	s.controllableClosed.Store(true)
}

// Transfer transfers the player to a server with the IP and port passed.
func (s *Session) Transfer(ip net.IP, port int) {
	s.writePacket(&packet.Transfer{
		Address: ip.String(),
		Port:    uint16(port),
	})
}

// SendCommandOutput sends the output of a command to the player. It will be shown to the caller of the
// command, which might be the player or a websocket server.
func (s *Session) SendCommandOutput(output *cmd.Output) {
	messages := make([]protocol.CommandOutputMessage, 0, output.MessageCount()+output.ErrorCount())
	for _, message := range output.Messages() {
		messages = append(messages, protocol.CommandOutputMessage{
			Success: true,
			Message: message,
		})
	}
	for _, err := range output.Errors() {
		messages = append(messages, protocol.CommandOutputMessage{
			Success: false,
			Message: err.Error(),
		})
	}

	s.writePacket(&packet.CommandOutput{
		CommandOrigin:  s.cmdOrigin,
		OutputType:     3,
		SuccessCount:   uint32(output.MessageCount()),
		OutputMessages: messages,
	})
}

// SendAvailableCommands sends all available commands of the server. Once sent, they will be visible in the
// /help list and will be auto-completed.
func (s *Session) SendAvailableCommands() {
	commands := cmd.Commands()
	pk := &packet.AvailableCommands{}
	for alias, c := range commands {
		if c.Name() != alias {
			// Don't add duplicate entries for aliases.
			continue
		}
		params := c.Params()
		overloads := make([]protocol.CommandOverload, len(params))
		for i, params := range params {
			for _, paramInfo := range params {
				t, enum := valueToParamType(paramInfo.Value)
				t |= protocol.CommandArgValid
				overloads[i].Parameters = append(overloads[i].Parameters, protocol.CommandParameter{
					Name:                paramInfo.Name,
					Type:                t,
					Optional:            paramInfo.Optional,
					CollapseEnumOptions: paramInfo.Value == false || paramInfo.Value == true,
					Enum:                enum,
					Suffix:              paramInfo.Suffix,
				})
			}
		}
		pk.Commands = append(pk.Commands, protocol.Command{
			Name:        c.Name(),
			Description: c.Description(),
			Aliases:     c.Aliases(),
			Overloads:   overloads,
		})
	}
	s.writePacket(pk)
}

// writePacket writes a packet to the session's connection if it is not Nop.
func (s *Session) writePacket(pk packet.Packet) {
	if s == Nop {
		return
	}
	_ = s.conn.WritePacket(pk)
}

// valueToParamType finds the command argument type of a value passed and returns it, in addition to creating
// an enum if applicable.
func valueToParamType(i interface{}) (t uint32, enum protocol.CommandEnum) {
	switch i.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return protocol.CommandArgTypeInt, enum
	case float32, float64:
		return protocol.CommandArgTypeFloat, enum
	case string:
		return protocol.CommandArgTypeString, enum
	case bool:
		return 0, protocol.CommandEnum{
			Type:    "bool",
			Options: []string{"true", "1", "false", "0"},
		}
	case mgl32.Vec3:
		return protocol.CommandArgTypePosition, enum
	}
	if param, ok := i.(cmd.Parameter); ok && param.Type() == "player" || param.Type() == "target" {
		return protocol.CommandArgTypeTarget, enum
	}
	if enum, ok := i.(cmd.Enum); ok {
		return 0, protocol.CommandEnum{
			Type:    enum.Type(),
			Options: enum.Options(),
		}
	}
	return protocol.CommandArgTypeValue, enum
}
