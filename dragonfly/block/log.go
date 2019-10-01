package block

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/item/inventory"
)

// Log is a naturally occurring block found in trees, primarily used to create planks. It comes in six
// species: oak, spruce, birch, jungle, acacia, and dark oak.
// Stripped log is a variant obtained by using an axe on a log.
type (
	OakLog     log
	SpruceLog  log
	BirchLog   log
	JungleLog  log
	AcaciaLog  log
	DarkOakLog log

	// log implements the base of each of the logs above.
	log struct {
		Axis     Axis
		Stripped bool
	}
)

func (l OakLog) Name() string {
	return stripped(log(l)) + "Oak Log"
}

func (l OakLog) Drops() []inventory.Item {
	return []inventory.Item{OakLog{Stripped: l.Stripped}}
}

func (l SpruceLog) Name() string {
	return stripped(log(l)) + "Spruce Log"
}

func (l SpruceLog) Drops() []inventory.Item {
	return []inventory.Item{SpruceLog{Stripped: l.Stripped}}
}

func (l BirchLog) Name() string {
	return stripped(log(l)) + "Birch Log"
}

func (l BirchLog) Drops() []inventory.Item {
	return []inventory.Item{BirchLog{Stripped: l.Stripped}}
}

func (l JungleLog) Name() string {
	return stripped(log(l)) + "Jungle Log"
}

func (l JungleLog) Drops() []inventory.Item {
	return []inventory.Item{JungleLog{Stripped: l.Stripped}}
}

func (l AcaciaLog) Name() string {
	return stripped(log(l)) + "Acacia Log"
}

func (l AcaciaLog) Drops() []inventory.Item {
	return []inventory.Item{AcaciaLog{Stripped: l.Stripped}}
}

func (l DarkOakLog) Name() string {
	return stripped(log(l)) + "Dark Oak Log"
}

func (l DarkOakLog) Drops() []inventory.Item {
	return []inventory.Item{DarkOakLog{Stripped: l.Stripped}}
}

// stripped returns the name prefix 'Stripped ' for a log if it is stripped, otherwise an empty string.
func stripped(l log) string {
	if l.Stripped {
		return "Stripped "
	}
	return ""
}
