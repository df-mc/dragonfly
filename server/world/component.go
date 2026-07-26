package world

import (
	"fmt"
	"image/color"
	"iter"
	"maps"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Component capabilities: a component attached to an entity may implement any
// of the interfaces below. All of them are called on the World's owner
// goroutine only, like all other entity state.
type (
	// TickerComponent is implemented by components that run logic on every
	// tick of the entity they are attached to. Components tick in attach
	// order, before the entity's own TickerEntity logic. A component may
	// attach or detach components, including itself, during its tick.
	TickerComponent interface {
		Tick(tx *Tx, e Entity, current int64)
	}
	// MetaSyncer is implemented by components that contribute to the
	// client-visible metadata of the entity they are attached to. SyncMeta is
	// called whenever the entity's metadata is built. Components run after
	// built-in metadata and overwrite it on conflict. After mutating a
	// MetaSyncer component through ComponentKey.Of, call MarkMetaDirty.
	MetaSyncer interface {
		SyncMeta(e Entity, m *EntityMetadata)
	}
	// NBTSaver is implemented by components that persist with the entity in
	// the world save. Components that do not implement NBTSaver are attached
	// at runtime only. LoadNBT is called on a fresh component value and must
	// tolerate missing or malformed data, e.g. by using the nbtconv helpers.
	NBTSaver interface {
		SaveNBT() map[string]any
		LoadNBT(m map[string]any)
	}
)

// ComponentID uniquely identifies a registered component type within the
// process. IDs are assigned by RegisterComponent.
type ComponentID uint32

// ComponentKey provides typed access to a component of type T on entities.
// Keys are created once per component type using RegisterComponent and are
// usually stored in a package-level variable.
type ComponentKey[T any] struct {
	id ComponentID
}

// componentInfo holds the runtime information of a registered component type.
type componentInfo struct {
	id   ComponentID
	name string
	typ  reflect.Type
	// new returns a fresh *T as any, used when loading components from NBT.
	new func() any
}

// components is the process-wide component registry. Registration is expected
// to happen during package initialisation; lookups on hot paths go through
// ComponentKey and do not touch this registry.
var components = struct {
	sync.Mutex
	byName map[string]*componentInfo
	byType map[reflect.Type]*componentInfo
	byID   []*componentInfo
}{
	byName: map[string]*componentInfo{},
	byType: map[reflect.Type]*componentInfo{},
}

// RegisterComponent registers T as an entity component under a stable,
// namespaced name such as "myplugin:charged" and returns the typed key used
// to access it. The name is used to persist the component if T implements
// NBTSaver. RegisterComponent is intended to be called from package-level
// variable initialisation and panics if the name or type is already
// registered, or if T is a pointer type.
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
		id:   ComponentID(len(components.byID)),
		name: name,
		typ:  typ,
		new:  func() any { return new(T) },
	}
	components.byName[name] = info
	components.byType[typ] = info
	components.byID = append(components.byID, info)
	return ComponentKey[T]{id: info.id}
}

// componentSlot is a single component attached to an entity. The value is
// always a pointer to the registered component type.
type componentSlot struct {
	id ComponentID
	v  any
}

// findComponent returns the slot index of a component ID in the component
// slice of an EntityData. The slice is small, so a linear scan is cheap.
func findComponent(data *EntityData, id ComponentID) (int, bool) {
	for i, slot := range data.components {
		if slot.id == id {
			return i, true
		}
	}
	return -1, false
}

// componentRegistrySnapshot returns the slice of registered component info,
// indexed by ComponentID. The registry only ever grows and existing entries
// are never modified, so the returned header stays valid after the lock is
// released.
func componentRegistrySnapshot() []*componentInfo {
	components.Lock()
	defer components.Unlock()
	return components.byID
}

// Of returns a pointer to the entity's component of type T for in-place
// mutation, or nil if the entity has no such component attached. Always
// nil-check the result for entities whose components are not your own. Of
// must be called from the World's owner goroutine, like all entity state
// access.
func (k ComponentKey[T]) Of(e Entity) *T {
	h := e.H()
	data := &h.data
	if i, ok := findComponent(data, k.id); ok {
		v := data.components[i].v.(*T)
		if _, persistent := any(v).(NBTSaver); persistent {
			markComponentPersistenceDirty(h)
		}
		return v
	}
	return nil
}

