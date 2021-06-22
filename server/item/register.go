package item

import (
	"github.com/df-mc/dragonfly/server/item/armour"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/item/tool"
	"github.com/df-mc/dragonfly/server/world"
)

//noinspection SpellCheckingInspection
func init() {
	for _, t := range tool.Tiers() {
		world.RegisterItem(Pickaxe{Tier: t})
		world.RegisterItem(Axe{Tier: t})
		world.RegisterItem(Shovel{Tier: t})
		world.RegisterItem(Sword{Tier: t})
		world.RegisterItem(Hoe{Tier: t})
	}
	for _, t := range armour.Tiers() {
		world.RegisterItem(Helmet{Tier: t})
		world.RegisterItem(Chestplate{Tier: t})
		world.RegisterItem(Leggings{Tier: t})
		world.RegisterItem(Boots{Tier: t})
	}

	world.RegisterItem(Bucket{})

	world.RegisterItem(Shears{})

	world.RegisterItem(Diamond{})
	world.RegisterItem(GlowstoneDust{})
	world.RegisterItem(LapisLazuli{})
	world.RegisterItem(Emerald{})
	world.RegisterItem(GoldIngot{})
	world.RegisterItem(GoldNugget{})
	world.RegisterItem(IronIngot{})
	world.RegisterItem(Coal{})
	world.RegisterItem(NetheriteIngot{})
	world.RegisterItem(ClayBall{})
	world.RegisterItem(NetherQuartz{})
	world.RegisterItem(Flint{})

	world.RegisterItem(Stick{})
	world.RegisterItem(MagmaCream{})

	world.RegisterItem(BoneMeal{})
	world.RegisterItem(Wheat{})
	world.RegisterItem(Beetroot{})
	world.RegisterItem(MelonSlice{})

	world.RegisterItem(Apple{})

	world.RegisterItem(Brick{})

	world.RegisterItem(Leather{})

	world.RegisterItem(GlassBottle{})
	for _, p := range potion.All() {
		world.RegisterItem(Potion{Type: p})
	}

	world.RegisterItem(FlintAndSteel{})

	world.RegisterItem(PrismarineCrystals{})

	world.RegisterItem(PoisonousPotato{})
	world.RegisterItem(GoldenApple{})
	world.RegisterItem(EnchantedApple{})
	world.RegisterItem(Pufferfish{})
	world.RegisterItem(Clock{})
	world.RegisterItem(Compass{})

	world.RegisterItem(CopperIngot{})
	world.RegisterItem(RawCopper{})
	world.RegisterItem(RawIron{})
	world.RegisterItem(RawGold{})
	world.RegisterItem(BlazePowder{})
	world.RegisterItem(BlazeRod{})
	world.RegisterItem(Bone{})
	world.RegisterItem(Book{})
	world.RegisterItem(Bowl{})
	world.RegisterItem(Charcoal{})
	world.RegisterItem(DragonBreath{})
	world.RegisterItem(DriedKelp{})
	world.RegisterItem(Feather{})
	world.RegisterItem(FermentedSpiderEye{})
	world.RegisterItem(GhastTear{})
	world.RegisterItem(GlowInkSac{})
	world.RegisterItem(Gunpowder{})
	world.RegisterItem(HeartOfTheSea{})
	world.RegisterItem(Honeycomb{})
	world.RegisterItem(InkSac{})
	world.RegisterItem(IronNugget{})
	world.RegisterItem(NautilusShell{})
	world.RegisterItem(NetherBrick{})
	world.RegisterItem(NetherStar{})
	world.RegisterItem(NetheriteScrap{})
	world.RegisterItem(Paper{})
	world.RegisterItem(PhantomMembrane{})
	world.RegisterItem(PrismarineShard{})
	world.RegisterItem(RabbitFoot{})
	world.RegisterItem(RabbitHide{})
	world.RegisterItem(Scute{})
	world.RegisterItem(ShulkerShell{})
	world.RegisterItem(Slimeball{})
	world.RegisterItem(SpiderEye{})
	world.RegisterItem(Sugar{})
	world.RegisterItem(BakedPotato{})
	world.RegisterItem(Bread{})
	world.RegisterItem(CookedChicken{})
	world.RegisterItem(CookedCod{})
	world.RegisterItem(CookedMutton{})
	world.RegisterItem(CookedPorkchop{})
	world.RegisterItem(CookedRabbit{})
	world.RegisterItem(CookedSalmon{})
	world.RegisterItem(Cookie{})
	world.RegisterItem(GoldenCarrot{})
	world.RegisterItem(PumpkinPie{})
	world.RegisterItem(RawBeef{})
	world.RegisterItem(RawChicken{})
	world.RegisterItem(RawCod{})
	world.RegisterItem(RawMutton{})
	world.RegisterItem(RawPorkchop{})
	world.RegisterItem(RawRabbit{})
	world.RegisterItem(RawSalmon{})
	world.RegisterItem(RottenFlesh{})
	world.RegisterItem(Steak{})
	world.RegisterItem(GlisteringMelonSlice{})
	world.RegisterItem(MushroomStew{})
	world.RegisterItem(BeetrootSoup{})
	world.RegisterItem(RabbitStew{})
}
