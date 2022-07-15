package item

import (
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
)

//noinspection SpellCheckingInspection
func init() {
	for _, t := range ToolTiers() {
		world.RegisterItem(Pickaxe{Tier: t})
		world.RegisterItem(Axe{Tier: t})
		world.RegisterItem(Shovel{Tier: t})
		world.RegisterItem(Sword{Tier: t})
		world.RegisterItem(Hoe{Tier: t})
	}
	for _, t := range ArmourTiers() {
		world.RegisterItem(Helmet{Tier: t})
		world.RegisterItem(Chestplate{Tier: t})
		world.RegisterItem(Leggings{Tier: t})
		world.RegisterItem(Boots{Tier: t})
	}
	world.RegisterItem(TurtleShell{})

	world.RegisterItem(Bucket{})

	world.RegisterItem(Shears{})

	world.RegisterItem(Snowball{})
	world.RegisterItem(EnderPearl{})

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
	world.RegisterItem(EnchantedBook{})

	world.RegisterItem(Stick{})
	world.RegisterItem(MagmaCream{})

	world.RegisterItem(BoneMeal{})
	world.RegisterItem(Wheat{})
	world.RegisterItem(Beetroot{})
	world.RegisterItem(MelonSlice{})

	world.RegisterItem(Apple{})

	world.RegisterItem(Brick{})

	world.RegisterItem(Leather{})

	world.RegisterItem(Bow{})
	world.RegisterItem(Arrow{})

	world.RegisterItem(GlassBottle{})
	for i, p := range potion.All() {
		if i > 4 {
			world.RegisterItem(Arrow{Tip: p})
		}
		world.RegisterItem(SplashPotion{Type: p})
		world.RegisterItem(Potion{Type: p})
	}

	world.RegisterItem(FlintAndSteel{})

	world.RegisterItem(CarrotOnAStick{})
	world.RegisterItem(WarpedFungusOnAStick{})

	world.RegisterItem(PrismarineCrystals{})

	world.RegisterItem(PoisonousPotato{})
	world.RegisterItem(GoldenApple{})
	world.RegisterItem(EnchantedApple{})
	world.RegisterItem(Pufferfish{})
	world.RegisterItem(Clock{})
	world.RegisterItem(Compass{})
	world.RegisterItem(Spyglass{})

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
	world.RegisterItem(Gunpowder{})
	world.RegisterItem(HeartOfTheSea{})
	world.RegisterItem(Honeycomb{})
	world.RegisterItem(InkSac{})
	world.RegisterItem(InkSac{Glowing: true})
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
	world.RegisterItem(Cookie{})
	world.RegisterItem(GoldenCarrot{})
	world.RegisterItem(PumpkinPie{})
	world.RegisterItem(Beef{})
	world.RegisterItem(Beef{Cooked: true})
	world.RegisterItem(Chicken{})
	world.RegisterItem(Chicken{Cooked: true})
	world.RegisterItem(Cod{})
	world.RegisterItem(Cod{Cooked: true})
	world.RegisterItem(Mutton{})
	world.RegisterItem(Mutton{Cooked: true})
	world.RegisterItem(Porkchop{})
	world.RegisterItem(Porkchop{Cooked: true})
	world.RegisterItem(Rabbit{})
	world.RegisterItem(Rabbit{Cooked: true})
	world.RegisterItem(Salmon{})
	world.RegisterItem(Salmon{Cooked: true})
	world.RegisterItem(RottenFlesh{})
	world.RegisterItem(GlisteringMelonSlice{})
	world.RegisterItem(MushroomStew{})
	world.RegisterItem(BeetrootSoup{})
	world.RegisterItem(RabbitStew{})
	world.RegisterItem(PoppedChorusFruit{})
	for _, dye := range AllDyes() {
		world.RegisterItem(dye)
	}
	world.RegisterItem(TropicalFish{})
	world.RegisterItem(AmethystShard{})
	world.RegisterItem(DiscFragment{})
	world.RegisterItem(EchoShard{})
}
