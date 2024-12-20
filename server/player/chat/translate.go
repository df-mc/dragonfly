package chat

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"golang.org/x/text/language"
)

var MessageJoin = Translate(str("%multiplayer.player.joined"), 1, "%v joined the game").Enc("<yellow>%v</yellow>")
var MessageQuit = Translate(str("%multiplayer.player.left"), 1, "%v left the game").Enc("<yellow>%v</yellow>")

type str string

func (s str) Resolve(language.Tag) string { return string(s) }

type TranslationString interface {
	Resolve(l language.Tag) string
}

func Translate(str TranslationString, params int, fallback string) Translatable {
	return Translatable{str: str, params: params, fallback: fallback}
}

type Translatable struct {
	str      TranslationString
	format   string
	params   int
	fallback string
}

func (t Translatable) Enc(format string) Translatable {
	t.format = format
	t.fallback = text.Colourf(format, t.fallback)
	return t
}

func (t Translatable) F(a ...any) Translation {
	if len(a) != t.params {
		panic(fmt.Sprintf("translation '%v' requires exactly %v parameters, got %v", t.format, t.params, len(a)))
	}
	params := make([]string, len(a))
	for i, arg := range a {
		params[i] = fmt.Sprint(arg)
	}
	return Translation{str: t.str, format: t.format, fallback: t.fallback, params: params, fallbackParams: a}
}

type Translation struct {
	str    TranslationString
	format string
	params []string

	fallback       string
	fallbackParams []any
}

func (t Translation) Format(l language.Tag) string {
	return text.Colourf(t.format, t.str.Resolve(l))
}

func (t Translation) Params() []string {
	return t.params
}

func (t Translation) String() string {
	return fmt.Sprintf(text.Colourf(t.format, t.fallback), t.fallbackParams...)
}
