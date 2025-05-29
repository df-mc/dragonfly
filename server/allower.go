package server

import (
	"net"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// Allower may be implemented to specifically allow or disallow players from
// joining a Server, by setting the specific Allower implementation through a
// call to Server.Allow.
type Allower interface {
	// Allow filters what connections are allowed to connect to the Server. The
	// address, protocol, identity data, and client data of the connection are passed. If
	// Allow returns false, the connection is closed with the string returned as
	// the disconnect message. WARNING: Use the client data at your own risk, it
	// cannot be trusted because it can be freely changed by the player
	// connecting.
	Allow(addr net.Addr, proto minecraft.Protocol, d login.IdentityData, c login.ClientData) (string, bool)
}

// allower is the standard Allower implementation. It accepts all connections.
type allower struct{}

// Allow always returns true.
func (allower) Allow(net.Addr, minecraft.Protocol, login.IdentityData, login.ClientData) (string, bool) {
	return "", true
}
