package projectile

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/world"
)

func init() {
	world.RegisterItem(Snowball{})
	creative.RegisterItem(item.NewStack(Snowball{}, 1))
}
