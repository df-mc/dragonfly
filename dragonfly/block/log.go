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

func (OakLog) Name() string {
	return "Oak Log"
}

func (l OakLog) Drops() []inventory.Item {
	return []inventory.Item{OakLog{Stripped: l.Stripped}}
}

func (SpruceLog) Name() string {
	return "Spruce Log"
}

func (l SpruceLog) Drops() []inventory.Item {
	return []inventory.Item{SpruceLog{Stripped: l.Stripped}}
}

func (BirchLog) Name() string {
	return "Birch Log"
}

func (l BirchLog) Drops() []inventory.Item {
	return []inventory.Item{BirchLog{Stripped: l.Stripped}}
}

func (JungleLog) Name() string {
	return "Jungle Log"
}

func (l JungleLog) Drops() []inventory.Item {
	return []inventory.Item{JungleLog{Stripped: l.Stripped}}
}

func (AcaciaLog) Name() string {
	return "Acacia Log"
}

func (l AcaciaLog) Drops() []inventory.Item {
	return []inventory.Item{AcaciaLog{Stripped: l.Stripped}}
}

func (DarkOakLog) Name() string {
	return "Dark Oak Log"
}

func (l DarkOakLog) Drops() []inventory.Item {
	return []inventory.Item{DarkOakLog{Stripped: l.Stripped}}
}
