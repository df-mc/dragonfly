package player

import (
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
)

// FreezingEffectStrength returns the player's freezing progress from zero to one.
func (p *Player) FreezingEffectStrength() float64 {
	return float64(p.frozenTicks) / 140
}

func (p *Player) tickFreezing(current int64) {
	immune := !p.GameMode().AllowsTakingDamage()
	for _, stack := range p.Armour().Items() {
		var tier item.ArmourTier
		switch armour := stack.Item().(type) {
		case item.Helmet:
			tier = armour.Tier
		case item.Chestplate:
			tier = armour.Tier
		case item.Leggings:
			tier = armour.Tier
		case item.Boots:
			tier = armour.Tier
		}
		if _, leather := tier.(item.ArmourTierLeather); leather {
			immune = true
			break
		}
	}
	if p.inPowderSnow && !immune {
		p.setFrozenTicks(min(140, p.frozenTicks+1))
	} else {
		p.setFrozenTicks(max(0, p.frozenTicks-2))
	}
	if !immune && p.frozenTicks == 140 && current%40 == 0 {
		p.Hurt(1, entity.FreezingDamageSource{})
	}
}

func (p *Player) setFrozenTicks(ticks int) {
	if p.frozenTicks == ticks {
		return
	}
	p.frozenTicks = ticks
	p.session().SendSpeed(max(0, p.speed-0.05*p.FreezingEffectStrength()))
	p.updateState()
}
