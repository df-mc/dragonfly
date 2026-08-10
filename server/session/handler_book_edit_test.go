package session

import (
	"strings"
	"testing"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// TestBookEditSignBounds verifies that the title and author a book is signed with are bounded. Both come from the
// client and are written into the item, which is then kept in the player's data and sent back out to every viewer.
func TestBookEditSignBounds(t *testing.T) {
	tests := []struct {
		name          string
		title, author string
		wantErr       bool
	}{
		{name: "a title and author a client can type", title: "Diary", author: "Steve"},
		{name: "a title at the maximum", title: strings.Repeat("a", 32), author: "Steve"},
		{name: "a title beyond the maximum", title: strings.Repeat("a", 33), author: "Steve", wantErr: true},
		{name: "an author beyond the maximum", title: "Diary", author: strings.Repeat("a", 33), wantErr: true},
		{name: "a megabyte of each", title: strings.Repeat("a", 1<<20), author: strings.Repeat("a", 1<<20), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{inv: inventory.New(36, nil)}
			if err := s.inv.SetItem(0, item.NewStack(item.BookAndQuill{}, 1)); err != nil {
				t.Fatalf("SetItem() = %v, want nil", err)
			}

			err := BookEditHandler{}.Handle(&packet.BookEdit{
				ActionType:    packet.BookActionSign,
				InventorySlot: 0,
				Title:         tt.title,
				Author:        tt.author,
			}, s, nil, nil)

			if tt.wantErr && err == nil {
				it, _ := s.inv.Item(0)
				book, _ := it.Item().(item.WrittenBook)
				t.Errorf("signing was accepted with a title of %v and an author of %v bytes, want an error (book title is now %v bytes)",
					len(tt.title), len(tt.author), len(book.Title))
			}
			if !tt.wantErr && err != nil {
				t.Errorf("signing = %v, want nil", err)
			}
		})
	}
}
