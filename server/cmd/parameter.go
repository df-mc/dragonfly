package cmd

import (
	"github.com/go-gl/mathgl/mgl64"
	"reflect"
	"strings"
)

// Parameter is an interface for a generic parameters. Users may have types as command parameters that
// implement this parameter.
type Parameter interface {
	// Parse takes an arbitrary amount of arguments from the command Line passed and parses it, so that it can
	// store it to value v. If the arguments cannot be parsed from the Line, an error should be returned.
	Parse(line *Line, v reflect.Value) error
	// Type returns the type of the parameter. It will show up in the usage of the command, and, if one of the
	// known type names, will also show up client-side.
	Type() string
}

// Enum is an interface for enum-type parameters. Users may have types as command parameters that implement
// this parameter in order to allow a specific set of options only.
// Enum implementations must be of the type string, for example:
//
//   type GameMode string
//   func (GameMode) Type() string { return "GameMode" }
//   func (GameMode) Options(Source) []string { return []string{"survival", "creative"} }
//
// Their values will then automatically be set to whichever option returned in Enum.Options is selected by
// the user.
type Enum interface {
	// Type returns the type of the enum. This type shows up client-side in the command usage, in the spot
	// where parameter types otherwise are.
	// Type names returned are used as an identifier for this enum type. Different Enum implementations must
	// return a different string in the Type method.
	Type() string
	// Options should return a list of options that show up on the client side. The command will ensure that
	// the argument passed to the enum parameter will be equal to one of these options. The provided Source
	// can also be used to change the enums for each player.
	Options(source Source) []string
}

// SubCommand represents a subcommand that may be added as a static value that must be written. Adding
// multiple Runnable implementations to the command in New with different SubCommand fields as the
// first parameter allows for commands with subcommands.
type SubCommand struct{}

// Varargs is an argument type that may be used to capture all arguments that follow. This is useful for,
// for example, messages and names.
type Varargs string

// Optional is an argument type that may be used to make any of the available parameter types optional. Optional command
// parameters may only occur at the end of the Runnable struct. No non-optional parameter is allowed after an optional
// parameter.
type Optional[T any] struct {
	val T
	set bool
}

// Load returns the value specified upon executing the command and a bool that is true if the parameter was filled out
// by the Source.
func (o Optional[T]) Load() (T, bool) {
	return o.val, o.set
}

// LoadOr returns the value specified upon executing the command, or a value 'or' if the parameter was not filled out
// by the Source.
func (o Optional[T]) LoadOr(or T) T {
	if o.set {
		return o.val
	}
	return or
}

// with returns an Optional[T] with the value passed. It also sets the 'set' field to true.
func (o Optional[T]) with(val any) any {
	return Optional[T]{val: val.(T), set: true}
}

// optionalT is used to identify a parameter of the Optional type.
type optionalT interface {
	with(val any) any
}

// typeNameOf returns a readable type name for the interface value passed. If none could be found, 'value'
// is returned.
func typeNameOf(i any, name string) string {
	switch i.(type) {
	case int, int8, int16, int32, int64:
		return "int"
	case uint, uint8, uint16, uint32, uint64:
		return "uint"
	case float32, float64:
		return "float"
	case string:
		return "string"
	case Varargs:
		return "text"
	case bool:
		return "bool"
	case mgl64.Vec3:
		return "x y z"
	case []Target:
		return "target"
	case SubCommand:
		return name
	}
	if param, ok := i.(Parameter); ok {
		return param.Type()
	}
	if enum, ok := i.(Enum); ok {
		return enum.Type()
	}
	return "value"
}

// unwrap returns the underlying reflect.Value of a reflect.Value, assuming it is of the Optional[T] type.
func unwrap(v reflect.Value) reflect.Value {
	if _, ok := v.Interface().(optionalT); ok {
		return reflect.New(v.Field(0).Type()).Elem()
	}
	return v
}

// optional checks if the reflect.Value passed implements the optionalT interface.
func optional(v reflect.Value) bool {
	_, ok := v.Interface().(optionalT)
	return ok
}

// suffix returns the suffix of the parameter as set in the struct field, if any.
func suffix(v reflect.StructField) string {
	_, str := tag(v)
	return str
}

// name returns the name of the parameter as set in the struct tag if it exists, or the field's name if not.
func name(v reflect.StructField) string {
	str, _ := tag(v)
	if str == "" {
		return v.Name
	}
	return str
}

// tag returns the name and suffix as specified in the 'cmd' tag, or empty strings if not present.
func tag(v reflect.StructField) (name string, suffix string) {
	t, _ := v.Tag.Lookup("cmd")
	a, b, _ := strings.Cut(t, ",")
	return a, b
}
