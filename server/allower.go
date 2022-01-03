package server

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"net"
)

// Allower may be implemented to specifically allow or disallow players from joining a Server, by setting the specific
// Allower implementation through a call to Server.Allow.
type Allower interface {
	// Allow filters what connections are allowed to connect to the Server. The address and identity data of the
	// connection are passed. If Admit returns false, the connection is closed with the string returned as disconnect
	// message.
	Allow(addr net.Addr, d login.IdentityData) (string, bool)
}

// allower is the standard Allower implementation. It accepts all connections.
type allower struct{}

// Allow always returns true.
func (allower) Allow(net.Addr, login.IdentityData) (string, bool) { return "", true }
