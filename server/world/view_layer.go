package world

import (
	"maps"
	"slices"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// layer stores the appearance overrides that a ViewLayer applies to an entity.
type layer struct {
	nameTag    *string
	scoreTag   *string
	visibility VisibilityLevel
}

// ViewLayerUpdater handles immediate updates after a ViewLayer changes how an entity is viewed.
type ViewLayerUpdater interface {
	// ViewLayerEntityChanged handles an entity whose view-layer overrides changed.
	ViewLayerEntityChanged(entity Entity)
}

// ViewLayerBlockUpdater may be implemented by a ViewLayerUpdater to handle immediate updates after a
// ViewLayer changes how a block is viewed.
type ViewLayerBlockUpdater interface {
	// ViewLayerBlockChanged handles a block whose view-layer override changed.
	ViewLayerBlockChanged(tx *Tx, pos cube.Pos)
}

type viewLayerViewer interface {
	ViewLayer() *ViewLayer
}

// ViewLayer holds overrides for how entities are viewed by a single viewer. It allows entities to be
// viewed differently by different players, such as with a different name tag or visibility state.
// Block overrides are scoped by world.
type ViewLayer struct {
	mu            sync.RWMutex
	entities      map[*EntityHandle]layer
	blocksByWorld map[*World]map[ChunkPos]map[cube.Pos]Block
	updater       ViewLayerUpdater
}

// NewViewLayer returns a new ViewLayer.
func NewViewLayer(updater ViewLayerUpdater) *ViewLayer {
	return &ViewLayer{
		entities:      map[*EntityHandle]layer{},
		blocksByWorld: map[*World]map[ChunkPos]map[cube.Pos]Block{},
		updater:       updater,
	}
}

// Entities returns the handles of all entities with overrides in the view layer.
func (v *ViewLayer) Entities() []*EntityHandle {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return slices.Collect(maps.Keys(v.entities))
}

// ViewNameTag overwrites the public name tag of the entity and allows this ViewLayer to view a different name tag.
// Passing an empty name tag removes the name tag for this ViewLayer.
func (v *ViewLayer) ViewNameTag(entity Entity, nameTag string) {
	v.update(entity, func(l *layer) {
		l.nameTag = &nameTag
	})
}

// ViewPublicNameTag removes the name tag override from the entity, causing the public name tag to be
// viewed again.
func (v *ViewLayer) ViewPublicNameTag(entity Entity) {
	v.update(entity, func(l *layer) {
		l.nameTag = nil
	})
}

// NameTag returns the overwritten name tag of the entity and whether an override was set.
func (v *ViewLayer) NameTag(entity Entity) (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	nameTag := v.entities[entity.H()].nameTag
	if nameTag == nil {
		return "", false
	}
	return *nameTag, true
}

// ViewScoreTag overwrites the public score tag of the entity and allows this ViewLayer to view a different score tag.
// Passing an empty score tag removes the score tag for this ViewLayer.
func (v *ViewLayer) ViewScoreTag(entity Entity, scoreTag string) {
	v.update(entity, func(l *layer) {
		l.scoreTag = &scoreTag
	})
}

// ViewPublicScoreTag removes the score tag override from the entity, causing the public score tag to be
// viewed again.
func (v *ViewLayer) ViewPublicScoreTag(entity Entity) {
	v.update(entity, func(l *layer) {
		l.scoreTag = nil
	})
}

// ScoreTag returns the overwritten score tag of the entity and whether an override was set.
func (v *ViewLayer) ScoreTag(entity Entity) (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	scoreTag := v.entities[entity.H()].scoreTag
	if scoreTag == nil {
		return "", false
	}
	return *scoreTag, true
}

// ViewVisibility overwrites the public visibility of the entity and allows this ViewLayer to view
// this entity as (in)visible depending on the VisibilityLevel.
func (v *ViewLayer) ViewVisibility(entity Entity, level VisibilityLevel) {
	v.update(entity, func(l *layer) {
		l.visibility = level
	})
}

// Visibility returns the visibility of the entity.
func (v *ViewLayer) Visibility(entity Entity) VisibilityLevel {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.entities[entity.H()].visibility
}

// ViewBlock overwrites the public block at the position passed in the transaction's world for this ViewLayer. Liquid or waterlogged
// state at layer 1 is not represented for overrides. Passing nil removes the block override, causing the
// public block to be viewed again.
func (v *ViewLayer) ViewBlock(tx *Tx, pos cube.Pos, b Block) {
	w := tx.World()
	v.mu.Lock()
	chunkPos := ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
	blocksByChunk := v.blocksByWorld[w]
	if b == nil {
		delete(blocksByChunk[chunkPos], pos)
		if len(blocksByChunk[chunkPos]) == 0 {
			delete(blocksByChunk, chunkPos)
		}
		if len(blocksByChunk) == 0 {
			delete(v.blocksByWorld, w)
		}
	} else {
		if blocksByChunk == nil {
			blocksByChunk = map[ChunkPos]map[cube.Pos]Block{}
			v.blocksByWorld[w] = blocksByChunk
		}
		if blocksByChunk[chunkPos] == nil {
			blocksByChunk[chunkPos] = map[cube.Pos]Block{}
		}
		blocksByChunk[chunkPos][pos] = b
	}
	v.mu.Unlock()

	v.refreshBlock(tx, pos)
}

// ViewPublicBlock removes the block override at the position passed in the transaction's world, causing the public block to be viewed again.
func (v *ViewLayer) ViewPublicBlock(tx *Tx, pos cube.Pos) {
	v.ViewBlock(tx, pos, nil)
}

// Block returns the overwritten block at the position passed in the world passed and whether an override was set.
func (v *ViewLayer) Block(w *World, pos cube.Pos) (Block, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	b, ok := v.blocksByWorld[w][ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}][pos]
	return b, ok
}

