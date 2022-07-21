// Package cmd implements a Minecraft specific command system, which may be used simply by 'plugging' it in
// and sending commands registered in an AvailableCommandsPacket.
//
// The cmd package handles commands in a specific way: It requires a struct to be passed to the cmd.New()
// function, which implements the Runnable interface. For every exported field in the struct, executing the
// command will result in the parsing of the arguments using the types of the fields of the struct, in the
// order that they appear in. Fields unexported or ignored using the `cmd:"-"` struct tag (see below) have
// their values copied but retained.
//
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
//
// Commands may be registered using the cmd.Register() method. By itself, this method will not ensure that the
// client will be able to use the command: The user of the cmd package must handle commands itself and run the
// appropriate one using the cmd.ByAlias function.
package cmd
