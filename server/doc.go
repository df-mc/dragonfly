// Package server implements a high-level implementation of a Minecraft server.
// Creating such a server may be done using the `server.New()` function.
// Additional configuration may be passed by using server.Config{...}.New().
// `Server.Listen()` may be called to start and run the server. It should be
// followed up by a loop as such:
//
//   for srv.Accept(nil) {
//	 }
//
// `Server.Accept()` blocks until a new player connects to the server and
// spawns in the default world, and calls the function passed to it once this
// happens. If `Server.Accept()` returns false, this means the server was
// closed.
package server
