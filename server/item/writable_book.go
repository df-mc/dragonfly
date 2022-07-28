package item

// WritableBook is an item used to write written books.
type WritableBook struct {
	// Pages represents the pages within the book.
	Pages []string
}

// Page returns a specific page from the book. If the page exists, it will return the content and true, otherwise
// it will return an empty string and false.
func (w WritableBook) Page(page int) (string, bool) {
	if page < 0 {
		panic("negative page number")
	}
	if len(w.Pages) <= page {
		return "", false
	}
	return w.Pages[page], true
}

// MaxCount always returns 1.
func (w WritableBook) MaxCount() int {
	return 1
}

// PageExists checks to see whether a page exists or not.
func (w WritableBook) PageExists(page int) bool {
	return page >= 0 && len(w.Pages) > page
}

// Set writes a page to the book, if the page doesn't exist it will be created. It will panic if the
// text is longer then 256 characters. It will return a new book representing this data.
func (w WritableBook) Set(page int, text string) WritableBook {
	if page < 0 {
		panic("negative page number")
	}
	if len(text) > 256 {
		panic("text longer then 256 bytes")
	}
	if !w.PageExists(page) {
		newPages := make([]string, page+1)
		copy(newPages, w.Pages)
		w.Pages = newPages
	}
	w.Pages[page] = text
	return w
}

// Swap swaps two different pages, it will panic if the largest of the two numbers doesn't exist. It will
// return the newly updated pages.
func (w WritableBook) Swap(page1, page2 int) WritableBook {
	if page1 < 0 || page2 < 0 {
		panic("negative page number")
	}
	if w.PageExists(max(page1, page2)) {
		panic("invalid page number")
	}
	content1 := w.Pages[page1]
	content2 := w.Pages[page2]
	w.Pages[page1] = content2
	w.Pages[page2] = content1
	return w
}

// DecodeNBT ...
func (w WritableBook) DecodeNBT(data map[string]any) any {
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
func (w WritableBook) EncodeNBT() map[string]any {
	if len(w.Pages) == 0 {
		return nil
	}
	data := map[string]any{}
	var pages []any
	for _, page := range w.Pages {
		pages = append(pages, map[string]any{
			"text": page,
		})
	}
	data["pages"] = pages
	return data
}

// max ...
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (w WritableBook) EncodeItem() (name string, meta int16) {
	return "minecraft:writable_book", 0
}
