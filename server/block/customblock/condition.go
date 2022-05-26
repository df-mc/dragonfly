package customblock

import "fmt"

// Condition is a MoLang condition used for custom blocks.
type Condition struct {
	expression string
}

// NoCondition represents a blank condition where nothing is required.
func NoCondition() Condition {
	return Condition{}
}

// ParseCondition parses a string MoLang condition into a Condition.
func ParseCondition(expression string) Condition {
	// TODO: Validation.
	return Condition{expression}
}

// PropertyQueryCondition is a Condition to query a block's properties.
func PropertyQueryCondition(property string, value any) Condition {
	return ParseCondition(fmt.Sprintf("query.block_property('%v') == %v", property, value))
}
