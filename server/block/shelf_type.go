package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"sync"
	"time"
)

// Shelf is a wooden block that can store and display up to three stacks of items.
type Shelf struct {
	solid
	bass
	sourceWaterDisplacer

	// Wood is the type of wood of the shelf.
	Wood WoodType
	// Bamboo specifies if the shelf is made of bamboo.
	Bamboo bool
	// Facing is the direction the shelf is facing.
	Facing cube.Direction
	// Powered specifies if the shelf is powered by redstone.
	Powered bool
	// PoweredType is the type of the powered shelf (0-3).
	PoweredType int

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}
}

// NewShelf crea una nueva estantería inicializada con su inventario de 3 slots.
func NewShelf(w WoodType, bamboo bool) Shelf {
	m := new(sync.RWMutex)
	v := make(map[ContainerViewer]struct{}, 1)
	return Shelf{
		Wood:   w,
		Bamboo: bamboo,
		inventory: inventory.New(3, func(slot int, _, item item.Stack) {
			m.RLock()
			defer m.RUnlock()
			for viewer := range v {
				viewer.ViewSlotChange(slot, item)
			}
		}),
		viewerMu: m,
		viewers:  v,
	}
}

// MaxCount ...
func (Shelf) MaxCount() int {
	return 64
}

// FlammabilityInfo ...
func (s Shelf) FlammabilityInfo() FlammabilityInfo {
	if s.Bamboo {
		return newFlammabilityInfo(5, 20, true)
	}
	if !s.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// FuelInfo permite que el Shelf se use como combustible (cocina 1.5 ítems).
func (s Shelf) FuelInfo() item.FuelInfo {
	if s.Bamboo || s.Wood.Flammable() {
		return newFuelInfo(time.Second * 15)
	}
	return item.FuelInfo{}
}

// BreakInfo suelta los ítems del inventario al romper el bloque.
func (s Shelf) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(Shelf{Wood: s.Wood, Bamboo: s.Bamboo})).withBlastResistance(3).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		if s.inventory != nil {
			for _, i := range s.inventory.Clear() {
				dropItem(tx, i, pos.Vec3Middle())
			}
		}
	})
}

// Inventory devuelve el inventario de 3 slots.
func (s Shelf) Inventory(*world.Tx, cube.Pos) *inventory.Inventory {
	return s.inventory
}

// AddViewer ...
func (s Shelf) AddViewer(v ContainerViewer, _ *world.Tx, _ cube.Pos) {
	s.viewerMu.Lock()
	defer s.viewerMu.Unlock()
	s.viewers[v] = struct{}{}
}

// RemoveViewer ...
func (s Shelf) RemoveViewer(v ContainerViewer, _ *world.Tx, _ cube.Pos) {
	s.viewerMu.Lock()
	defer s.viewerMu.Unlock()
	delete(s.viewers, v)
}

// Activate maneja la interacción (intercambio de ítems o hotbar si hay redstone).
func (s Shelf) Activate(pos cube.Pos, face cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	if sneaking, ok := u.(interface{ Sneaking() bool }); ok && sneaking.Sneaking() {
		return false
	}
	held, _ := u.HeldItems()
	if _, ok := held.Item().(Shelf); ok {
		return false
	}

	if s.inventory == nil {
		s = NewShelf(s.Wood, s.Bamboo)
	}

	if s.Powered {
		s.swapHotbar(pos, tx, u)
		return true
	}

	if face != s.Facing.Face() {
		return false
	}

	// Intentamos usar la lógica de 3 slots por posición.
	slot := s.calculateSlotFromView(pos, u)
	slotItem, _ := s.inventory.Item(slot)

	// Aplicamos la lógica de intercambio que el usuario confirmó que funciona.
	switch {
	case !held.Empty() && slotItem.Empty():
		_ = s.inventory.SetItem(slot, held.Grow(-held.Count()+1))
		ctx.SubtractFromCount(1)
		tx.PlaySound(pos.Vec3Middle(), sound.ItemAdd{})
		tx.SetBlock(pos, s, &world.SetOpts{DisableBlockUpdates: true})
		return true
	case held.Empty() && !slotItem.Empty():
		ctx.NewItem = slotItem
		_ = s.inventory.SetItem(slot, item.Stack{})
		tx.PlaySound(pos.Vec3Middle(), sound.Click{})
		tx.SetBlock(pos, s, &world.SetOpts{DisableBlockUpdates: true})
		return true
	case !held.Empty() && !slotItem.Empty() && held.Count() == 1:
		_ = s.inventory.SetItem(slot, held)
		ctx.NewItem = slotItem
		ctx.SubtractFromCount(1)
		tx.PlaySound(pos.Vec3Middle(), sound.Click{})
		tx.SetBlock(pos, s, &world.SetOpts{DisableBlockUpdates: true})
		return true
	}
	return false
}

