package md

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	tag = "markdown"

	tagOmitField      = "-"
	tagObfuscateField = "obfuscate"
)

var (
	marshalerType = reflect.TypeOf(new(Marshaler)).Elem()
)

// Marshal returns the Slack Markdown encoding of v.
//
// Marshal traverses the value v recursively, similar to
// how encoding/json does.
//
// Struct field names will be printed in bold. Values will
// follow the default Golang enconding, except for pointers,
// which will be automatically dereferenced, and for objects
// implementing the Marshaler interface.
//
// Two tags are provided to facilitate proper serilializing.
//
//   // Field is ignored by this package.
//   Field int `json:"-"`
//
//   // Field appears, but at most the last 4 characters are shown
//   Field int `json:"obfuscate"`
//
// Those tags will preventing sensitive that from showing in your
// Slack channels.
//
func Marshal(v interface{}) ([]byte, error) {
	return marshal(
		v,
		marshalOpts{indentLevel: 0},
	)
}

type marshalOpts struct {
	indentLevel int
	obfuscate   bool
}

func (o marshalOpts) withIncrementedIndentLevel() marshalOpts {
	o.indentLevel = o.indentLevel + 1

	return o
}

func marshal(v interface{}, opts marshalOpts) ([]byte, error) {
	t := reflect.TypeOf(v)

	if t.Implements(marshalerType) {
		return v.(Marshaler).MarshalMarkdown()
	}

	switch t.Kind() {
	case reflect.String:
		return marshalStr(v, opts)
	case reflect.Struct:
		return marshalStruct(v, opts)
	case reflect.Ptr:
		return marshalPrt(v, opts)
	default:
		return []byte(fmt.Sprintf("%v", v)), nil
	}
}

func marshalStr(v interface{}, opts marshalOpts) ([]byte, error) {
	valueStr := v.(string)

	if opts.obfuscate {
		valueStr = obfuscate(valueStr)
	}

	return []byte(valueStr), nil
}

func obfuscate(str string) string {
	length := len(str)

	if length <= 4 {
		return str
	}

	return strings.Repeat("*", length-4) + str[length-4:]
}

func marshalPrt(v interface{}, opts marshalOpts) ([]byte, error) {
	value := reflect.ValueOf(v)

	if value.IsNil() {
		return []byte("null"), nil
	}

	return marshal(value.Elem().Interface(), opts)
}

func marshalStruct(v interface{}, opts marshalOpts) ([]byte, error) {
	lines := []string{}

	value := reflect.ValueOf(v)
	t := reflect.TypeOf(v)

	for i := 0; i < value.NumField(); i++ {
		fieldType := t.Field(i)
		fieldValue := value.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		innerOpts := opts.withIncrementedIndentLevel()
		if fieldType.Tag.Get(tag) == tagObfuscateField {
			innerOpts.obfuscate = true
		}

		marshaledStructValue, err := marshal(fieldValue.Interface(), innerOpts)
		if err != nil {
			return nil, err
		}

		if fieldType.Tag.Get(tag) == tagOmitField {
			continue
		}

		line := fmt.Sprintf("- **%s**: %s", fieldType.Name, marshaledStructValue)
		lines = append(lines, line)
	}

	lines = applyIndentation(lines, opts.indentLevel)

	formattedLines := strings.Join(lines, "\n")

	return []byte(formattedLines), nil
}

func applyIndentation(lines []string, level int) []string {
	for i, line := range lines {
		lines[i] = strings.Repeat("\t", level) + line
	}

	if len(lines) > 0 && level > 0 {
		lines[0] = "\n" + lines[0] // inner structs should start on a new line
	}

	return lines
}
