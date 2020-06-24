package block

//lint:file-ignore ST1005 Errors returned by this file are intentionally readable for an end user.

import (
	"fmt"
	"github.com/df-mc/dragonfly/dragonfly/internal/block_internal"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/sahilm/fuzzy"
	"go/ast"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// Parse attempts to parse a string passed into a slice of blocks. The string passed is a list of either one
// or more blocks, separated using a comma.
// If parsing the string was not successful, the slice returned is nil and an error is returned explaining the
// reason the string could not be parsed.
//
// Parse parses strings in a rather specific format. The strings accepted may look like the following:
// "andesite"
// "andesite,log[wood=oak]"
// "log,leaves[wood=spruce,persistent=true]"
// The properties, which are optional, are identical to the fields of the blocks registered, except for being
// fully lowercase.
//
// Errors returned by Parse are explicitly user-friendly. They are fit to be displayed to the end user
// supplying the string.
//noinspection GoErrorStringFormat
func Parse(s string) ([]world.Block, error) {
	s = strings.Replace(strings.Replace(s, "	", "", -1), "\n", "", -1)
	p := &parser{r: strings.NewReader(s)}
	var blocks []world.Block

	for p.r.Len() != 0 {
		name, err := p.block()
		if err != nil {
			return nil, err
		}
		if name == "" {
			return nil, fmt.Errorf("Invalid block name encountered: All names must be at least one letter long.")
		}
		b, ok := block_internal.BlockByTypeName(name)
		if !ok {
			matches := fuzzy.Find(name, block_internal.BlockNames())
			if len(matches) == 0 {
				// No matches found, there is no block that even remotely resembles what was typed.
				return nil, fmt.Errorf("Invalid block name encountered: '%v' is not a known block name.", name)
			}
			return nil, fmt.Errorf("Invalid block name encountered: '%v' is not a known block name. Did you mean '%v'?", name, matches[0].Str)
		}

		if p.propertiesFollowing {
			props, err := p.properties()
			if err != nil {
				return nil, err
			}
			val := reflect.New(reflect.TypeOf(b)).Elem()
			for k, v := range props {
				field := val.FieldByNameFunc(func(s string) bool {
					return titleStrToUnderscores(s) == k && ast.IsExported(s)
				})
				if !field.IsValid() {
					matches := fuzzy.Find(k, fieldNames(val))
					if len(matches) == 0 {
						// No matches found, there is no property that even remotely resembles what was typed.
						return nil, fmt.Errorf("Invalid block property encountered in block '%v': '%v' is not a known property.", name, k)
					}
					return nil, fmt.Errorf("Invalid block property encountered in block '%v': '%v' is not a known property. Did you mean '%v'?", name, k, matches[0].Str)
				}
				if err := setStringToField(field, v); err != nil {
					return nil, fmt.Errorf("Invalid block property value '%v' for property '%v' in block '%v': %v", v, k, name, err)
				}
			}
			b = val.Interface().(world.Block)
		}

		p.propertiesFollowing = false
		p.endOfProperties = false
		blocks = append(blocks, b)
	}
	return blocks, nil
}

// setStringToField attempts to set a string to a reflect.Value of a struct field passed.
func setStringToField(val reflect.Value, s string) error {
	switch v := val.Interface().(type) {
	case FromStringer:
		newValue, err := v.FromString(s)
		if err != nil {
			return fmt.Errorf("cannot parse '%v': %w", s, err)
		}
		val.Set(reflect.ValueOf(newValue))
	case bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("cannot parse '%v' as bool", s)
		}
		val.SetBool(v)
	case string:
		val.SetString(s)
	case int8, int16, int32, int64, int:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse '%v' as int", s)
		}
		val.SetInt(v.(int64))
	case uint8, uint16, uint32, uint64, uint:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse '%v' as uint", s)
		}
		val.SetUint(v.(uint64))
	default:
		panic("unknown block property type " + fmt.Sprintf("%T", val.Interface()))
	}
	return nil
}

// FromStringer represents a type that is able to return a specific variant of itself by reading the string
// passed.
type FromStringer interface {
	FromString(s string) (interface{}, error)
}

// titleStrToUnderscores converts a string like 'HelloWorld' to 'hello_world'.
func titleStrToUnderscores(s string) string {
	var name strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i != 0 {
				name.WriteByte('_')
			}
			name.WriteRune(unicode.ToLower(r))
			continue
		}
		name.WriteRune(r)
	}
	return name.String()
}

