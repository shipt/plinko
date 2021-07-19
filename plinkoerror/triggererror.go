// Copyright 2021, Shipt. All rights reserved.
// Licensed under the Apache License

package plinkoerror

import "github.com/shipt/plinko"

type PlinkoTriggerError struct {
	plinko.Trigger
	ErrorMessage string
}

func (e *PlinkoTriggerError) Error() string {
	return e.ErrorMessage
}

func CreatePlinkoTriggerError(trigger plinko.Trigger, errorMessage string) error {
	return &PlinkoTriggerError{
		Trigger:      trigger,
		ErrorMessage: errorMessage,
	}
}
