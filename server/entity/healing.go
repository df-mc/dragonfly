package entity

type (
	// FoodHealingSource is a healing source used for when an entity regenerates health automatically when their food
	// bar is at least 90% filled.
	FoodHealingSource struct{}
)

func (FoodHealingSource) HealingSource() {}
