package world

import (
	"encoding/binary"
	"hash/fnv"
	"sort"
	"strings"
)

var bufNetworkhash []byte = make([]byte, 0xff)

func networkBlockHash(name string, properties map[string]any) uint32 {
	if name == "minecraft:unknown" {
		return 0xfffffffe // -2
	}

	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	bufNetworkhash = bufNetworkhash[:0]
	var data = bufNetworkhash
	writeString := func(str string) {
		data = binary.LittleEndian.AppendUint16(data, uint16(len(str)))
		data = append(data, []byte(str)...)
	}

	data = append(data, 10) // compound
	data = append(data, 0)
	data = append(data, 0)

	data = append(data, 8) // string
	writeString("name")
	writeString(name)

	data = append(data, 10) // compound
	writeString("states")
	for _, k := range keys {
		v := properties[k]
		switch v := v.(type) {
		case string:
			data = append(data, 8) // string
			writeString(k)
			writeString(v)

		case uint8:
			data = append(data, 1) // tagByte
			writeString(k)
			data = append(data, byte(v))
		case int8:
			data = append(data, 1) // tagByte
			writeString(k)
			data = append(data, byte(v))
		case bool:
			b := 0
			if v {
				b = 1
			}
			data = append(data, 1) // tagByte
			writeString(k)
			data = append(data, byte(b))

		case uint16:
			data = append(data, 2) // tagInt16
			writeString(k)
			data = binary.LittleEndian.AppendUint16(data, uint16(v))
		case int16:
			data = append(data, 2) // tagInt16
			writeString(k)
			data = binary.LittleEndian.AppendUint16(data, uint16(v))

		case uint32:
			data = append(data, 3) // tagInt32
			writeString(k)
			data = binary.LittleEndian.AppendUint32(data, uint32(v))
		case int32:
			data = append(data, 3) // tagInt32
			writeString(k)
			data = binary.LittleEndian.AppendUint32(data, uint32(v))
		default:
			panic("unhandled nbt type")
		}
	}
	data = append(data, 0) // end
	data = append(data, 0) // end

	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

func splitNamespace(identifier string) (ns, name string) {
	ns_name := strings.Split(identifier, ":")
	return ns_name[0], ns_name[len(ns_name)-1]
}
