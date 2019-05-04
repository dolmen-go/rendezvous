// Package rendezvous provides synchronization utilities.
package rendezvous

import (
	"fmt"
	"sync"
)

// Func is just a task returning a runtime error.
type Func func() error

// WaitAll launches each Func in a separate goroutine and waits indefinitely until all goroutines finish.
// This is called a "rendez-vous".
//
// The result is the unordered list of non-nil errors returned by any func.
// Panic occuring inside a goroutine are caught and converted as errors.
func WaitAll(funcs ...Func) []error {
	if len(funcs) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	errChan := make(chan error, len(funcs))
	wg.Add(len(funcs))

	for _, f := range funcs {
		if f == nil {
			wg.Done()
			continue
		}
		go func(f Func) {
			var err error
			defer func() {
				if err != nil {
					errChan <- err
				}
				wg.Done()
			}()
			defer catchPanicAsError(&err)
			err = f()
		}(f)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func catchPanicAsError(perr *error) {
	if p := recover(); p != nil {
		if e, isError := p.(error); isError {
			*perr = e
		} else {
			*perr = fmt.Errorf("panic: %v", p)
		}
	}
}
