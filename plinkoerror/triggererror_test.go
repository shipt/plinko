// Copyright 2021, Shipt. All rights reserved.
// Licensed under the Apache License
package plinkoerror

import (
	"errors"
	"testing"

	"github.com/shipt/plinko"
	"github.com/stretchr/testify/assert"
)

func TestCreatePlinkoTriggerError(t *testing.T) {
	var e *PlinkoTriggerError
	err := CreatePlinkoTriggerError("foo", "set")

	if errors.As(err, &e) {
		assert.Equal(t, plinko.Trigger("foo"), e.Trigger)
		assert.Equal(t, "set", e.Error())
	} else {
		assert.Fail(t, "error not returning properly")
	}
}
