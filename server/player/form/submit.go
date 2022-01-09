package form

// Submittable is a structure which may be submitted by sending it as a form using form.New(). When filled out
// and submitted, the struct will have its Submit method called and its fields will have the values that the
// Submitter passed filled out.
// The fields of a Submittable struct must be either unexported or have a type of one of those that implement
// the form.Element interface.
type Submittable interface {
	// Submit is called when the Submitter submits the form sent to it. Once this method is called, all fields
	// in the struct will have their values filled out as filled out by the Submitter.
	Submit(submitter Submitter)
}

// MenuSubmittable is a structure which may be submitted by sending it as a form using form.NewMenu(), much
// like a Submittable. The struct will have its Submit method called with the button pressed.
// A struct that implements the MenuSubmittable interface must only have exported fields with the type
// form.Button.
type MenuSubmittable interface {
	// Submit is called when the Submitter submits the menu form sent to it. The method is called with the
	// button that was pressed. It may be compared with buttons in the MenuSubmittable struct to check which
	// button was pressed.
	Submit(submitter Submitter, pressed Button)
}

// ModalSubmittable is a structure which may be submitted by sending it as a form using form.NewModal(), much
// like a Submittable and a MenuSubmittable. The struct will have its Submit method called with the button
// pressed.
// A struct that implements the ModalSubmittable interface must have exactly two exported fields with the type
// form.Button, which may be used to specify the text of the Modal form's buttons. Unlike with a Menu form,
// buttons on a Modal form will not have images.
type ModalSubmittable MenuSubmittable

// Closer represents a form which has special logic when being closed by a Submitter.
type Closer interface {
	// Close is called when the Submitter closes a form.
	Close(submitter Submitter)
}

// Submitter is an entity that is able to submit a form sent to it. It is able to fill out fields in the form
// which will then be present when handled.
type Submitter interface {
	SendForm(form Form)
}
