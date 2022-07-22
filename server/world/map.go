package world

import (
	"image/color"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type MapData struct {
	Pixels        [][]color.RGBA
	TrackEntities map[Entity]struct{}
	TrackBlocks   map[cube.Pos]struct{}
	// Scale should be 0 to 4.
	Scale byte
	// Locked map should not be affected by world content (block) changes.
	Locked bool
}

type MapDataViewer interface {
	ViewMapDataChange(*ViewableMapData)
}

type ViewableMapData struct {
	// MapID is the unique identifier of a map data. For both runtime and when in disk.
	MapID int64
	world *World

	viewersMu sync.RWMutex
	viewers   map[MapDataViewer]struct{}

	pixelsMu, trackEntitiesMu, trackBlocksMu, scaleMu, lockedMu sync.RWMutex

	data MapData
}

// ChangePixels broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagTexture.
func (d *ViewableMapData) ChangePixels(pixels [][]color.RGBA) {
	d.pixelsMu.Lock()
	defer d.pixelsMu.Unlock()

	d.data.Pixels = pixels
	d.change(packet.MapUpdateFlagTexture)
}

// AddTrackEntity broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) AddTrackEntity(e Entity) {
	d.trackEntitiesMu.Lock()
	defer d.trackEntitiesMu.Unlock()

	s := struct{}{}
	if d.data.TrackEntities == nil {
		d.data.TrackEntities = map[Entity]struct{}{e: s}
	} else {
		d.data.TrackEntities[e] = s
	}
	d.change(packet.MapUpdateFlagDecoration)
}

// RemoveTrackEntity broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) RemoveTrackEntity(e Entity) {
	d.trackEntitiesMu.Lock()
	defer d.trackEntitiesMu.Unlock()

	if d.data.TrackEntities != nil {
		delete(d.data.TrackEntities, e)
		d.change(packet.MapUpdateFlagDecoration)
	}
}

// AddTrackBlock broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) AddTrackBlock(pos cube.Pos) {
	d.trackBlocksMu.Lock()
	defer d.trackBlocksMu.Unlock()

	s := struct{}{}
	if d.data.TrackBlocks == nil {
		d.data.TrackBlocks = map[cube.Pos]struct{}{pos: s}
	} else {
		d.data.TrackBlocks[pos] = s
	}
	d.change(packet.MapUpdateFlagDecoration)
}

// RemoveTrackBlock broadcast *packet.ClientBoundMapItemData to viewers with packet.MapUpdateFlagDecoration.
func (d *ViewableMapData) RemoveTrackBlock(pos cube.Pos) {
	d.trackBlocksMu.Lock()
	defer d.trackBlocksMu.Unlock()

	if d.data.TrackBlocks != nil {
		delete(d.data.TrackBlocks, pos)
		d.change(packet.MapUpdateFlagDecoration)
	}
}

// SetScale does not broadcast *packet.ClientBoundMapItemData.
// Scale of existed map data should not be changed, please create a new map data instead.
// Scale should be 0 to 4.
func (d *ViewableMapData) SetScale(scale byte) {
	d.scaleMu.Lock()
	defer d.scaleMu.Unlock()

	d.data.Scale = scale
}

// SetLocked does not broadcast *packet.ClientBoundMapItemData.
// Lock status of existed map data should not be changed, please create a new map data instead.
func (d *ViewableMapData) SetLocked(locked bool) {
	d.lockedMu.Lock()
	defer d.lockedMu.Unlock()

	d.data.Locked = locked
}

func (d *ViewableMapData) change(updateFlag byte) {
	d.broadcast(updateFlag)
	d.save()
}

func (d *ViewableMapData) broadcast(updateFlag byte) {
	d.viewersMu.RLock()
	defer d.viewersMu.RUnlock()

	for viewer := range d.viewers {
		viewer.ViewMapDataChange(d)
	}
}

func (d *ViewableMapData) save() {
	// TODO: save()
}

// AddViewer ...
func (m *ViewableMapData) AddViewer(v MapDataViewer) {
	m.viewersMu.Lock()
	defer m.viewersMu.Unlock()

	s := struct{}{}
	if m.viewers == nil {
		m.viewers = map[MapDataViewer]struct{}{v: s}
	} else {
		m.viewers[v] = s
	}
}

// RemoveViewer ...
func (m *ViewableMapData) RemoveViewer(v MapDataViewer) {
	m.viewersMu.Lock()
	defer m.viewersMu.Unlock()

	if m.viewers != nil {
		delete(m.viewers, v)
	}
}
