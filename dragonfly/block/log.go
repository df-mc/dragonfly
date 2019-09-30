package block

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/item/inventory"
)

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

// Name ...
func (l OakLog) Name() string {
	return "Oak Log"
}

// Drops ...
func (l OakLog) Drops() []inventory.Item {
	return []inventory.Item{OakLog{Stripped: l.Stripped}}
}

// Name ...
func (l SpruceLog) Name() string {
	return "Spruce Log"
}

// Drops ...
func (l SpruceLog) Drops() []inventory.Item {
	return []inventory.Item{SpruceLog{Stripped: l.Stripped}}
}

// Name ...
func (l BirchLog) Name() string {
	return "Birch Log"
}

// Drops ...
func (l BirchLog) Drops() []inventory.Item {
	return []inventory.Item{BirchLog{Stripped: l.Stripped}}
}

// Name ...
func (l JungleLog) Name() string {
	return "Jungle Log"
}

// Drops ...
func (l JungleLog) Drops() []inventory.Item {
	return []inventory.Item{JungleLog{Stripped: l.Stripped}}
}

// Name ...
func (l AcaciaLog) Name() string {
	return "Acacia Log"
}

// Drops ...
func (l AcaciaLog) Drops() []inventory.Item {
	return []inventory.Item{AcaciaLog{Stripped: l.Stripped}}
}

// Name ...
func (l DarkOakLog) Name() string {
	return "Dark Oak Log"
}

// Drops ...
func (l DarkOakLog) Drops() []inventory.Item {
	return []inventory.Item{DarkOakLog{Stripped: l.Stripped}}
}
