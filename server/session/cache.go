package session

import (
	"sync"

	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

var (
	craftingCacheMu      sync.Mutex
	craftingDataCache    packet.Packet
	craftingRecipesCache map[uint32]recipe.Recipe
)
