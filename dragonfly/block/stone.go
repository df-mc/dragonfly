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

func (Stone) Name() string {
	return "Stone"
}

func (Granite) Name() string {
	return "Granite"
}

func (PolishedGranite) Name() string {
	return "Polished Granite"
}

func (Diorite) Name() string {
	return "Diorite"
}

func (PolishedDiorite) Name() string {
	return "Polished Diorite"
}

func (Andesite) Name() string {
	return "Andesite"
}

func (PolishedAndesite) Name() string {
	return "Polished Andesite"
}
