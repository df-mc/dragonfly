package entity

import "github.com/df-mc/dragonfly/server/world"

// init registers all entities that can be saved in a world.World, so that they can be loaded when found in the world
// save.
func init() {
	world.RegisterEntity(&AreaEffectCloud{})
	world.RegisterEntity(&Arrow{})
	world.RegisterEntity(&BottleOfEnchanting{})
	world.RegisterEntity(&Egg{})
	world.RegisterEntity(&EnderPearl{})
	world.RegisterEntity(&ExperienceOrb{})
	world.RegisterEntity(&FallingBlock{})
	world.RegisterEntity(&Firework{})
	world.RegisterEntity(&Item{})
	world.RegisterEntity(&Lightning{})
	world.RegisterEntity(&LingeringPotion{})
	world.RegisterEntity(&Snowball{})
	world.RegisterEntity(&SplashPotion{})
	world.RegisterEntity(&TNT{})
	world.RegisterEntity(&Text{})
}
