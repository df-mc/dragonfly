package intintmap

import "testing"

func TestMap(t *testing.T) {
	m := New(1, 0.999)
	m.Put(5, 10)
	if v, ok := m.Get(5); !ok || v != 10 {
		t.Fatalf("expected map value 10, got %v (%v)", v, ok)
	}
	m.Del(5)
	if _, ok := m.Get(5); ok {
		t.Fatal("expected deleted key to be absent")
	}
}
