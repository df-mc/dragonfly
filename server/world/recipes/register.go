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

// Recipes returns each recipe in an array.
func All() []Recipe {
	return recipes
}
