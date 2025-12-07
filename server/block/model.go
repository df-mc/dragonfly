package block

import (
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// solid represents a block that is fully solid. It always returns a model.Solid when Model is called.
type solid struct{}

// Model ...
func (solid) Model() world.BlockModel {
	return model.Solid{}
}

// empty represents a block that is fully empty/transparent, such as air or a plant. It always returns a
// model.Empty when Model is called.
type empty struct{}

// Model ...
func (empty) Model() world.BlockModel {
	return model.Empty{}
}

// chest represents a block that has a model of a chest.
type chest struct{}

// Model ...
func (chest) Model() world.BlockModel {
	return model.Chest{}
}

// carpet represents a block that has a model of a carpet.
type carpet struct{}

// Model ...
func (carpet) Model() world.BlockModel {
	return model.Carpet{}
}

// tilledGrass represents a block that has a model of farmland or dirt paths.
type tilledGrass struct{}

// Model ...
func (tilledGrass) Model() world.BlockModel {
	return model.TilledGrass{}
}

// leaves represents a block that has a model of leaves. A full block but with no solid faces.
type leaves struct{}

// Model ...
func (leaves) Model() world.BlockModel {
	return model.Leaves{}
}

// thin represents a thin, partial block such as a glass pane or an iron bar, that connects to nearby solid faces.
type thin struct{}

// Model ...
func (thin) Model() world.BlockModel {
	return model.Thin{}
}
