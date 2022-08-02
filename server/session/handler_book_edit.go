package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/exp/slices"
)

// BookEditHandler handles the BookEdit Packet
type BookEditHandler struct{}

// Handle ...
func (b BookEditHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.BookEdit)
	it, err := s.inv.Item(int(pk.InventorySlot))
	if err != nil {
		return fmt.Errorf("invalid inventory slot index %v", pk.InventorySlot)
	}
	book, ok := it.Item().(item.WritableBook)
	if !ok {
		return fmt.Errorf("inventory slot %v does not contain a writable book", pk.InventorySlot)
	}
	// Checks page number and text beforehand to reduce repetition, shouldn't matter as the default values
	// match for the data matches these constraints.
	page := int(pk.PageNumber)
	if page >= 50 || page < 0 {
		return fmt.Errorf("page number %v is out of bounds", pk.PageNumber)
	}
	if len(pk.Text) > 256 {
		return fmt.Errorf("text can not be longer than 256 bytes")
	}
	slot := int(pk.InventorySlot)
	switch pk.ActionType {
	case packet.BookActionReplacePage:
		book = book.SetPage(page, pk.Text)
	case packet.BookActionAddPage:
		if !book.PageExists(page) {
			return fmt.Errorf("may only come before a page which already exists")
		}
		if len(book.Pages) >= 50 {
			return fmt.Errorf("unable to add page beyond 50")
		}
		book = item.WritableBook{Pages: slices.Insert(book.Pages, int(page), pk.Text)}
	case packet.BookActionDeletePage:
		if !book.PageExists(page) {
			return fmt.Errorf("page number %v does not exist", pk.PageNumber)
		}
		book = item.WritableBook{Pages: slices.Delete(book.Pages, int(page), int(page+1))}
	case packet.BookActionSwapPages:
		if pk.SecondaryPageNumber >= 50 {
			return fmt.Errorf("page number out of bounds")
		}
		if !book.PageExists(page) || !book.PageExists(int(pk.SecondaryPageNumber)) {
			return fmt.Errorf("page numbers do not exist")
		}
		book = book.SwapPages(page, int(pk.SecondaryPageNumber))
	case packet.BookActionSign:
		// Error does not need to be handled as it's confirmed at the beginning that this slot contains a writable book.
		s.inv.SetItem(slot, item.NewStack(item.WrittenBook{Title: pk.Title, Author: pk.Author, Pages: book.Pages, Generation: item.OriginalGeneration()}, 1))
		return nil
	}
	// Error does not need to be handled as it's confirmed at the beginning that this slot contains a writable book.
	s.inv.SetItem(slot, item.NewStack(book, 1))
	return nil
}
