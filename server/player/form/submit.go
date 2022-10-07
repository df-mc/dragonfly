package form

// Submitter is an entity that is able to submit a form sent to it. It is able to fill out fields in the form
// which will then be present when handled.
type Submitter interface {
	SendForm(form Form)
}

// Closer represents a form which has special logic when being closed by a Submitter.
type Closer = func(Submitter)
