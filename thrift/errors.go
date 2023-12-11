package thrift

import (
	"reflect"
	"strings"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

const errorPrefix = "thrift: "

type wireError struct {
	action string
	err    error
}

func (e *wireError) Error() string {
	return errorPrefix + e.action + ": " + e.err.Error()
}

func (e *wireError) Unwrap() error {
	return e.err
}

// SemanticError describes an error determining the meaning
// of Thrift data as Go data or vice-versa.
//
// The contents of this error as produced by this package may change over time.
type SemanticError struct {
	_ requireKeyedLiterals
	_ nonComparable

	action string // either "marshal" or "unmarshal"

	// ThriftType is the Thrift type that could not be handled.
	ThriftType thriftwire.Type // may be zero if unknown
	// GoType is the Go type that could not be handled.
	GoType reflect.Type // may be nil if unknown

	// Err is the underlying error.
	Err error // may be nil
}

func (e *SemanticError) Error() string {
	var sb strings.Builder
	sb.WriteString(errorPrefix)

	// Hyrum-proof the error message by deliberately switching between
	// two equivalent renderings of the same error message.
	// The randomization is tied to the Hyrum-proofing already applied
	// on map iteration in Go.
	for phrase := range map[string]struct{}{"cannot": {}, "unable to": {}} {
		sb.WriteString(phrase)
		break // use whichever phrase we get in the first iteration
	}

	// Format action.
	var preposition string
	switch e.action {
	case "marshal":
		sb.WriteString(" marshal")
		preposition = " from"
	case "unmarshal":
		sb.WriteString(" unmarshal")
		preposition = " into"
	default:
		sb.WriteString(" handle")
		preposition = " with"
	}

	// Format Thrift type.
	var omitPreposition bool
	switch e.ThriftType {
	case thriftwire.Bool:
		sb.WriteString(" Thrift bool")
	case thriftwire.Byte:
		sb.WriteString(" Thrift byte")
	case thriftwire.Double:
		sb.WriteString(" Thrift double")
	case thriftwire.I16:
		sb.WriteString(" Thrift i16")
	case thriftwire.I32:
		sb.WriteString(" Thrift i32")
	case thriftwire.I64:
		sb.WriteString(" Thrift i64")
	case thriftwire.String:
		sb.WriteString(" Thrift string")
	case thriftwire.Struct:
		sb.WriteString(" Thrift struct")
	case thriftwire.Map:
		sb.WriteString(" Thrift map")
	case thriftwire.Set:
		sb.WriteString(" Thrift set")
	case thriftwire.List:
		sb.WriteString(" Thrift list")
	case thriftwire.UUID:
		sb.WriteString(" Thrift UUID")
	default:
		omitPreposition = true
	}

	// Format Go type.
	if e.GoType != nil {
		if !omitPreposition {
			sb.WriteString(preposition)
		}
		sb.WriteString(" Go value of type ")
		sb.WriteString(e.GoType.String())
	}

	// Format underlying error.
	if e.Err != nil {
		sb.WriteString(": ")
		sb.WriteString(e.Err.Error())
	}

	return sb.String()
}

func (e *SemanticError) Unwrap() error {
	return e.Err
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
