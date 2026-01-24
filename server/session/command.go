package session

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/text/language"
)

// SendCommandOutput sends the output of a command to the player. It will be shown to the caller of the
// command, which might be the player or a websocket server.
func (s *Session) SendCommandOutput(output *cmd.Output, l language.Tag) {
	if s == Nop {
		return
	}
	messages := make([]protocol.CommandOutputMessage, 0, output.MessageCount()+output.ErrorCount())
	for _, message := range output.Messages() {
		om := protocol.CommandOutputMessage{Success: true, Message: message.String()}
		if t, ok := message.(translation); ok {
			om.Message, om.Parameters = t.Resolve(l), t.Params(l)
		}
		messages = append(messages, om)
	}
	for _, err := range output.Errors() {
		om := protocol.CommandOutputMessage{Message: err.Error()}
		if t, ok := err.(translation); ok {
			om.Message, om.Parameters = t.Resolve(l), t.Params(l)
		}
		messages = append(messages, om)
	}

	s.writePacket(&packet.CommandOutput{
		CommandOrigin:  s.handlers[packet.IDCommandRequest].(*CommandRequestHandler).origin,
		OutputType:     packet.CommandOutputTypeAllOutput,
		SuccessCount:   uint32(output.MessageCount()),
		OutputMessages: messages,
	})
}

type translation interface {
	Resolve(l language.Tag) string
	Params(l language.Tag) []string
}

// sendAvailableCommands sends all available commands of the server. Once sent, they will be visible in the
// /help list and will be auto-completed.
func (s *Session) sendAvailableCommands(co Controllable) map[string]map[int]cmd.Runnable {
	commands := cmd.Commands()
	m := make(map[string]map[int]cmd.Runnable, len(commands))

	for alias, c := range commands {
		if run := c.Runnables(co); len(run) > 0 {
			m[alias] = run
		}
	}
	s.writePacket(cmd.BuildAvailableCommands(commands, co))
	return m
}

// resendCommands resends all commands that a Session has access to if the map of runnable commands passed does not
// match with the commands that the Session is currently allowed to execute.
// True is returned if the commands were resent.
func (s *Session) resendCommands(before map[string]map[int]cmd.Runnable, co Controllable) (map[string]map[int]cmd.Runnable, bool) {
	commands := cmd.Commands()
	m := make(map[string]map[int]cmd.Runnable, len(commands))

	for alias, c := range commands {
		if c.Name() == alias {
			if run := c.Runnables(co); len(run) > 0 {
				m[alias] = run
			}
		}
	}
	if len(before) != len(m) {
		return s.sendAvailableCommands(co), true
	}
	// First check for commands that were newly added.
	for name, r := range m {
		for k := range r {
			if _, ok := before[name][k]; !ok {
				return s.sendAvailableCommands(co), true
			}
		}
	}
	// Then check for commands that a player could execute before, but no longer can.
	for name, r := range before {
		for k := range r {
			if _, ok := m[name][k]; !ok {
				return s.sendAvailableCommands(co), true
			}
		}
	}
	return m, false
}

// enums returns a map of all enums exposed to the Session and records the values those enums currently hold.
func (s *Session) enums(co Controllable) (map[string]cmd.Enum, map[string][]string) {
	enums, enumValues := make(map[string]cmd.Enum), make(map[string][]string)
	for alias, c := range cmd.Commands() {
		if c.Name() == alias {
			for _, params := range c.Params(co) {
				for _, paramInfo := range params {
					if enum, ok := paramInfo.Value.(cmd.Enum); ok {
						enums[enum.Type()] = enum
						enumValues[enum.Type()] = enum.Options(co)
					}
				}
			}
		}
	}
	return enums, enumValues
}

// resendEnums checks the options of the enums passed against the values that were previously recorded. If they do not
// match, the enum is resent to the client and the values are updated in the before map.
func (s *Session) resendEnums(enums map[string]cmd.Enum, before map[string][]string, c Controllable) {
	for name, enum := range enums {
		valuesBefore := before[name]
		values := enum.Options(c)
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
