package cmd

import "reflect"

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
// multiple Runnable implementations to the command in New with different SubCommand implementations as the
// first parameter allows for commands with subcommands.
type SubCommand interface {
	// SubName returns the value that must be entered by the user when executing the subcommand, such as
	// 'kill' for a command such as /entity kill <target>.
	SubName() string
}

// optional checks if a struct field is considered optional.
func optional(v reflect.StructField) bool {
	if _, ok := v.Tag.Lookup("optional"); ok {
		return true
	}
	return false
}

// suffix returns the suffix of the parameter as set in the struct field, if any.
func suffix(v reflect.StructField) string {
	return v.Tag.Get("suffix")
}

// name returns the name of the parameter as set in the struct tag if it exists, or the field's name if not.
func name(v reflect.StructField) string {
	if name, ok := v.Tag.Lookup("name"); ok {
		return name
	}
	return v.Name
}
