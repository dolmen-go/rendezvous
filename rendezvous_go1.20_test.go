// Tests that require errors.Is to work with errors.Join.

//go:build go1.20

package rendezvous_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dolmen-go/rendezvous"
)

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
