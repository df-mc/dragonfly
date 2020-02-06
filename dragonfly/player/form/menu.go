package form

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"reflect"
)

// Menu represents a menu form. These menus are made up of a title and a body, with a number of buttons which
// come below the body. These buttons may also have buttons on the side of them.
type Menu struct {
	title, body string
	submittable MenuSubmittable
}

// Button represents a button added to a Menu form. The button has text on it and an optional image, which
// may be either retrieved from a website or the local assets of the game.
type Button struct {
	// Text holds the text displayed on the button. It may use Minecraft formatting codes and may have
	// newlines.
	Text string
	// Image holds a path to an image for the button. The Image may either be an URL pointing to an image,
	// such as 'https://someimagewebsite.com/someimage.png', or a path pointing to a local asset, such as
	// 'textures/blocks/grass_carried'.
	Image string
}

// NewMenu creates a new Menu form using the MenuSubmittable passed to handle the output of the form. The
// title passed is formatted following the rules of fmt.Sprintln.
func NewMenu(submittable MenuSubmittable, title ...interface{}) Menu {
	t := reflect.TypeOf(submittable)
	if t.Kind() != reflect.Struct {
		panic("submittable must be struct")
	}
	m := Menu{title: format(title), submittable: submittable}
	m.verify()
	return m
}

// WithBody creates a copy of the Menu form and changes its body to the body passed, after which the new Menu
// form is returned. The text is formatted following the rules of fmt.Sprintln.
func (m Menu) WithBody(body ...interface{}) Menu {
	m.body = format(body)
	return m
}

// Title returns the formatted title passed to the menu upon construction using NewMenu().
func (m Menu) Title() string {
	return m.title
}

// Body returns the formatted text in the body passed to the menu using WithBody().
func (m Menu) Body() string {
	return m.body
}

// Buttons returns a list of all buttons of the MenuSubmittable. It parses them from the fields using
// reflection and returns them.
func (m Menu) Buttons() []Button {
	v := reflect.ValueOf(m.submittable)
	t := reflect.TypeOf(m.submittable)

	buttons := make([]Button, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		fieldT := t.Field(i)
		fieldV := v.Field(i)
		if !ast.IsExported(fieldT.Name) {
			continue
		}
		// Each exported field is guaranteed to be of type Button.
		buttons = append(buttons, fieldV.Interface().(Button))
	}
	return buttons
}

// SubmitJSON submits a JSON value to the menu, containing the index of the button clicked.
func (m Menu) SubmitJSON(b []byte, submitter Submitter) error {
	var index uint
	err := json.Unmarshal(b, &index)
	if err != nil {
		return fmt.Errorf("cannot parse button index as int: %v", err)
	}
	buttons := m.Buttons()
	if index >= uint(len(buttons)) {
		return fmt.Errorf("button index points to inexistent button: %v (only %v buttons present)", index, len(buttons))
	}
	m.submittable.Submit(submitter, buttons[index])
	return nil
}

// verify verifies if the form is valid, checking all fields are of the type Button. It panics if the form is
//not valid.
func (m Menu) verify() {
	v := reflect.ValueOf(m.submittable)
	t := reflect.TypeOf(m.submittable)
	for i := 0; i < v.NumField(); i++ {
		fieldT := t.Field(i)
		if !ast.IsExported(fieldT.Name) {
			continue
		}
		if _, ok := v.Field(i).Interface().(Button); !ok {
			panic("all exported fields must be of the type form.Button")
		}
	}
}

func (m Menu) __() {}
