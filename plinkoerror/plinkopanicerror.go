package plinkoerror

import (
	"fmt"

	"github.com/shipt/plinko"
)

func CreatePlinkoPanicError(pn interface{}, t plinko.TransitionInfo, step int) error {
	if err, ok := pn.(error); ok {
		return &PlinkoPanicError{
			TransitionInfo: t,
			StepNumber:     step,
			InnerError:     err,
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
}

func (ce *PlinkoPanicError) Error() string {
	return fmt.Sprintf("%+v", *ce)
}
