package plinkoerror

import (
	"fmt"

	"github.com/shipt/plinko"
)

func CreatePlinkoPanicError(pn interface{}, t plinko.TransitionInfo, step int, name string, stack string) error {
	if err, ok := pn.(error); ok {
		return &PlinkoPanicError{
			TransitionInfo: t,
			StepNumber:     step,
			StepName:       name,
			InnerError:     err,
			Stack:          stack,
		}
	}

	return &PlinkoPanicError{
		TransitionInfo:    t,
		StepNumber:        step,
		StepName:          name,
		UnknownInnerError: pn,
		Stack:             stack,
	}
}

type PlinkoPanicError struct {
	plinko.TransitionInfo
	StepNumber        int
	StepName          string
	InnerError        error
	UnknownInnerError interface{}
	Stack             string
}

func (ce *PlinkoPanicError) Error() string {
	return fmt.Sprintf("%+v", *ce)
}
