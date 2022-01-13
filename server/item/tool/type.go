package tool

// Type represents the type of tool. This decides the type of blocks that the tool is used for.
type Type struct {
	t
}

// TypeNone is the type of items that are not tools.
var TypeNone = Type{-1}

// TypePickaxe is the type for pickaxes.
var TypePickaxe = Type{0}

// TypeAxe is the type for axes.
var TypeAxe = Type{1}

// TypeHoe is the type for hoes.
var TypeHoe = Type{2}

// TypeShovel is the type for shovels.
var TypeShovel = Type{3}

// TypeShears is the type for shears.
var TypeShears = Type{4}

// TypeSword is the type for swords.
var TypeSword = Type{5}

// t represents the type of a Tool.
type t int
