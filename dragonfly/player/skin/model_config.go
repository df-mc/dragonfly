package skin

import (
	"encoding/json"
)

// ModelConfig specifies the way that the model (geometry data) is used to form the complete skin. It does
// this by setting model names for specific keys found in the struct.
type ModelConfig struct {
	// Default is the 'default' model to use. This model is essentially the model of the skin that will be
	// used at all times, when nothing special is being done. (For example, an animation)
	// The field holds the name of one of the models present in the JSON of the skin's model.
	// This field should always be filled out.
	Default string `json:"default"`
	// AnimatedFace is the model of an animation played over the face. This field should be set if the model
	// contains the model of an animation, in which case this field should hold the name of that model.
	AnimatedFace string `json:"animated_face,omitempty"`
}

// modelConfigContainer is a container of the model config data when encoded.
type modelConfigContainer struct {
	Geometry ModelConfig `json:"geometry"`
}

// Encode encodes a ModelConfig into its JSON representation.
func (cfg ModelConfig) Encode() []byte {
	b, _ := json.Marshal(modelConfigContainer{Geometry: cfg})
	return b
}

// DecodeModelConfig attempts to decode a ModelConfig from the JSON data passed. If not successful, an error
// is returned.
func DecodeModelConfig(b []byte) (ModelConfig, error) {
	var m modelConfigContainer
	err := json.Unmarshal(b, &m)
	return m.Geometry, err
}
