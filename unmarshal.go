package darknut

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/structtag"
	"github.com/ttab/newsdoc"
)

// UnmarshalDocument unmarshals a NewsDoc document into a struct.
func UnmarshalDocument(doc newsdoc.Document, o any) error {
	return unmarshal(documentBlocks(&doc), documentAttributes(&doc), o)
}

const (
	metaBlocks    = "meta"
	linkBlocks    = "links"
	contentBlocks = "content"
)

type (
	blockFunc func(string) []newsdoc.Block
	attrFunc  func(string) (string, bool)
)

func unmarshal(blocks blockFunc, attributes attrFunc, o any) error {
	t := reflect.TypeOf(o)

	if t.Kind() != reflect.Pointer {
		return errors.New("must be a pointer")
	}

	e := t.Elem()

	if e.Kind() != reflect.Struct {
		return errors.New("must be a pointer to a struct")
	}

	for i := 0; i < e.NumField(); i++ {
		field := e.Field(i)
		fieldRef := reflect.ValueOf(o).Elem().Field(i)

		tag, err := structtag.Parse(string(field.Tag))
		if err != nil {
			return fmt.Errorf("failed to parse tags for %s.%s: %w",
				t.Name(), field.Name, err,
			)
		}

		if tag == nil {
			continue
		}

		ndtag, err := tag.Get("newsdoc")
		if err != nil {
			return fmt.Errorf("invalid newsdoc tag for %s.%s: %w",
				t.Name(), field.Name, err,
			)
		}

		switch ndtag.Name {
		case metaBlocks, linkBlocks, contentBlocks:
			ok, err := recurse(blocks(ndtag.Name), fieldRef, ndtag)
			if err != nil {
				return fmt.Errorf("failed to unmarshal blocks for %s.%s: %w",
					t.Name(), field.Name, err,
				)
			}

			if !ok && fieldRef.Kind() != reflect.Pointer {
				return fmt.Errorf("missing required block for %s.%s",
					t.Name(), field.Name,
				)
			}
		default:
			err = readValue(attributes, fieldRef, ndtag)
			if err != nil {
				return fmt.Errorf("failed to unmarshal %s.%s: %w",
					t.Name(), field.Name, err,
				)
			}
		}
	}

	return nil
}

func readValue(
	attributes attrFunc,
	field reflect.Value,
	ndtag *structtag.Tag,
) error {
	pointer := field.Kind() == reflect.Pointer

	value, ok := attributes(ndtag.Name)
	if !ok && !pointer && !slices.Contains(ndtag.Options, "optional") {
		return errors.New("missing required value")
	}

	if !ok {
		return nil
	}

	valueType := field.Type()
	if pointer {
		valueType = valueType.Elem()
	}

	kind := valueType.Kind()

	switch kind { //nolint:exhaustive
	case reflect.String:
		if pointer {
			field.Set(reflect.ValueOf(&value))
		} else {
			field.SetString(value)
		}

		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool: %w", err)
		}

		if pointer {
			field.Set(reflect.ValueOf(&b))
		} else {
			field.SetBool(b)
		}

		return nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		i, err := strconv.ParseUint(value, 10, bitSize(kind))
		if err != nil {
			return fmt.Errorf("invalid uint: %w", err)
		}

		if pointer {
			ip := reflect.New(valueType)
			field.Set(ip)
			field.Elem().SetUint(i)
		} else {
			field.SetUint(i)
		}

		return nil
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		i, err := strconv.ParseInt(value, 10, bitSize(kind))
		if err != nil {
			return fmt.Errorf("invalid int: %w", err)
		}

		if pointer {
			ip := reflect.New(valueType)
			field.Set(ip)
			field.Elem().SetInt(i)
		} else {
			field.SetInt(i)
		}

		return nil
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, bitSize(kind))
		if err != nil {
			return fmt.Errorf("invalid float: %w", err)
		}

		if pointer {
			ip := reflect.New(valueType)
			field.Set(ip)
			field.Elem().SetFloat(f)
		} else {
			field.SetFloat(f)
		}

		return nil
	case reflect.Struct:
		tt := reflect.TypeOf(time.Time{})

		if valueType.AssignableTo(tt) {
			t, err := time.Parse(timeFormat(ndtag), value)
			if err != nil {
				return fmt.Errorf("invalid time: %w", err)
			}

			if pointer {
				field.Set(reflect.ValueOf(&t))
			} else {
				field.Set(reflect.ValueOf(t))
			}

			return nil
		}
	}

	unmarshaller, ok := field.Addr().Interface().(encoding.TextUnmarshaler)
	if ok {
		err := unmarshaller.UnmarshalText([]byte(value))
		if err != nil {
			return fmt.Errorf("failed to unmarshal text: %w", err)
		}

		return nil
	}

	return nil
}

