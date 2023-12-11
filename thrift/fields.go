package thrift

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type isZeroer interface {
	IsZero() bool
}

var isZeroerType = reflect.TypeOf((*isZeroer)(nil)).Elem()

type structFields struct {
	sorted []structField
	byID   map[int16]*structField
}

type structField struct {
	index  []int
	typ    reflect.Type
	fncs   *arshaler
	isZero func(addressableValue) bool
	fieldOptions
}

func makeStructFields(t reflect.Type) (structFields, *SemanticError) {
	var allFields []structField

	var hasAnyThriftTag bool   // whether any Go struct field has a `thrift` tag
	var hasAnyThriftField bool // whether any Thrift serializable fields exist in current struct
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		_, hasTag := sf.Tag.Lookup("thrift")
		hasAnyThriftTag = hasAnyThriftTag || hasTag
		options, ignored, err := parseFieldOptions(sf)
		if err != nil {
			return structFields{}, &SemanticError{GoType: t, Err: err}
		} else if ignored {
			continue
		}
		hasAnyThriftField = true

		f := structField{
			index:        sf.Index,
			typ:          sf.Type,
			fieldOptions: options,
		}

		// Provide a function that uses a type's IsZero method.
		switch {
		case sf.Type.Kind() == reflect.Interface && sf.Type.Implements(isZeroerType):
			f.isZero = func(va addressableValue) bool {
				// Avoid panics calling IsZero on a nil interface or
				// non-nil interface with nil pointer.
				return va.IsNil() || (va.Elem().Kind() == reflect.Pointer && va.Elem().IsNil()) || va.Interface().(isZeroer).IsZero()
			}
		case sf.Type.Kind() == reflect.Pointer && sf.Type.Implements(isZeroerType):
			f.isZero = func(va addressableValue) bool {
				// Avoid panics calling IsZero on nil pointer.
				return va.IsNil() || va.Interface().(isZeroer).IsZero()
			}
		case sf.Type.Implements(isZeroerType):
			f.isZero = func(va addressableValue) bool { return va.Interface().(isZeroer).IsZero() }
		case reflect.PointerTo(sf.Type).Implements(isZeroerType):
			f.isZero = func(va addressableValue) bool { return va.Addr().Interface().(isZeroer).IsZero() }
		}

		f.fncs = lookupArshaler(sf.Type)
		allFields = append(allFields, f)
	}

	isEmptyStruct := t.NumField() == 0
	if !isEmptyStruct && !hasAnyThriftTag && !hasAnyThriftField {
		err := errors.New("Go struct has no exported fields")
		return structFields{}, &SemanticError{GoType: t, Err: err}
	}

	sort.Slice(allFields, func(i, j int) bool {
		return allFields[i].id < allFields[j].id
	})

	fs := structFields{
		sorted: allFields,
		byID:   make(map[int16]*structField),
	}
	for i := range fs.sorted {
		f := &fs.sorted[i]
		fs.byID[f.id] = f
	}

	return fs, nil
}

type fieldOptions struct {
	id       int16
	name     string
	required bool
}

func parseFieldOptions(sf reflect.StructField) (out fieldOptions, ignored bool, err error) {
	out.name = sf.Name

	tag, hasTag := sf.Tag.Lookup("thrift")

	// Check whether this field is explicitly ignored.
	if tag == "-" {
		return fieldOptions{}, true, nil
	}

	// Check whether this field is unexported.
	if !sf.IsExported() {
		// Tag options specified on an unexported field suggests user error.
		if hasTag {
			err = firstError(err, fmt.Errorf("unexported Go struct field %s cannot have non-ignored `thrift:%q` tag", sf.Name, tag))
		}
		return fieldOptions{}, true, err
	}

	i := strings.IndexByte(tag, ',')
	if i == -1 {
		i = len(tag)
	}
	id64, idErr := strconv.ParseInt(tag[:i], 10, 16)
	tag = tag[i:]
	if idErr != nil {
		return fieldOptions{}, true, firstError(err, fmt.Errorf("Go struct field %s has malformed `thrift` tag: invalid field ID: %w", sf.Name, idErr))
	}
	out.id = int16(id64)

	// Handle any additional tag options (if any).
	seenOpts := make(map[string]bool)
	for len(tag) > 0 {
		// Consume comma delimiter.
		if tag[0] != ',' {
			err = firstError(err, fmt.Errorf("Go struct field %s has malformed `thrift` tag: invalid character %q before next option (expecting ',')", sf.Name, tag[0]))
		} else {
			tag = tag[len(","):]
			if len(tag) == 0 {
				err = firstError(err, fmt.Errorf("Go struct field %s has malformed `thrift` tag: invalid trailing ',' character", sf.Name))
				break
			}
		}

		// Consume and process the tag option.
		opt, n, err2 := consumeTagOption(tag)
		if err2 != nil {
			err = firstError(err, fmt.Errorf("Go struct field %s has malformed `thrift` tag: %v", sf.Name, err2))
		}
		rawOpt := tag[:n]
		tag = tag[n:]

		switch opt {
		case "required":
			out.required = true
		default:
			// Reject keys that resemble one of the supported options.
			// This catches invalid mutants such as "omitEmpty" or "omit_empty".
			normOpt := strings.ReplaceAll(strings.ToLower(opt), "_", "")
			switch normOpt {
			case "required":
				err = firstError(err, fmt.Errorf("Go struct field %s has invalid appearance of `%s` tag option; specify `%s` instead", sf.Name, opt, normOpt))
			}

			// NOTE: Everything else is ignored. This does not mean it is
			// forward compatible to insert arbitrary tag options since
			// a future version of this package may understand that tag.
		}

		// Reject duplicates.
		switch {
		case seenOpts[opt]:
			err = firstError(err, fmt.Errorf("Go struct field %s has duplicate appearance of `%s` tag option", sf.Name, rawOpt))
		}
		seenOpts[opt] = true

	}
	return out, false, err
}

func consumeTagOption(in string) (string, int, error) {
	i := strings.IndexByte(in, ',')
	if i < 0 {
		i = len(in)
	}

	switch r, _ := utf8.DecodeRuneInString(in); {
	// Option as a Go identifier.
	case r == '_' || unicode.IsLetter(r):
		n := len(in) - len(strings.TrimLeftFunc(in, isLetterOrDigit))
		return in[:n], n, nil
	case len(in) == 0:
		return in[:i], i, io.ErrUnexpectedEOF
	default:
		return in[:i], i, fmt.Errorf("invalid character %q at start of option (expecting Unicode letter or single quote)", r)
	}
}

func isLetterOrDigit(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
}
