package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/go-gl/mathgl/mgl64"
)

type EventBlockBurn struct {
	World    *World
	Position cube.Pos
	*event.Context
}

type EventClose struct {
	World *World
}

type EventEntityDespawn struct {
	World  *World
	Entity Entity
}

type EventEntitySpawn struct {
	World  *World
	Entity Entity
}

type EventFireSpread struct {
	World *World
	From  cube.Pos
	To    cube.Pos
	*event.Context
}

type EventLiquidDecay struct {
	World    *World
	Position cube.Pos
	From     Liquid
	To       Liquid
	*event.Context
}

type EventLiquidFlow struct {
	World    *World
	From     cube.Pos
	To       cube.Pos
	Liquid   Liquid
	Replaced Block
	*event.Context
}

type EventLiquidHarden struct {
	World          *World
	Position       cube.Pos
	LiquidHardened Block
	OtherLiquid    Block
	NewBlock       Block
	*event.Context
}

type EventSound struct {
	World    *World
	Position mgl64.Vec3
	Sound    Sound
	*event.Context
}
