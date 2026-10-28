package session

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// personaPieceTypes maps the persona piece type names sent at login to the protocol's PieceType values.
var personaPieceTypes = map[string]uint32{
	"persona_unknown":        protocol.PieceTypeUnknown,
	"persona_skeleton":       protocol.PieceTypeSkeleton,
	"persona_body":           protocol.PieceTypeBody,
	"persona_skin":           protocol.PieceTypeSkin,
	"persona_bottom":         protocol.PieceTypeBottom,
	"persona_feet":           protocol.PieceTypeFeet,
	"persona_dress":          protocol.PieceTypeDress,
	"persona_top":            protocol.PieceTypeTop,
	"persona_high_pants":     protocol.PieceTypeHighPants,
	"persona_hand":           protocol.PieceTypeHands,
	"persona_outerwear":      protocol.PieceTypeOuterwear,
	"persona_facial_hair":    protocol.PieceTypeFacialHair,
	"persona_mouth":          protocol.PieceTypeMouth,
	"persona_eyes":           protocol.PieceTypeEyes,
	"persona_hair":           protocol.PieceTypeHair,
	"persona_hood":           protocol.PieceTypeHood,
	"persona_back":           protocol.PieceTypeBack,
	"persona_face_accessory": protocol.PieceTypeFaceAccessory,
	"persona_head":           protocol.PieceTypeHead,
	"persona_legs":           protocol.PieceTypeLegs,
	"persona_left_leg":       protocol.PieceTypeLeftLeg,
	"persona_right_leg":      protocol.PieceTypeRightLeg,
	"persona_arms":           protocol.PieceTypeArms,
	"persona_left_arm":       protocol.PieceTypeLeftArm,
	"persona_right_arm":      protocol.PieceTypeRightArm,
	"persona_capes":          protocol.PieceTypeCapes,
	"persona_classic_skin":   protocol.PieceTypeClassicSkin,
	"persona_emote":          protocol.PieceTypeEmote,
	"unsupported":            protocol.PieceTypeUnsupported,
}

// personaPieceNames is the reverse of personaPieceTypes.
var personaPieceNames = func() map[uint32]string {
	names := make(map[uint32]string, len(personaPieceTypes))
	for name, t := range personaPieceTypes {
		names[t] = name
	}
	return names
}()

// personaPieceType returns the PieceType of a persona piece type name, or PieceTypeUnknown.
func personaPieceType(name string) uint32 {
	return personaPieceTypes[name]
}

// personaPieceName returns the persona piece type name of a PieceType, or "persona_unknown".
func personaPieceName(t uint32) string {
	if name, ok := personaPieceNames[t]; ok {
		return name
	}
	return "persona_unknown"
}
