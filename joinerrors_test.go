package rendezvous

import (
	"context"
	"errors"
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
	var me myError
	if !errors.As(joinErrors(context.Canceled, myError{}), &me) {
		t.Error("errors.As failure")
	}
}
