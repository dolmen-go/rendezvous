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
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/dolmen-go/rendezvous"
)

func ExampleTaskValueCtx() {
	aChan, aTask := rendezvous.TaskValueCtx(func(ctx context.Context) (int, error) {
		return 42, nil
	})
	bChan, bTask := rendezvous.TaskValueCtx(func(ctx context.Context) (string, error) {
		return "ok", nil
	})
	err := rendezvous.WaitFirstError(context.Background(), aTask, bTask)
	if err != nil {
		log.Fatal(err)
	}
	a, b := <-aChan, <-bChan
	fmt.Println(a, b)
	// Output:
	// 42 ok
}

func ExampleTaskValue() {
	aChan, aTask := rendezvous.TaskValue(func() (int, error) {
		return 42, nil
	})
	bChan, bTask := rendezvous.TaskValue(func() (string, error) {
		return "ok", nil
	})
	errs := rendezvous.WaitAll(aTask, bTask)
	if errs != nil {
		log.Fatal(errs)
	}
	a, b := <-aChan, <-bChan
	fmt.Println(a, b)
	// Output:
	// 42 ok
}

func TestTaskValueCtxFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	waitStart := make(chan struct{})

	aChan, makeA := rendezvous.TaskValueCtx(func(ctx context.Context) (int, error) {
		waitStart <- struct{}{}
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("makeA failure: %w", ctx.Err())
		case <-time.After(100 * time.Millisecond):
			return 42, nil
		}
	})

	bChan, makeB := rendezvous.TaskValueCtx(func(ctx context.Context) (string, error) {
		return "hello", nil
	})

	go func() {
		<-waitStart
		cancel()
	}()

	err := rendezvous.WaitFirstError(ctx, makeA, makeB)

	if err == nil {
		t.Fatal("error expected")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error context.Canceled expected, got %q", err)
	}

	_, ok := <-aChan

	if ok {
		t.Fatal("aChan should be empty because aChan has failed")
	}

	b, ok := <-bChan
	if !ok {
		t.Fatal("bChan should not be empty")
	}
	if b != "hello" {
		t.Error("b value expected")
	}
}
