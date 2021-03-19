package plinkoerror

import (
	"fmt"

	"github.com/shipt/plinko"
)

func CreatePlinkoPanicError(pn interface{}, t plinko.TransitionInfo, step int, stack string) error {
	if err, ok := pn.(error); ok {
		return &PlinkoPanicError{
			TransitionInfo: t,
			StepNumber:     step,
			InnerError:     err,
			Stack:          stack,
		}
	}

	return &PlinkoPanicError{
		TransitionInfo:    t,
		StepNumber:        step,
		UnknownInnerError: pn,
	}
}

type PlinkoPanicError struct {
	plinko.TransitionInfo
	StepNumber        int
	InnerError        error
	UnknownInnerError interface{}
	Stack             string
}

func (ce *PlinkoPanicError) Error() string {
	return fmt.Sprintf("%+v", *ce)
}
