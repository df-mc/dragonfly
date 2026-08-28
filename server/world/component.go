package world

import (
	"fmt"
	"image/color"
	"maps"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/df-mc/dragonfly/server/internal/colour"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type (
	// TickerComponent runs before the entity on each tick. Components run in
	// attachment order and may attach or detach components while ticking.
	TickerComponent interface {
		Tick(tx *Tx, e Entity, current int64)
	}
	// MetaSyncer adds client-visible entity metadata. Component values override
	// built-in values. Call MarkMetaDirty after changing metadata state.
	MetaSyncer interface {
		SyncMeta(e Entity, m *EntityMetadata)
	}
	// NBTSaver persists a component with its entity. LoadNBT must handle
	// missing or malformed fields.
	NBTSaver interface {
		SaveNBT() map[string]any
		LoadNBT(m map[string]any)
	}
)

type componentID uint32

// ComponentKey provides typed access to a registered component.
type ComponentKey[T any] struct {
	id componentID
}

type componentInfo struct {
	id   componentID
	name string
	typ  reflect.Type
	new  func() any
}

var components = struct {
	sync.Mutex
	byName map[string]*componentInfo
	byType map[reflect.Type]*componentInfo
	byID   []*componentInfo
}{
	byName: map[string]*componentInfo{},
	byType: map[reflect.Type]*componentInfo{},
	byID:   []*componentInfo{nil},
}

// RegisterComponent registers T under a namespaced name and returns its key.
// The name is used in saved data. RegisterComponent panics if the name or type
// is already registered, or if T is a pointer.
func RegisterComponent[T any](name string) ComponentKey[T] {
	if !strings.Contains(name, ":") {
		panic("world.RegisterComponent: name must be namespaced, e.g. 'myplugin:charged', got " + name)
	}
	typ := reflect.TypeFor[T]()
	if typ.Kind() == reflect.Pointer {
		panic(fmt.Sprintf("world.RegisterComponent: %v: register the element type, not a pointer type", typ))
	}
	components.Lock()
	defer components.Unlock()

	if _, ok := components.byName[name]; ok {
		panic("world.RegisterComponent: name " + name + " registered twice")
	}
	if info, ok := components.byType[typ]; ok {
		panic(fmt.Sprintf("world.RegisterComponent: type %v already registered as %v", typ, info.name))
	}
	info := &componentInfo{
		id:   componentID(len(components.byID)),
		name: name,
		typ:  typ,
		new:  func() any { return new(T) },
	}
	components.byName[name] = info
	components.byType[typ] = info
	components.byID = append(components.byID, info)
	return ComponentKey[T]{id: info.id}
}

// info returns the registry entry represented by k. It panics with a clear
// diagnostic for zero, stale, or type-mismatched keys.
func (k ComponentKey[T]) info() *componentInfo {
	byID := componentRegistrySnapshot()
	if k.id == 0 || int(k.id) >= len(byID) {
		panic("world.ComponentKey: zero or unregistered key")
	}
	info := byID[k.id]
	if info == nil || info.typ != reflect.TypeFor[T]() {
		panic(fmt.Sprintf("world.ComponentKey: key does not identify component type %v", reflect.TypeFor[T]()))
	}
	return info
}

type componentSlot struct {
	id componentID
	v  any
}

func findComponent(data *EntityData, id componentID) (int, bool) {
	for i, slot := range data.components {
		if slot.id == id {
			return i, true
		}
	}
	return -1, false
}

// componentRegistrySnapshot is safe after unlocking because entries never
// change or move.
func componentRegistrySnapshot() []*componentInfo {
	components.Lock()
	defer components.Unlock()
	return components.byID
}

// Of returns the entity's component, or nil if it is not attached. The
// returned pointer is valid only for the duration of tx's callback. Call
// MarkDirty after mutating a persistent component through it.
func (k ComponentKey[T]) Of(tx *Tx, e Entity) *T {
	k.info()
	h, _ := ownedComponentEntity(tx, e)
	data := &h.data
	if i, ok := findComponent(data, k.id); ok {
		return data.components[i].v.(*T)
	}
	return nil
}

// Attach adds or replaces the entity's component in the world owned by tx.
func (k ComponentKey[T]) Attach(tx *Tx, e Entity, v T) {
	k.info()
	h, current := ownedComponentEntity(tx, e)
	attachSlot(&h.data, componentSlot{id: k.id, v: &v})
	if _, persistent := any(&v).(NBTSaver); persistent {
		markComponentPersistenceDirty(tx, h)
	}
	markMetaDirty(tx, current)
}

