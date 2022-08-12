package item

import "golang.org/x/exp/slices"

// BookAndQuill is an item used to write WrittenBook(s).
type BookAndQuill struct {
	// Pages represents the pages within the book.
	Pages []string
}

// MaxCount always returns 1.
func (b BookAndQuill) MaxCount() int {
	return 1
}

// Page returns a specific page from the book and true when the page exists. It will otherwise return an empty string
// and false.
func (b BookAndQuill) Page(page int) (string, bool) {
	if page < 0 || len(b.Pages) <= page {
		return "", false
	}
	return b.Pages[page], true
}

// DeletePage attempts to delete a page from the book.
func (b BookAndQuill) DeletePage(page int) BookAndQuill {
	if page < 0 || page >= 50 {
		panic("invalid page number")
	}
	if _, ok := b.Page(page); !ok {
		panic("cannot delete nonexistent page")
	}
	b.Pages = slices.Delete(b.Pages, page, page+1)
	return b
}

// InsertPage attempts to insert a page within the book
func (b BookAndQuill) InsertPage(page int, text string) BookAndQuill {
	if page < 0 || page >= 50 {
		panic("invalid page number")
	}
	if len(text) > 256 {
		panic("text longer then 256 bytes")
	}
	if page > len(b.Pages) {
		panic("unable to insert page at invalid position")
	}
	b.Pages = slices.Insert(b.Pages, page, text)
	return b
}

// SetPage writes a page to the book, if the page doesn't exist it will be created. It will panic if the
// text is longer then 256 characters. It will return a new book representing this data.
func (b BookAndQuill) SetPage(page int, text string) BookAndQuill {
	if page < 0 || page >= 50 {
		panic("invalid page number")
	}
	if len(text) > 256 {
		panic("text longer then 256 bytes")
	}
	if _, ok := b.Page(page); !ok {
		pages := make([]string, page+1)
		copy(pages, b.Pages)
		b.Pages = pages
	}
	b.Pages[page] = text
	return b
}

// SwapPages swaps two different pages, it will panic if the largest of the two numbers doesn't exist. It will
// return the newly updated pages.
func (b BookAndQuill) SwapPages(pageOne, pageTwo int) BookAndQuill {
	if pageOne < 0 || pageTwo < 0 {
		panic("negative page number")
	}
	if _, ok := b.Page(max(pageOne, pageTwo)); !ok {
		panic("invalid page number")
	}
	temp := b.Pages[pageOne]
	b.Pages[pageOne] = b.Pages[pageTwo]
	b.Pages[pageTwo] = temp
	return b
}

// DecodeNBT ...
func (b BookAndQuill) DecodeNBT(data map[string]any) any {
	if pages, ok := data["pages"].([]any); ok {
		for _, page := range pages {
			if pageData, ok := page.(map[string]any); ok {
				if text, ok := pageData["text"].(string); ok {
					b.Pages = append(b.Pages, text)
				}
			}
		}
	}
	return b
}

// EncodeNBT ...
func (b BookAndQuill) EncodeNBT() map[string]any {
	pages := make([]any, 0, len(b.Pages))
	for _, page := range b.Pages {
		pages = append(pages, map[string]any{"text": page})
	}
	return map[string]any{"pages": pages}
}

// EncodeItem ...
func (b BookAndQuill) EncodeItem() (name string, meta int16) {
	return "minecraft:writable_book", 0
}

// max ...
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
