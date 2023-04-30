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

// Constructor creates a new instance of an object
type Constructor[T any] func(context.Context, *Container) (T, error)

// Provider provides a value of a given type.
type Provider[T any] interface {
	Resolve(context.Context, *Container) (T, error)
}

type valueProvider[T any] struct {
	value T
}

// NewValueProvider returns a provider that will always return the given value. This allows objects to be directly added
// to the Container without needing a constructor.
func NewValueProvider[T any](value T) Provider[T] {
	return &valueProvider[T]{value: value}
}

// RegisterValue adds a value to the dependency injection container. It is equivalent to calling
// `Register(ctx, container, NewValueProvider(value))`
func RegisterValue[T any](ctx context.Context, container *Container, value T) error {
	return Register(ctx, container, NewValueProvider(value))
}

// MustRegisterValue adds a value to the dependency injection container. The method panics on error.
func MustRegisterValue[T any](ctx context.Context, container *Container, value T) {
	must(RegisterValue(ctx, container, value))
}

func (p *valueProvider[T]) Resolve(context.Context, *Container) (T, error) {
	return p.value, nil
}

type factoryProvider[T any] struct {
	typ reflect.Type
	fn  Constructor[T]
}

// NewFactoryProvider returns a provider that calls the provided constructor to generate the value every time is is
// required.
func NewFactoryProvider[T any](fn Constructor[T]) Provider[T] {
	return &factoryProvider[T]{
		typ: typeof[T](),
		fn:  fn,
	}
}

// RegisterFactory adds a factory to the dependency injection container. It is equivalent to calling
// `Register(ctx, container, NewFactoryProvider(fn))`
func RegisterFactory[T any](ctx context.Context, container *Container, fn Constructor[T]) error {
	return Register(ctx, container, NewFactoryProvider(fn))
}

// MustRegisterFactory adds a factory to the dependency injection container. The method panics on error.
func MustRegisterFactory[T any](ctx context.Context, container *Container, fn Constructor[T]) {
	must(RegisterFactory(ctx, container, fn))
}

func (p *factoryProvider[T]) Resolve(ctx context.Context, container *Container) (T, error) {
	// Call the factory function
	value, err := p.fn(ctx, container)
	if err != nil {
		return value, errors.Enrich(err,
			errors.Wrapf("could not resolve dependency %s", p.typ),
		)
	}

	// Return the value
	return value, nil
}

type singletonProvider[T any] struct {
	cond  sync.Cond
	typ   reflect.Type
	fn    Constructor[T]
	value T
	state singletonProviderState
}

type singletonProviderState uint8

const (
	singletonProviderStateUnresolved singletonProviderState = iota
	singletonProviderStateResolving
	singletonProviderStateResolved
)

// NewSingletonProvider returns a provider that calls the provided constructor once to generate the value and provides
// the same value on subsequent resolves. The provider is thread safe.
func NewSingletonProvider[T any](fn Constructor[T]) Provider[T] {
	return &singletonProvider[T]{
		cond: sync.Cond{L: &sync.Mutex{}},
		typ:  typeof[T](),
		fn:   fn,
	}
}

// RegisterSingleton adds a singleton to the dependency injection container. It is equivalent to calling
// `Register(ctx, container, NewSingletonProvider(fn))`
func RegisterSingleton[T any](ctx context.Context, container *Container, fn Constructor[T]) error {
	return Register(ctx, container, NewSingletonProvider(fn))
}

// MustRegisterSingleton adds a singleton to the dependency injection container. The method panics on error.
func MustRegisterSingleton[T any](ctx context.Context, container *Container, fn Constructor[T]) {
	must(RegisterSingleton(ctx, container, fn))
}

func (p *singletonProvider[T]) Resolve(ctx context.Context, container *Container) (T, error) {
	// Lock the condition
	p.cond.L.Lock()

	// Loop waiting for the resolve to complete
	for {
		switch p.state {
		case singletonProviderStateUnresolved:
			// Update the state
			p.state = singletonProviderStateResolving

			// Unlock now we have updated the state
			p.cond.L.Unlock()

			// Resolve the value
			value, err := p.fn(ctx, container)
			if err != nil {
				// Mark the dependency as unresolved
				p.cond.L.Lock()
				p.state = singletonProviderStateUnresolved
				p.cond.L.Unlock()

				// Alert all waiting resolves so they do not sit blocked
				p.cond.Broadcast()

				// Return the error
				return *new(T), errors.Enrich(err,
					errors.Wrapf("could not resolve dependency %s", p.typ),
				)
			}

			// Update the value and mark as resolved
			p.cond.L.Lock()
			p.state = singletonProviderStateResolved
			p.value = value
			p.cond.L.Unlock()

			// Alert all waiting resolves that the value is now resolved
			p.cond.Broadcast()

			// Return the value
			return value, nil
		case singletonProviderStateResolving:
			p.cond.Wait()
		case singletonProviderStateResolved:
			p.cond.L.Unlock()
			return p.value, nil
		}
	}
}
