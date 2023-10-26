//go:build go1.20

package rendezvous

import "errors"

func joinErrors(errs ...error) error {
	return errors.Join(errs...)
}
