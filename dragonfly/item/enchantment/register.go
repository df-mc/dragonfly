package enchantment

import "git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"

func init() {
	item.RegisterEnchantment(0, Protection{})
	item.RegisterEnchantment(1, FireProtection{})
	item.RegisterEnchantment(3, BlastProtection{})
	item.RegisterEnchantment(4, ProjectileProtection{})
}