func timeFormat(ndtag *structtag.Tag) string {
	for _, opt := range ndtag.Options {
		k, v, ok := strings.Cut(opt, "=")
		if ok && k == "format" {
			return v
		}
	}

	return time.RFC3339
}

func bitSize(k reflect.Kind) int {
	switch k { //nolint:exhaustive
	case reflect.Int, reflect.Int64,
		reflect.Uint, reflect.Uint64,
		reflect.Float64:
		return 64
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		return 32
	case reflect.Int16, reflect.Uint16:
		return 16
	case reflect.Int8, reflect.Uint8:
		return 8
	}

	return 0
}

func recurse(
	blocks []newsdoc.Block,
	field reflect.Value,
	ndtag *structtag.Tag,
) (bool, error) {
	pattern := make(map[string]string)

	for _, opt := range ndtag.Options {
		k, v, _ := strings.Cut(opt, "=")
		pattern[k] = v
	}

	slice := field.Kind() == reflect.Slice

bloop:
	for i := range blocks {
		ba := blockAttributes(&blocks[i])

		for k, v := range pattern {
			bv, _ := ba(k)
			if bv != v {
				continue bloop
			}
		}

		var target reflect.Value

		switch field.Kind() { //nolint:exhaustive
		case reflect.Struct:
			target = field.Addr()
		case reflect.Pointer:
			target = reflect.New(field.Type().Elem())
			field.Set(target)
		case reflect.Slice:
			eType := field.Type().Elem()

			if eType.Kind() == reflect.Pointer {
				return false, errors.New(
					"slice element  cannot be a pointer")
			}

			target = reflect.New(eType)
		default:
			return false, fmt.Errorf(
				"invalid field type %q", field.Type().Name())
		}

		err := unmarshal(
			blockBlocks(&blocks[i]),
			blockAttributes(&blocks[i]),
			target.Interface(),
		)
		if err != nil {
			return false, err
		}

		if slice {
			field.Set(reflect.Append(field, target.Elem()))

			continue
		}

		return true, nil
	}

	return slice, nil
}

func documentBlocks(d *newsdoc.Document) func(string) []newsdoc.Block {
	return func(s string) []newsdoc.Block {
		switch s {
		case metaBlocks:
			return d.Meta
		case linkBlocks:
			return d.Links
		case contentBlocks:
			return d.Content
		}

		return nil
	}
}

const (
	docAttrType     = "type"
	docAttrLanguage = "language"
	docAttrTitle    = "title"
	docAttrUUID     = "uuid"
	docAttrURI      = "uri"
	docAttrURL      = "url"
)

func documentAttributes(d *newsdoc.Document) attrFunc {
	return func(name string) (string, bool) {
		switch name {
		case docAttrUUID:
			return d.UUID, true
		case docAttrType:
			return d.Type, true
		case docAttrURI:
			return d.URI, true
		case docAttrURL:
			return d.URL, true
		case docAttrTitle:
			return d.Title, true
		case docAttrLanguage:
			return d.Language, true
		}

		return "", false
	}
}

func blockBlocks(b *newsdoc.Block) func(string) []newsdoc.Block {
	return func(s string) []newsdoc.Block {
		switch s {
		case "meta":
			return b.Meta
		case "links":
			return b.Links
		case "content":
			return b.Content
		}

		return nil
	}
}

const (
	blockAttrID          = "id"
	blockAttrUUID        = "uuid"
	blockAttrType        = "type"
	blockAttrURI         = "uri"
	blockAttrURL         = "url"
	blockAttrTitle       = "title"
	blockAttrRel         = "rel"
	blockAttrName        = "name"
	blockAttrValue       = "value"
	blockAttrContentType = "contenttype"
	blockAttrRole        = "role"
)

func blockAttributes(block *newsdoc.Block) attrFunc {
	return func(name string) (string, bool) {
		key, isData := strings.CutPrefix(name, "data.")
		if isData {
			if block.Data == nil {
				return "", false
			}

			v, ok := block.Data[key]

			return v, ok
		}

		switch name {
		case blockAttrID:
			return block.ID, true
		case blockAttrUUID:
			return block.UUID, true
		case blockAttrType:
			return block.Type, true
		case blockAttrURI:
			return block.URI, true
		case blockAttrURL:
			return block.URL, true
		case blockAttrTitle:
			return block.Title, true
		case blockAttrRel:
			return block.Rel, true
		case blockAttrName:
			return block.Name, true
		case blockAttrValue:
			return block.Value, true
		case blockAttrContentType:
			return block.Contenttype, true
		case blockAttrRole:
			return block.Role, true
		}

		return "", false
	}
}
