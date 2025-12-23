package session

import (
	"math"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/go-gl/mathgl/mgl64"
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

	pk := &packet.AvailableCommands{}
	var enums []commandEnum
	enumIndices := map[string]uint32{}

	var dynamicEnums []commandEnum
	dynamicEnumIndices := map[string]uint32{}

	suffixIndices := map[string]uint32{}

	for alias, c := range commands {
		if c.Name() != alias {
			// Don't add duplicate entries for aliases.
			continue
		}
		if run := c.Runnables(co); len(run) > 0 {
			m[alias] = run
		} else {
			continue
		}

		params := c.Params(co)
		overloads := make([]protocol.CommandOverload, len(params))

		aliasesIndex := uint32(math.MaxUint32)
		if len(c.Aliases()) > 0 {
			aliasesIndex = uint32(len(enumIndices))
			enumIndices[c.Name()+"Aliases"] = aliasesIndex
			enums = append(enums, commandEnum{Type: c.Name() + "Aliases", Options: c.Aliases()})
		}

		for i, params := range params {
			for _, paramInfo := range params {
				t, enum := valueToParamType(paramInfo, co)
				t |= protocol.CommandArgValid
				suffix := paramInfo.Suffix

				opt := byte(0)
				if _, ok := paramInfo.Value.(bool); ok {
					opt |= protocol.ParamOptionCollapseEnum
				}
				if len(enum.Options) > 0 || enum.Type != "" {
					if !enum.Dynamic {
						index, ok := enumIndices[enum.Type]
						if !ok {
							index = uint32(len(enums))
							enumIndices[enum.Type] = index
							enums = append(enums, enum)
						}
						t |= protocol.CommandArgEnum | index
					} else {
						index, ok := dynamicEnumIndices[enum.Type]
						if !ok {
							index = uint32(len(dynamicEnums))
							dynamicEnumIndices[enum.Type] = index
							dynamicEnums = append(dynamicEnums, enum)
						}
						t |= protocol.CommandArgSoftEnum | index
					}
				}
				if suffix != "" {
					index, ok := suffixIndices[suffix]
					if !ok {
						index = uint32(len(pk.Suffixes))
						suffixIndices[suffix] = index
						pk.Suffixes = append(pk.Suffixes, suffix)
					}
					t |= protocol.CommandArgSuffixed | index
				}
				overloads[i].Parameters = append(overloads[i].Parameters, protocol.CommandParameter{
					Name:     paramInfo.Name,
					Type:     t,
					Optional: paramInfo.Optional,
					Options:  opt,
				})
			}
		}
		pk.Commands = append(pk.Commands, protocol.Command{
			Name:            c.Name(),
			Description:     c.Description(),
			AliasesOffset:   aliasesIndex,
			PermissionLevel: protocol.CommandPermissionLevelAny,
			Overloads:       overloads,
		})
	}
	pk.DynamicEnums = make([]protocol.DynamicEnum, 0, len(dynamicEnums))
	for _, e := range dynamicEnums {
		pk.DynamicEnums = append(pk.DynamicEnums, protocol.DynamicEnum{Type: e.Type, Values: e.Options})
	}

	enumValueIndices := make(map[string]uint32, len(enums)*3)
	pk.EnumValues = make([]string, 0, len(enumValueIndices))

	pk.Enums = make([]protocol.CommandEnum, 0, len(enums))
	for _, enum := range enums {
		protoEnum := protocol.CommandEnum{Type: enum.Type}
		for _, opt := range enum.Options {
			index, ok := enumValueIndices[opt]
			if !ok {
				index = uint32(len(pk.EnumValues))
				enumValueIndices[opt] = index
				pk.EnumValues = append(pk.EnumValues, opt)
			}
			protoEnum.ValueIndices = append(protoEnum.ValueIndices, index)
		}
		pk.Enums = append(pk.Enums, protoEnum)
	}
	s.writePacket(pk)
	return m
}

type commandEnum struct {
	Type    string
	Options []string
	Dynamic bool
}

// valueToParamType finds the command argument type of the value passed and returns it, in addition to creating
// an enum if applicable.
func valueToParamType(i cmd.ParamInfo, source cmd.Source) (t uint32, enum commandEnum) {
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
		return 0, commandEnum{
			Type:    "bool",
			Options: []string{"true", "1", "false", "0"},
		}
	case mgl64.Vec3:
		return protocol.CommandArgTypePosition, enum
	case cmd.SubCommand:
		return 0, commandEnum{
			Type:    "SubCommand" + i.Name,
			Options: []string{i.Name},
		}
	}
	if enum, ok := i.Value.(cmd.Enum); ok {
		return 0, commandEnum{
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
