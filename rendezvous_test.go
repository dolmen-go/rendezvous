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

package rendezvous_test

import (
	"context"
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/dolmen-go/rendezvous"
)

func checkNil(t *testing.T, errs []error) {
	if errs != nil {
		t.Error("nil []error expected")
	}
}

func checkEquals(t *testing.T, got []error, expected []error) {
	if reflect.DeepEqual(got, expected) {
		return
	}
	if len(got) != len(expected) {
		t.Errorf("count of errors mismatch: %d got, %d expected", len(got), len(expected))
	}
	for i, exp := range expected {
		if got[i] == exp {
			continue
		}
		typeGot := reflect.TypeOf(got[i])
		typeExp := reflect.TypeOf(exp)
		if typeGot != typeExp {
			t.Errorf("error %d: got %v, expecting %v", i, got[i], exp)
			return
		}
	}
	t.Errorf("got %v", got)
}

func noError() error {
	return nil
}

var myErr = errors.New("my error")

func withError() error {
	return myErr
}

func withPanic() error {
	panic(myErr)
}

func withStringPanic() error {
	panic("OK")
}

func withDelay(d time.Duration, f func() error) func() error {
	return func() error {
		time.Sleep(d)
		return f()
	}
}

func TestNone(t *testing.T) {
	t.Parallel()

	checkNil(t, rendezvous.WaitAll())

	checkNil(t, rendezvous.WaitAll(nil))
}

func TestOne(t *testing.T) {
	t.Parallel()

	checkNil(t, rendezvous.WaitAll(noError))

	checkEquals(t, rendezvous.WaitAll(withError), []error{myErr})

	checkEquals(t, rendezvous.WaitAll(withPanic), []error{myErr})

	errs := rendezvous.WaitAll(withStringPanic)
	if len(errs) != 1 || errs[0] == nil || errs[0].Error() != "panic: OK" {
		t.Errorf("got %v", errs)
	}
}

func TestTwo(t *testing.T) {
	t.Parallel()

	checkNil(t, rendezvous.WaitAll(noError, noError))

	checkEquals(t, rendezvous.WaitAll(withError, withError), []error{myErr, myErr})

	checkEquals(t, rendezvous.WaitAll(noError, withError), []error{myErr})
	checkEquals(t, rendezvous.WaitAll(withError, noError), []error{myErr})

	checkEquals(t, rendezvous.WaitAll(noError, withPanic), []error{myErr})
	checkEquals(t, rendezvous.WaitAll(withPanic, noError), []error{myErr})
}

func TestManyRandom(t *testing.T) {
	t.Parallel()

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := rand.Intn(50)
	var tasks []rendezvous.Task
	var countErrors int
	for i := 0; i < n; i++ {
		var f func() error
		if rand.Intn(2) == 1 {
			f = withError
			countErrors++
		} else {
			f = noError
		}
		f = withDelay(time.Duration(rand.Intn(500))*time.Millisecond, f)
		tasks = append(tasks, f)
	}

	errs := rendezvous.WaitAll(tasks...)
	if len(errs) != countErrors {
		t.Errorf("got: %v", errs)
	}
}

func TestContextNone(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	if rendezvous.WaitFirstError(ctx) != nil {
		t.Error("nil expected")
	}
	if rendezvous.WaitFirstError(ctx, nil) != nil {
		t.Error("nil expected")
	}
	if rendezvous.WaitFirstError(ctx, nil, nil) != nil {
		t.Error("nil expected")
	}
}

func TestContextOK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	err := rendezvous.WaitFirstError(ctx,
		func(ctx context.Context) error {
			return nil
		},
	)
	if err != nil {
		t.Error("nil expected")
	}

	err = rendezvous.WaitFirstError(ctx,
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
	)
	if err != nil {
		t.Error("nil expected")
	}
}
