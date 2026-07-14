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
)

// Component capabilities: a component attached to an entity may implement any
// of the interfaces below. All of them are called on the World's owner
// goroutine only, like all other entity state.
type (
	// TickerComponent is implemented by components that run logic on every
	// tick of the entity they are attached to. Components tick in attach
	// order, before the entity's main behaviour. A component may attach or
	// detach components, including itself, during its tick. Ticking runs for
	// entities built on entity.Ent and for players; other world.Entity
	// implementations must run EntityHandle.TickerComponents themselves.
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

// Of returns a pointer to the entity's component of type T for in-place
// mutation, or nil if the entity has no such component attached. Always
// nil-check the result for entities whose components are not your own. Of
// must be called from the World's owner goroutine, like all entity state
// access.
func (k ComponentKey[T]) Of(e Entity) *T {
	data := &e.H().data
	if i, ok := findComponent(data, k.id); ok {
		return data.components[i].v.(*T)
	}
	return nil
}

// Attach adds a component of type T to the entity, replacing any existing
// one, and resends the entity's metadata to viewers if T contributes to it.
func (k ComponentKey[T]) Attach(e Entity, v T) {
	h := e.H()
	attachSlot(&h.data, componentSlot{id: k.id, v: &v})
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
		h.data.tickers = nil
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

// attachSlot adds a slot to the component slice of an EntityData, replacing
// an existing component with the same ID in place, so attach order is kept.
func attachSlot(data *EntityData, slot componentSlot) {
	if i, ok := findComponent(data, slot.id); ok {
		data.components[i] = slot
	} else {
		data.components = append(data.components, slot)
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
// Values yielded are pointers to the registered component types. It must
// only be used from the World's owner goroutine.
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

// encodeComponentsNBT encodes all components implementing NBTSaver, along
// with any component data of unknown types read earlier, so unknown
// components round-trip losslessly through the world save.
func (data *EntityData) encodeComponentsNBT() map[string]any {
	if len(data.components) == 0 && len(data.unknownComponents) == 0 {
		return nil
	}
	components.Lock()
	// byID only ever grows and existing entries are never modified, so a
	// snapshot of the slice header is safe to read after unlocking.
	byID := components.byID
	components.Unlock()

	m := make(map[string]any, len(data.components)+len(data.unknownComponents))
	for _, slot := range data.components {
		if saver, ok := slot.v.(NBTSaver); ok {
			m[byID[slot.id].name] = saver.SaveNBT()
		}
	}
	for name, raw := range data.unknownComponents {
		m[name] = raw
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// decodeComponentsNBT restores components from the "Components" compound of
// an entity's saved NBT. Entries of unregistered component names or with
// malformed values are retained verbatim for the next save. Components are
// restored in sorted name order, keeping attach order deterministic.
func (data *EntityData) decodeComponentsNBT(m map[string]any) {
	retain := func(name string, raw any) {
		if data.unknownComponents == nil {
			data.unknownComponents = make(map[string]any)
		}
		data.unknownComponents[name] = raw
	}
	for _, name := range slices.Sorted(maps.Keys(m)) {
		sub, ok := m[name].(map[string]any)
		if !ok {
			retain(name, m[name])
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
		if saver, ok := v.(NBTSaver); ok {
			saver.LoadNBT(sub)
		}
		attachSlot(data, componentSlot{id: info.id, v: v})
	}
}

// EntityMetadata holds client-visible metadata contributed by components
// through MetaSyncer. Keys and flag bits mirror the protocol's actor
// metadata; the session layer merges them into the packets it sends.
type EntityMetadata struct {
	values map[uint32]any
	flags  map[uint32]int64
}

// NewEntityMetadata returns an empty EntityMetadata.
func NewEntityMetadata() *EntityMetadata {
	return &EntityMetadata{values: map[uint32]any{}, flags: map[uint32]int64{}}
}

// EntityMetaFlag is a single boolean actor flag that components may set
// through EntityMetadata.SetFlag.
type EntityMetaFlag struct {
	key uint32
	bit uint8
}

// The most commonly used actor flags. Values mirror the protocol's actor
// flags.
var (
	MetaFlagOnFire    = EntityMetaFlag{bit: 0}
	MetaFlagSneaking  = EntityMetaFlag{bit: 1}
	MetaFlagSprinting = EntityMetaFlag{bit: 3}
	MetaFlagInvisible = EntityMetaFlag{bit: 5}
	MetaFlagCritical  = EntityMetaFlag{bit: 13}
	MetaFlagNoAI      = EntityMetaFlag{bit: 16}
	MetaFlagLingering = EntityMetaFlag{bit: 47}
	MetaFlagEnchanted = EntityMetaFlag{bit: 52}
)

// SetFlag sets an actor flag, combined with the flags already present on the
// entity.
func (m *EntityMetadata) SetFlag(f EntityMetaFlag) {
	m.flags[f.key] |= 1 << f.bit
}

// SetFlagBit sets a single flag bit under a raw protocol actor data key, for
// flags without an EntityMetaFlag value.
func (m *EntityMetadata) SetFlagBit(key uint32, bit uint8) {
	m.flags[key] |= 1 << bit
}

// SetScoreTag sets the text displayed below the entity's name tag.
func (m *EntityMetadata) SetScoreTag(s string) {
	m.values[84] = s // protocol.EntityDataKeyScore
}

// SetVariant sets the entity's variant, used by clients to pick textures and,
// for some entity types, block appearances.
func (m *EntityMetadata) SetVariant(v int32) {
	m.values[2] = v // protocol.EntityDataKeyVariant
}

// SetScale sets the entity's render scale.
func (m *EntityMetadata) SetScale(s float64) {
	m.values[38] = float32(s) // protocol.EntityDataKeyScale
}

// SetColour sets the entity's effect colour, used for tints such as potion
// swirls.
func (m *EntityMetadata) SetColour(c color.RGBA) {
	// protocol.EntityDataKeyEffectColor, encoded as ARGB.
	m.values[8] = int32(uint32(c.A)<<24 | uint32(c.R)<<16 | uint32(c.G)<<8 | uint32(c.B))
}

// Set sets a metadata value under a raw protocol actor data key, for values
// without a typed setter. Values are converted to their protocol-encodable
// equivalents where needed; Set panics on values the protocol cannot encode,
// as they would otherwise fail every metadata packet sent for the entity.
func (m *EntityMetadata) Set(key uint32, value any) {
	switch v := value.(type) {
	case byte, int16, int32, float32, int64, string, map[string]any:
		m.values[key] = v
	case int:
		m.values[key] = int32(v)
	case uint32:
		m.values[key] = int32(v)
	case float64:
		m.values[key] = float32(v)
	case bool:
		var b byte
		if v {
			b = 1
		}
		m.values[key] = b
	default:
		panic(fmt.Sprintf("world.EntityMetadata: value of type %T cannot be encoded as actor metadata", value))
	}
}

// Values returns the metadata values set. The returned map is a read-only
// view for the session layer.
func (m *EntityMetadata) Values() map[uint32]any { return m.values }

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
