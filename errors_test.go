package errors

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestError_E(t *testing.T) {
	errDummy := New("foo")
	tc := []struct {
		name   string
		args   []interface{}
		expect error
	}{
		{
			name: "only error",
			args: []interface{}{
				errDummy,
			},
			expect: &Error{
				cause: errDummy,
			},
		},
		{
			name: "error msg",
			args: []interface{}{
				errDummy,
				"network latency",
			},
			expect: &Error{
				s:     "network latency",
				cause: errDummy,
			},
		},

		{
			name: "error kind",
			args: []interface{}{
				errDummy,
				IO,
			},
			expect: &Error{
				cause: errDummy,
				Kind:  IO,
			},
		},
		{
			name: "error msg kind",
			args: []interface{}{
				errDummy,
				"network latency",
				IO,
			},
			expect: &Error{
				s:     "network latency",
				cause: errDummy,
				Kind:  IO,
			},
		},
		{
			name: "underlayin errors.Error type",
			args: []interface{}{
				"geting foo from db",
				E("network latency", IO, errDummy),
			},
			expect: &Error{
				s:     "geting foo from db",
				cause: E("network latency", IO, errDummy),
				Kind:  IO,
			},
		},
		{
			name: "underlayin errors.Meda",
			args: []interface{}{
				"geting foo from db",
				E("network latency", IO, errDummy, MetaData{"1": 1}),
			},
			expect: &Error{
				s:     "geting foo from db",
				cause: E("network latency", IO, errDummy, MetaData{"1": 1}),
				Kind:  IO,
				Meta:  MetaData{"1": 1},
			},
		},
		{
			name: "nil error",
			args: []interface{}{
				"network latency",
				IO,
			},
			expect: nil,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			err := E(tt.args...)
			if !reflect.DeepEqual(tt.expect, err) {
				t.Errorf("expected: %+v, got: %+v", tt.expect, err)
			}
		})
	}
}

func TestError_Msg(t *testing.T) {
	tc := []struct {
		name   string
		err    error
		expect string
	}{
		{
			name:   "simple error",
			err:    E(New("foo"), "try to get foo"),
			expect: "try to get foo",
		},
		{
			name:   "simple error no description",
			err:    E(New("foo")),
			expect: "foo",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			err, ok := tt.err.(*Error)
			if !ok {
				t.Error("invalid error, should be of type errors.Error")
				return
			}

			msg := err.Msg()
			if tt.expect != msg {
				t.Errorf("\nexpected: %s\n     got: %s", tt.expect, msg)
			}
		})
	}
}

func TestError_MarshalJSON(t *testing.T) {
	errDummy := New("foo")
	er := E("network latency", IO, errDummy, MetaData{"foo": "bar"})
	b, err := json.Marshal(er)
	if err != nil {
		t.Error(err)
		return
	}

	expect := `{"detail":{"foo":"bar"},"type":"I/O error","error":"network latency","code":3}`
	if string(b) != expect {
		t.Errorf("\nexpected: %s\n     got: %s", expect, string(b))
	}

}

func TestError_Error(t *testing.T) {
	errIO := E("io error", IO, New("network unreachable"))
	errUnmarshal := E("can't unmarshal bar", Unmarshal, errIO)
	errDecrypt := E("invalid key", Decrypt, errUnmarshal)
	megaError := E("no part of group", Permission, errDecrypt)

	tc := []struct {
		name   string
		err    error
		expect string
	}{
		{
			name:   "no msg",
			err:    E(New("foo")),
			expect: "foo",
		},
		{
			name:   "with msg",
			err:    E(New("foo"), "bar"),
			expect: "bar: foo",
		},
		{
			name:   "multiple underlaying errors",
			err:    megaError,
			expect: "no part of group: invalid key: can't unmarshal bar: io error: network unreachable",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			err, ok := tt.err.(*Error)
			if !ok {
				t.Error("invalid error, should be of type errors.Error")
				return
			}

			msg := err.Error()
			if tt.expect != msg {
				t.Errorf("\nexpected: %s\n     got: %s", tt.expect, msg)
			}
		})
	}
}

func TestIsKind(t *testing.T) {
	tc := []struct {
		name   string
		err    error
		kind   Kind
		expect bool
	}{
		{
			name:   "no error",
			err:    nil,
			kind:   Duplicated,
			expect: false,
		},
		{
			name:   "std error",
			err:    fmt.Errorf("some error"),
			kind:   Duplicated,
			expect: false,
		},
		{
			name:   "wrong kind",
			err:    E(Duplicated, "hi", fmt.Errorf("hi")),
			kind:   Invalid,
			expect: false,
		},
		{
			name:   "correct kind",
			err:    E(Duplicated, "hi", fmt.Errorf("hi")),
			kind:   Duplicated,
			expect: true,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			got := IsKind(tt.err, tt.kind)
			if tt.expect != got {
				t.Errorf("\ntest: %s\nexpected: %t\n     got: %t", tt.name, tt.expect, got)
			}
		})
	}
}
