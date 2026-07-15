package server

import (
	"reflect"
	"testing"
)

func TestNetherNetConfigDoesNotOwnTLSCertificates(t *testing.T) {
	typ := reflect.TypeOf(UserConfig{}.Network.NetherNet)
	for _, field := range []string{"CertificateFile", "KeyFile"} {
		if _, ok := typ.FieldByName(field); ok {
			t.Errorf("Network.NetherNet must not expose %s", field)
		}
	}
}
