package item

// Vanilla item tags that may be added to an item using the minecraft:tags item component. Only vanilla item
// tags may use the "minecraft:" namespace. Custom tags may use any other namespace.
const (
	// Armour tags.
	TagArmor         = "minecraft:is_armor"
	TagHorseArmor    = "minecraft:horse_armor"
	TagNautilusArmor = "minecraft:nautilus_armor"
	TagHarness       = "minecraft:harness"

	// Food tags.
	TagIsMeat   = "minecraft:is_meat"
	TagIsCooked = "minecraft:is_cooked"
	TagIsFood   = "minecraft:is_food"

	// Piglin bartering tags.
	TagPiglinLoved      = "minecraft:piglin_loved"
	TagPiglinRepellents = "minecraft:piglin_repellents"

	// Smithing table tags.
	TagTransformableItems = "minecraft:transformable_items"
	TagTransformMaterials = "minecraft:transform_materials"
	TagTransformTemplates = "minecraft:transform_templates"

	// Sulfur cube archetype tags.
	TagSulfurCubeArchetypeBouncy         = "minecraft:sulfur_cube_archetype_bouncy"
	TagSulfurCubeArchetypeRegular        = "minecraft:sulfur_cube_archetype_regular"
	TagSulfurCubeArchetypeSlowFlat       = "minecraft:sulfur_cube_archetype_slow_flat"
	TagSulfurCubeArchetypeFastFlat       = "minecraft:sulfur_cube_archetype_fast_flat"
	TagSulfurCubeArchetypeLight          = "minecraft:sulfur_cube_archetype_light"
	TagSulfurCubeArchetypeFastSliding    = "minecraft:sulfur_cube_archetype_fast_sliding"
	TagSulfurCubeArchetypeSlowSliding    = "minecraft:sulfur_cube_archetype_slow_sliding"
	TagSulfurCubeArchetypeSticky         = "minecraft:sulfur_cube_archetype_sticky"
	TagSulfurCubeArchetypeHighResistance = "minecraft:sulfur_cube_archetype_high_resistance"
	TagSulfurCubeArchetypeExplosive      = "minecraft:sulfur_cube_archetype_explosive"

	// Tier tags.
	TagChainmailTier = "minecraft:chainmail_tier"
	TagCopperTier    = "minecraft:copper_tier"
	TagDiamondTier   = "minecraft:diamond_tier"
	TagGoldenTier    = "minecraft:golden_tier"
	TagIronTier      = "minecraft:iron_tier"
	TagLeatherTier   = "minecraft:leather_tier"
	TagNetheriteTier = "minecraft:netherite_tier"
	TagStoneTier     = "minecraft:stone_tier"
	TagWoodenTier    = "minecraft:wooden_tier"

	// Tool tags.
	TagDigger    = "minecraft:digger"
	TagIsAxe     = "minecraft:is_axe"
	TagIsHoe     = "minecraft:is_hoe"
	TagIsPickaxe = "minecraft:is_pickaxe"
	TagIsShears  = "minecraft:is_shears"
	TagIsShovel  = "minecraft:is_shovel"
	TagIsSpear   = "minecraft:is_spear"
	TagIsSword   = "minecraft:is_sword"
	TagIsTool    = "minecraft:is_tool"
	TagIsTrident = "minecraft:is_trident"

	// Trim tags.
	TagTrimmableArmors = "minecraft:trimmable_armors"
	TagTrimMaterials   = "minecraft:trim_materials"
	TagTrimTemplates   = "minecraft:trim_templates"

	// Woodset tags.
	TagBoat         = "minecraft:boat"
	TagBoats        = "minecraft:boats"
	TagChestBoat    = "minecraft:chest_boat"
	TagCrimsonStems = "minecraft:crimson_stems"
	TagDoor         = "minecraft:door"
	TagHangingActor = "minecraft:hanging_actor"
	TagHangingSign  = "minecraft:hanging_sign"
	TagLogs         = "minecraft:logs"
	TagLogsThatBurn = "minecraft:logs_that_burn"
	TagMangroveLogs = "minecraft:mangrove_logs"
	TagPlanks       = "minecraft:planks"
	TagSign         = "minecraft:sign"
	TagWarpedStems  = "minecraft:warped_stems"
	TagWoodenSlabs  = "minecraft:wooden_slabs"

	// Miscellaneous tags.
	TagArrow                  = "minecraft:arrow"
	TagBanner                 = "minecraft:banner"
	TagCoals                  = "minecraft:coals"
	TagEgg                    = "minecraft:egg"
	TagIsFish                 = "minecraft:is_fish"
	TagLecternBooks           = "minecraft:lectern_books"
	TagIsMinecart             = "minecraft:is_minecart"
	TagMusicDisc              = "minecraft:music_disc"
	TagSand                   = "minecraft:sand"
	TagSoulFireBaseBlocks     = "minecraft:soul_fire_base_blocks"
	TagSpawnEgg               = "minecraft:spawn_egg"
	TagStoneBricks            = "minecraft:stone_bricks"
	TagStoneCraftingMaterials = "minecraft:stone_crafting_materials"
	TagStoneToolMaterials     = "minecraft:stone_tool_materials"
	TagVibrationDamper        = "minecraft:vibration_damper"
	TagWool                   = "minecraft:wool"
	TagBookshelfBooks         = "minecraft:bookshelf_books"
	TagDecoratedPotSherds     = "minecraft:decorated_pot_sherds"
	TagMetalNuggets           = "minecraft:metal_nuggets"
)

// Tagged represents an item that has one or more item tags. These tags may be used by the client for various
// purposes, such as determining the tier of an item or checking if an item is food.
type Tagged interface {
	// Tags returns the tags of the item.
	Tags() []string
}
