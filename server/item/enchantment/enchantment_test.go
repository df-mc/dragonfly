package enchantment

import "testing"

func TestPiercingAndMultishotAreMutuallyExclusive(t *testing.T) {
	if Piercing.CompatibleWithEnchantment(Multishot) {
		t.Fatal("expected Piercing to reject Multishot")
	}
	if Multishot.CompatibleWithEnchantment(Piercing) {
		t.Fatal("expected Multishot to reject Piercing")
	}
}
