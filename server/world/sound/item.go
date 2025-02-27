package sound

import "github.com/df-mc/dragonfly/server/world"

// ItemBreak is a sound played when an item in the inventory is broken, such as when a tool reaches 0
// durability and breaks.
type ItemBreak struct{ sound }

// ItemThrow is a sound played when a player throws an item, such as a snowball.
type ItemThrow struct{ sound }

// ItemUseOn is a sound played when a player uses its item on a block. An example of this is when a player
// uses a shovel to turn grass into dirt path. Note that in these cases, the Block is actually the new block,
// not the old one.
type ItemUseOn struct {
	// Block is generally the block that was created by using the item on a block. The sound played differs
	// depending on this field.
	Block world.Block

	sound
}

// EquipItem is a sound played when the player fast equips an item by using it.
type EquipItem struct {
	// Item is the item that was equipped. The sound played differs depending on this field.
	Item world.Item

	sound
}

// BucketFill is a sound played when a bucket is filled using a liquid source block from the world.
type BucketFill struct {
	// Liquid is the liquid that the bucket is filled up with.
	Liquid world.Liquid

	sound
}

// BucketEmpty is a sound played when a bucket with a liquid in it is placed into the world.
type BucketEmpty struct {
	// Liquid is the liquid that the bucket places into the world.
	Liquid world.Liquid

	sound
}

// BowShoot is a sound played when a bow is shot.
type BowShoot struct{ sound }

// CrossbowLoad is a sound when a crossbow is being loaded.
type CrossbowLoad struct {
	// Stage is the stage of the crossbow.
	Stage int
	// QuickCharge returns if the item being used has quick charge enchantment.
	QuickCharge bool

	sound
}

// CrossbowShoot is a sound played when a crossbow is shot.
type CrossbowShoot struct{ sound }

const (
	// CrossbowLoadingStart is the stage where a crossbows start to load.
	CrossbowLoadingStart = iota
	// CrossbowLoadingMiddle is the stage where a crossbow loading but not yet
	// finished loading.
	CrossbowLoadingMiddle
	// CrossbowLoadingEnd is the stage where a crossbow is finished loading.
	CrossbowLoadingEnd
)

// ArrowHit is a sound played when an arrow hits ground.
type ArrowHit struct{ sound }

// Teleport is a sound played upon teleportation of an enderman, or teleportation of a player by an ender pearl or a chorus fruit.
type Teleport struct{ sound }

// UseSpyglass is a sound played when a player uses a spyglass.
type UseSpyglass struct{ sound }

// StopUsingSpyglass is a sound played when a player stops using a spyglass.
type StopUsingSpyglass struct{ sound }

// GoatHorn is a sound played when a player uses a goat horn.
type GoatHorn struct {
	// Horn is the type of the goat horn.
	Horn Horn

	sound
}

// FireCharge is a sound played when a player lights a block on fire with a fire charge, or when a dispenser or a
// blaze shoots a fireball.
type FireCharge struct{ sound }

// Totem is a sound played when a player uses a totem.
type Totem struct{ sound }
