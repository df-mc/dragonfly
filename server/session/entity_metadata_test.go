package session

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type metadataBlockingEntity struct{}

func (metadataBlockingEntity) Blocking() bool { return true }

type metadataShieldBlockingEntity struct{}

func (metadataShieldBlockingEntity) ShieldBlocking() bool { return true }

func TestEntityMetadataDoesNotTreatBroadBlockingAsShieldBlocking(t *testing.T) {
	m := protocol.NewEntityMetadata()
	var s Session

	s.addSpecificMetadata(metadataBlockingEntity{}, m)

	if m.Flag(protocol.EntityDataKeyFlagsTwo, protocol.EntityDataFlagBlocking&63) {
		t.Fatal("expected broad Blocking method not to set shield blocking metadata")
	}
}

func TestEntityMetadataSetsShieldBlockingFlag(t *testing.T) {
	m := protocol.NewEntityMetadata()
	var s Session

	s.addSpecificMetadata(metadataShieldBlockingEntity{}, m)

	if !m.Flag(protocol.EntityDataKeyFlagsTwo, protocol.EntityDataFlagBlocking&63) {
		t.Fatal("expected ShieldBlocking method to set shield blocking metadata")
	}
}
