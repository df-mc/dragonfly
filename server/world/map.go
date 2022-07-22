package world

import (
	"image/color"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

var (
	mapDataMu/*, persistedMapDataMu*/ sync.RWMutex
	mapData = map[int64]*ViewableMapData{}
	// persistedMapData              = map[int64]*ViewableMapData{}
)

func NewMapData() *ViewableMapData {
	mapDataMu.Lock()
	defer mapDataMu.Unlock()

	d := &ViewableMapData{
		mapID:   int64(len(mapData)),
		viewers: map[MapDataViewer]struct{}{},
		data: MapData{
			TrackEntities: map[Entity]struct{}{},
			TrackBlocks:   map[cube.Pos]struct{}{},
		},
	}
	mapData[d.mapID] = d

	return d
}

type MapData struct {
	Pixels        [][]color.RGBA
	TrackEntities map[Entity]struct{}
	TrackBlocks   map[cube.Pos]struct{}
	// Scale should be 0 to 4. TODO: verify.
	Scale byte
	// Locked map should not be affected by world content (block) changes.
	Locked bool
}

type MapDataViewer interface {
	ViewMapDataChange(updateFlag uint32, mapID int64, pixelsChunk MapPixelsChunk, data *ViewableMapData)
}

type ViewableMapData struct {
	mapID/*, persistentID*/ int64
	world *World

	viewersMu sync.RWMutex
	viewers   map[MapDataViewer]struct{}

	pixelsMu, trackEntitiesMu, trackBlocksMu sync.RWMutex

	data MapData
}

type MapPixelsChunk struct {
	XOffset, YOffset, Height, Width int32
	Pixels                          [][]color.RGBA
}

// ChangePixels broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagTexture.
// Returns the calculated offsets and lengths.
func (d *ViewableMapData) ChangePixels(pixels [][]color.RGBA) MapPixelsChunk {
	d.pixelsMu.Lock()
	defer d.pixelsMu.Unlock()

	old := d.data.Pixels
	pixelsChunk := MapPixelsChunk{Pixels: pixels, Height: int32(len(pixels))}
	for y, row := range pixels {
		for x, pixel := range row {
			if old[y][x] != pixel {
				if pixelsChunk.YOffset == 0 {
					pixelsChunk.YOffset = int32(y)
				}
				if pixelsChunk.XOffset == 0 {
					pixelsChunk.XOffset = int32(x)
				}
			}

			old[y][x] = pixel
		}

		rowLen := int32(len(row))
		if rowLen > pixelsChunk.Width {
			pixelsChunk.Width = rowLen
		}
	}

	d.change(packet.MapUpdateFlagTexture, pixelsChunk)
	return pixelsChunk
}

// AddTrackEntity broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) AddTrackEntity(e Entity) {
	d.trackEntitiesMu.Lock()
	defer d.trackEntitiesMu.Unlock()

	s := struct{}{}
	d.data.TrackEntities[e] = s
	d.change(packet.MapUpdateFlagDecoration, MapPixelsChunk{})
}

// RemoveTrackEntity broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) RemoveTrackEntity(e Entity) {
	d.trackEntitiesMu.Lock()
	defer d.trackEntitiesMu.Unlock()

	delete(d.data.TrackEntities, e)
	d.change(packet.MapUpdateFlagDecoration, MapPixelsChunk{})
}

// AddTrackBlock broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) AddTrackBlock(pos cube.Pos) {
	d.trackBlocksMu.Lock()
	defer d.trackBlocksMu.Unlock()

	s := struct{}{}
	d.data.TrackBlocks[pos] = s
	d.change(packet.MapUpdateFlagDecoration, MapPixelsChunk{})
}

// RemoveTrackBlock broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) RemoveTrackBlock(pos cube.Pos) {
	d.trackBlocksMu.Lock()
	defer d.trackBlocksMu.Unlock()

	delete(d.data.TrackBlocks, pos)
	d.change(packet.MapUpdateFlagDecoration, MapPixelsChunk{})
}

// MapData ...
func (d *ViewableMapData) MapData() MapData {
	d.pixelsMu.RLock()
	d.trackEntitiesMu.RLock()
	d.trackBlocksMu.RLock()
	defer d.pixelsMu.RUnlock()
	defer d.trackEntitiesMu.RUnlock()
	defer d.trackBlocksMu.RUnlock()

	return d.data
}

func (d *ViewableMapData) change(updateFlag uint32, pixelsChunk MapPixelsChunk) {
	d.viewersMu.RLock()
	defer d.viewersMu.RUnlock()

	for viewer := range d.viewers {
		viewer.ViewMapDataChange(updateFlag, d.mapID, pixelsChunk, d)
	}

	// TODO: async save().
}

// AddViewer ...
func (d *ViewableMapData) AddViewer(v MapDataViewer) {
	d.viewersMu.Lock()
	defer d.viewersMu.Unlock()

	s := struct{}{}
	d.viewers[v] = s
}

// RemoveViewer ...
func (d *ViewableMapData) RemoveViewer(v MapDataViewer) {
	d.viewersMu.Lock()
	defer d.viewersMu.Unlock()

	delete(d.viewers, v)
}

// EncodeNBT provides value of field map ID, scale and is scaling for item.BaseMap.EncodeNBT().
// Returns empty map if nil.
func (d *ViewableMapData) EncodeItemNBT() map[string]any {
	if d == nil {
		return map[string]any{}
	}

	data := d.MapData()
	return map[string]any{
		"map_uuid":       d.mapID,
		"map_scale":      data.Scale,
		"map_is_scaling": data.Scale > 0,
	}
}

// World returns the map's belonging world.
// This is for the map's dimension.
// And filter tracked blocks that are not in the same world as viewer.
func (d *ViewableMapData) World() *World {
	return d.world
}

// MapIDEquals ...
func (d *ViewableMapData) MapIDEquals(mapID int64) bool {
	return d.mapID == mapID
}
