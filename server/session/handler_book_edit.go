package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/exp/slices"
)

// BookEditHandler handles the BookEdit packet.
type BookEditHandler struct{}

// Handle ...
func (b BookEditHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.BookEdit)

	it, err := s.inv.Item(int(pk.InventorySlot))
	if err != nil {
		return fmt.Errorf("invalid inventory slot index %v", pk.InventorySlot)
	}
	book, ok := it.Item().(item.BookAndQuill)
	if !ok {
		return fmt.Errorf("inventory slot %v does not contain a writable book", pk.InventorySlot)
	}

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
		if len(book.Pages) >= 50 {
			return fmt.Errorf("unable to add page beyond 50")
		}
		if page >= len(book.Pages) && page <= len(book.Pages)+2 {
			book = book.SetPage(page, "")
			break
		}
		if _, ok := book.Page(page); !ok {
			return fmt.Errorf("unable to insert page at %v", pk.PageNumber)
		}
		book = item.BookAndQuill{Pages: slices.Insert(book.Pages, page, pk.Text)}
	case packet.BookActionDeletePage:
		// We break here instead of returning an error because the client can be a page or two ahead in the UI then
		// the actual pages representation server side. The client still sends the deletion indexes.
		if _, ok := book.Page(page); !ok {
			break
		}
		book = item.BookAndQuill{Pages: slices.Delete(book.Pages, page, page+1)}
	case packet.BookActionSwapPages:
		if pk.SecondaryPageNumber >= 50 {
			return fmt.Errorf("page number out of bounds")
		}
		_, ok := book.Page(page)
		_, ok2 := book.Page(int(pk.SecondaryPageNumber))
		// We break here instead of returning an error because the client can try to swap pages that don't exist.
		// This happens as a result of the client being a page or two ahead in the UI then the actual pages
		// representation server side. The client still sends the swap indexes.
		if !ok || !ok2 {
			break
		}
		book = book.SwapPages(page, int(pk.SecondaryPageNumber))
	case packet.BookActionSign:
		_ = s.inv.SetItem(slot, item.NewStack(item.WrittenBook{Title: pk.Title, Author: pk.Author, Pages: book.Pages, Generation: item.OriginalGeneration()}, 1))
		return nil
	}
	_ = s.inv.SetItem(slot, item.NewStack(book, 1))
	return nil
}
