package gen

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

const (
	structureNBTTagEnd byte = iota
	structureNBTTagByte
	structureNBTTagShort
	structureNBTTagInt
	structureNBTTagLong
	structureNBTTagFloat
	structureNBTTagDouble
	structureNBTTagByteArray
	structureNBTTagString
	structureNBTTagList
	structureNBTTagCompound
	structureNBTTagIntArray
	structureNBTTagLongArray
)

const structureTemplateMaxDepth = 4096

type structureTemplateNBTReader struct {
	r io.Reader
}

func decodeStructureTemplateFallback(data []byte) (StructureTemplate, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return StructureTemplate{}, err
	}
	defer reader.Close()

	nbtReader := structureTemplateNBTReader{r: reader}
	rootTag, err := nbtReader.readByte()
	if err != nil {
		return StructureTemplate{}, err
	}
	if rootTag != structureNBTTagCompound {
		return StructureTemplate{}, fmt.Errorf("expected root compound tag, got %d", rootTag)
	}
	if _, err := nbtReader.readString(); err != nil {
		return StructureTemplate{}, err
	}
	rootValue, err := nbtReader.readTagPayload(rootTag, 0)
	if err != nil {
		return StructureTemplate{}, err
	}
	root, ok := rootValue.(map[string]any)
	if !ok {
		return StructureTemplate{}, fmt.Errorf("expected root compound payload")
	}

	var out StructureTemplate
	sizeValues := anySlice(root["size"])
	if len(sizeValues) >= 3 {
		out.Size = [3]int{anyInt(sizeValues[0]), anyInt(sizeValues[1]), anyInt(sizeValues[2])}
	}

	paletteValues := anySlice(root["palette"])
	if len(paletteValues) == 0 {
		if palettes := anySlice(root["palettes"]); len(palettes) > 0 {
			paletteValues = anySlice(palettes[0])
		}
	}
	out.Palette = make([]StructureTemplateBlockState, 0, len(paletteValues))
	for _, value := range paletteValues {
		entry := anyMap(value)
		out.Palette = append(out.Palette, StructureTemplateBlockState{
			Name:       anyString(entry["Name"]),
			Properties: anyMap(entry["Properties"]),
		})
	}

	blockValues := anySlice(root["blocks"])
	out.Blocks = make([]StructureTemplateBlock, 0, len(blockValues))
	for _, value := range blockValues {
		entry := anyMap(value)
		posValues := anySlice(entry["pos"])
		if len(posValues) < 3 {
			continue
		}
		out.Blocks = append(out.Blocks, StructureTemplateBlock{
			Pos:   [3]int{anyInt(posValues[0]), anyInt(posValues[1]), anyInt(posValues[2])},
			State: anyInt(entry["state"]),
			NBT:   anyMap(entry["nbt"]),
		})
	}
	return out, nil
}

func (r structureTemplateNBTReader) readTagPayload(tag byte, depth int) (any, error) {
	if depth >= structureTemplateMaxDepth {
		return nil, fmt.Errorf("structure template NBT exceeded depth %d", structureTemplateMaxDepth)
	}
	switch tag {
	case structureNBTTagEnd:
		return nil, nil
	case structureNBTTagByte:
		v, err := r.readByte()
		return int64(int8(v)), err
	case structureNBTTagShort:
		v, err := r.readInt16()
		return int64(v), err
	case structureNBTTagInt:
		v, err := r.readInt32()
		return int64(v), err
	case structureNBTTagLong:
		v, err := r.readInt64()
		return v, err
	case structureNBTTagFloat:
		v, err := r.readFloat32()
		return float64(v), err
	case structureNBTTagDouble:
		return r.readFloat64()
	case structureNBTTagByteArray:
		length, err := r.readInt32()
		if err != nil {
			return nil, err
		}
		return r.readRaw(int(length))
	case structureNBTTagString:
		return r.readString()
	case structureNBTTagList:
		elemTag, err := r.readByte()
		if err != nil {
			return nil, err
		}
		length, err := r.readInt32()
		if err != nil {
			return nil, err
		}
		if length < 0 {
			return nil, fmt.Errorf("negative list length %d", length)
		}
		out := make([]any, 0, length)
		for i := int32(0); i < length; i++ {
			value, err := r.readTagPayload(elemTag, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, value)
		}
		return out, nil
	case structureNBTTagCompound:
		out := make(map[string]any)
		for {
			nextTag, err := r.readByte()
			if err != nil {
				return nil, err
			}
			if nextTag == structureNBTTagEnd {
				return out, nil
			}
			name, err := r.readString()
			if err != nil {
				return nil, err
			}
			value, err := r.readTagPayload(nextTag, depth+1)
			if err != nil {
				return nil, err
			}
			out[name] = value
		}
	case structureNBTTagIntArray:
		length, err := r.readInt32()
		if err != nil {
			return nil, err
		}
		if length < 0 {
			return nil, fmt.Errorf("negative int array length %d", length)
		}
		out := make([]int32, length)
		for i := int32(0); i < length; i++ {
			value, err := r.readInt32()
			if err != nil {
				return nil, err
			}
			out[i] = value
		}
		return out, nil
	case structureNBTTagLongArray:
		length, err := r.readInt32()
		if err != nil {
			return nil, err
		}
		if length < 0 {
			return nil, fmt.Errorf("negative long array length %d", length)
		}
		out := make([]int64, length)
		for i := int32(0); i < length; i++ {
			value, err := r.readInt64()
			if err != nil {
				return nil, err
			}
			out[i] = value
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported NBT tag %d", tag)
	}
}

func (r structureTemplateNBTReader) readRaw(length int) ([]byte, error) {
	if length < 0 {
		return nil, fmt.Errorf("negative byte length %d", length)
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(r.r, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (r structureTemplateNBTReader) readByte() (byte, error) {
	var value [1]byte
	_, err := io.ReadFull(r.r, value[:])
	return value[0], err
}

func (r structureTemplateNBTReader) readInt16() (int16, error) {
	var value int16
	err := binary.Read(r.r, binary.BigEndian, &value)
	return value, err
}

func (r structureTemplateNBTReader) readInt32() (int32, error) {
	var value int32
	err := binary.Read(r.r, binary.BigEndian, &value)
	return value, err
}

func (r structureTemplateNBTReader) readInt64() (int64, error) {
	var value int64
	err := binary.Read(r.r, binary.BigEndian, &value)
	return value, err
}

func (r structureTemplateNBTReader) readFloat32() (float32, error) {
	var bits uint32
	if err := binary.Read(r.r, binary.BigEndian, &bits); err != nil {
		return 0, err
	}
	return math.Float32frombits(bits), nil
}

func (r structureTemplateNBTReader) readFloat64() (float64, error) {
	var bits uint64
	if err := binary.Read(r.r, binary.BigEndian, &bits); err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

func (r structureTemplateNBTReader) readString() (string, error) {
	length, err := r.readInt16()
	if err != nil {
		return "", err
	}
	if length < 0 {
		return "", fmt.Errorf("negative string length %d", length)
	}
	data, err := r.readRaw(int(length))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func anySlice(value any) []any {
	switch v := value.(type) {
	case []any:
		return v
	case []int32:
		out := make([]any, len(v))
		for i, value := range v {
			out[i] = int64(value)
		}
		return out
	case []int64:
		out := make([]any, len(v))
		for i, value := range v {
			out[i] = value
		}
		return out
	default:
		return nil
	}
}

func anyMap(value any) map[string]any {
	switch v := value.(type) {
	case map[string]any:
		return v
	default:
		return nil
	}
}

func anyString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func anyInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}
