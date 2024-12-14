package item

// SmithingTemplate is an item used in smithing tables to alter tools and
// armour. They are consumed when used, but can be duplicated using an existing
// template, its material and diamonds.
type SmithingTemplate struct {
	// Template the upgrade item used in smithing tables.
	Template SmithingTemplateType
}

// EncodeItem ...
func (t SmithingTemplate) EncodeItem() (name string, meta int16) {
	if t.Template == TemplateNetheriteUpgrade() {
		return "minecraft:netherite_upgrade_smithing_template", 0
	}
	return "minecraft:" + t.Template.String() + "_armor_trim_smithing_template", 0
}