// fieldNames returns a list of all field names of the reflect.Value representing a struct passed.
func fieldNames(v reflect.Value) []string {
	m := make([]string, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		if !ast.IsExported(v.Type().Field(i).Name) {
			continue
		}
		m = append(m, titleStrToUnderscores(v.Type().Field(i).Name))
	}
	return m
}

// parser handles the parsing of a string containing blocks.
type parser struct {
	r                   *strings.Reader
	propertiesFollowing bool
	endOfProperties     bool
}

// block reads a block name from the input string.
//noinspection GoErrorStringFormat
func (p *parser) block() (string, error) {
	var n strings.Builder
	for {
		b, err := p.r.ReadByte()
		if err != nil {
			break
		}
		if b == ' ' {
			// Skip spaces.
			continue
		}
		r := rune(b)
		if unicode.IsLetter(r) || unicode.IsNumber(r) || b == '_' {
			// We've got a normal letter, number or underscore, so we add it to the name and continue.
			n.WriteByte(b)
			continue
		} else if b == ',' {
			// We've got a comma, so we've hit the end of this block. Break right away.
			break
		} else if b == '[' {
			// We've got an opening bracket, so we've hit the end of the block name and properties are
			// starting. We let the caller know properties are following and break.
			p.propertiesFollowing = true
			break
		}
		return "", fmt.Errorf("Invalid character '%v' encountered: \"%v>>%v<<%v\".", string(b), n.String(), string(b), p.leftover())
	}
	return n.String(), nil
}

// properties reads a list of properties separated by commas and enclosed by brackets from the input string.
// If the first character is not a '[', it will return an empty map.
//noinspection GoErrorStringFormat
func (p *parser) properties() (map[string]string, error) {
	m := make(map[string]string)
	for !p.endOfProperties {
		name, err := p.readPropertyName()
		if err != nil {
			return nil, err
		}
		if name == "" {
			return nil, fmt.Errorf("Invalid property name encountered: Properties must always be at least one character long.")
		}
		val, err := p.readPropertyValue()
		if err != nil {
			return nil, err
		}
		m[name] = val
	}
	return m, nil
}

// readPropertyName reads a property name, which must be ended with an '='.
//noinspection GoErrorStringFormat
func (p *parser) readPropertyName() (string, error) {
	var n strings.Builder
	for {
		b, err := p.r.ReadByte()
		if err != nil {
			return "", fmt.Errorf("Unexpected end of property name: Property names must have a value supplied after a '=' and must be closed using a ']'.")
		}
		if b == ' ' {
			// Skip spaces.
			continue
		}
		r := rune(b)
		if unicode.IsLetter(r) || unicode.IsNumber(r) || b == '_' {
			// We've got a normal letter, number or underscore, so we add it to the name and continue.
			n.WriteByte(b)
			continue
		} else if b == '=' {
			// We've got a comma, so we've hit the end of this block. Break right away.
			break
		} else if b == ']' {
			return "", fmt.Errorf("Unexpected ']' after property name '%v': Each property name must have a value.", n.String())
		}
		return "", fmt.Errorf("Invalid character '%v' encountered in property name: \"%v>>%v<<%v\".", string(b), n.String(), string(b), p.leftover())
	}
	return n.String(), nil
}

// readPropertyValue reads a property value, which must be ended with either a ']' or a ','.
//noinspection GoErrorStringFormat
func (p *parser) readPropertyValue() (string, error) {
	var n strings.Builder
	for {
		b, err := p.r.ReadByte()
		if err != nil {
			return "", fmt.Errorf("Unexpected end of property value: The properties must be ended with a ']'.")
		}
		if b == ']' {
			// We've got a closing bracket, so we've hit the end of the properties. We let the caller know
			// we've hit the end of the properties.
			// Make sure to remove the comma that follows.
			if comma, err := p.r.ReadByte(); comma != ',' && err == nil {
				return "", fmt.Errorf("Unexpected character '%v' encountered after ']'. Expecting a comma or end of the string.", string(comma))
			}
			p.endOfProperties = true
			break
		} else if b == ',' {
			// We've hit a comma, meaning a new property is incoming.
			break
		}
		// For the values, we accept anything, provided strings may contain any character.
		n.WriteByte(b)
	}
	return strings.Trim(n.String(), " "), nil
}

// leftover returns the leftover string in the parser.
func (p *parser) leftover() string {
	m := make([]byte, p.r.Len())
	_, _ = p.r.Read(m)
	return string(m)
}
