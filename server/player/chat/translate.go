package chat

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"golang.org/x/text/language"
)

var MessageJoin = Translate(str("%multiplayer.player.joined"), 1, "%v joined the game").Enc("<yellow>%v</yellow>")
var MessageQuit = Translate(str("%multiplayer.player.left"), 1, "%v left the game").Enc("<yellow>%v</yellow>")

type str string

// Resolve returns the translation identifier as a string.
func (s str) Resolve(language.Tag) string { return string(s) }

// TranslationString is a value that can resolve a translated version of itself
// for a language.Tag passed.
type TranslationString interface {
	// Resolve finds a suitable translated version for a translation string for
	// a specific language.Tag.
	Resolve(l language.Tag) string
}

// Translate returns a Translatable for a TranslationString. The required number
// of parameters specifies how many arguments may be passed to Translatable.F.
// The fallback string should be a 'standard' translation of the string, which
// is used when Translation.String is called on the Translation that results
// from a call to Translatable.F. This fallback string should have as many
// formatting identifiers (like in fmt.Sprintf) as the number of params.
func Translate(str TranslationString, params int, fallback string) Translatable {
	return Translatable{str: str, params: params, fallback: fallback, format: "%v"}
}

// Translatable represents a TranslationString with additional formatting, that
// may be filled out by calling F on it with a list of arguments for the
// translation.
type Translatable struct {
	str      TranslationString
	format   string
	params   int
	fallback string
}

// Enc encapsulates the translation string into the format passed. This format
// should have exactly one formatting identifier, %v, to specify where the
// translation string should go, such as 'Translation: %v'.
// Enc accepts colouring formats parsed by text.Colourf.
func (t Translatable) Enc(format string) Translatable {
	t.format = format
	return t
}

// F takes arguments for a translation string passed and returns a filled out
// Translation that may be sent to players. The number of arguments passed must
// be exactly equal to the number specified in Translate. If not, F will panic.
func (t Translatable) F(a ...any) Translation {
	if len(a) != t.params {
		panic(fmt.Sprintf("translation '%v' requires exactly %v parameters, got %v", t.format, t.params, len(a)))
	}
	params := make([]string, len(a))
	for i, arg := range a {
		params[i] = fmt.Sprint(arg)
	}
	return Translation{t: t, params: params, fallbackParams: a}
}

// Translation is a translation string with its arguments filled out. Format may
// be called to obtain the translated version of the translation string and
// Params may be called to obtain the parameters passed in Translatable.F.
type Translation struct {
	t              Translatable
	params         []string
	fallbackParams []any
}

// Format translates the TranslationString of the Translation to the language
// passed and returns it.
func (t Translation) Format(l language.Tag) string {
	return text.Colourf(t.t.format, t.t.str.Resolve(l))
}

// Params returns a slice of values that are used to parameterise the
// translation returned by Format.
func (t Translation) Params() []string {
	return t.params
}

// String formats and returns the fallback value of the Translation.
func (t Translation) String() string {
	return fmt.Sprintf(text.Colourf(t.t.format, t.t.fallback), t.fallbackParams...)
}
