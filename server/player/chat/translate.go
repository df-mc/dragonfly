package chat

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
)

var JoinMessage = translate("%multiplayer.player.joined", 1, "%v joined the game")
var QuitMessage = translate("%multiplayer.player.left", 1, "%v left the game")

func translate(format string, params int, fallback string) Translatable {
	return Translatable{format: format, params: params, fallback: fallback}
}

type Translatable struct {
	format   string
	params   int
	fallback string
}

func (t Translatable) F(a ...any) Translation {
	if len(a) != t.params {
		panic(fmt.Sprintf("translation '%v' requires exactly %v parameters, got %v", t.format, t.params, len(a)))
	}
	params := make([]string, len(a))
	for i, arg := range a {
		params[i] = fmt.Sprint(arg)
	}
	return Translation{format: t.format, fallback: t.fallback, params: params}
}

type Translation struct {
	format   string
	fallback string
	params   []string
}

func (t Translation) String() string {
	return fmt.Sprintf(t.fallback, sliceutil.Convert[any](t.params)...)
}
