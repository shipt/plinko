package plinkoerror

import "github.com/shipt/plinko"

type PlinkoStateError struct {
	plinko.State
	ErrorMessage string
}

func (e *PlinkoStateError) Error() string {
	return e.ErrorMessage
}

func CreatePlinkoStateError(state plinko.State, errorMessage string) error {
	return &PlinkoStateError{
		State:        state,
		ErrorMessage: errorMessage,
	}
}
