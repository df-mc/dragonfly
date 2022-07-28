package item

type WritableBook struct {
	Pages []string
}

// Page returns a specific page from the book. If the page exists, it will return the content and true, otherwise
// it will return an empty string and false.
func (w WritableBook) Page(page uint) (string, bool) {
	p := int(page)
	if len(w.Pages) <= p {
		return "", false
	}
	return w.Pages[p], true
}

// MaxCount always returns 1.
func (w WritableBook) MaxCount() int {
	return 1
}

// Exists checks to see weather a page exists or not
func (w WritableBook) Exists(page uint) bool {
	return len(w.Pages) > int(page)
}

// Set writes a page to the book, if the page doesn't exist it will be created. It will panic if the
// text is longer then 256 characters. It will return the newly updated pages.
func (w WritableBook) Set(page uint, text string) []string {
	if len(text) > 256 {
		panic("text longer then 256 bytes")
	}
	pages := w.Pages
	if !w.Exists(page) {
		newPages := make([]string, page+1)
		copy(newPages, pages)
		pages = newPages
	}
	pages[page] = text
	return pages
}

// Swap swaps two different pages, it will panic if the largest of the two numbers doesn't exist. it will
// return the newly updated pages.
func (w WritableBook) Swap(page1, page2 uint) []string {
	pages := w.Pages
	if w.Exists(max(page1, page2)) {
		panic("invalid page number")
	}
	content1 := pages[page1]
	content2 := pages[page2]
	pages[page1] = content2
	pages[page2] = content1
	return pages
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
func max(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}

func (w WritableBook) EncodeItem() (name string, meta int16) {
	return "minecraft:writable_book", 0
}
