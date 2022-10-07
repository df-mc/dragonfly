package form

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Form represents a form that may be sent to a Submitter. The three types of forms, custom forms, menu forms
// and modal forms implement this interface.
type Form interface {
	json.Marshaler
	SubmitJSON(b []byte, submitter Submitter) error
	__()
}
type Data = struct{}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end.
func format(a []any) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}
