package item

type Template struct {
	Template TemplateType
}

// EncodeItem ...
func (t Template) EncodeItem() (name string, meta int16) {
	if t.Template.Name == "netherite_upgrade" {
		return "minecraft:netherite_upgrade_smithing_template", 0
	}
	return "minecraft:" + t.Template.Name + "_armor_trim_smithing_template", 0
}