// Attach adds a component of type T to the entity, replacing any existing
// one, and resends the entity's metadata to viewers if T contributes to it.
func (k ComponentKey[T]) Attach(e Entity, v T) {
	h := e.H()
	attachSlot(&h.data, componentSlot{id: k.id, v: &v})
	if _, persistent := any(&v).(NBTSaver); persistent {
		markComponentPersistenceDirty(h)
	}
	if _, ok := any(&v).(MetaSyncer); ok {
		MarkMetaDirty(e)
	}
}

// Detach removes the component of type T from the entity, if attached, and
// resends the entity's metadata to viewers if T contributed to it.
func (k ComponentKey[T]) Detach(e Entity) {
	h := e.H()
	if i, ok := findComponent(&h.data, k.id); ok {
		v := h.data.components[i].v
		h.data.components = slices.Delete(h.data.components, i, i+1)
		name := componentRegistrySnapshot()[k.id].name
		if _, retained := h.data.unknownComponents[name]; !retained {
			h.data.componentOrder = slices.DeleteFunc(h.data.componentOrder, func(entry string) bool {
				return entry == name
			})
		}
		h.data.tickers = nil
		if _, persistent := v.(NBTSaver); persistent {
			markComponentPersistenceDirty(h)
		}
		if _, ok := v.(MetaSyncer); ok {
			MarkMetaDirty(e)
		}
	}
}

// MarkMetaDirty resends the entity's metadata to its viewers. Call it after
// mutating a MetaSyncer component through ComponentKey.Of; Attach and Detach
// call it automatically.
func MarkMetaDirty(e Entity) {
	h := e.H()
	if h.w == nil {
		return
	}
	for _, viewer := range h.w.viewersOf(h.data.Pos) {
		viewer.ViewEntityState(e)
	}
}

// markComponentPersistenceDirty marks the chunk containing h for persistence.
// Component values are mutable pointers, so persistent components are marked
// conservatively when accessed or ticked as well as when attached or detached.
func markComponentPersistenceDirty(h *EntityHandle) {
	if h.w == nil {
		return
	}
	pos, ok := h.w.entities[h]
	if !ok {
		return
	}
	if c, ok := h.w.loadedChunk(pos); ok {
		c.modified = true
	}
}

// attachSlot adds a slot to the component slice of an EntityData, replacing
// an existing component with the same ID in place, so attach order is kept.
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

// AttachComponent attaches a component value of any registered component type
// to an EntityData. It is intended for entity construction, i.e. in
// EntityConfig.Apply or EntityType.DecodeNBT implementations; use
// ComponentKey.Attach for entities that are in a world. AttachComponent
// panics if the value's type was not registered with RegisterComponent.
// Pointer values are attached as-is and must not be shared between entities.
func AttachComponent(data *EntityData, v any) {
	attachSlot(data, anySlot(v))
}

// AttachComponentIfAbsent attaches a component like AttachComponent, but
// keeps an existing component of the same type, if present. It returns true
// if the component was attached.
func AttachComponentIfAbsent(data *EntityData, v any) bool {
	slot := anySlot(v)
	if _, ok := findComponent(data, slot.id); ok {
		return false
	}
	attachSlot(data, slot)
	return true
}

// ValidateComponents checks that every value passed is of a registered
// component type and that no type occurs twice, returning an error describing
// the first violation found. It allows entity specs to be validated eagerly
// at registration time, so mistakes surface at init instead of at first
// spawn.
func ValidateComponents(values ...any) error {
	seen := make(map[ComponentID]struct{}, len(values))
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

// lookupComponentType resolves the componentInfo of a value passed either as
// T or *T.
func lookupComponentType(v any) (*componentInfo, bool) {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	components.Lock()
	info, ok := components.byType[typ]
	components.Unlock()
	return info, ok
}

// anySlot converts an untyped component value, passed either as T or *T, to
// a componentSlot holding a *T. Non-pointer values are copied to a fresh
// pointer, so a shared prototype value is never aliased between entities.
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

// Components yields all components attached to the entity in attach order.
// Values yielded are pointers to the registered component types and must not
// be modified. Use ComponentKey.Of to obtain a component for mutation.
// Components must only be used from the World's owner goroutine.
func (e *EntityHandle) Components() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, slot := range e.data.components {
			if !yield(slot.v) {
				return
			}
		}
	}
}

