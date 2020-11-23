package plugin

import (
	"github.com/df-mc/dragonfly/dragonfly/cmd"
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"github.com/df-mc/dragonfly/dragonfly/entity/healing"
	"github.com/df-mc/dragonfly/dragonfly/event"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/player"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"net"
)

// Plugin can be used to create an implementation of Handler (player/handlers.go), without implementing
// all Handler methods, but only with necessary
type Plugin struct {
	handlers []interface{}
}

func MakePlugin(handler interface{}) *Plugin {
	return &Plugin{
		handlers: []interface{}{
			handler,
		},
	}
}

func JoinPlugins(handlers []interface{}) *Plugin {
	return &Plugin{
		handlers: handlers,
	}
}

func (plugin Plugin) HandleMove(p *player.Player, ctx *event.Context, newPos mgl64.Vec3, newYaw, newPitch float64) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.MoveHandler); ok {
			h.HandleMove(p, ctx, newPos, newYaw, newPitch)
		}
	}
}

func (plugin Plugin) HandleTeleport(p *player.Player, ctx *event.Context, pos mgl64.Vec3) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.TeleportHandler); ok {
			h.HandleTeleport(p, ctx, pos)
		}
	}
}

func (plugin Plugin) HandleChat(p *player.Player, ctx *event.Context, message *string) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.ChatHandler); ok {
			h.HandleChat(p, ctx, message)
		}
	}
}

func (plugin Plugin) HandleFoodLoss(p *player.Player, ctx *event.Context, from, to int) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.FoodLossHandler); ok {
			h.HandleFoodLoss(p, ctx, from, to)
		}
	}
}

func (plugin Plugin) HandleHeal(p *player.Player, ctx *event.Context, health *float64, src healing.Source) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.HealHandler); ok {
			h.HandleHeal(p, ctx, health, src)
		}
	}
}

func (plugin Plugin) HandleHurt(p *player.Player, ctx *event.Context, damage *float64, src damage.Source) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.HurtHandler); ok {
			h.HandleHurt(p, ctx, damage, src)
		}
	}
}

func (plugin Plugin) HandleDeath(p *player.Player, src damage.Source) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.DeathHandler); ok {
			h.HandleDeath(p, src)
		}
	}
}

func (plugin Plugin) HandleRespawn(p *player.Player, pos *mgl64.Vec3) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.RespawnHandler); ok {
			h.HandleRespawn(p, pos)
		}
	}
}

func (plugin Plugin) HandleStartBreak(p *player.Player, ctx *event.Context, pos world.BlockPos) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.StartBreakHandler); ok {
			h.HandleStartBreak(p, ctx, pos)
		}
	}
}

func (plugin Plugin) HandleBlockBreak(p *player.Player, ctx *event.Context, pos world.BlockPos) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.BlockBreakHandler); ok {
			h.HandleBlockBreak(p, ctx, pos)
		}
	}
}

func (plugin Plugin) HandleBlockPlace(p *player.Player, ctx *event.Context, pos world.BlockPos, b world.Block) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.BlockPlaceHandler); ok {
			h.HandleBlockPlace(p, ctx, pos, b)
		}
	}
}

func (plugin Plugin) HandleBlockPick(p *player.Player, ctx *event.Context, pos world.BlockPos, b world.Block) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.BlockPickHandler); ok {
			h.HandleBlockPick(p, ctx, pos, b)
		}
	}
}

func (plugin Plugin) HandleItemUse(p *player.Player, ctx *event.Context) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.ItemUseHandler); ok {
			h.HandleItemUse(p, ctx)
		}
	}
}

func (plugin Plugin) HandleItemUseOnBlock(p *player.Player, ctx *event.Context, pos world.BlockPos, face world.Face, clickPos mgl64.Vec3) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.ItemUseOnBlockHandler); ok {
			h.HandleItemUseOnBlock(p, ctx, pos, face, clickPos)
		}
	}
}

func (plugin Plugin) HandleItemUseOnEntity(p *player.Player, ctx *event.Context, e world.Entity) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.ItemUseOnEntityHandler); ok {
			h.HandleItemUseOnEntity(p, ctx, e)
		}
	}
}

func (plugin Plugin) HandleAttackEntity(p *player.Player, ctx *event.Context, e world.Entity) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.AttackEntityHandler); ok {
			h.HandleAttackEntity(p, ctx, e)
		}
	}
}

func (plugin Plugin) HandleItemDamage(p *player.Player, ctx *event.Context, i item.Stack, damage int) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.ItemDamageHandler); ok {
			h.HandleItemDamage(p, ctx, i, damage)
		}
	}
}

func (plugin Plugin) HandleItemPickup(p *player.Player, ctx *event.Context, i item.Stack) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.ItemPickupHandler); ok {
			h.HandleItemPickup(p, ctx, i)
		}
	}
}

func (plugin Plugin) HandleItemDrop(p *player.Player, ctx *event.Context, e *entity.Item) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.ItemDropHandler); ok {
			h.HandleItemDrop(p, ctx, e)
		}
	}
}

func (plugin Plugin) HandleTransfer(p *player.Player, ctx *event.Context, addr *net.UDPAddr) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.TransferHandler); ok {
			h.HandleTransfer(p, ctx, addr)
		}
	}
}

func (plugin Plugin) HandleCommandExecution(p *player.Player, ctx *event.Context, command cmd.Command, args []string) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.CommandExecutionHandler); ok {
			h.HandleCommandExecution(p, ctx, command, args)
		}
	}
}

func (plugin Plugin) HandleQuit(p *player.Player) {
	for _, handler := range plugin.handlers {
		if h, ok := handler.(player.QuitHandler); ok {
			h.HandleQuit(p)
		}
	}
}
