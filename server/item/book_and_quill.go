package item

// BookAndQuill is an item used to write WrittenBook(s).
type BookAndQuill struct {
	// Pages represents the pages within the book.
	Pages []string
}

// MaxCount always returns 1.
func (w BookAndQuill) MaxCount() int {
	return 1
}

// ValidPage checks to see whether a page exists or not.
func (w BookAndQuill) ValidPage(page int) bool {
	return page >= 0 && len(w.Pages) > page
}

// Page returns a specific page from the book. If the page does not exist
func (w BookAndQuill) Page(page int) (string, bool) {
	if page < 0 {
		panic("negative page number")
	}
	if len(w.Pages) <= page {
		return "", false
	}
	return w.Pages[page], true
}

// SetPage writes a page to the book, if the page doesn't exist it will be created. It will panic if the
// text is longer then 256 characters. It will return a new book representing this data.
func (w BookAndQuill) SetPage(page int, text string) BookAndQuill {
	if page < 0 {
		panic("negative page number")
	}
	if len(text) > 256 {
		panic("text longer then 256 bytes")
	}
	if !w.ValidPage(page) {
		pages := make([]string, page+1)
		copy(pages, w.Pages)
		w.Pages = pages
	}
	w.Pages[page] = text
	return w
}

// SwapPages swaps two different pages, it will panic if the largest of the two numbers doesn't exist. It will
// return the newly updated pages.
func (w BookAndQuill) SwapPages(pageOne, pageTwo int) BookAndQuill {
	if pageOne < 0 || pageTwo < 0 {
		panic("negative page number")
	}
	if !w.ValidPage(max(pageOne, pageTwo)) {
		panic("invalid page number")
	}
	contentOne := w.Pages[pageOne]
	contentTwo := w.Pages[pageTwo]
	w.Pages[pageOne] = contentTwo
	w.Pages[pageTwo] = contentOne
	return w
}

// DecodeNBT ...
func (w BookAndQuill) DecodeNBT(data map[string]any) any {
	if pages, ok := data["pages"].([]any); ok {
		for _, page := range pages {
			if pageData, ok := page.(map[string]any); ok {
				if text, ok := pageData["text"].(string); ok {
					w.Pages = append(w.Pages, text)
				}
			}
		}
	}
	return w
}

// EncodeNBT ...
func (w BookAndQuill) EncodeNBT() map[string]any {
	pages := make([]any, 0, len(w.Pages))
	for _, page := range w.Pages {
		pages = append(pages, map[string]any{"text": page})
	}
	return map[string]any{"pages": pages}
}

// EncodeItem ...
func (w BookAndQuill) EncodeItem() (name string, meta int16) {
	return "minecraft:writable_book", 0
}

// max ...
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
