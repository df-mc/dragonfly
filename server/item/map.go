package item

import (
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
	m.IsInit, _ = data["map_is_init"].(bool)
	m.NameIndex, _ = data["map_name_index"].(int32)
	m.DisplayPlayers, _ = data["map_display_players"].(bool)

	if id, ok := data["map_uuid"].(int64); ok {
		id = id
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
