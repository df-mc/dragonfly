package block

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/block/encoder"
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
func (logEncoder) DecodeBlock(id string, meta int16, nbt []byte) encoder.Block {
	switch id {
	default:
		switch meta & 0x2 {
		default:
			return OakLog{Axis: axisFromInt16(meta >> 2)}
		case 1:
			return SpruceLog{Axis: axisFromInt16(meta >> 2)}
		case 2:
			return BirchLog{Axis: axisFromInt16(meta >> 2)}
		case 3:
			return JungleLog{Axis: axisFromInt16(meta >> 2)}
		}
	case "minecraft:log2":
		switch meta & 0x2 {
		default:
			return AcaciaLog{Axis: axisFromInt16(meta >> 2)}
		case 1:
			return DarkOakLog{Axis: axisFromInt16(meta >> 2)}
		}
	case "minecraft:stripped_oak_log":
		return OakLog{Stripped: true, Axis: axisFromInt16(meta)}
	case "minecraft:stripped_spruce_log":
		return SpruceLog{Stripped: true, Axis: axisFromInt16(meta)}
	case "minecraft:stripped_birch_log":
		return BirchLog{Stripped: true, Axis: axisFromInt16(meta)}
	case "minecraft:stripped_jungle_log":
		return JungleLog{Stripped: true, Axis: axisFromInt16(meta)}
	case "minecraft:stripped_acacia_log":
		return AcaciaLog{Stripped: true, Axis: axisFromInt16(meta)}
	case "minecraft:stripped_dark_oak_log":
		return DarkOakLog{Stripped: true, Axis: axisFromInt16(meta)}
	}
}

// EncodeBlock ...
func (logEncoder) EncodeBlock(b encoder.Block) (id string, meta int16, nbt []byte) {
	switch log := b.(type) {
	case OakLog:
		if log.Stripped {
			return "minecraft:stripped_oak_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log", axisToInt16(log.Axis) << 2, nil
	case SpruceLog:
		if log.Stripped {
			return "minecraft:stripped_spruce_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log", 1 | axisToInt16(log.Axis)<<2, nil
	case BirchLog:
		if log.Stripped {
			return "minecraft:stripped_birch_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log", 2 | axisToInt16(log.Axis)<<2, nil
	case JungleLog:
		if log.Stripped {
			return "minecraft:stripped_jungle_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log", 3 | axisToInt16(log.Axis)<<2, nil
	case AcaciaLog:
		if log.Stripped {
			return "minecraft:stripped_acacia_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log2", axisToInt16(log.Axis) << 2, nil
	case DarkOakLog:
		if log.Stripped {
			return "minecraft:stripped_dark_oak_log", axisToInt16(log.Axis), nil
		}
		return "minecraft:log2", 1 | axisToInt16(log.Axis)<<2, nil
	default:
		return // Never happens.
	}
}
