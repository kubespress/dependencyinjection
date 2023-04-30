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
	"runtime"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kubespress/dependencyinjection"
)

var _ = Describe("ValueProvider", func() {
	type Type1 string

	var (
		ctx      context.Context
		provider dependencyinjection.Provider[Type1]
		value    Type1 = "value1"
	)

	BeforeEach(func() {
		ctx = context.Background()
		provider = dependencyinjection.NewValueProvider(value)
	})

	It("should resolve the provided value", func() {
		resolved, err := provider.Resolve(ctx, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(resolved).To(Equal(value))
	})
})

var _ = Describe("FactoryProvider", func() {
	type Type1 string

	var (
		ctx      context.Context
		provider dependencyinjection.Provider[Type1]
		value    Type1 = "value1"
		calls    uint32
	)

	BeforeEach(func() {
		ctx = context.Background()
		calls = 0
		provider = dependencyinjection.NewFactoryProvider(func(ctx context.Context, c *dependencyinjection.Container) (Type1, error) {
			atomic.AddUint32(&calls, 1)
			return value, nil
		})
	})

	It("should resolve the provided value", func() {
		resolved, err := provider.Resolve(ctx, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(resolved).To(Equal(value))
	})

	It("should call the constructor every resolve", func() {
		resolved, err := provider.Resolve(ctx, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(resolved).To(Equal(value))
		Expect(calls).To(BeEquivalentTo(1))

		resolved, err = provider.Resolve(ctx, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(resolved).To(Equal(value))
		Expect(calls).To(BeEquivalentTo(2))
	})
})

var _ = Describe("SingletonProvider", func() {
	type Type1 string

	var (
		ctx      context.Context
		provider dependencyinjection.Provider[Type1]
		value    Type1 = "value1"
		calls    uint32
	)

	BeforeEach(func() {
		ctx = context.Background()
		calls = 0
		provider = dependencyinjection.NewSingletonProvider(func(ctx context.Context, c *dependencyinjection.Container) (Type1, error) {
			atomic.AddUint32(&calls, 1)
			return value, nil
		})
	})

	It("should resolve the provided value", func() {
		resolved, err := provider.Resolve(ctx, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(resolved).To(Equal(value))
	})

	It("should call the constructor once only", func() {
		resolved, err := provider.Resolve(ctx, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(resolved).To(Equal(value))
		Expect(calls).To(BeEquivalentTo(1))

		resolved, err = provider.Resolve(ctx, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(resolved).To(Equal(value))
		Expect(calls).To(BeEquivalentTo(1))
	})

	Context("with multiple concurrent calls", func() {
		var (
			resolver chan Type1
		)

		BeforeEach(func() {
			resolver = make(chan Type1)
			provider = dependencyinjection.NewSingletonProvider(func(ctx context.Context, c *dependencyinjection.Container) (Type1, error) {
				atomic.AddUint32(&calls, 1)
				return <-resolver, nil
			})
		})

		It("the constructor should run once", func() {
			var (
				expected = Type1("foo")
				result1  = make(chan Type1)
				result2  = make(chan Type1)
			)

			go func() {
				result, err := provider.Resolve(ctx, nil)
				Expect(err).NotTo(HaveOccurred())
				result1 <- result
			}()

			go func() {
				result, err := provider.Resolve(ctx, nil)
				Expect(err).NotTo(HaveOccurred())
				result2 <- result
			}()

			go func() {
				runtime.Gosched()
				resolver <- expected
			}()

			Eventually(result1, "10s").Should(Receive(&expected))
			Eventually(result2, "10s").Should(Receive(&expected))
			Expect(calls).To(BeEquivalentTo(1))
		})
	})
})
