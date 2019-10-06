package block

import (
	// Usually imports like this are not a good idea, but in this particular case it makes the following code
	// look a little more readable.
	. "github.com/dragonfly-tech/dragonfly/dragonfly/block/encoder"
)

// init registers all blocks implemented by Dragonfly.
func init() {
	Register(BasicEncoder{ID: "minecraft:air", Block: Air{}}, Air{})
	Register(stoneEncoder{}, Stone{}, Granite{}, Diorite{}, Andesite{})
	Register(BasicEncoder{ID: "minecraft:grass", Block: Grass{}}, Grass{})
	Register(dirtEncoder{}, Dirt{}, CoarseDirt{})
	Register(logEncoder{}, OakLog{}, SpruceLog{}, BirchLog{}, JungleLog{}, AcaciaLog{}, DarkOakLog{})
	Register(BasicEncoder{ID: "minecraft:bedrock", Block: Bedrock{}}, Bedrock{})
}
