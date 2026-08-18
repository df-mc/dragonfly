package session

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestSetShieldBlockState(t *testing.T) {
	tests := []struct {
		name             string
		blocked, damaged bool
	}{
		{name: "blocked", blocked: true},
		{name: "damaged", damaged: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := protocol.NewEntityMetadata()
			setShieldBlockState(metadata, tt.blocked, tt.damaged)
			if got := metadata.Flag(protocol.EntityDataKeyFlagsTwo, protocol.EntityDataFlagBlockedUsingShield&63); got != tt.blocked {
				t.Fatalf("blocked shield flag = %v, want %v", got, tt.blocked)
			}
			if got := metadata.Flag(protocol.EntityDataKeyFlagsTwo, protocol.EntityDataFlagBlockedUsingDamagedShield&63); got != tt.damaged {
				t.Fatalf("damaged shield flag = %v, want %v", got, tt.damaged)
			}
		})
	}
}
