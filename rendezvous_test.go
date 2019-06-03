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
	return nil
}

func withNilPanic() error {
	panic(nil)
	return nil
}

func withStringPanic() error {
	panic("OK")
	return nil
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

	checkNil(t, rendezvous.WaitAll(withNilPanic))

	checkEquals(t, rendezvous.WaitAll(withPanic), []error{myErr})

	errs := rendezvous.WaitAll(withStringPanic)
	if len(errs) != 1 || errs[0] == nil || errs[0].Error() != "panic: OK" {
		t.Errorf("got %v", errs)
	}
}

func TestTwo(t *testing.T) {
	t.Parallel()

	checkNil(t, rendezvous.WaitAll(noError, noError))
	checkNil(t, rendezvous.WaitAll(noError, withNilPanic))
	checkNil(t, rendezvous.WaitAll(withNilPanic, noError))

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
	var funcs []rendezvous.Func
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
		funcs = append(funcs, f)
	}

	errs := rendezvous.WaitAll(funcs...)
	if len(errs) != countErrors {
		t.Errorf("got: %v", errs)
	}
}

func TestContextNone(t *testing.T) {
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

func TestContextErr(t *testing.T) {
	t.Parallel()

	var (
		err1 = errors.New("error 1")
		err2 = errors.New("error 2")
		ctx  = context.Background()
	)

	err := rendezvous.WaitFirstError(ctx,
		func(ctx context.Context) error {
			return err1
		},
	)
	if !errors.Is(err, err1) {
		t.Error("err1 expected")
	}

	err = rendezvous.WaitFirstError(ctx,
		func(ctx context.Context) error {
			return err1
		},
		func(ctx context.Context) error {
			return nil
		},
	)
	if !errors.Is(err, err1) {
		t.Error("err1 expected")
	}

	err = rendezvous.WaitFirstError(ctx,
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return err2
		},
	)
	if !errors.Is(err, err2) {
		t.Error("err2 expected")
	}

	err = rendezvous.WaitFirstError(ctx,
		func(ctx context.Context) error {
			return err1
		},
		func(ctx context.Context) error {
			return err2
		},
	)
	if !errors.Is(err, err1) && !errors.Is(err, err2) {
		t.Error("both err1 and err2 expected")
	}
}

func TestContextCanceled(t *testing.T) {
	t.Parallel()

	var (
		err1 = errors.New("error 1")
		err2 = errors.New("error 2")
	)

	t.Run("no-error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())

		err := rendezvous.WaitFirstError(ctx,
			func(_ context.Context) error {
				return nil
			},
		)
		if err != nil {
			t.Error("nil expected")
		}

		cancel()
	})

	t.Run("before-start", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())

		// Cancel before starting
		cancel()

		err := rendezvous.WaitFirstError(ctx,
			func(_ context.Context) error {
				panic(errors.New("this should not happen because the task should not even start"))
			},
		)
		t.Logf("%q", err)
		if !errors.Is(err, context.Canceled) {
			t.Error("context.Canceled expected because we have not started any task")
		}
	})

	t.Run("canceled-error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())

		ch := make(chan struct{})
		go func() {
			// Wait that the goroutine starts...
			<-ch
			cancel()
		}()

		err := rendezvous.WaitFirstError(ctx,
			func(_ context.Context) error {
				close(ch)
				<-ctx.Done()
				return ctx.Err()
			},
		)
		if !errors.Is(err, context.Canceled) {
			t.Error("context.Canceled expected")
		}
	})

	t.Run("1-error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())

		ch := make(chan struct{})
		go func() {
			// Wait for the FuncCtx below to start...
			<-ch
			cancel()
		}()

		err := rendezvous.WaitFirstError(ctx,
			func(_ context.Context) error {
				close(ch)
				<-ctx.Done()
				return err1
			},
		)
		t.Logf("%q", err)
		// both errors are expected
		if !errors.Is(err, context.Canceled) {
			t.Error("context.Canceled expected because the parent context is canceled")
		}
		if !errors.Is(err, err1) {
			t.Error("err1 expected")
		}
	})

	t.Run("2-errors", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())

		ch := make(chan struct{})
		go func() {
			// Wait for the FuncCtx below to start...
			<-ch
			cancel()
		}()
		err := rendezvous.WaitFirstError(ctx,
			func(childCtx context.Context) error {
				close(ch)
				<-ctx.Done()
				return err1
			},
			func(_ context.Context) error {
				return err2
			},
		)
		t.Logf("%q", err)
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Error("context is supposed to be canceled")
		}
		// both errors are expected
		if !errors.Is(err, context.Canceled) {
			t.Error("context.Canceled expected because the parent context is canceled")
		}
		if !errors.Is(err, err1) {
			t.Error("err1 expected")
		}
		// func2 may have not even been started
		if !errors.Is(err, err2) {
			t.Log("f2 not started")
		}

	})
}
