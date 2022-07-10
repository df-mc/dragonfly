package session

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// SendCommandOutput sends the output of a command to the player. It will be shown to the caller of the
// command, which might be the player or a websocket server.
func (s *Session) SendCommandOutput(output *cmd.Output) {
	if s == Nop {
		return
	}
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
		CommandOrigin:  s.handlers[packet.IDCommandRequest].(*CommandRequestHandler).origin,
		OutputType:     packet.CommandOutputTypeAllOutput,
		SuccessCount:   uint32(output.MessageCount()),
		OutputMessages: messages,
	})
}

// sendAvailableCommands sends all available commands of the server. Once sent, they will be visible in the
// /help list and will be auto-completed.
func (s *Session) sendAvailableCommands() map[string]map[int]cmd.Runnable {
	commands := cmd.Commands()
	m := make(map[string]map[int]cmd.Runnable, len(commands))

	pk := &packet.AvailableCommands{}
	for alias, c := range commands {
		if c.Name() != alias {
			// Don't add duplicate entries for aliases.
			continue
		}
		m[alias] = c.Runnables(s.c)

		params := c.Params(s.c)
		overloads := make([]protocol.CommandOverload, len(params))
		for i, params := range params {
			for _, paramInfo := range params {
				t, enum := valueToParamType(paramInfo, s.c)
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
	s.writePacket(pk)
	return m
}

// valueToParamType finds the command argument type of the value passed and returns it, in addition to creating
// an enum if applicable.
func valueToParamType(i cmd.ParamInfo, source cmd.Source) (t uint32, enum protocol.CommandEnum) {
	switch i.Value.(type) {
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
	case cmd.SubCommand:
		return 0, protocol.CommandEnum{
			Type:    "SubCommand" + i.Name,
			Options: []string{i.Name},
		}
	}
	if enum, ok := i.Value.(cmd.Enum); ok {
		return 0, protocol.CommandEnum{
			Type:    enum.Type(),
			Options: enum.Options(source),
			Dynamic: true,
		}
	}
	return protocol.CommandArgTypeValue, enum
}

// resendCommands resends all commands that a Session has access to if the map of runnable commands passed does not
// match with the commands that the Session is currently allowed to execute.
// True is returned if the commands were resent.
func (s *Session) resendCommands(before map[string]map[int]cmd.Runnable) (map[string]map[int]cmd.Runnable, bool) {
	commands := cmd.Commands()
	m := make(map[string]map[int]cmd.Runnable, len(commands))

	for alias, c := range commands {
		if c.Name() == alias {
			m[alias] = c.Runnables(s.c)
		}
	}
	if len(before) != len(m) {
		return s.sendAvailableCommands(), true
	}
	for name, r := range m {
		for k := range r {
			if _, ok := before[name][k]; !ok {
				return s.sendAvailableCommands(), true
			}
		}
	}
	return m, false
}

// enums returns a map of all enums exposed to the Session and records the values those enums currently hold.
func (s *Session) enums() (map[string]cmd.Enum, map[string][]string) {
	enums, enumValues := make(map[string]cmd.Enum), make(map[string][]string)
	for alias, c := range cmd.Commands() {
		if c.Name() == alias {
			for _, params := range c.Params(s.c) {
				for _, paramInfo := range params {
					if enum, ok := paramInfo.Value.(cmd.Enum); ok {
						enums[enum.Type()] = enum
						enumValues[enum.Type()] = enum.Options(s.c)
					}
				}
			}
		}
	}
	return enums, enumValues
}

// resendEnums checks the options of the enums passed against the values that were previously recorded. If they do not
// match, the enum is resent to the client and the values are updated in the before map.
func (s *Session) resendEnums(enums map[string]cmd.Enum, before map[string][]string) {
	for name, enum := range enums {
		valuesBefore := before[name]
		values := enum.Options(s.c)
		before[name] = values

		if len(valuesBefore) != len(values) {
			s.writePacket(&packet.UpdateSoftEnum{EnumType: name, Options: values, ActionType: packet.SoftEnumActionSet})
			continue
		}
		for k, v := range values {
			if valuesBefore[k] != v {
				s.writePacket(&packet.UpdateSoftEnum{EnumType: name, Options: values, ActionType: packet.SoftEnumActionSet})
				break
			}
		}
	}
}
