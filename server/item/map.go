package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

type baseMap struct {
	// IsInit has unknown functionality (referring to the Minecraft Wiki).
	IsInit bool
	// NameIndex is the index of the map's name.
	NameIndex int32
	// DisplayPlayers controls whether the map displays player markers (depends on Decorations and TrackedObjects in the map data).
	DisplayPlayers bool

	data *world.ViewableMapData
}

// GetDimension ...
func (m *baseMap) GetDimension() world.Dimension {
	// TODO: Check where the data is stored if map is pesistent.
	return world.Overworld
}

// DecodeNBT ...
func (m *baseMap) DecodeNBT(data map[string]any) any {
	return m
}

// EncodeNBT ...
func (m *baseMap) EncodeNBT() map[string]any {
	return map[string]any{
		"map_is_init":         m.IsInit,
		"map_uuid":            m.data.MapID,
		"map_name_index":      m.NameIndex,
		"map_display_players": m.DisplayPlayers,
		"map_scale":           int32(m.data.Scale),
		"map_is_scaling":      m.IsScaling,
	}
}
