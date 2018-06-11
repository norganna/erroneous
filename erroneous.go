package erroneous

import (
	"encoding/json"
	"fmt"
	"runtime"
)

// ErrFields is a map[string]interface{}.
type ErrFields map[string]interface{}

// Erroneous is an error which keeps track of the line it was generated with.
type Erroneous struct {
	msg    string
	err    error
	fields ErrFields
	file   string
	line   int
	depth  int
}

// Error makes Erroneous an error.
func (e *Erroneous) Error() string {
	if e == nil {
		return "unknown error"
	}

	msg := e.Message()

	if e.file != "" {
		msg += fmt.Sprintf(" [%s:%d]", e.file, e.line)
	}

	if e.fields != nil {
		data, _ := json.Marshal(e.fields)
		if len(data) > 0 {
			msg += "  " + string(data)
		}
	}

	return msg
}

// Message returns the message string for the error.
func (e *Erroneous) Message() string {
	msg := e.msg

	if e.err != nil {
		if msg != "" {
			msg += ": "
		}
		msg += e.err.Error()
	}
	return msg
}

// Source returns the file and line for the error.
func (e *Erroneous) Source() (string, int) {
	return e.file, e.line
}

// Fields returns the embedded fields for the error.
func (e *Erroneous) Fields() ErrFields {
	return e.fields
}

// ErrOpts are functions which can manipulate an error.
type ErrOpts func(e *Erroneous) error

// Msg attaches a message and fields to the error.
func Msg(msg string, fields ErrFields) ErrOpts {
	return func(e *Erroneous) error {
		e.msg = msg
		if fields != nil {
			e.fields = fields

			if v, ok := fields["error"]; ok {
				if err, ok := v.(error); ok {
					return Err(err)(e)
				}
			}
		}
		return nil
	}
}

// Fields attaches a fields to the error.
func Fields(fields ErrFields) ErrOpts {
	return func(e *Erroneous) error {
		e.fields = fields
		return nil
	}
}

// Err attaches an existing error to the error.
func Err(err error) ErrOpts {
	return func(e *Erroneous) error {
		if _, ok := err.(*Erroneous); ok {
			return err
		}
		e.err = err
		return nil
	}
}

// Source sets the source file and line for the error.
func Source(file string, line int) ErrOpts {
	return func(e *Erroneous) error {
		e.file = file
		e.line = line
		return nil
	}
}

// Depth sets the lookup depth for a caller file and line.
func Depth(depth int) ErrOpts {
	return func(e *Erroneous) error {
		e.depth = depth
		return nil
	}
}

// New returns a new Erroneous error.
func New(opts ...ErrOpts) error {
	e := &Erroneous{
		depth: 2,
	}
	for _, fn := range opts {
		err := fn(e)
		if err != nil {
			return err
		}
	}

	if e.file == "" {
		_, file, line, ok := runtime.Caller(e.depth)
		if ok {
			e.file = file
			e.line = line
		}

	}

	return e
}
