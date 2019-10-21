package block

import (
	r "github.com/dragonfly-tech/dragonfly/dragonfly/block/encoder"
)

// init registers all blocks implemented by Dragonfly.
func init() {
	r.Register(r.BasicEncoder{ID: "minecraft:air", Block: Air{}}, Air{})
	r.Register(stoneEncoder{}, Stone{}, Granite{}, Diorite{}, Andesite{})
	r.Register(r.BasicEncoder{ID: "minecraft:grass", Block: Grass{}}, Grass{})
	r.Register(dirtEncoder{}, Dirt{}, CoarseDirt{})
	r.Register(logEncoder{}, OakLog{}, SpruceLog{}, BirchLog{}, JungleLog{}, AcaciaLog{}, DarkOakLog{})
	r.Register(r.BasicEncoder{ID: "minecraft:bedrock", Block: Bedrock{}}, Bedrock{})
}
