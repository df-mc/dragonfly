package block

// Bedrock is a block that is indestructible in survival.
type Bedrock struct{}

func (Bedrock) Name() string {
	return "Bedrock"
}
