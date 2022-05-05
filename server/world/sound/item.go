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

// ArrowHit is a sound played when an arrow hits ground.
type ArrowHit struct{ sound }

// Teleport is a sound played upon teleportation of an enderman, or teleportation of a player by an ender pearl or a chorus fruit.
type Teleport struct{ sound }

// UseSpyglass is a sound played when a player uses a spyglass.
type UseSpyglass struct{ sound }

// StopUsingSpyglass is a sound played when a player stops using a spyglass.
type StopUsingSpyglass struct{ sound }
