package recipes

import (
	_ "embed"
)

// recipes is a list of each recipe.
var recipes []Recipe

// Register registers a new recipe.
func Register(recipe Recipe) {
	recipes = append(recipes, recipe)
}

// All returns each recipe in a slice.
func All() []Recipe {
	return recipes
}
