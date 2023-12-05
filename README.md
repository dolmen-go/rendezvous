# rendezvous - Synchronization utilities for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/dolmen-go/rendezvous.svg)](https://pkg.go.dev/github.com/dolmen-go/rendezvous)
[![CI](https://github.com/dolmen-go/rendezvous/actions/workflows/go.yml/badge.svg)](https://github.com/dolmen-go/rendezvous/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/dolmen-go/rendezvous)](https://goreportcard.com/report/github.com/dolmen-go/rendezvous)

## Status

Production ready. 100% code coverage.

## Features

* Never leak goroutines: all functions will wait until the termination of all launched goroutines, whatever happen.

* Panics in goroutines are caught and propagated as errors.

* Task cancellation and timeout via [context.Context](https://pkg.go.dev/context#Context).

## API

* [`type Task = func() error`](https://pkg.go.dev/github.com/dolmen-go/rendezvous#Task)

* [`WaitAll(...Task) []error`](https://pkg.go.dev/github.com/dolmen-go/rendezvous#WaitAll)

* [`type TaskCtx = func(context.Context) error`](https://pkg.go.dev/github.com/dolmen-go/rendezvous#TaskCtx)

* [`WaitFirstError(...TaskCtx) error`](https://pkg.go.dev/github.com/dolmen-go/rendezvous#WaitFirstError)

## See also

* [Go issue #57534](https://github.com/golang/go/issues/57534)

* [`golang.org/x/sync/errgroup.WithContext`](https://pkg.go.dev/golang.org/x/sync/errgroup#WithContext)

## License

Copyright 2019-2023 Olivier Mengué

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.