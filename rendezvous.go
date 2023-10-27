/*
   Copyright 2023 Olivier Mengué

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package rendezvous provides synchronization utilities.
package rendezvous

import (
	"context"
	"fmt"
	"sync"
)

// Task is just a func returning a runtime error.
type Task = func() error

// WaitAll launches each [Task] in a separate goroutine and waits indefinitely until all goroutines finish.
// This is called a "rendez-vous".
//
// The result is the unordered list of non-nil errors returned by any task.
// Panic occuring inside a goroutine are caught and converted as errors.
func WaitAll(tasks ...Task) []error {
	if len(tasks) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))
	wg.Add(len(tasks))

	for _, t := range tasks {
		if t == nil {
			wg.Done()
			continue
		}
		go func(t Task) {
			var err error
			defer func() {
				if err != nil {
					errChan <- err
				}
				wg.Done()
			}()
			defer catchPanicAsError(&err)
			err = t()
		}(t)
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

// TaskCtx is a task with a cancelable execution context.
type TaskCtx = func(ctx context.Context) error

// WaitFirstError runs each task in a goroutine and waits for all to terminate.
//
// The first error returned by a task triggers the cancellation of the context of the others.
// In any case, return happens only after all launched goroutines are done.
//
// The returned error, if not nil, wraps the list of errors, in no particular order. Use this to unwrap:
//
//	var errs interface { Unwrap() []error }
//	if errors.As(err, &errs) {
//		... errs.Unwrap() ...
//	}
//
// Notes:
//   - if the context is cancelled, there is no builtin way to know which task was launched and succeeded.
//   - when abort happens, some tasks may not have even been launched.
func WaitFirstError(ctx context.Context, tasks ...TaskCtx) error {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	var earlyStop bool
	errChan := make(chan error, len(tasks))

launch:
	for _, t := range tasks {
		if t == nil {
			continue
		}
		select {
		case <-ctx.Done():
			earlyStop = true
			break launch
		case <-childCtx.Done():
			earlyStop = true
			break launch
		default:
			wg.Add(1)
			go func(t TaskCtx) {
				defer wg.Done()
				var err error
				defer func() {
					if err != nil {
						errChan <- err
						cancel()
					}
				}()
				defer catchPanicAsError(&err)
				err = t(childCtx)
			}(t)
		}
	}

	wg.Wait()
	close(errChan)

	var errs []error

	// Context cancelled?
	errCtx := ctx.Err()
	if errCtx != nil {
		if earlyStop {
			// We stopped launching tasks because of context cancellation
			// so we must report that cause.
			errs = append(errs, errCtx)
		} else {
			// All tasks were launched.
			// We will add errCtx later only if some errors happened in tasks.
			errs = append(errs, nil)
		}
	}

	for err := range errChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Add the error that has probably triggered the bad termination of some tasks.
	// If no errors happened, the cancellation had no impact, so don't fail.
	if len(errs) > 1 && errs[0] == nil {
		errs[0] = errCtx
	}

	return joinErrors(errs...)
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
