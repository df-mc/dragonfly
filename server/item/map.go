package item

import (
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
)

type MapItem interface {
	GetBaseMap() BaseMap
}

type BaseMap struct {
	// IsInit has unknown functionality (referring to the Minecraft Wiki).
	IsInit bool
	// NameIndex is the index of the map's name.
	NameIndex int32
	// DisplayPlayers controls whether the map displays player markers (depends on Decorations and TrackedObjects in the map data).
	DisplayPlayers bool

	*world.ViewableMapData
}

// DecodeNBT ...
func (m BaseMap) DecodeNBT(data map[string]any) any {
	m.IsInit = nbtconv.Map[bool](data, "map_is_init")
	m.NameIndex = nbtconv.Map[int32](data, "map_name_index")
	m.DisplayPlayers = nbtconv.Map[bool](data, "map_display_players")

	if id, ok := data["map_uuid"].(int64); ok {
		// TODO: load map data.
	}

	return m
}

// EncodeNBT ...
func (m BaseMap) EncodeNBT() map[string]any {
	data := m.ViewableMapData.EncodeItemNBT()
	data["map_is_init"] = m.IsInit
	data["map_name_index"] = m.NameIndex
	data["map_display_players"] = m.DisplayPlayers
	return data
}
