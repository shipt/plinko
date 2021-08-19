/**
 * Copyright (c) Shipt.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package plinkoerror

import (
	"errors"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicErrorCreate(t *testing.T) {
	a := func(a int) (retErr error) {
		defer func() {
			if err1 := recover(); err1 != nil {
				stack := string(debug.Stack())
				retErr = CreatePlinkoPanicError(err1, nil, 0, "name", stack)
			} else {
				stack := string(debug.Stack())
				retErr = CreatePlinkoPanicError(nil, nil, 0, "name", stack)
			}
		}()

		panic(errors.New("dd"))
	}

	e := a(5)

	assert.NotNil(t, e)

}

func TestPanicErrorCreateWithNilError(t *testing.T) {
	a := func(a int) (retErr error) {
		defer func() {
			if err1 := recover(); err1 != nil {
				stack := string(debug.Stack())
				retErr = CreatePlinkoPanicError(err1, nil, 0, "name", stack)
			} else {
				stack := string(debug.Stack())
				retErr = CreatePlinkoPanicError(nil, nil, 0, "name", stack)
			}
		}()

		panic(nil)
	}

	e := a(5)

	assert.NotNil(t, e)
	assert.Equal(t, "{TransitionInfo:<nil>", e.Error()[:21])

}
