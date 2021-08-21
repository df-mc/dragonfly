package item

// Armoured represents an entity that is able to wear armour in a specific armour inventory. These entities
// typically include human-like entities such as zombies.
type Armoured interface {
	// Armour returns the armour inventory of the entity.
	Armour() ArmourContainer
}

// ArmourContainer represents a container of armour. Generally, entities will want to use the inventory.Armour
// type.
type ArmourContainer interface {
	// SetHelmet sets the item stack passed as the helmet in the inventory.
	SetHelmet(helmet Stack)
	// Helmet returns the item stack set as helmet in the inventory.
	Helmet() Stack
	// SetChestplate sets the item stack passed as the chestplate in the inventory.
	SetChestplate(chestplate Stack)
	// Chestplate returns the item stack set as chestplate in the inventory.
	Chestplate() Stack
	// SetLeggings sets the item stack passed as the leggings in the inventory.
	SetLeggings(leggings Stack)
	// Leggings returns the item stack set as leggings in the inventory.
	Leggings() Stack
	// SetBoots sets the item stack passed as the boots in the inventory.
	SetBoots(boots Stack)
	// Boots returns the item stack set as boots in the inventory.
	Boots() Stack
	// Clear clears all items in the inventory.
	Clear()
}
