//go:build !go1.20

package rendezvous

import "errors"

func joinErrors(errs ...error) error {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	e := &joinError{
		errs: make([]error, 0, n),
	}
	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
	return e
}

type joinError struct {
	errs []error
}

func (e *joinError) Error() string {
	var b []byte
	for i, err := range e.errs {
		if i > 0 {
			b = append(b, '\n')
		}
		b = append(b, err.Error()...)
	}
	return string(b)
}

func (e *joinError) Unwrap() []error {
	return e.errs
}

func (e *joinError) Is(target error) bool {
	for _, f := range e.errs {
		if errors.Is(f, target) {
			return true
		}
	}
	return false
}

func (e *joinError) As(target interface{}) bool {
	for _, f := range e.errs {
		if errors.As(f, target) {
			return true
		}
	}
	return false
}
