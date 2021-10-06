package session

import (
	"bytes"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"time"
)

// startCommandTicking starts a ticker that will check every minute for changes in the command data,
// and if so sync this with the client.
func (s *Session) startCommandTicking() {
	ticker := time.NewTicker(time.Minute)
	stop := make(chan struct{})
	s.commandSync = stop
	go func() {
		select {
		case <-ticker.C:
			oldCommands := s.lastCommands
			newCommands := s.buildAvailableCommands()

			oldBuff, newBuff := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})

			oldCommands.Marshal(protocol.NewWriter(oldBuff, 0))
			newCommands.Marshal(protocol.NewWriter(newBuff, 0))

			if bytes.Compare(oldBuff.Bytes(), newBuff.Bytes()) != 0 {
				s.writePacket(newCommands)
			}
		case _, _ = <-stop:
			ticker.Stop()
			return
		}
	}()
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

	h := s.handlers[packet.IDCommandRequest]
	if h == nil { // This will be nil if the player has been disconnected
		return
	}
	s.writePacket(&packet.CommandOutput{
		CommandOrigin:  h.(*CommandRequestHandler).origin,
		OutputType:     3,
		SuccessCount:   uint32(output.MessageCount()),
		OutputMessages: messages,
	})
}

// SendAvailableCommands sends all available commands of the server. Once sent, they will be visible in the
// /help list and will be auto-completed.
func (s *Session) SendAvailableCommands() {
	pk := s.buildAvailableCommands()
	s.lastCommands = pk
	s.writePacket(pk)
}

// buildAvailableCommands packet builds a new up-to-date packet containing all command data such as aliases,
// enums, etc.
func (s *Session) buildAvailableCommands() *packet.AvailableCommands {
	commands := cmd.Commands()
	pk := &packet.AvailableCommands{}
	for alias, c := range commands {
		if c.Name() != alias {
			// Don't add duplicate entries for aliases.
			continue
		}
		params := c.Params(s.c)
		overloads := make([]protocol.CommandOverload, len(params))
		for i, params := range params {
			for _, paramInfo := range params {
				t, enum := valueToParamType(paramInfo.Value, s.c)
				t |= protocol.CommandArgValid

				opt := byte(0)
				if _, ok := paramInfo.Value.(bool); ok {
					opt |= protocol.ParamOptionCollapseEnum
				}
				overloads[i].Parameters = append(overloads[i].Parameters, protocol.CommandParameter{
					Name:     paramInfo.Name,
					Type:     t,
					Optional: paramInfo.Optional,
					Options:  opt,
					Enum:     enum,
					Suffix:   paramInfo.Suffix,
				})
			}
		}
		if len(params) > 0 {
			pk.Commands = append(pk.Commands, protocol.Command{
				Name:        c.Name(),
				Description: c.Description(),
				Aliases:     c.Aliases(),
				Overloads:   overloads,
			})
		}
	}
	return pk
}

// valueToParamType finds the command argument type of a value passed and returns it, in addition to creating
// an enum if applicable.
func valueToParamType(i interface{}, source cmd.Source) (t uint32, enum protocol.CommandEnum) {
	switch i.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return protocol.CommandArgTypeInt, enum
	case float32, float64:
		return protocol.CommandArgTypeFloat, enum
	case string:
		return protocol.CommandArgTypeString, enum
	case cmd.Varargs:
		return protocol.CommandArgTypeRawText, enum
	case cmd.Target, []cmd.Target:
		return protocol.CommandArgTypeTarget, enum
	case bool:
		return 0, protocol.CommandEnum{
			Type:    "bool",
			Options: []string{"true", "1", "false", "0"},
		}
	case mgl64.Vec3:
		return protocol.CommandArgTypePosition, enum
	}
	if sub, ok := i.(cmd.SubCommand); ok {
		return 0, protocol.CommandEnum{
			Type:    "SubCommand" + sub.SubName(),
			Options: []string{sub.SubName()},
		}
	}
	if enum, ok := i.(cmd.Enum); ok {
		return 0, protocol.CommandEnum{
			Type:    enum.Type(),
			Options: enum.Options(source),
		}
	}
	return protocol.CommandArgTypeValue, enum
}
