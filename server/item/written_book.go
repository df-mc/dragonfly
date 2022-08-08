package item

// WrittenBook is an item created after a book and quill is signed. It appears the same as a regular book, but
// without the quill, and has an enchanted-looking glint.
type WrittenBook struct {
	// Title is the title of the book
	Title string
	// Author is the author of the book
	Author string
	// Generation is the generation of the book. The copy tier of the book. 0 = original, 1 = copy of original,
	// 2 = copy of copy.
	Generation WrittenBookGeneration
	// Pages represents the pages within the book.
	Pages []string
}

// MaxCount always returns 1.
func (w WrittenBook) MaxCount() int {
	return 1
}

// Page returns a specific page from the book. If the page exists, it will return the content and true, otherwise
// it will return an empty string and false.
func (w WrittenBook) Page(page int) (string, bool) {
	if page < 0 {
		panic("negative page number")
	}
	if len(w.Pages) <= page {
		return "", false
	}
	return w.Pages[page], true
}

// DecodeNBT ...
func (w WrittenBook) DecodeNBT(data map[string]any) any {
	if pages, ok := data["pages"].([]any); ok {
		w.Pages = make([]string, len(pages))
		for i, page := range pages {
			w.Pages[i] = page.(map[string]any)["text"].(string)
		}
	}
	if v, ok := data["title"].(string); ok {
		w.Title = v
	}
	if v, ok := data["author"].(string); ok {
		w.Author = v
	}
	if v, ok := data["generation"].(uint8); ok {
		switch v {
		case 0:
			w.Generation = OriginalGeneration()
		case 1:
			w.Generation = CopyGeneration()
		case 2:
			w.Generation = CopyOfCopyGeneration()
		}
	}
	return w
}

// EncodeNBT ...
func (w WrittenBook) EncodeNBT() map[string]any {
	pages := make([]any, 0, len(w.Pages))
	for _, page := range w.Pages {
		pages = append(pages, map[string]any{"text": page})
	}
	return map[string]any{
		"pages":      pages,
		"author":     w.Author,
		"title":      w.Title,
		"generation": w.Generation.Uint8(),
	}
}

// EncodeItem ...
func (w WrittenBook) EncodeItem() (name string, meta int16) {
	return "minecraft:written_book", 0
}
