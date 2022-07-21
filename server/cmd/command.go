package cmd

import (
	"encoding/csv"
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

// Runnable represents a Command that may be run by a Command source. The Command must be a struct type and
// its fields represent the parameters of the Command. When the Run method is called, these fields are set
// and may be used for behaviour in the Command. Fields unexported or ignored using the `cmd:"-"` struct tag (see
// below) have their values copied but retained.
// A Runnable may have exported fields only of the following types:
// int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint,
// float32, float64, string, bool, mgl64.Vec3, Varargs, []Target, cmd.SubCommand, Optional[T] (to make a parameter
// optional), or a type that implements the cmd.Parameter or cmd.Enum interface. cmd.Enum implementations must be of the
// type string.
// Fields in the Runnable struct may have `cmd:` struct tag to specify the name and suffix of a parameter as such:
//   type T struct {
//       Param int `cmd:"name,suffix"`
//   }
// If no name is set, the field name is used. Additionally, the name as specified in the struct tag may be '-' to make
// the parser ignore the field. In this case, the field does not have to be of one of the types above.
type Runnable interface {
	// Run runs the Command, using the arguments passed to the Command. The source is passed to the method,
	// which is the source of the execution of the Command, and the output is passed, to which messages may be
	// added which get sent to the source.
	Run(src Source, o *Output)
}

// Allower may be implemented by a type also implementing Runnable to limit the sources that may run the
// command.
type Allower interface {
	// Allow checks if the Source passed is allowed to execute the command. True is returned if the Source is
	// allowed to execute the command.
	Allow(src Source) bool
}

// Command is a wrapper around a Runnable. It provides additional identity and utility methods for the actual
// runnable command so that it may be identified more easily.
type Command struct {
	v           []reflect.Value
	name        string
	description string
	usage       string
	aliases     []string
}

// New returns a new Command using the name and description passed. The Runnable passed must be a
// (pointer to a) struct, with its fields representing the parameters of the command.
// When the command is run, the Run method of the Runnable will be called, after all fields have their values
// from the parsed command set.
// If r is not a struct or a pointer to a struct, New panics.
func New(name, description string, aliases []string, r ...Runnable) Command {
	usages := make([]string, len(r))
	runnableValues := make([]reflect.Value, len(r))

	if len(aliases) > 0 {
		namePresent := false
		for _, alias := range aliases {
			if alias == name {
				namePresent = true
			}
		}
		if !namePresent {
			aliases = append(aliases, name)
		}
	}

	for i, runnable := range r {
		t := reflect.TypeOf(runnable)
		if t.Kind() != reflect.Struct && (t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct) {
			panic(fmt.Sprintf("Runnable r must be struct or pointer to struct, but got %v", t.Kind()))
		}
		original := reflect.ValueOf(runnable)
		if t.Kind() == reflect.Ptr {
			original = original.Elem()
		}

		cp := reflect.New(original.Type()).Elem()
		if err := verifySignature(cp); err != nil {
			panic(err.Error())
		}
		runnableValues[i], usages[i] = original, parseUsage(name, cp)
	}

	return Command{name: name, description: description, aliases: aliases, v: runnableValues, usage: strings.Join(usages, "\n")}
}

// Name returns the name of the command. The name is guaranteed to be lowercase and will never have spaces in
// it. This name is used to call the command, and is shown in the /help list.
func (cmd Command) Name() string {
	return cmd.name
}

// Description returns the description of the command. The description is shown in the /help list, and
// provides information on the functionality of a command.
func (cmd Command) Description() string {
	return cmd.description
}

// Usage returns the usage of the command. The usage will be roughly equal to the one showed by the client
// in-game.
func (cmd Command) Usage() string {
	return cmd.usage
}

// Aliases returns a list of aliases for the command. In addition to the name of the command, the command may
// be called using one of these aliases.
func (cmd Command) Aliases() []string {
	return cmd.aliases
}

// Execute executes the Command as a source with the args passed. The args are parsed assuming they do not
// start with the command name. Execute will attempt to parse and execute one Runnable at a time. If one of
// the Runnable was able to parse args correctly, it will be executed and no more Runnables will be attempted
// to be run.
// If parsing of all Runnables was unsuccessful, a command output with an error message is sent to the Source
// passed, and the Run method of the Runnables are not called.
// The Source passed must not be nil. The method will panic if a nil Source is passed.
func (cmd Command) Execute(args string, source Source) {
	if source == nil {
		panic("execute: invalid command source: source must not be nil")
	}
	output := &Output{}
	defer source.SendCommandOutput(output)

	var leastErroneous error
	leastArgsLeft := len(strings.Split(args, " "))

	for _, v := range cmd.v {
		cp := reflect.New(v.Type())
		cp.Elem().Set(v)
		line, err := cmd.executeRunnable(cp, args, source, output)
		if err == nil {
			// Command was executed successfully: We won't execute any of the other Runnable values passed, as
			// we've already found an overload that works.
			return
		}
		if line == nil {
			// This Runnable was not runnable by the source passed. Only if no error was yet set, we set an
			// error for the wrong source.
			if leastErroneous == nil {
				leastErroneous = err
			}
			continue
		}
		if line.Len() <= leastArgsLeft {
			// If the line had less (or equal) arguments left than the previous lowest, we update the error,
			// so that we can return an error that applies for the most successful Runnable.
			leastErroneous = err
			leastArgsLeft = line.Len()
		}
	}
	// No working Runnable found for the arguments passed. We add the most applicable error to the output and
	// stop there.
	output.Error(leastErroneous)
}

// ParamInfo holds the information of a parameter in a Runnable. Information of a parameter may be obtained
// by calling Command.Params().
type ParamInfo struct {
	Name     string
	Value    any
	Optional bool
	Suffix   string
}

// Params returns a list of all parameters of the runnables. No assumptions should be done on the values that
// they hold: Only the types are guaranteed to be consistent.
func (cmd Command) Params(src Source) [][]ParamInfo {
	params := make([][]ParamInfo, 0, len(cmd.v))
	for _, runnable := range cmd.v {
		elem := reflect.New(runnable.Type()).Elem()
		elem.Set(runnable)

		if allower, ok := runnable.Interface().(Allower); ok && !allower.Allow(src) {
			// This source cannot execute this runnable.
			continue
		}

		var fields []ParamInfo
		for _, t := range exportedFields(elem) {
			field := elem.FieldByName(t.Name)
			fields = append(fields, ParamInfo{
				Name:     name(t),
				Value:    unwrap(field).Interface(),
				Optional: optional(field),
				Suffix:   suffix(t),
			})
		}
		params = append(params, fields)
	}
	return params
}

// Runnables returns a map of all Runnable implementations of the Command that a Source can execute.
func (cmd Command) Runnables(src Source) map[int]Runnable {
	m := make(map[int]Runnable, len(cmd.v))
	for i, runnable := range cmd.v {
		v := runnable.Interface().(Runnable)
		if allower, ok := v.(Allower); !ok || allower.Allow(src) {
			m[i] = v
		}
	}
	return m
}

// String returns the usage of the command. The usage will be roughly equal to the one showed by the client
// in-game.
func (cmd Command) String() string {
	return cmd.usage
}

// executeRunnable executes a Runnable v, by parsing the args passed using the source and output obtained. If
// parsing was not successful or the Runnable could not be run by this source, an error is returned, and the
// leftover command line.
func (cmd Command) executeRunnable(v reflect.Value, args string, source Source, output *Output) (*Line, error) {
	if a, ok := v.Interface().(Allower); ok && !a.Allow(source) {
		//lint:ignore ST1005 Error string is capitalised because it is shown to the player.
		//goland:noinspection GoErrorStringFormat
		return nil, fmt.Errorf("You cannot execute this command.")
	}

	var argFrags []string
	if args != "" {
		r := csv.NewReader(strings.NewReader(args))
		r.Comma = ' '
		r.LazyQuotes = true
		record, err := r.Read()
		if err != nil {
			return nil, fmt.Errorf("error parsing command string: %w", err)
		}
		argFrags = record
	}
	parser := parser{}
	arguments := &Line{args: argFrags, src: source}

	// We iterate over all the fields of the struct: Each of the fields will have an argument parsed to
	// produce its value.
	signature := v.Elem()
	for _, t := range exportedFields(signature) {
		field := signature.FieldByName(t.Name)
		parser.currentField = t.Name
		opt := optional(field)

		val := field
		if opt {
			val = reflect.New(field.Field(0).Type()).Elem()
		}

		err, success := parser.parseArgument(arguments, val, opt, name(t), source)
		if err != nil {
			// Parsing was not successful, we return immediately as we don't need to call the Runnable.
			return arguments, err
		}
		if success && opt {
			field.Set(reflect.ValueOf(field.Interface().(optionalT).with(val.Interface())))
		}
	}
	if arguments.Len() != 0 {
		return arguments, fmt.Errorf("unexpected '%v'", strings.Join(arguments.args, " "))
	}

	v.Interface().(Runnable).Run(source, output)
	return arguments, nil
}

// parseUsage parses the usage of a command found in value v using the name passed. It accounts for optional
// parameters and converts types to a more friendly representation.
func parseUsage(commandName string, command reflect.Value) string {
	parts := make([]string, 0, command.NumField()+1)
	parts = append(parts, "/"+commandName)

	for _, t := range exportedFields(command) {
		field := command.FieldByName(t.Name)

		typeName := typeNameOf(field.Interface(), name(t))
		if _, ok := field.Interface().(optionalT); ok {
			typeName = typeNameOf(reflect.New(field.Field(0).Type()).Elem().Interface(), name(t))
		}
		if _, ok := field.Interface().(SubCommand); ok {
			parts = append(parts, typeName)
			continue
		}
		if optional(field) {
			parts = append(parts, "["+name(t)+": "+typeName+"]"+suffix(t))
			continue
		}
		parts = append(parts, "<"+name(t)+": "+typeName+">"+suffix(t))
	}
	return strings.Join(parts, " ")
}

// verifySignature verifies the passed struct pointer value signature to ensure it is a valid command,
// checking things such as the validity of the optional struct tags.
// If not valid, an error is returned.
func verifySignature(command reflect.Value) error {
	optionalField := false
	for _, t := range exportedFields(command) {
		field := command.FieldByName(t.Name)
		if _, ok := field.Interface().(Enum); ok && field.Kind() != reflect.String {
			return fmt.Errorf("parameters implementing Enum must be of the type string")
		}
		// If the field is not optional, while the last field WAS optional, we return an error, as this is
		// not parsable in an expected way.
		opt := optional(field)
		if !opt && optionalField {
			return fmt.Errorf("command must only have optional parameters at the end")
		}
		optionalField = opt
	}
	return nil
}

// exportedFields returns all exported struct fields of the reflect.Value passed. It returns the fields as returned by
// reflect.VisibleFields, but filters out unexported fields, anonymous fields and fields that have a name value in the
// 'cmd' tag of '-'.
func exportedFields(command reflect.Value) []reflect.StructField {
	visible := reflect.VisibleFields(command.Type())
	fields := make([]reflect.StructField, 0, len(visible))

	for _, t := range visible {
		if !ast.IsExported(t.Name) || name(t) == "-" || t.Anonymous {
			continue
		}
		field := command.FieldByName(t.Name)
		if !field.CanSet() {
			continue
		}
		fields = append(fields, t)
	}
	return fields
}