// noTickers is the cached ticker slice of entities without any, so their
// cache does not read as invalidated on every tick.
var noTickers = make([]TickerComponent, 0)

// TickerComponents returns the entity's components implementing
// TickerComponent in attach order. The returned slice is cached and must not
// be modified. Attach and Detach invalidate the cache, so ticking components
// may attach and detach components, including themselves: the tick pipeline
// finishes the snapshot it started with.
func (e *EntityHandle) TickerComponents() []TickerComponent {
	if e.data.tickers == nil {
		e.data.tickers = noTickers
		for _, slot := range e.data.components {
			if t, ok := slot.v.(TickerComponent); ok {
				e.data.tickers = append(e.data.tickers, t)
			}
		}
	}
	return e.data.tickers
}

// tickComponents ticks the entity's component snapshot until it finishes or a
// component removes or closes the entity. It reports whether the entity may
// continue through its own tick logic.
func tickComponents(tx *Tx, e Entity, current int64) bool {
	h := e.H()
	for _, ticker := range h.TickerComponents() {
		if _, persistent := ticker.(NBTSaver); persistent {
			markComponentPersistenceDirty(h)
		}
		ticker.Tick(tx, e, current)
		if h.Closed() || h.w != tx.World() {
			return false
		}
	}
	return true
}

// encodeComponentsNBT encodes all components implementing NBTSaver, along
// with any component data of unknown types read earlier, so unknown
// components round-trip losslessly through the world save.
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

// orderedNames returns the keys of m ordered by the primary sequence first
// (skipping names absent from m and duplicates), then any remaining keys in
// sorted order. It gives persisted components a deterministic order that
// follows attachment where known and stays stable across saves written before
// order was stored.
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

// encodeComponentOrderNBT returns attached and retained component names in
// attachment order. Runtime-only component names are included when another
// component has state to persist so their relative position survives reload.
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

// decodedComponentOrder preserves every name in the persisted order,
// including runtime-only components without NBT and temporarily unknown
// components, then appends payload names absent from older order data.
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

