package enchantment

import "github.com/df-mc/dragonfly/server/item"

func init() {
	item.RegisterEnchantment(0, Protection{})
	item.RegisterEnchantment(1, FireProtection{})
	item.RegisterEnchantment(2, FeatherFalling{})
	// TODO: (3) Blast Protection.
	item.RegisterEnchantment(4, ProjectileProtection{})
	// TODO: (5) Thorns.
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
	// TODO: (19) Power.
	// TODO: (20) Punch.
	// TODO: (21) Flame.
	// TODO: (22) Infinity.
	// TODO: (23) Luck of the Sea.
	// TODO: (24) Lure.
	// TODO: (25) Frost Walker.
	// TODO: (26) Mending.
	// TODO: (27) Curse of Binding.
	// TODO: (28) Curse of Vanishing.
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
