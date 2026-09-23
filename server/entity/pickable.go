package entity

import "github.com/df-mc/dragonfly/server/item"

// Pickable represents an entity that may be picked by a player.
type Pickable interface {
	// Pick returns the item that is picked when the entity is picked.
	Pick() item.Stack
}