// Detach removes the entity's component if it is attached in the world owned
// by tx.
func (k ComponentKey[T]) Detach(tx *Tx, e Entity) {
	info := k.info()
	h, current := ownedComponentEntity(tx, e)
	if i, ok := findComponent(&h.data, k.id); ok {
		v := h.data.components[i].v
		h.data.components = slices.Delete(h.data.components, i, i+1)
		name := info.name
		if _, retained := h.data.unknownComponents[name]; !retained {
			h.data.componentOrder = slices.DeleteFunc(h.data.componentOrder, func(entry string) bool {
				return entry == name
			})
		}
		h.data.tickers = nil
		if _, persistent := v.(NBTSaver); persistent {
			markComponentPersistenceDirty(tx, h)
		}
		markMetaDirty(tx, current)
	}
}

// MarkDirty marks a persistent component as changed so that its chunk is
// included in the next save. It does nothing for absent or runtime-only
// components.
func (k ComponentKey[T]) MarkDirty(tx *Tx, e Entity) {
	k.info()
	h, _ := ownedComponentEntity(tx, e)
	if i, ok := findComponent(&h.data, k.id); ok {
		if _, persistent := h.data.components[i].v.(NBTSaver); persistent {
			markComponentPersistenceDirty(tx, h)
		}
	}
}

// MarkMetaDirty sends the entity's latest metadata to its viewers. Attach and
// Detach call it automatically.
func MarkMetaDirty(tx *Tx, e Entity) {
	_, current := ownedComponentEntity(tx, e)
	markMetaDirty(tx, current)
}

// MarkEntityDirty marks the entity's chunk modified so that entity-owned
// persistent state is included in the next save.
func MarkEntityDirty(tx *Tx, e Entity) {
	h, _ := ownedComponentEntity(tx, e)
	markComponentPersistenceDirty(tx, h)
}

// markMetaDirty sends current entity metadata through viewers owned by tx.
func markMetaDirty(tx *Tx, e Entity) {
	for _, viewer := range tx.Viewers(e.Position()) {
		viewer.ViewEntityState(e)
	}
}

// ownedComponentEntity resolves e through tx and rejects cross-world access.
func ownedComponentEntity(tx *Tx, e Entity) (*EntityHandle, Entity) {
	if e == nil || e.H() == nil {
		panic("world component operation: nil entity")
	}
	h := e.H()
	current, ok := h.Entity(tx)
	if !ok {
		panic("world component operation: entity is not owned by transaction world")
	}
	return h, current
}

// markComponentPersistenceDirty marks h's current chunk modified through tx.
func markComponentPersistenceDirty(tx *Tx, h *EntityHandle) {
	w := tx.World()
	pos, ok := w.entities[h]
	if !ok {
		return
	}
	if c, ok := w.loadedChunk(pos); ok {
		c.modified = true
	}
}

func attachSlot(data *EntityData, slot componentSlot) {
	byID := componentRegistrySnapshot()
	name := byID[slot.id].name
	if i, ok := findComponent(data, slot.id); ok {
		data.components[i] = slot
	} else {
		orderIndex := slices.Index(data.componentOrder, name)
		if orderIndex == -1 {
			data.componentOrder = append(data.componentOrder, name)
			data.components = append(data.components, slot)
		} else {
			insertAt := len(data.components)
			for i, attached := range data.components {
				if slices.Index(data.componentOrder, byID[attached.id].name) > orderIndex {
					insertAt = i
					break
				}
			}
			data.components = slices.Insert(data.components, insertAt, slot)
		}
	}
	if _, persistent := slot.v.(NBTSaver); persistent && data.unknownComponents != nil {
		delete(data.unknownComponents, name)
	}
	data.tickers = nil
}

// AttachComponent adds a registered component while constructing or loading
// an entity. Use ComponentKey.Attach for entities already in a world.
// Do not share pointer components between entities.
func AttachComponent(data *EntityData, v any) {
	attachSlot(data, anySlot(v))
}

// AttachComponentIfAbsent adds a component only when it is not already
// attached. It reports whether the component was added.
func AttachComponentIfAbsent(data *EntityData, v any) bool {
	slot := anySlot(v)
	if _, ok := findComponent(data, slot.id); ok {
		return false
	}
	attachSlot(data, slot)
	return true
}

