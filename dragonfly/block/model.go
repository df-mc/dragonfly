package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
)

// solid represents a block that is fully solid. It always returns a model.Solid when Model is called.
type solid struct{}

// Model ...
func (solid) Model() model.Model {
	return model.Solid{}
}

// empty represents a block that is fully empty/transparent, such as air or a plant. It always returns a
// model.Empty when Model is called.
type empty struct{}

// Model ...
func (empty) Model() model.Empty {
	return model.Empty{}
}
