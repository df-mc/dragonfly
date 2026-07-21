package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestValueToParamTypeUsesSemanticParserSymbols(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  uint32
	}{
		{"float", float64(0), protocol.CommandArgTypeValue},
		{"generic value", struct{}{}, protocol.CommandArgTypeRValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := valueToParamType(cmd.ParamInfo{Value: tt.value}, nil)
			if got != tt.want {
				t.Fatalf("valueToParamType() = %d, want %d", got, tt.want)
			}
		})
	}
}
