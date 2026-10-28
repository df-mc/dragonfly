package skin

// PersonaPiece is one of the pieces a persona skin is built from. Each piece names marketplace content by ID
// rather than carrying pixel or mesh data.
type PersonaPiece struct {
	// PieceID is a UUID identifying this piece.
	PieceID string
	// PieceType is the kind of piece, in the persona_* form the client sends at login, such as
	// "persona_body" or "persona_facial_hair".
	PieceType string
	// PackID is a UUID identifying the pack the piece belongs to.
	PackID string
	// Default specifies whether the piece is one of the default pieces of a Steve or Alex skin.
	Default bool
	// ProductID is a UUID identifying the piece for purchases. It is empty for default pieces.
	ProductID string
}

// PersonaPieceTintColour holds the tint colours applied to one persona piece, such as persona_mouth,
// persona_eyes or persona_hair.
type PersonaPieceTintColour struct {
	// PieceType is the piece the tints apply to, in the same form as PersonaPiece.PieceType.
	PieceType string
	// Colours holds four ARGB colours in hex notation, such as "#ffa12722". Unused entries are "#0".
	Colours [4]string
}