// ValidateComponents checks that each component is registered and appears
// once.
func ValidateComponents(values ...any) error {
	seen := make(map[componentID]struct{}, len(values))
	for _, v := range values {
		info, ok := lookupComponentType(v)
		if !ok {
			return fmt.Errorf("component type %T is not registered; call world.RegisterComponent first", v)
		}
		if _, dup := seen[info.id]; dup {
			return fmt.Errorf("component type %v (%v) occurs twice", info.typ, info.name)
		}
		seen[info.id] = struct{}{}
	}
	return nil
}

func lookupComponentType(v any) (*componentInfo, bool) {
	typ := reflect.TypeOf(v)
	if typ == nil {
		return nil, false
	}
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	components.Lock()
	info, ok := components.byType[typ]
	components.Unlock()
	return info, ok
}

func anySlot(v any) componentSlot {
	info, ok := lookupComponentType(v)
	if !ok {
		panic(fmt.Sprintf("world.AttachComponent: component type %T is not registered; call world.RegisterComponent first", v))
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		p := reflect.New(rv.Type())
		p.Elem().Set(rv)
		rv = p
	}
	return componentSlot{id: info.id, v: rv.Interface()}
}

func (e *EntityHandle) tickerComponents() []componentSlot {
	if e.data.tickers == nil {
		e.data.tickers = []componentSlot{}
		for _, slot := range e.data.components {
			if _, ok := slot.v.(TickerComponent); ok {
				e.data.tickers = append(e.data.tickers, slot)
			}
		}
	}
	return e.data.tickers
}

func tickComponents(tx *Tx, e Entity, current int64) bool {
	h := e.H()
	for _, slot := range h.tickerComponents() {
		i, attached := findComponent(&h.data, slot.id)
		if !attached || h.data.components[i].v != slot.v {
			continue
		}
		ticker := slot.v.(TickerComponent)
		ticker.Tick(tx, e, current)
		if h.Closed() || h.w != tx.World() {
			return false
		}
	}
	return true
}

