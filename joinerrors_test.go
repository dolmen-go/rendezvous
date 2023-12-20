package rendezvous

import (
	"context"
	"errors"
	"regexp"
	"testing"
)

type myError struct{}

func (myError) Error() string {
	return "myError"
}

func TestJoinErrors(t *testing.T) {
	if !errors.Is(joinErrors(context.Canceled), context.Canceled) {
		t.Error("errors.Is failure")
	}
	if errors.Is(joinErrors(context.Canceled), myError{}) {
		t.Error("errors.Is failure")
	}
	var me myError
	if !errors.As(joinErrors(context.Canceled, myError{}), &me) {
		t.Error("errors.As failure")
	}
	if errors.As(joinErrors(context.Canceled), &me) {
		t.Error("errors.As failure")
	}

	err := joinErrors(myError{}, myError{})
	if err == nil {
		t.Fatal("nil unexpected")
	}
	var errs []error
	if err := err.(interface{ Unwrap() []error }); err == nil {
		t.Fatal("Unwrap() []error expected")
	} else {
		errs = err.Unwrap()
	}
	if len(errs) != 2 {
		t.Fatal("len 2 expected")
	}

	_ = errs[0].(myError)
	_ = errs[1].(myError)

	str := err.Error()
	re := regexp.MustCompile(regexp.QuoteMeta(myError{}.Error()))
	if len(re.FindAllString(str, -1)) != 2 {
		t.Fatal("2 occurences of (myError{}).Error() expected")
	}
}
