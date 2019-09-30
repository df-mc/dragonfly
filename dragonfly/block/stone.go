package block

type (
	Stone            struct{}
	Granite          struct{}
	PolishedGranite  struct{}
	Diorite          struct{}
	PolishedDiorite  struct{}
	Andesite         struct{}
	PolishedAndesite struct{}
)

// Name ...
func (l Stone) Name() string {
	return "Stone"
}

// Name ...
func (l Granite) Name() string {
	return "Granite"
}

// Name ...
func (l PolishedGranite) Name() string {
	return "Polished Granite"
}

// Name ...
func (l Diorite) Name() string {
	return "Diorite"
}

// Name ...
func (l PolishedDiorite) Name() string {
	return "Polished Diorite"
}

// Name ...
func (l Andesite) Name() string {
	return "Andesite"
}

// Name ...
func (l PolishedAndesite) Name() string {
	return "Polished Andesite"
}
