package session

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PlayerSkinHandler handles the PlayerSkin packet.
type PlayerSkinHandler struct{}

// Handle ...
func (PlayerSkinHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.PlayerSkin)

	playerSkin := skin.New(int(pk.Skin.SkinImageWidth), int(pk.Skin.SkinImageHeight))

	if _, err := base64.StdEncoding.Decode(playerSkin.Pix, pk.Skin.SkinData); err != nil {
		return fmt.Errorf("error decoding SkinData base64 data: %v", err)
	} else if len(playerSkin.Pix) != int(pk.Skin.SkinImageHeight*pk.Skin.SkinImageWidth*4) {
		return fmt.Errorf("invalid SkinData size: got %v, but expected %v", len(playerSkin.Pix), int(pk.Skin.SkinImageHeight*pk.Skin.SkinImageWidth*4))
	}

	playerSkin.Cape = skin.NewCape(int(pk.Skin.CapeImageWidth), int(pk.Skin.CapeImageHeight))

	if _, err := base64.StdEncoding.Decode(playerSkin.Cape.Pix, pk.Skin.CapeData); err != nil {
		return fmt.Errorf("error decoding CapeData base64 data: %v", err)
	} else if len(playerSkin.Cape.Pix) != int(pk.Skin.CapeImageHeight*pk.Skin.CapeImageWidth*4) {
		return fmt.Errorf("invalid CapeData size: got %v, but expected %v", len(playerSkin.Cape.Pix), int(pk.Skin.CapeImageHeight*pk.Skin.CapeImageWidth*4))
	}

	if _, err := base64.StdEncoding.Decode(playerSkin.Model, pk.Skin.SkinGeometry); err != nil {
		return fmt.Errorf("SkinGeometry was not a valid base64 string: %v", err)
	} else if len(playerSkin.Model) != 0 {
		m := make(map[string]interface{})
		if err := json.Unmarshal(playerSkin.Model, &m); err != nil {
			return fmt.Errorf("SkinGeometry base64 decoded was not a valid JSON string: %v", err)
		}
	}

	var skinResourcePatch []byte
	if _, err := base64.StdEncoding.Decode(skinResourcePatch, pk.Skin.SkinResourcePatch); err != nil {
		return fmt.Errorf("SkinResourcePatch was not a valid base64 string: %v", err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(skinResourcePatch, &m); err != nil {
		return fmt.Errorf("SkinResourcePatch base64 decoded was not a valid JSON string: %v", err)
	}
	if pk.Skin.SkinID == "" {
		return fmt.Errorf("SkinID must not be an empty string")
	}

	modelConfig, _ := skin.DecodeModelConfig(skinResourcePatch)

	playerSkin.Persona = pk.Skin.PersonaSkin
	playerSkin.ModelConfig = modelConfig
	playerSkin.PlayFabID = pk.Skin.PlayFabID

	for _, animation := range pk.Skin.Animations {
		var t skin.AnimationType
		switch animation.AnimationType {
		case protocol.SkinAnimationHead:
			t = skin.AnimationHead
		case protocol.SkinAnimationBody32x32:
			t = skin.AnimationBody32x32
		case protocol.SkinAnimationBody128x128:
			t = skin.AnimationBody128x128
		default:
			return fmt.Errorf("invalid animation type: %v", animation.AnimationType)
		}

		anim := skin.NewAnimation(int(animation.ImageWidth), int(animation.ImageHeight), int(animation.ExpressionType), t)
		anim.FrameCount = int(animation.FrameCount)

		if _, err := base64.StdEncoding.Decode(anim.Pix, animation.ImageData); err != nil {
			return fmt.Errorf("error decoding animation base64 data: %v", err)
		} else if len(anim.Pix) != int(animation.ImageHeight*animation.ImageWidth*4) {
			return fmt.Errorf("animation size is invalid: got %v, but expected %v", len(anim.Pix), int(animation.ImageHeight*animation.ImageWidth*4))
		}

		playerSkin.Animations = append(playerSkin.Animations, anim)
	}

	s.c.SetSkin(playerSkin)

	return nil
}
