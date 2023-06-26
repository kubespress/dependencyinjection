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
	"errors"
	"fmt"
	"reflect"
)

// ErrDependencyNotRegistered is the type of error returned when a requested
// type is not registered with the container
type ErrDependencyNotRegistered struct {
	typ reflect.Type
}

func (e ErrDependencyNotRegistered) Error() string {
	return fmt.Sprintf("type not registered with container: %s", e.typ)
}

// IsDependencyNotRegistered returns true if the error is a
// ErrDependencyNotRegistered
func IsDependencyNotRegistered(err error) bool {
	return errors.As(err, &ErrDependencyNotRegistered{})
}

// ErrDependencyAlreadyRegistered is the type of error returned when type that
// is already registered with a container is re-registered.
type ErrDependencyAlreadyRegistered struct {
	typ reflect.Type
}

func (e ErrDependencyAlreadyRegistered) Error() string {
	return fmt.Sprintf("type already registered with container: %s", e.typ)
}

// IsDependencyAlreadyRegistered returns true if the error is a
// ErrDependencyAlreadyRegistered
func IsDependencyAlreadyRegistered(err error) bool {
	return errors.As(err, &ErrDependencyAlreadyRegistered{})
}

// ErrDependencyAlreadyRegistered is the type of error returned when dependency
// cycle is detected when constructing dependencies
type ErrDependencyCycle struct {
	typ reflect.Type
}

func (e ErrDependencyCycle) Error() string {
	return fmt.Sprintf("dependency cycle detected: %s", e.typ)
}

// IsDependencyCycle returns true if the error is a
// ErrDependencyCycle
func IsDependencyCycle(err error) bool {
	return errors.As(err, &ErrDependencyCycle{})
}

// ErrNilReceiver is the type of error returned when a nil receiver is used in
// the Inject method
type ErrNilReceiver struct{}

func (e ErrNilReceiver) Error() string {
	return "nil receiver for dependency"
}

// IsNilReceiver returns true if the error is a ErrNilReceiver
func IsNilReceiver(err error) bool {
	return errors.As(err, &ErrNilReceiver{})
}
