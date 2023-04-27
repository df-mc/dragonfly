package mcdb_test

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"testing"
)

func TestOpen(t *testing.T) {
	db, err := mcdb.Open("Test World")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", db.Settings())
	iter := db.NewColumnIterator(nil)
	for iter.Next() {
		fmt.Println(iter.Position(), iter.Dimension())
		fmt.Printf("%#v\n", iter.Column().BlockEntities)
	}

	pos := world.ChunkPos{-2, -4}
	col, err := db.LoadColumn(pos, world.Overworld)
	fmt.Printf("%#v\n", col.Entities)
	fmt.Println(col, err)
}
