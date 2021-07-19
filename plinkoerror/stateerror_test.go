// Copyright 2021, Shipt. All rights reserved.
// Licensed under the Apache License
package plinkoerror

import (
	"errors"
	"testing"

	"github.com/shipt/plinko"
	"github.com/stretchr/testify/assert"
)

func TestCreatePlinkoStateError(t *testing.T) {
	var e *PlinkoStateError
	err := CreatePlinkoStateError("foo", "set")

	if errors.As(err, &e) {
		assert.Equal(t, plinko.State("foo"), e.State)
		assert.Equal(t, "set", e.Error())
	} else {
		assert.Fail(t, "error not returning properly")
	}

}
