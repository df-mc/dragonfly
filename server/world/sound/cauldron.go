package sound

import "image/color"

// CauldronFillWater is played when water is added to a cauldron.
type CauldronFillWater struct {
	sound
	// Colour is the colour of the water in the cauldron.
	Colour color.RGBA
}

// CauldronTakeWater is played when water is taken from a cauldron.
type CauldronTakeWater struct {
	sound
	// Colour is the colour of the water in the cauldron.
	Colour color.RGBA
}

// CauldronFillPotion is played when a potion is added to a cauldron.
type CauldronFillPotion struct {
	sound
	// Colour is the colour of the potion in the cauldron.
	Colour color.RGBA
}

// CauldronTakePotion is played when a potion is taken from a cauldron.
type CauldronTakePotion struct {
	sound
	// Colour is the colour of the potion in the cauldron.
	Colour color.RGBA
}

// CauldronFillLava is played when lava is added to a cauldron.
type CauldronFillLava struct{ sound }

// CauldronTakeLava is played when lava is taken from a cauldron.
type CauldronTakeLava struct{ sound }

// CauldronFillPowderSnow is played when powder snow is added to a cauldron.
type CauldronFillPowderSnow struct{ sound }

// CauldronTakePowderSnow is played when powder snow is taken from a cauldron.
type CauldronTakePowderSnow struct{ sound }

// CauldronAddDye is played when dye is added to the water in a cauldron.
type CauldronAddDye struct {
	sound
	// Colour is the colour of the dyed water.
	Colour color.RGBA
}

// CauldronDyeArmour is played when leather armour is dyed in a cauldron.
type CauldronDyeArmour struct {
	sound
	// Colour is the colour of the dyed water.
	Colour color.RGBA
}

// CauldronCleanArmour is played when leather armour is washed in a cauldron.
type CauldronCleanArmour struct {
	sound
	// Colour is the colour of the water in the cauldron.
	Colour color.RGBA
}

// CauldronCleanBanner is played when a banner pattern is washed in a cauldron.
type CauldronCleanBanner struct {
	sound
	// Colour is the colour of the water in the cauldron.
	Colour color.RGBA
}

// CauldronExplode is played when incompatible liquids are mixed in a cauldron.
type CauldronExplode struct {
	sound
	// Colour is the colour of the liquid being removed from the cauldron.
	Colour color.RGBA
}
