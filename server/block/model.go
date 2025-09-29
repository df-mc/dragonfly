package block

import (
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// solid represents a block that is fully solid. It always returns a model.Solid when Model is called.
type solid struct{}

func (solid) Model() world.BlockModel {
	return model.Solid{}
}

// empty represents a block that is fully empty/transparent, such as air or a plant. It always returns a
// model.Empty when Model is called.
type empty struct{}

func (empty) Model() world.BlockModel {
	return model.Empty{}
}

// chest represents a block that has a model of a chest.
type chest struct{}

func (chest) Model() world.BlockModel {
	return model.Chest{}
}

// carpet represents a block that has a model of a carpet.
type carpet struct{}

func (carpet) Model() world.BlockModel {
	return model.Carpet{}
}

// tilledGrass represents a block that has a model of farmland or dirt paths.
type tilledGrass struct{}

func (tilledGrass) Model() world.BlockModel {
	return model.TilledGrass{}
}

// leaves represents a block that has a model of leaves. A full block but with no solid faces.
type leaves struct{}

func (leaves) Model() world.BlockModel {
	return model.Leaves{}
}

// thin represents a thin, partial block such as a glass pane or an iron bar, that connects to nearby solid faces.
type thin struct{}

func (thin) Model() world.BlockModel {
	return model.Thin{}
}
