package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

type MapInterface interface {
	GetMapID() int64
	// UpdateData updates the map's tracked entites, blocks or a chunk (with offset) of pixels.
	UpdateData(MapDataUpdate)
	AddViewer(MapDataViewer)
	// IsPresisted refers to whether or not the map and its data will be saved to disk.
	IsPersisted() bool // TODO: Reset map item ID if not persisted. Because all data is gone on server's next life cycle.
}

type MapDataViewer interface {
	ViewMapDataChange(MapDataUpdate)
}

// MapDataUpdate is world.MapData but with X and Y offsets.
// Therefore, not all pixels need to be resent when updating a chunk of pixels.
type MapDataUpdate struct {
	XOffset, YOffset int32

	world.MapData
}

type baseMap struct {
	// IsInit has unknown functionality (referring to the Minecraft Wiki).
	// However in Dragonfly, this indicates whether a player has access to the persisted map data corresponding to Uuid.
	IsInit bool
	// Uuid is the numeric identifier of the map's linked MapData.
	Uuid int64
	// NameIndex is the index of the map's name.
	NameIndex int32
	// DisplayPlayers controls whether the map displays player markers (depends on Decorations and TrackedObjects in the map data).
	DisplayPlayers bool
	// Scale should be 0 to 4.
	Scale int32
	// IsScaling has unknown functionality (referring to the Minecraft Wiki).
	IsScaling bool

	viewers   map[MapDataViewer]struct{}
	data      *world.MapData
	persisted bool
}

// DecodeNBT ...
func (m baseMap) DecodeNBT(data map[string]any) any {
	return m
}

// EncodeNBT ...
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

// GetMapID ...
func (m baseMap) GetMapID() int64 {
	return m.Uuid
}

// UpdateData ...
func (m baseMap) UpdateData(u MapDataUpdate) {
	if m.data == nil || m.viewers == nil {
		return
	}

	for viewer := range m.viewers {
		viewer.ViewMapDataChange(u)
	}

	// TODO: Update to disk if map is persistent.
}

// AddViewer ...
func (m *baseMap) AddViewer(v MapDataViewer) {
	s := struct{}{}
	if m.viewers == nil {
		m.viewers = map[MapDataViewer]struct{}{v: s}
	} else {
		m.viewers[v] = s
	}
}

// IsPersisted ...
func (m baseMap) IsPersisted() bool {
	return m.persisted
}