// calculateSlotFromView usa trigonometría para saber qué parte del frente del bloque miras.
func (s Shelf) calculateSlotFromView(pos cube.Pos, u item.User) int {
	eyePos := u.Position().Add(mgl64.Vec3{0, 1.62, 0})
	rot := u.Rotation()
	yaw, pitch := rot.Yaw(), rot.Pitch()

	direction := mgl64.Vec3{
		-math.Sin(yaw*math.Pi/180) * math.Cos(pitch*math.Pi/180),
		-math.Sin(pitch * math.Pi / 180),
		math.Cos(yaw*math.Pi/180) * math.Cos(pitch*math.Pi/180),
	}

	var planeNormal mgl64.Vec3
	var planePoint mgl64.Vec3
	switch s.Facing {
	case cube.North:
		planeNormal = mgl64.Vec3{0, 0, -1}
		planePoint = pos.Vec3()
	case cube.South:
		planeNormal = mgl64.Vec3{0, 0, 1}
		planePoint = pos.Vec3().Add(mgl64.Vec3{0, 0, 1})
	case cube.West:
		planeNormal = mgl64.Vec3{-1, 0, 0}
		planePoint = pos.Vec3()
	case cube.East:
		planeNormal = mgl64.Vec3{1, 0, 0}
		planePoint = pos.Vec3().Add(mgl64.Vec3{1, 0, 0})
	}

	denom := direction.Dot(planeNormal)
	if math.Abs(denom) < 1e-6 {
		return 1
	}
	t := (planePoint.Sub(eyePos)).Dot(planeNormal) / denom
	hitPoint := eyePos.Add(direction.Mul(t))
	relativeHit := hitPoint.Sub(pos.Vec3())

	var x float64
	switch s.Facing {
	case cube.North:
		x = relativeHit[0]
	case cube.South:
		x = 1 - relativeHit[0]
	case cube.West:
		x = 1 - relativeHit[2]
	case cube.East:
		x = relativeHit[2]
	}

	if x < 1.0/3.0 {
		return 0
	} else if x < 2.0/3.0 {
		return 1
	}
	return 2
}

// swapHotbar intercambia el contenido del Shelf con la hotbar del jugador.
func (s Shelf) swapHotbar(pos cube.Pos, tx *world.Tx, u item.User) {
	connected := s.findConnectedShelves(pos, tx)
	numShelves := len(connected)
	if numShelves > 3 {
		numShelves = 3
	}

	if invHolder, ok := u.(interface{ Inventory() *inventory.Inventory }); ok {
		inv := invHolder.Inventory()
		numSlots := numShelves * 3
		startIndex := 9 - numSlots
		if startIndex < 0 {
			startIndex = 0
		}

		tx.PlaySound(pos.Vec3Middle(), sound.Click{})

		slotOffset := 0
		for _, shelf := range connected[:numShelves] {
			if shelf.inventory == nil {
				continue
			}
			for i := 0; i < 3; i++ {
				shelfItem, _ := shelf.inventory.Item(i)
				hotbarItem, _ := inv.Item(startIndex + slotOffset)

				_ = shelf.inventory.SetItem(i, hotbarItem)
				_ = inv.SetItem(startIndex+slotOffset, shelfItem)
				slotOffset++
			}
		}
	}
}