// encodeComponentsNBT preserves saved data for unknown components.
func (data *EntityData) encodeComponentsNBT() map[string]any {
	if len(data.components) == 0 && len(data.unknownComponents) == 0 {
		return nil
	}
	byID := componentRegistrySnapshot()
	m := make(map[string]any, len(data.components)+len(data.unknownComponents))
	for _, slot := range data.components {
		if saver, ok := slot.v.(NBTSaver); ok {
			m[byID[slot.id].name] = saver.SaveNBT()
		}
	}
	for name, raw := range data.unknownComponents {
		if _, known := m[name]; !known {
			m[name] = raw
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// orderedNames keeps the saved order and appends new names alphabetically.
func orderedNames(primary []string, m map[string]any) []string {
	names := make([]string, 0, len(m))
	seen := make(map[string]struct{}, len(m))
	for _, name := range primary {
		if _, ok := m[name]; !ok {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		names = append(names, name)
		seen[name] = struct{}{}
	}
	for _, name := range slices.Sorted(maps.Keys(m)) {
		if _, ok := seen[name]; !ok {
			names = append(names, name)
		}
	}
	return names
}

// encodeComponentOrderNBT includes runtime-only positions when component data
// is saved.
func (data *EntityData) encodeComponentOrderNBT() []string {
	byID := componentRegistrySnapshot()
	present := make(map[string]any, len(data.components)+len(data.unknownComponents))
	for _, slot := range data.components {
		present[byID[slot.id].name] = nil
	}
	for name := range data.unknownComponents {
		present[name] = nil
	}
	return orderedNames(data.componentOrder, present)
}

// decodedComponentOrder restores saved order and appends missing names.
func decodedComponentOrder(order []string, m map[string]any) []string {
	names := make([]string, 0, len(order)+len(m))
	seen := make(map[string]struct{}, len(order)+len(m))
	for _, name := range order {
		if name == "" {
			continue
		}
		if _, duplicate := seen[name]; duplicate {
			continue
		}
		names = append(names, name)
		seen[name] = struct{}{}
	}
	for _, name := range slices.Sorted(maps.Keys(m)) {
		if _, exists := seen[name]; !exists {
			names = append(names, name)
		}
	}
	return names
}

// decodeComponentsNBT keeps unknown or malformed data for the next save.
func (data *EntityData) decodeComponentsNBT(m map[string]any, order []string) {
	data.componentOrder = decodedComponentOrder(order, m)
	retain := func(name string, raw any) {
		if data.unknownComponents == nil {
			data.unknownComponents = make(map[string]any)
		}
		data.unknownComponents[name] = raw
	}
	for _, name := range data.componentOrder {
		raw, ok := m[name]
		if !ok {
			continue
		}
		sub, ok := raw.(map[string]any)
		if !ok {
			retain(name, raw)
			continue
		}
		components.Lock()
		info, ok := components.byName[name]
		components.Unlock()
		if !ok {
			retain(name, sub)
			continue
		}
		v := info.new()
		saver, ok := v.(NBTSaver)
		if !ok {
			retain(name, sub)
			continue
		}
		saver.LoadNBT(sub)
		attachSlot(data, componentSlot{id: info.id, v: v})
	}
}

type metadata = protocol.EntityMetadata

// EntityMetadata stores metadata contributed by a component.
type EntityMetadata struct {
	metadata
	resets protocol.EntityMetadata
}

func newEntityMetadata() *EntityMetadata {
	return &EntityMetadata{
		metadata: protocol.NewEntityMetadata(),
		resets:   protocol.EntityMetadata{},
	}
}

// SetScoreTag sets the text below the entity's name tag.
func (m *EntityMetadata) SetScoreTag(s string) {
	m.metadata[protocol.EntityDataKeyScore] = s
	m.resets[protocol.EntityDataKeyScore] = ""
}

// SetVariant sets the entity's visual variant.
func (m *EntityMetadata) SetVariant(v int32) {
	m.metadata[protocol.EntityDataKeyVariant] = v
	m.resets[protocol.EntityDataKeyVariant] = int32(0)
}

// SetScale sets the entity's render scale.
func (m *EntityMetadata) SetScale(s float64) {
	m.metadata[protocol.EntityDataKeyScale] = float32(s)
	m.resets[protocol.EntityDataKeyScale] = float32(1)
}

// SetColour sets the entity's effect colour.
func (m *EntityMetadata) SetColour(c color.RGBA) {
	m.metadata[protocol.EntityDataKeyEffectColor] = colour.Int32FromRGBA(c)
	m.resets[protocol.EntityDataKeyEffectColor] = int32(0)
}

// AddComponentMetadata merges metadata contributed by e's components into dst.
// It returns the current values and the values used to reset removed metadata.
// Call AddComponentMetadata only from the world's owner goroutine.
func AddComponentMetadata(e Entity, dst protocol.EntityMetadata) (current, resets protocol.EntityMetadata) {
	var metadata *EntityMetadata
	for _, component := range e.H().data.components {
		if syncer, ok := component.v.(MetaSyncer); ok {
			if metadata == nil {
				metadata = newEntityMetadata()
			}
			syncer.SyncMeta(e, metadata)
		}
	}
	if metadata == nil {
		return nil, nil
	}
	mergeComponentMetadata(dst, metadata.metadata)
	return metadata.metadata, metadata.resets
}

func mergeComponentMetadata(dst, src protocol.EntityMetadata) {
	for key, value := range src {
		switch key {
		case protocol.EntityDataKeyPlayerFlags:
			var flags byte
			if existing, ok := dst[key]; ok {
				flags = existing.(byte)
			}
			dst[key] = flags | value.(byte)
		case protocol.EntityDataKeyFlags, protocol.EntityDataKeyFlagsTwo:
			var flags int64
			if existing, ok := dst[key]; ok {
				flags = existing.(int64)
			}
			dst[key] = flags | value.(int64)
		default:
			dst[key] = value
		}
	}
}

type spawner interface {
	EntityType
	New(opts EntitySpawnOpts, components ...any) *EntityHandle
}

// SpawnEntity creates and adds a registered entity type. It returns false if
// the type is unknown or cannot be spawned this way. Invalid components panic.
func (tx *Tx) SpawnEntity(name string, opts EntitySpawnOpts, components ...any) (Entity, bool) {
	t, ok := tx.World().EntityRegistry().Lookup(name)
	if !ok {
		return nil, false
	}
	spawner, ok := t.(spawner)
	if !ok {
		return nil, false
	}
	return tx.AddEntity(spawner.New(opts, components...)), true
}
