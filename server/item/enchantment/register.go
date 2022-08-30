package enchantment

import "github.com/df-mc/dragonfly/server/item"

func init() {
	item.RegisterEnchantment(0, Protection{})
	item.RegisterEnchantment(1, FireProtection{})
	item.RegisterEnchantment(2, FeatherFalling{})
	item.RegisterEnchantment(3, BlastProtection{})
	item.RegisterEnchantment(4, ProjectileProtection{})
	item.RegisterEnchantment(5, Thorns{})
	item.RegisterEnchantment(6, Respiration{})
	item.RegisterEnchantment(7, DepthStrider{})
	item.RegisterEnchantment(8, AquaAffinity{})
	item.RegisterEnchantment(9, Sharpness{})
	// TODO: (10) Smite. (Requires undead mobs)
	// TODO: (11) Bane of Arthropods. (Requires arthropod mobs)
	item.RegisterEnchantment(12, KnockBack{})
	item.RegisterEnchantment(13, FireAspect{})
	// TODO: (14) Looting.
	item.RegisterEnchantment(15, Efficiency{})
	item.RegisterEnchantment(16, SilkTouch{})
	item.RegisterEnchantment(17, Unbreaking{})
	// TODO: (18) Fortune.
	item.RegisterEnchantment(19, Power{})
	item.RegisterEnchantment(20, Punch{})
	item.RegisterEnchantment(21, Flame{})
	item.RegisterEnchantment(22, Infinity{})
	// TODO: (23) Luck of the Sea.
	// TODO: (24) Lure.
	// TODO: (25) Frost Walker.
	item.RegisterEnchantment(26, Mending{})
	// TODO: (27) Curse of Binding.
	item.RegisterEnchantment(28, CurseOfVanishing{})
	// TODO: (29) Impaling.
	// TODO: (30) Riptide.
	// TODO: (31) Loyalty.
	// TODO: (32) Channeling.
	// TODO: (33) Multishot.
	// TODO: (34) Piercing.
	// TODO: (35) Quick Charge.
	item.RegisterEnchantment(36, SoulSpeed{})
	item.RegisterEnchantment(37, SwiftSneak{})
}
