package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/exp/slices"
)

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
	// check page number and text beforehand to reduce repetition, shouldn't matter as the default values
	// match these constraints
	page := uint(pk.PageNumber)
	if page >= 50 {
		return fmt.Errorf("page number %v is out of bounds", pk.PageNumber)
	}
	if len(pk.Text) > 256 {
		return fmt.Errorf("text can not be longer than 256 bytes")
	}
	var pages []string
	slot := int(pk.InventorySlot)
	switch pk.ActionType {
	case packet.BookActionReplacePage:
		pages = book.Set(page, pk.Text)
		break
	case packet.BookActionAddPage:
		if !book.Exists(page) {
			return fmt.Errorf("may only come before a page which already exists")
		}
		if len(book.Pages) >= 50 {
			return fmt.Errorf("unable to add page beyond 50")
		}
		pages = slices.Insert(book.Pages, int(page), pk.Text)
		break
	case packet.BookActionDeletePage:
		if !book.Exists(page) {
			return fmt.Errorf("page number %v does not exist", pk.PageNumber)
		}
		pages = slices.Delete(book.Pages, int(page), int(page+1))
		break
	case packet.BookActionSwapPages:
		if pk.SecondaryPageNumber >= 50 {
			return fmt.Errorf("page number out of bounds")
		}
		if !book.Exists(page) || !book.Exists(uint(pk.SecondaryPageNumber)) {
			return fmt.Errorf("page numbers do not exist")
		}
		pages = book.Swap(page, uint(pk.SecondaryPageNumber))
		break
	case packet.BookActionSign:
		// Error does not need to be handled as it's confirmed at the begging that this slot contains a writable book.
		s.inv.SetItem(slot, item.NewStack(item.WrittenBook{Title: pk.Title, Author: pk.Author, Pages: book.Pages, Generation: 0}, 1))
		return nil
	}
	// Error does not need to be handled as it's confirmed at the begging that this slot contains a writable book.
	s.inv.SetItem(slot, item.NewStack(item.WritableBook{Pages: pages}, 1))
	return nil
}
