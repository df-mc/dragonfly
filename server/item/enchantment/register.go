package enchantment

import "github.com/df-mc/dragonfly/server/item"

func init() {
	item.RegisterEnchantment(0, Protection{})
	item.RegisterEnchantment(1, FireProtection{})
	item.RegisterEnchantment(3, BlastProtection{})
	item.RegisterEnchantment(4, ProjectileProtection{})
	item.RegisterEnchantment(16, SilkTouch{})
}
