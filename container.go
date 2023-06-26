/*
Copyright 2023 Kubespress Authors.

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

package dependencyinjection

import (
	"context"
	"reflect"
	"sync"

	"github.com/kubespress/errors"
)

// Global is the default global dependency container
var Global = &Container{}

// Container holds all dependencies, resolving them as required
type Container struct {
	mu        sync.RWMutex
	providers map[reflect.Type]any
}

// Register registers a new value provider in the dependency injection container.
func Register[T any](container *Container, provider Provider[T]) error {
	// Lock the map
	container.mu.Lock()
	defer container.mu.Unlock()

	// Get type
	typ := typeof[T]()

	// Initialize the map if required
	if container.providers == nil {
		container.providers = make(map[reflect.Type]any)
	}

	// If the provider already exists, return an error
	if _, ok := container.providers[typ]; ok {
		return ErrDependencyAlreadyRegistered{typ}
	}

	// Add the provider
	container.providers[typ] = provider

	// Return no error
	return nil
}

// MustRegister registers a new value provider in the dependency injection container. The method panics on error.
func MustRegister[T any](container *Container, provider Provider[T]) {
	must(Register(container, provider))
}

// Get retrieves an object from the container, constructing it if necessary
func Get[T any](ctx context.Context, container *Container) (T, error) {
	// Lock the map
	container.mu.RLock()
	defer container.mu.RUnlock()

	// Get type
	typ := typeof[T]()

	// Detect dependency cycles
	ctx, err := cycleDetect(ctx, typ)
	if err != nil {
		return *new(T), err
	}

	// Resolve the value
	if r, exists := container.providers[typ]; exists {
		return r.(Provider[T]).Resolve(ctx, container)
	}

	// Unknown type
	return *new(T), ErrDependencyNotRegistered{typ}
}

// Inject gets a dependency into a pointer receiver
func Inject[T any](ctx context.Context, container *Container, into *T) (err error) {
	if into == nil {
		return errors.Enrich(ErrNilReceiver{}, errors.WithStack())
	}

	*into, err = Get[T](ctx, container)
	return err
}

func typeof[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// cycleDetect detects dependency cycles by adding a value to the context marking what types have been requested, if a
// type is requested twice, an error is returned
func cycleDetect(ctx context.Context, typ reflect.Type) (context.Context, error) {
	// Define marker used to detect dependency cycles
	type cycleMarker struct {
		typ reflect.Type
	}

	// Detect dependency cycles
	if ctx.Value(cycleMarker{typ}) != nil {
		return nil, ErrDependencyCycle{typ}
	}

	// Return context with no error
	return context.WithValue(ctx, cycleMarker{typ}, struct{}{}), nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