// Blocks returns all block overrides in the world passed.
func (v *ViewLayer) Blocks(w *World) map[cube.Pos]Block {
	v.mu.RLock()
	defer v.mu.RUnlock()

	blocks := make(map[cube.Pos]Block)
	for _, chunkBlocks := range v.blocksByWorld[w] {
		maps.Copy(blocks, chunkBlocks)
	}
	return blocks
}

// ChunkBlocks returns all block overrides in a chunk in the world passed.
func (v *ViewLayer) ChunkBlocks(w *World, pos ChunkPos) map[cube.Pos]Block {
	v.mu.RLock()
	defer v.mu.RUnlock()

	blocks := v.blocksByWorld[w][pos]
	if len(blocks) == 0 {
		return nil
	}
	return maps.Clone(blocks)
}

// Remove removes all overrides for the entity from the ViewLayer.
func (v *ViewLayer) Remove(entity Entity) {
	if v.remove(entity) {
		v.refresh(entity)
	}
}

// remove removes all overrides for the entity from the ViewLayer without refreshing entity metadata. It returns
// whether any overrides were removed.
func (v *ViewLayer) remove(entity Entity) bool {
	handle := entity.H()

	v.mu.Lock()
	_, ok := v.entities[handle]
	delete(v.entities, handle)
	v.mu.Unlock()
	return ok
}

// update applies a mutation to the entity's layer, removes the entry if no overrides remain, and refreshes
// the entity for the layer's viewer.
func (v *ViewLayer) update(entity Entity, update func(*layer)) {
	handle := entity.H()

	v.mu.Lock()
	l := v.entities[handle]
	update(&l)
	if l.empty() {
		delete(v.entities, handle)
	} else {
		v.entities[handle] = l
	}
	v.mu.Unlock()

	v.refresh(entity)
}

// Close closes the view layer.
func (v *ViewLayer) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	clear(v.entities)
	clear(v.blocksByWorld)
	return nil
}

// empty checks if the layer does not override any public entity metadata.
func (l layer) empty() bool {
	return l.nameTag == nil && l.scoreTag == nil && l.visibility == PublicVisibility()
}

func (v *ViewLayer) refresh(entity Entity) {
	if v.updater != nil {
		v.updater.ViewLayerEntityChanged(entity)
	}
}

func (v *ViewLayer) refreshBlock(tx *Tx, pos cube.Pos) {
	if updater, ok := v.updater.(ViewLayerBlockUpdater); ok {
		updater.ViewLayerBlockChanged(tx, pos)
	}
}
