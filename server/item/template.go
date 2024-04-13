package item

// Template is an item used in smithing tables to alter tools and armor.
// They are consumed when used, but can be duplicated using an existing template, its material and diamonds.
type Template struct {
	// Template the upgrade item used in smithing tables.
	Template ArmourTrimTemplate
}

// EncodeItem ...
func (t Template) EncodeItem() (name string, meta int16) {
	if t.Template == TemplateNetheriteUpgrade() {
		return "minecraft:netherite_upgrade_smithing_template", 0
	}
	return "minecraft:" + t.Template.Name + "_armor_trim_smithing_template", 0
}
