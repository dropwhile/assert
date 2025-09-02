// Copyright (c) 2025 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
//
// Inspiration from https://github.com/nalgeon/be

package assert

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// TestingT is the subset of [testing.T] (see also [testing.TB]) used by the assert package.
type TestingT interface {
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

type helperT interface {
	Helper()
}

type equaler[T any] interface {
	Equal(T) bool
}

func Equal[T any](t TestingT, got, want T, msgAndArgs ...any) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}

	if !isEqual(got, want) {
		t.Errorf("got: %#v; want: %#v;%s", got, want, formatMsg(msgAndArgs...))
	}
}

func NotEqual[T any](t TestingT, got, want T, msgAndArgs ...any) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}

	if isEqual(got, want) {
		t.Errorf("got: %#v; expected values to be different;%s", got, formatMsg(msgAndArgs...))
	}
}

func Nil(t TestingT, got any, msgAndArgs ...any) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}

	if !isNil(got) {
		t.Errorf("got: %#v; want: <nil>;%s", got, formatMsg(msgAndArgs...))
	}
}

func NotNil(t TestingT, got any, msgAndArgs ...any) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}

	if isNil(got) {
		t.Errorf("got: <nil>; expected non-nil;%s", formatMsg(msgAndArgs...))
	}
}

func ErrorIs(t TestingT, got error, want any, msgAndArgs ...any) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}

	switch w := want.(type) {
	case nil:
		if got != nil {
			t.Errorf("unexpected error: %s;%s", got, formatMsg(msgAndArgs...))
		}
	case string:
		if !strings.Contains(got.Error(), w) {
			t.Errorf("got: %q; want: %q;%s", got, want, formatMsg(msgAndArgs...))
		}
	case error:
		if !errors.Is(got, w) {
			if isNil(got) {
				t.Errorf("got: <nil>; want: %T(%v);%s", w, w, formatMsg(msgAndArgs...))
			} else {
				t.Errorf("got: %T(%v); want: %T(%v);%s", got, got, w, w, formatMsg(msgAndArgs...))
			}
		}
	case reflect.Type:
		target := reflect.New(w).Interface()
		if !errors.As(got, target) {
			t.Errorf("got: %T; want: %v;%s", got, w, formatMsg(msgAndArgs...))
		}
	default:
		t.Fatalf("unsupported want type: %T", want)
	}
}

func ErrorAs(t TestingT, got error, target any, msgAndArgs ...any) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}

	if got == nil {
		t.Errorf("got: nil; want assignable to: %T;%s", target, formatMsg(msgAndArgs...))
		return
	}
	if !errors.As(got, target) {
		t.Errorf("got: %v#; want assignable to: %T(%#v);%s", got, target, formatMsg(msgAndArgs...))
	}
}

func MatchesRegexp(t TestingT, got, pattern string, msgAndArgs ...any) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}

	matched, err := regexp.MatchString(pattern, got)
	if err != nil {
		t.Fatalf("unable to parse regexp pattern %s: %s", pattern, err.Error())
		return
	}
	if !matched {
		t.Errorf("got: %q; want to match %q;%s", got, pattern, formatMsg(msgAndArgs...))
	}
}

func isEqual[T any](got, want T) bool {
	if isNil(got) && isNil(want) {
		return true
	}

	if equalable, ok := any(got).(equaler[T]); ok {
		return equalable.Equal(want)
	}

	// Special case for byte slices.
	if aBytes, ok := any(got).([]byte); ok {
		bBytes := any(want).([]byte)
		return bytes.Equal(aBytes, bBytes)
	}

	// Fallback to reflective comparison.
	return reflect.DeepEqual(got, want)
}

func isNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

func formatMsg(msgAndArgs ...any) string {
	var b strings.Builder
	switch len(msgAndArgs) {
	case 0:
	case 1:
		b.WriteString(" ")
		if s, ok := msgAndArgs[0].(string); ok {
			b.WriteString(s)
		} else {
			fmt.Fprint(&b, msgAndArgs[0])
		}
	default:
		b.WriteString(" ")
		if s, ok := msgAndArgs[0].(string); ok {
			fmt.Fprintf(&b, s, msgAndArgs[1:]...)
		} else {
			fmt.Fprint(&b, msgAndArgs...)
		}
	}
	return b.String()
}