// findConnectedShelves busca estanterías encendidas a la izquierda.
func (s Shelf) findConnectedShelves(pos cube.Pos, tx *world.Tx) []Shelf {
	connected := []Shelf{s}
	leftFace := s.Facing.RotateLeft().Face()
	currPos := pos
	for i := 0; i < 2; i++ {
		currPos = currPos.Side(leftFace)
		if b, ok := tx.Block(currPos).(Shelf); ok && b.Powered && b.Facing == s.Facing {
			connected = append(connected, b)
		} else {
			break
		}
	}
	return connected
}

// UseOnBlock coloca el bloque y lo orienta opuesto al jugador.
func (s Shelf) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, s)
	if !used {
		return false
	}
	newS := NewShelf(s.Wood, s.Bamboo)
	newS.Facing = user.Rotation().Direction().Opposite()
	place(tx, pos, newS, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick actualiza el estado de Redstone.
func (s Shelf) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	powered := tx.RedstonePower(pos, cube.FaceDown, true) > 0 ||
		tx.RedstonePower(pos, cube.FaceUp, true) > 0 ||
		tx.RedstonePower(pos, cube.FaceNorth, true) > 0 ||
		tx.RedstonePower(pos, cube.FaceSouth, true) > 0 ||
		tx.RedstonePower(pos, cube.FaceEast, true) > 0 ||
		tx.RedstonePower(pos, cube.FaceWest, true) > 0

	if powered != s.Powered {
		s.Powered = powered
		if powered {
			tx.PlaySound(pos.Vec3Middle(), sound.PowerOn{})
		} else {
			tx.PlaySound(pos.Vec3Middle(), sound.PowerOff{})
		}
		tx.SetBlock(pos, s, nil)
	}
}

// EncodeItem ...
func (s Shelf) EncodeItem() (name string, meta int16) {
	if s.Bamboo {
		return "minecraft:bamboo_shelf", 0
	}
	return "minecraft:" + s.Wood.String() + "_shelf", 0
}

// EncodeBlock ...
func (s Shelf) EncodeBlock() (name string, properties map[string]any) {
	if s.Bamboo {
		name = "minecraft:bamboo_shelf"
	} else {
		name = "minecraft:" + s.Wood.String() + "_shelf"
	}
	return name, map[string]any{
		"minecraft:cardinal_direction": s.Facing.String(),
		"powered_bit":                  s.Powered,
		"powered_shelf_type":           int32(s.PoweredType),
	}
}

// DecodeNBT recupera el inventario guardado.
func (s Shelf) DecodeNBT(data map[string]any) any {
	wood, bamboo := s.Wood, s.Bamboo
	s = NewShelf(wood, bamboo)
	s.Facing = cube.Direction(nbtconv.Int32(data, "Facing"))
	s.Powered = nbtconv.Bool(data, "Powered")
	s.PoweredType = int(nbtconv.Int32(data, "PoweredType"))
	nbtconv.InvFromNBT(s.inventory, nbtconv.Slice(data, "Items"))
	return s
}

// EncodeNBT guarda el inventario.
func (s Shelf) EncodeNBT() map[string]any {
	if s.inventory == nil {
		wood, bamboo, facing, powered, pType := s.Wood, s.Bamboo, s.Facing, s.Powered, s.PoweredType
		s = NewShelf(wood, bamboo)
		s.Facing, s.Powered, s.PoweredType = facing, powered, pType
	}
	return map[string]any{
		"id":          "Shelf",
		"Items":       nbtconv.InvToNBT(s.inventory),
		"Facing":      int32(s.Facing),
		"Powered":     s.Powered,
		"PoweredType": int32(s.PoweredType),
	}
}

// allShelves ...
func allShelves() (shelves []world.Block) {
	woods := WoodTypes()
	for _, w := range woods {
		for _, d := range cube.Directions() {
			for _, p := range []bool{false, true} {
				for pt := 0; pt < 4; pt++ {
					shelves = append(shelves, Shelf{
						Wood:        w,
						Facing:      d,
						Powered:     p,
						PoweredType: pt,
					})
				}
			}
		}
	}
	for _, d := range cube.Directions() {
		for _, p := range []bool{false, true} {
			for pt := 0; pt < 4; pt++ {
				shelves = append(shelves, Shelf{
					Bamboo:      true,
					Facing:      d,
					Powered:     p,
					PoweredType: pt,
				})
			}
		}
	}
	return
}
