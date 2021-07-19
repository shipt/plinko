// Copyright 2021, Shipt. All rights reserved.
// Licensed under the Apache License
package plinkoerror

import (
	"errors"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicErrorCreate(t *testing.T) {
	var a func(a int) error
	a = func(a int) (retErr error) {
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
	var a func(a int) error
	a = func(a int) (retErr error) {
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
