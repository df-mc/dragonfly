package form

// Submitter is an entity that is able to submit a form sent to it. It is able to fill out fields in the form
// which will then be present when handled.
type Submitter interface {
	SendForm(form Form)
}

// Handler is closure used in form closing, button click
type Handler func(Submitter)

func (c Handler) Call(s Submitter) {
	if c != nil {
		c(s)
	}
}
