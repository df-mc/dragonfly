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
	Generation uint
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
	if pages, ok := data["pages"].([]map[string]string); ok {
		for i, page := range pages {
			w.Pages[i] = page["text"]
		}
	}
	if v, ok := data["title"].(string); ok {
		w.Title = v
	}
	if v, ok := data["author"].(string); ok {
		w.Author = v
	}
	return w
}

// EncodeNBT ...
func (w WrittenBook) EncodeNBT() map[string]any {
	data := map[string]any{}
	var pages []map[string]string
	for _, page := range w.Pages {
		pages = append(pages, map[string]string{
			"text":      page,
			"photoname": "",
		})
	}
	data["pages"] = pages
	data["author"] = w.Author
	data["title"] = w.Title
	data["generation"] = byte(w.Generation)
	return data
}

// EncodeItem ...
func (w WrittenBook) EncodeItem() (name string, meta int16) {
	return "minecraft:written_book", 0
}
