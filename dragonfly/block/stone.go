package block

type (
	// Stone is a block found underground in the Overworld or on mountains.
	Stone struct{}
	// Granite is a type of igneous rock.
	Granite polishable
	// Diorite is a type of igneous rock.
	Diorite polishable
	// Andesite is a type of igneous rock.
	Andesite polishable

	// polishable forms the base of blocks that may be polished.
	polishable struct {
		Polished bool
	}
)

func (Stone) Name() string {
	return "Stone"
}

func (g Granite) Name() string {
	if g.Polished {
		return "Polished Granite"
	}
	return "Granite"
}

func (d Diorite) Name() string {
	if d.Polished {
		return "Polished Diorite"
	}
	return "Diorite"
}

func (a Andesite) Name() string {
	if a.Polished {
		return "Polished Andesite"
	}
	return "Andesite"
}
