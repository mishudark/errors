package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// MetaData is used to store aditional info about the underlaying error,
// this info will be serialized by MarshalJSON under the key "errors"
type MetaData map[string]interface{}

// Error is the type that implements the error interface.
// It contains a number of fields, each of different type.
// An Error value may leave some values unset.
type Error struct {
	// Kind is the class of error, such as permission failure,
	// or "Other" if its class is unknown or irrelevant.
	Kind Kind
	// The underlying error that triggered this one, if any.
	cause error
	// The msg to end user
	s string
	// Metadata about the underlaying error
	Meta MetaData
}

var _ json.Marshaler = (*Error)(nil)

// Kind defines the kind of error this is, mostly for use by systems
type Kind uint8

// Kinds of errors.
//
// The values of the error kinds are common between both
// clients and servers. Do not reorder this list or remove
// any items since that will change their values.
// New items must be added only to the end.
const (
	Unknown       Kind = iota // Unclassified error.
	Invalid                   // Invalid operation for this type of item.
	Permission                // Permission denied.
	IO                        // External I/O error such as network failure.
	Duplicated                // Duplicated item.
	NotExist                  // Item does not exist.
	Private                   // Information withheld.
	Internal                  // Internal error or inconsistency.
	Decrypt                   // Ivalid encription info.
	Unmarshal                 // Invalid input data.
	Transient                 // A transient error.
	Unsupported               // An unsupported media type.
	NotAcceptable             // We cannot accept the provided media types.
)

// String transforms enums into string, useful for encoders
func (k Kind) String() string {
	switch k {
	case Unknown:
		return "Unknown error"
	case Invalid:
		return "invalid operation"
	case Permission:
		return "permission denied"
	case IO:
		return "I/O error"
	case Duplicated:
		return "item already exists"
	case NotExist:
		return "item does not exist"
	case Private:
		return "information withheld"
	case Internal:
		return "internal error"
	case Decrypt:
		return "invalid encryption"
	case Unmarshal:
		return "invalid data"
	case Transient:
		return "transient error"
	case Unsupported:
		return "unsupported"
	case NotAcceptable:
		return "not accepted"
	}
	return "unknown error kind"
}

// StatusCode transform kind to http.StatusCode
func (k Kind) StatusCode() int {
	switch k {
	case Invalid,
		Decrypt,
		Unmarshal:
		return http.StatusBadRequest
	case Permission,
		Private:
		return http.StatusUnauthorized
	case Transient:
		return http.StatusServiceUnavailable
	case NotExist:
		return http.StatusNotFound
	case Unsupported:
		return http.StatusUnsupportedMediaType
	case NotAcceptable:
		return http.StatusNotAcceptable
	case Duplicated:
		return http.StatusConflict
	case Unknown:
	case Internal:
	case IO:
	}

	return http.StatusInternalServerError
}

// E builds an error value from its arguments.
// There must be at least one argument.
// The type of each argument determines its meaning.
// If more than one argument of a given type is presented,
// only the last one is recorded.
//
// If the error is nil, nil will be returned.
//
// The types are:
//	string
//		The msg to help to trace error callers, the last msg used
//		will be available via Msg method
//	errors.Kind
//		The class of error, such as permission failure.
//	error
//		The underlying error that triggered this one.
//
// If Kind is not specified or Unknown, we set it to the Kind of
// the underlying error.
// If MetaData is not defined, we use the underlaying MetaData
func E(args ...interface{}) error {
	e := &Error{}

	for _, arg := range args {
		switch opt := arg.(type) {
		case string:
			e.s = opt
		case Kind:
			e.Kind = opt
		case MetaData:
			e.Meta = opt
		case map[string]interface{}:
			meta := make(MetaData)
			for k, v := range opt {
				meta[k] = v
			}

			e.Meta = meta
		case *Error:
			// Make a copy
			copy := *opt
			e.cause = &copy
		case error:
			e.cause = opt
			//default:
			//	return Errorf("unknown type %T, value %v in error call", arg, arg)
		}
	}

	if e.cause == nil {
		return nil
	}

	// Fill missing fileds in case previous error is a Error type
	if err, ok := e.cause.(*Error); ok {
		if e.Kind == Unknown {
			e.Kind = err.Kind
		}

		if e.Meta == nil {
			e.Meta = err.Meta
		}
	}

	return e
}

// Msg returns the last known error msg, this is used to
// show a friendly msg to end user ,instead the full trace
// and to avoid leak of internal info
func (e *Error) Msg() string {
	str := e.Error()
	pos := strings.Index(str, ":")
	if pos == -1 {
		pos = len(str)
	}

	return str[:pos]
}

// Error format the output, joining all previous errors
func (e *Error) Error() string {
	str := e.s
	// skip ':' if e.s it's empty
	if str != "" {
		str += ": "
	}
	return str + e.cause.Error()
}

// Cause returns the underlaying error
func (e *Error) Cause() error {
	return e.cause
}

// StatusCode returns an http.StatusCode based on error kind
func (e *Error) StatusCode() int {
	return e.Kind.StatusCode()
}

// MarshalJSON determines how the error will be serialized
// if data is available, it will be serialized as is
// otherwise the error will be serialized as dict with key "error"
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Detail MetaData `json:"detail,omitempty"`
		Type   string   `json:"type"`
		Error  string   `json:"error"`
		Code   Kind     `json:"code"`
	}{
		Error:  e.Msg(),
		Detail: e.Meta,
		Type:   e.Kind.String(),
		Code:   e.Kind,
	})
}

// Recreate the errors.New functionality of the standard Go errors package
// so we can create simple text errors when needed.

// New returns an error that formats as the given text. It is intended to
// be used as the error-typed argument to the E function.
func New(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// Errorf is equivalent to errors.Errorf, but allows clients to import only this
// package for all error handling.
func Errorf(format string, args ...interface{}) error {
	return &errorString{fmt.Sprintf(format, args...)}
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//     type causer interface {
//            Cause() error
//     }
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

// IsKind is a convenience function that determines if the
// kind of the provided error value matches that of the
// provided kind.
func IsKind(err error, kind Kind) bool {
	if e, ok := err.(*Error); ok {
		return e.Kind == kind
	}
	return false
}
