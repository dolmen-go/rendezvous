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

package rendezvous

import "context"

// TaskValueCtx wraps a function that returns a value and an error to make it a [TaskCtx]
// for [WaitFirstError]. The value is returned via a channel, so the values of successful
// tasks can be retrieved even if one task fails by checking for each task's channel if it
// has a value.
func TaskValueCtx[T any](fetchT func(context.Context) (T, error)) (<-chan T, TaskCtx) {
	ch := make(chan T, 1)
	return ch, func(ctx context.Context) error {
		defer close(ch)
		v, err := fetchT(ctx)
		if err == nil {
			ch <- v
		}
		return err
	}
}

// TaskValue wraps a function that returns a value and an error to make it a [Task]
// for [WaitAll]. The value is returned via a channel, so the values of successful
// tasks can be retrieved even if one task fails by checking for each task's channel if it
// has a value.
func TaskValue[T any](makeT func() (T, error)) (<-chan T, Task) {
	ch := make(chan T, 1)
	return ch, func() error {
		defer close(ch)
		v, err := makeT()
		if err == nil {
			ch <- v
		}
		return err
	}
}
