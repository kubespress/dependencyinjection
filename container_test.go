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

package dependencyinjection_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kubespress/dependencyinjection"
)

type MockConstructor[T any] struct {
	value  T
	called bool
}

func NewMockConstructor[T any](value T) MockConstructor[T] {
	return MockConstructor[T]{value: value}
}

func (f *MockConstructor[T]) Fn(context.Context, *dependencyinjection.Container) (T, error) {
	f.called = true
	return f.value, nil
}

func (f *MockConstructor[T]) Called() bool {
	return f.called
}

var _ = Describe("Register", func() {
	type Type1 string

	var (
		container   *dependencyinjection.Container
		ctx         context.Context
		dependency1 Type1 = "dep1"
	)

	BeforeEach(func() {
		container = &dependencyinjection.Container{}
		ctx = context.Background()
	})

	Context("with new type", func() {
		var (
			err error
		)

		BeforeEach(func() {
			err = dependencyinjection.Register(ctx, container, dependencyinjection.NewValueProvider(dependency1))
		})

		It("should return no error", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("with type already registered", func() {
		var (
			err error
		)

		BeforeEach(func() {
			err = dependencyinjection.Register(ctx, container, dependencyinjection.NewValueProvider(dependency1))
			Expect(err).ShouldNot(HaveOccurred())
			err = dependencyinjection.Register(ctx, container, dependencyinjection.NewValueProvider(dependency1))
		})

		It("should return no error", func() {
			Expect(err).Should(HaveOccurred())
			Expect(err).Should(MatchError("type already registered with container: dependencyinjection_test.Type1"))
		})
	})
})

var _ = Describe("Get", func() {
	type Type1 string
	type Type2 string

	var (
		container   *dependencyinjection.Container
		ctx         context.Context
		dependency1 Type1 = "dep1"
		dependency2 Type2 = "dep2"
	)

	BeforeEach(func() {
		container = &dependencyinjection.Container{}
		ctx = context.Background()
	})

	Context("with known type", func() {
		BeforeEach(func() {
			err := dependencyinjection.Register(ctx, container, dependencyinjection.NewValueProvider(dependency1))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should return no error", func() {
			value, err := dependencyinjection.Get[Type1](ctx, container)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(value).To(Equal(dependency1))
		})
	})

	Context("with unknown type", func() {
		It("should return an error", func() {
			_, err := dependencyinjection.Get[Type1](ctx, container)
			Expect(err).Should(HaveOccurred())
			Expect(err).Should(MatchError("type not registered with container: dependencyinjection_test.Type1"))
		})
	})

	Context("with dependency cycle", func() {
		BeforeEach(func() {
			err := dependencyinjection.Register(ctx, container, dependencyinjection.NewFactoryProvider(func(ctx context.Context, c *dependencyinjection.Container) (Type1, error) {
				if _, err := dependencyinjection.Get[Type2](ctx, c); err != nil {
					return "", err
				}
				return dependency1, nil
			}))
			Expect(err).ShouldNot(HaveOccurred())

			err = dependencyinjection.Register(ctx, container, dependencyinjection.NewFactoryProvider(func(ctx context.Context, c *dependencyinjection.Container) (Type2, error) {
				if _, err := dependencyinjection.Get[Type1](ctx, c); err != nil {
					return "", err
				}
				return dependency2, nil
			}))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should return an error", func() {
			_, err := dependencyinjection.Get[Type1](ctx, container)
			Expect(err).Should(HaveOccurred())
			Expect(err).Should(MatchError("could not resolve dependency dependencyinjection_test.Type1: could not resolve dependency dependencyinjection_test.Type2: dependency cycle detected: dependencyinjection_test.Type1"))
		})
	})
})

// var _ = Describe("A dependency container", func() {
// 	type Type1 string
// 	type Type2 string

// 	var (
// 		container *dependencyinjection.Container
// 		ctx       context.Context
// 	)

// 	BeforeEach(func() {
// 		container = &dependencyinjection.Container{}
// 		ctx = context.Background()
// 	})

// 	When("a value is added to the container", func() {
// 		var expected Type1

// 		BeforeEach(func() {
// 			expected = Type1("foo")
// 			err := dependencyinjection.Add(ctx, container, expected)
// 			Expect(err).ToNot(HaveOccurred())
// 		})

// 		It("should be resolvable within the container", func() {
// 			result, err := dependencyinjection.Get[Type1](ctx, container)
// 			Expect(err).ToNot(HaveOccurred())
// 			Expect(result).To(Equal(expected))
// 		})
// 	})

// 	When("a constructor is added to the container", func() {
// 		var expected Type1

// 		BeforeEach(func() {
// 			expected = Type1("foo")
// 			err := dependencyinjection.Register(ctx, container, func(ctx context.Context, c *dependencyinjection.Container) (Type1, error) {
// 				return expected, nil
// 			})
// 			Expect(err).ToNot(HaveOccurred())
// 		})

// 		It("should be resolvable within the container", func() {
// 			result, err := dependencyinjection.Get[Type1](ctx, container)
// 			Expect(err).ToNot(HaveOccurred())
// 			Expect(result).To(Equal(expected))
// 		})
// 	})

// 	When("a dependency cycle exists", func() {
// 		BeforeEach(func() {
// 			err := dependencyinjection.Register(ctx, container, func(ctx context.Context, c *dependencyinjection.Container) (Type1, error) {
// 				_, err := dependencyinjection.Get[Type2](ctx, c)
// 				return Type1("foo"), err
// 			})
// 			Expect(err).ToNot(HaveOccurred())

// 			err = dependencyinjection.Register(ctx, container, func(ctx context.Context, c *dependencyinjection.Container) (Type2, error) {
// 				_, err := dependencyinjection.Get[Type1](ctx, c)
// 				return Type2("foo"), err
// 			})
// 			Expect(err).ToNot(HaveOccurred())
// 		})

// 		It("should return an error when resolved", func() {
// 			_, err := dependencyinjection.Get[Type1](ctx, container)
// 			Expect(err).To(HaveOccurred())
// 			Expect(err).To(MatchError("could not resolve dependency dependencyinjection_test.Type1: could not resolve dependency dependencyinjection_test.Type2: dependency cycle detected: dependencyinjection_test.Type1"))
// 		})
// 	})

// 	When("multiple go-routines are waiting on the same resolver", func() {
// 		var (
// 			resolver  chan Type1
// 			callCount int32
// 		)

// 		BeforeEach(func() {
// 			resolver = make(chan Type1)
// 			callCount = 0

// 			err := dependencyinjection.Register(ctx, container, func(ctx context.Context, c *dependencyinjection.Container) (Type1, error) {
// 				atomic.AddInt32(&callCount, 1)
// 				return <-resolver, nil
// 			})
// 			Expect(err).ToNot(HaveOccurred())
// 		})

// 		It("the constructor should run once", func() {
// 			var (
// 				expected = Type1("foo")
// 				result1  = make(chan Type1)
// 				result2  = make(chan Type1)
// 			)

// 			go func() {
// 				result, err := dependencyinjection.Get[Type1](ctx, container)
// 				Expect(err).NotTo(HaveOccurred())
// 				result1 <- result
// 			}()

// 			go func() {
// 				result, err := dependencyinjection.Get[Type1](ctx, container)
// 				Expect(err).NotTo(HaveOccurred())
// 				result2 <- result
// 			}()

// 			go func() {
// 				runtime.Gosched()
// 				resolver <- expected
// 			}()

// 			Eventually(result1, "10s").Should(Receive(&expected))
// 			Eventually(result2, "10s").Should(Receive(&expected))
// 			Expect(callCount).To(Equal(int32(1)))
// 		})
// 	})

// 	When("a type is added twice to the container", func() {
// 		It("should return an error", func() {
// 			err := dependencyinjection.Add(ctx, container, Type1("foo"))
// 			Expect(err).ToNot(HaveOccurred())
// 			err = dependencyinjection.Add(ctx, container, Type1("bar"))
// 			Expect(err).To(HaveOccurred())
// 			Expect(err).To(MatchError("type already registered with container: dependencyinjection_test.Type1"))
// 		})
// 	})

// 	When("a type not added to the container is requested", func() {
// 		It("should return an error", func() {
// 			err := dependencyinjection.Add(ctx, container, Type1("foo"))
// 			Expect(err).ToNot(HaveOccurred())
// 			_, err = dependencyinjection.Get[Type2](ctx, container)
// 			Expect(err).To(HaveOccurred())
// 			Expect(err).To(MatchError("type not registered with container: dependencyinjection_test.Type2"))
// 		})
// 	})
// })
