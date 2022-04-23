package item

type MapInterface interface {
	GetMapID() int64
}

type baseMap struct {
	// IsInit has unknown functionality (referring to the Minecraft Wiki).
	IsInit bool
	// Uuid is the numeric identifier of the map (data) used in this item.
	// Sending packet.ClientBoundMapItemData can create or update map data at client side.
	Uuid int64
	// NameIndex is the index of the map's name.
	NameIndex int32
	// DisplayPlayers controls whether the map displays player markers (depends on Decorations and TrackedObjects in the map data).
	DisplayPlayers bool
	// Scale should be 0 to 4.
	Scale int32
	// IsScaling has unknown functionality (referring to the Minecraft Wiki).
	IsScaling bool
}

func (m baseMap) DecodeNBT(data map[string]any) any {
	return m
}

func (m baseMap) EncodeNBT() map[string]any {
	return map[string]any{
		"map_is_init":         m.IsInit,
		"map_uuid":            m.Uuid,
		"map_name_index":      m.NameIndex,
		"map_display_players": m.DisplayPlayers,
		"map_scale":           m.Scale,
		"map_is_scaling":      m.IsScaling,
	}
}

func (m baseMap) GetMapID() int64 {
	return m.Uuid
}

type FilledMap struct {
	baseMap
}

// EncodeItem ...
func (m FilledMap) EncodeItem() (name string, meta int16) {
	return "minecraft:filled_map", 0
}
