package chunk

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"os"
	"testing"
)

func TestA(t *testing.T) {
	data, err := os.ReadFile("biomes.out")
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(data)
	c := New(0, cube.Range{-64, 319})
	err = decodeBiomes(buf, c, NetworkEncoding)
	fmt.Println(err)
}
