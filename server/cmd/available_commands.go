package cmd

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// BuildAvailableCommands builds an AvailableCommands packet for the commands passed.
// The input map may contain aliases. Only commands that the Source can execute are included.
func BuildAvailableCommands(commands map[string]Command, src Source) *packet.AvailableCommands {
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
		if run := c.Runnables(src); len(run) == 0 {
			continue
		}

		params := c.Params(src)
		overloads := make([]protocol.CommandOverload, len(params))

		aliasesIndex := uint32(math.MaxUint32)
		if len(c.Aliases()) > 0 {
			aliasesIndex = uint32(len(enumIndices))
			enumIndices[c.Name()+"Aliases"] = aliasesIndex
			enums = append(enums, commandEnum{Type: c.Name() + "Aliases", Options: c.Aliases()})
		}

		for i, params := range params {
			for _, paramInfo := range params {
				t, enum := valueToParamType(paramInfo, src)
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
	return pk
}

type commandEnum struct {
	Type    string
	Options []string
	Dynamic bool
}

// valueToParamType finds the command argument type of the value passed and returns it, in addition to creating an enum
// if applicable.
func valueToParamType(i ParamInfo, source Source) (t uint32, enum commandEnum) {
	switch i.Value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return protocol.CommandArgTypeInt, enum
	case float32, float64:
		return protocol.CommandArgTypeFloat, enum
	case string:
		return protocol.CommandArgTypeString, enum
	case Varargs:
		return protocol.CommandArgTypeRawText, enum
	case Target, []Target:
		return protocol.CommandArgTypeTarget, enum
	case bool:
		return 0, commandEnum{
			Type:    "bool",
			Options: []string{"true", "1", "false", "0"},
		}
	case mgl64.Vec3:
		return protocol.CommandArgTypePosition, enum
	case SubCommand:
		return 0, commandEnum{
			Type:    "SubCommand" + i.Name,
			Options: []string{i.Name},
		}
	}
	if enum, ok := i.Value.(Enum); ok {
		return 0, commandEnum{
			Type:    enum.Type(),
			Options: enum.Options(source),
			Dynamic: true,
		}
	}
	return protocol.CommandArgTypeValue, enum
}
