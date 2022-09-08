package dialogue

// Dialogue represents a singular dialog scene. Scenes are used to within NewMenu to create a new Dialog Menu. Submit is
// called when a button is submitted by a Submitter.
type Dialogue interface {
	Menu() Menu
	// Submit is called when a Submitter submits a Button on the specified Menu.
	Submit(submitter Submitter, pressed Button)
}
