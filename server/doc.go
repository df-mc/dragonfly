// Package server implements a high-level implementation of a Minecraft server. Creating such a server may be done
// using the `server.New()` function, which accepts a `*Config` and a logger implementation. Nil may be passed as config
// to use the default config.
// After creation of the server, `Server.Start()` may be called to start and run the server. It should be followed up
// by a loop as such:
//
//   for {
//        p, err := srv.Accept()
//		  if err != nil {
//			  return
//		  }
//        // Use p
//	  }
//
// `Server.Accept()` blocks until a new player connects to the server and spawns in the default world and returns it
// once this happens. If `Server.Accept()` returns an error, this means the server was closed.
package server
