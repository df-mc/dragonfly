package block

// TorchColour ...
type TorchColour struct {
	colouredTorch
}

type colouredTorch uint8

// TorchColourPurple ...
func TorchColourPurple() TorchColour {
	return TorchColour{colouredTorch(0)}
}

// TorchColourBlue ...
func TorchColourBlue() TorchColour {
	return TorchColour{colouredTorch(1)}
}

// TorchColourGreen ...
func TorchColourGreen() TorchColour {
	return TorchColour{colouredTorch(2)}
}

// TorchColourRed ...
func TorchColourRed() TorchColour {
	return TorchColour{colouredTorch(3)}
}

// Name ...
func (c colouredTorch) Name() string {
	switch c {
	case 0:
		return "Purple Torch"
	case 1:
		return "Blue Torch"
	case 2:
		return "Green Torch"
	case 3:
		return "Red Torch"
	}
	panic("unknown torch colour")
}

// String ...
func (c colouredTorch) String() string {
	switch c {
	case 0:
		return "purple"
	case 1:
		return "blue"
	case 2:
		return "green"
	case 3:
		return "red"
	}
	panic("unknown torch colour")
}

// Uint8 ...
func (c colouredTorch) Uint8() uint8 {
	return uint8(c)
}

// TorchColours ...
func TorchColours() []TorchColour {
	return []TorchColour{TorchColourPurple(), TorchColourBlue(), TorchColourGreen(), TorchColourRed()}
}
