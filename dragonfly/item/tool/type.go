package tool

// Type represents the type of a tool. This decides the type of blocks that the tool is used for.
type Type struct {
	t
}

// TypeNone is the type of items that are not tools.
var TypeNone = Type{t(-1)}

// TypePickaxe is the type for pickaxes.
var TypePickaxe = Type{t(0)}

// TypeAxe is the type for axes.
var TypeAxe = Type{t(1)}

// TypeHoe is the type for hoes.
var TypeHoe = Type{t(2)}

// TypeShovel is the type for shovels.
var TypeShovel = Type{t(3)}

// TypeShears is the type for shears.
var TypeShears = Type{t(4)}

// TypeSword is the type for swords.
var TypeSword = Type{t(5)}

// t represents the type of a tool.
type t int
