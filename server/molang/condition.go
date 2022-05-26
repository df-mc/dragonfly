package molang

// TODO: Improve!

import "fmt"

// Condition is a MoLang condition used for custom blocks.
type Condition struct {
	expression string
}

// String returns the condition's underlying expression.
func (c Condition) String() string {
	return c.expression
}

// Exists returns true if the condition is not blank.
func (c Condition) Exists() bool {
	return len(c.expression) > 0
}

// NoCondition represents a blank condition where nothing is required.
func NoCondition() Condition {
	return Condition{}
}

// ParseCondition parses a string MoLang condition into a Condition.
func ParseCondition(expression string) Condition {
	return Condition{expression}
}

// PropertyQueryCondition is a Condition to query a block's properties.
func PropertyQueryCondition(property string, value any) Condition {
	return ParseCondition(fmt.Sprintf("query.block_property('%v') == %v", property, value))
}

// TODO: Add more utility functions
