// Package cmd implements a Minecraft specific command system, which may be used simply by 'plugging' it in
// and sending commands registered in an AvailableCommandsPacket.
//
// The cmd package handles commands in a specific way: It requires a struct to be passed to the cmd.New()
// function, which implements the Runnable interface. For every exported field in the struct, executing the
// command will result in the parsing of the arguments using the types of the fields of the struct, in the
// order that they appear in.
//
// A Runnable may have exported fields only of the following types:
// int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint,
// float32, float64, string, bool, mgl64.Vec3, Varargs, []Target
// or a type that implements the cmd.Parameter, cmd.Enum or cmd.SubCommand interface. cmd.Enum implementations
// must be of the type string.
// Fields in the Runnable struct may have the `optional:""` struct tag to mark them as an optional parameter,
// the `suffix:"$suffix"` struct tag to add a suffix to the parameter in the usage, and the `name:"name"` tag
// to specify a name different from the field name for the parameter.
//
// Commands may be registered using the cmd.Register() method. By itself, this method will not ensure that the
// client will be able to use the command: The user of the cmd package must handle commands itself and run the
// appropriate one using the cmd.ByAlias function.
package cmd
