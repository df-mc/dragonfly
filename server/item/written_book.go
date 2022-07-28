package item

type WrittenBook struct {
	Title      string
	Author     string
	Generation uint
	Glinted
	Pages []string
}

// MaxCount always returns 1.
func (w WrittenBook) MaxCount() int {
	return 1
}

// Page returns a specific page from the book. If the page exists, it will return the content and true, otherwise
// it will return an empty string and false.
func (w WrittenBook) Page(page uint) (string, bool) {
	p := int(page)
	if len(w.Pages) <= p {
		return "", false
	}
	return w.Pages[p], true
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