// decodeComponentsNBT restores components from the "Components" compound of
// an entity's saved NBT. Entries of unregistered component names or with
// malformed values are retained verbatim for the next save. Components are
// restored in the persisted order, with names missing from order appended in
// sorted order for compatibility with saves written before order was stored.
func (data *EntityData) decodeComponentsNBT(m map[string]any, order []string) {
	data.componentOrder = decodedComponentOrder(order, m)
	retain := func(name string, raw any) {
		if data.unknownComponents == nil {
			data.unknownComponents = make(map[string]any)
		}
		data.unknownComponents[name] = raw
	}
	// componentOrder already contains every key of m (decodedComponentOrder
	// appends any missing ones), so iterating it restores components in the
	// persisted order. Names without a payload are runtime-only positions with
	// nothing to load; attachSlot keeps data.components ordered regardless.
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

// EntityMetadata holds client-visible metadata contributed by components
// through MetaSyncer. Keys and flag bits mirror the protocol's actor
// metadata; the session layer merges them into the packets it sends.
type EntityMetadata struct {
	values   map[uint32]any
	defaults map[uint32]any
	flags    map[uint32]int64
}

// NewEntityMetadata returns an empty EntityMetadata.
func NewEntityMetadata() *EntityMetadata {
	return &EntityMetadata{
		values:   map[uint32]any{},
		defaults: map[uint32]any{},
		flags:    map[uint32]int64{},
	}
}

// SetFlag sets an actor flag using protocol.EntityDataKey* and
// protocol.EntityDataFlag* constants. The flag is combined with flags already
// present on the entity.
func (m *EntityMetadata) SetFlag(key uint32, bit uint8) {
	switch key {
	case protocol.EntityDataKeyPlayerFlags:
		if bit >= 8 {
			panic("world.EntityMetadata: player flag bit must be less than 8")
		}
		m.defaults[key] = byte(0)
	case protocol.EntityDataKeyFlags, protocol.EntityDataKeyFlagsTwo:
		if bit >= 64 {
			panic("world.EntityMetadata: actor flag bit must be less than 64")
		}
		m.defaults[key] = int64(0)
	default:
		panic(fmt.Sprintf("world.EntityMetadata: key %d is not an actor flag key", key))
	}
	m.flags[key] |= 1 << bit
}

// SetScoreTag sets the text displayed below the entity's name tag.
func (m *EntityMetadata) SetScoreTag(s string) {
	m.values[protocol.EntityDataKeyScore] = s
	m.defaults[protocol.EntityDataKeyScore] = ""
}

// SetVariant sets the entity's variant, used by clients to pick textures and,
// for some entity types, block appearances.
func (m *EntityMetadata) SetVariant(v int32) {
	m.values[protocol.EntityDataKeyVariant] = v
	m.defaults[protocol.EntityDataKeyVariant] = int32(0)
}

// SetScale sets the entity's render scale.
func (m *EntityMetadata) SetScale(s float64) {
	m.values[protocol.EntityDataKeyScale] = float32(s)
	m.defaults[protocol.EntityDataKeyScale] = float32(1)
}

// SetColour sets the entity's effect colour, used for tints such as potion
// swirls.
func (m *EntityMetadata) SetColour(c color.RGBA) {
	// Encoded as ARGB.
	m.values[protocol.EntityDataKeyEffectColor] = int32(uint32(c.A)<<24 | uint32(c.R)<<16 | uint32(c.G)<<8 | uint32(c.B))
	m.defaults[protocol.EntityDataKeyEffectColor] = int32(0)
}

// Set sets a metadata value under a raw protocol actor data key, for values
// without a typed setter. Values are converted to their protocol-encodable
// equivalents where needed; Set panics on values the protocol cannot encode,
// as they would otherwise fail every metadata packet sent for the entity.
func (m *EntityMetadata) Set(key uint32, value any) {
	switch v := value.(type) {
	case byte:
		m.values[key] = v
		m.defaults[key] = byte(0)
	case int16:
		m.values[key] = v
		m.defaults[key] = int16(0)
	case int32:
		m.values[key] = v
		m.defaults[key] = int32(0)
	case float32:
		m.values[key] = v
		m.defaults[key] = float32(0)
	case int64:
		m.values[key] = v
		m.defaults[key] = int64(0)
	case string:
		m.values[key] = v
		m.defaults[key] = ""
	case map[string]any:
		m.values[key] = v
		m.defaults[key] = map[string]any{}
	case protocol.BlockPos:
		m.values[key] = v
		m.defaults[key] = protocol.BlockPos{}
	case mgl32.Vec3:
		m.values[key] = v
		m.defaults[key] = mgl32.Vec3{}
	case int:
		m.values[key] = int32(v)
		m.defaults[key] = int32(0)
	case uint32:
		m.values[key] = int32(v)
		m.defaults[key] = int32(0)
	case float64:
		m.values[key] = float32(v)
		m.defaults[key] = float32(0)
	case bool:
		var b byte
		if v {
			b = 1
		}
		m.values[key] = b
		m.defaults[key] = byte(0)
	default:
		panic(fmt.Sprintf("world.EntityMetadata: value of type %T cannot be encoded as actor metadata", value))
	}
}

// Values returns the metadata values set. The returned map is a read-only
// view for the session layer.
func (m *EntityMetadata) Values() map[uint32]any { return m.values }

// Defaults returns the values used to reset metadata keys when a component no
// longer contributes them. The returned map is a read-only view for the
// session layer.
func (m *EntityMetadata) Defaults() map[uint32]any { return m.defaults }

// Flags returns the flag bits set per key. The returned map is a read-only
// view for the session layer.
func (m *EntityMetadata) Flags() map[uint32]int64 { return m.flags }

// Spawner is implemented by EntityTypes that can construct entities from
// spawn options and a list of extra components, such as types created with
// the entity package's Spec registry. It enables Tx.SpawnEntity.
type Spawner interface {
	EntityType
	New(opts EntitySpawnOpts, components ...any) *EntityHandle
}

// SpawnEntity spawns an entity of the EntityType registered under the name
// passed in the World of the Tx, with any extra components attached. It
// returns false if no type with that name is registered in the World's
// EntityRegistry or if the type does not implement Spawner. Passing a
// component of an unregistered type is a programmer error and panics, like
// ComponentKey.Attach.
func (tx *Tx) SpawnEntity(name string, opts EntitySpawnOpts, components ...any) (Entity, bool) {
	t, ok := tx.World().EntityRegistry().Lookup(name)
	if !ok {
		return nil, false
	}
	spawner, ok := t.(Spawner)
	if !ok {
		return nil, false
	}
	return tx.AddEntity(spawner.New(opts, components...)), true
}
