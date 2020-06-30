package definitions

import (
	"github.com/shipt/plinko/interfaces"
	"github.com/shipt/plinko/types"
)

type CallbackDefinitions struct {
	OnEntryFn func(pp interfaces.PlinkoPayload, transitionInfo TransitionInfo) (interfaces.PlinkoPayload, error)
	OnExitFn  func(pp interfaces.PlinkoPayload, transitionInfo TransitionInfo) (interfaces.PlinkoPayload, error)
}

type TransitionInfo interface {
	GetSource() types.State
	GetDestination() types.State
	GetTrigger() types.Trigger
}
