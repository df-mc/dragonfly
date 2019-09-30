package encoder

import (
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/block"
)

// logEncoder implements the encoding and decoding of log type blocks.
type logEncoder struct{}

// BlocksHandled ...
func (logEncoder) BlocksHandled() []string {
	return []string{
		"minecraft:log", "minecraft:log2", "minecraft:stripped_oak_log", "minecraft:stripped_spruce_log",
		"minecraft:stripped_birch_log", "minecraft:stripped_jungle_log", "minecraft:stripped_acacia_log",
		"minecraft:stripped_dark_oak_log",
	}
}

// DecodeBlock ...
func (logEncoder) DecodeBlock(id string, meta int16, nbt []byte) Block {
	switch id {
	default:
		switch meta & 0x2 {
		default:
			return block.OakLog{Axis: axisFromInt16(meta >> 2)}
		case 1:
			return block.SpruceLog{Axis: axisFromInt16(meta >> 2)}
		case 2:
			return block.BirchLog{Axis: axisFromInt16(meta >> 2)}
		case 3:
			return block.JungleLog{Axis: axisFromInt16(meta >> 2)}
		}
	case "minecraft:log2":
		switch meta & 0x2 {
		default:
			return block.AcaciaLog{Axis: axisFromInt16(meta >> 2)}
		case 1:
			return block.DarkOakLog{Axis: axisFromInt16(meta >> 2)}
		}
	case "minecraft:stripped_oak_log":
		return block.OakLog{Stripped: true, Axis: axisFromInt16(meta)}
	case "minecraft:stripped_spruce_log":
		return block.SpruceLog{Stripped: true, Axis: axisFromInt16(meta)}
	case "minecraft:stripped_birch_log":
		return block.BirchLog{Stripped: true, Axis: axisFromInt16(meta)}
	case "minecraft:stripped_jungle_log":
		return block.JungleLog{Stripped: true, Axis: axisFromInt16(meta)}
	case "minecraft:stripped_acacia_log":
		return block.AcaciaLog{Stripped: true, Axis: axisFromInt16(meta)}
	case "minecraft:stripped_dark_oak_log":
		return block.DarkOakLog{Stripped: true, Axis: axisFromInt16(meta)}
	}
}

// EncodeBlock ...
func (logEncoder) EncodeBlock(b Block) (id string, meta int16, nbt []byte) {
	switch log := b.(type) {
	case block.OakLog:
		if log.Stripped {
			return "minecraft:stripped_oak_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log", axisToInt16(log.Axis) << 2, nil
	case block.SpruceLog:
		if log.Stripped {
			return "minecraft:stripped_spruce_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log", 1 | axisToInt16(log.Axis)<<2, nil
	case block.BirchLog:
		if log.Stripped {
			return "minecraft:stripped_birch_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log", 2 | axisToInt16(log.Axis)<<2, nil
	case block.JungleLog:
		if log.Stripped {
			return "minecraft:stripped_jungle_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log", 3 | axisToInt16(log.Axis)<<2, nil
	case block.AcaciaLog:
		if log.Stripped {
			fmt.Println(log.Axis, axisToInt16(log.Axis))
			return "minecraft:stripped_acacia_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log2", axisToInt16(log.Axis) << 2, nil
	case block.DarkOakLog:
		if log.Stripped {
			return "minecraft:stripped_dark_oak_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log2", 1 | axisToInt16(log.Axis)<<2, nil
	default:
		return // Never happens.
	}
}
