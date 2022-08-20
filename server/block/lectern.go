package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Lectern is a librarian's job site block found in villages. It is used to hold books for multiple players to read in
// multiplayer.
// TODO: Redstone functionality.
type Lectern struct {
	bass
	sourceWaterDisplacer

	// Facing represents the direction the Lectern is facing.
	Facing cube.Direction
	// Book is the book currently held by the Lectern.
	Book item.Stack
	// Page is the page the Lectern is currently on in the book.
	Page int
}

// Model ...
func (Lectern) Model() world.BlockModel {
	//TODO implement me
	return model.Solid{}
}

// FuelInfo ...
func (Lectern) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// SideClosed ...
func (Lectern) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// BreakInfo ...
func (l Lectern) BreakInfo() BreakInfo {
	d := []item.Stack{item.NewStack(Lectern{}, 1)}
	if !l.Book.Empty() {
		d = append(d, l.Book)
	}
	return newBreakInfo(2, alwaysHarvestable, axeEffective, simpleDrops(d...))
}

// UseOnBlock ...
func (l Lectern) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, l)
	if !used {
		return false
	}
	l.Facing = user.Facing().Opposite()
	place(w, pos, l, user, ctx)
	return placed(ctx)
}

// readableBook represents a book that can be read through a lectern.
type readableBook interface {
	// TotalPages returns the total number of pages in the book.
	TotalPages() int
	// Page returns a specific page from the book and true when the page exists. It will otherwise return an empty string
	// and false.
	Page(page int) (string, bool)
}

// Activate ...
func (l Lectern) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, ctx *item.UseContext) bool {
	if !l.Book.Empty() {
		// We can't put a book on the lectern if it's full.
		return true
	}
	held, _ := u.HeldItems()
	if _, ok := held.Item().(readableBook); ok {
		l.Page = 0
		l.Book = held

		w.SetBlock(pos, l, nil)
		w.PlaySound(pos.Vec3Centre(), sound.LecternBookPlace{})
		ctx.SubtractFromCount(1)
	}
	return true
}

// Punch ...
func (l Lectern) Punch(pos cube.Pos, _ cube.Face, w *world.World, _ item.User) {
	if l.Book.Empty() {
		// We can't remove a book from the lectern if there isn't one.
		return
	}

	dropItem(w, l.Book, pos.Side(cube.FaceUp).Vec3Middle())

	l.Book = item.Stack{}
	w.SetBlock(pos, l, nil)
	w.PlaySound(pos.Vec3Centre(), sound.Attack{})
}

// TurnPage updates the page the lectern is currently on to the page given.
func (l Lectern) TurnPage(pos cube.Pos, w *world.World, page int) error {
	if page == l.Page {
		// We're already on the correct page, so ignore this packet.
		return nil
	}
	if l.Book.Empty() {
		return fmt.Errorf("lectern at %v is empty", pos)
	}
	if r, ok := l.Book.Item().(readableBook); ok && (page >= r.TotalPages() || page < 0) {
		return fmt.Errorf("page number %d is out of bounds", page)
	}
	l.Page = page
	w.SetBlock(pos, l, nil)
	return nil
}

// EncodeNBT ...
func (l Lectern) EncodeNBT() map[string]any {
	m := map[string]any{
		"hasBook": boolByte(!l.Book.Empty()),
		"page":    int32(l.Page),
		"id":      "Lectern",
	}
	if r, ok := l.Book.Item().(readableBook); ok {
		m["book"] = nbtconv.WriteItem(l.Book, true)
		m["totalPages"] = int32(r.TotalPages())
	}
	return m
}

// DecodeNBT ...
func (l Lectern) DecodeNBT(m map[string]any) any {
	l.Page = int(nbtconv.Map[int32](m, "page"))
	l.Book = nbtconv.MapItem(m, "book")
	return l
}

// EncodeItem ...
func (Lectern) EncodeItem() (name string, meta int16) {
	return "minecraft:lectern", 0
}

// EncodeBlock ...
func (l Lectern) EncodeBlock() (string, map[string]any) {
	return "minecraft:lectern", map[string]any{
		"direction":   int32(horizontalDirection(l.Facing)),
		"powered_bit": uint8(0), // We don't support redstone, anyway.
	}
}

// allLecterns ...
func allLecterns() (lecterns []world.Block) {
	for _, d := range cube.Directions() {
		lecterns = append(lecterns, Lectern{Facing: d})
	}
	return
}
