package sideeffects

import "github.com/shipt/plinko/pkg/plinko"

// AllowAllSideEffects is a convenience constant for registering a global
const AllowAllSideEffects = plinko.AllowBeforeTransition | plinko.AllowAfterTransition | plinko.AllowBetweenStates

// SideEffectDefinition holds the callback and filtering characteristics describing when the sideeffect is signaled.
type SideEffectDefinition struct {
	SideEffect plinko.SideEffect
	Filter     plinko.SideEffectFilter
}

func getFilterDefinition(stateAction plinko.StateAction) plinko.SideEffectFilter {

	switch stateAction {
	case plinko.BeforeTransition:
		return plinko.AllowBeforeTransition
	case plinko.BetweenStates:
		return plinko.AllowBetweenStates
	case plinko.AfterTransition:
		return plinko.AllowAfterTransition
	}

	return 0
}

// TransitionDef is used to notify the registered function of a transition occuring
type TransitionDef struct {
	Source      plinko.State
	Destination plinko.State
	Trigger     plinko.Trigger
}

// GetSource returns the Source / Starting state
func (td TransitionDef) GetSource() plinko.State {
	return td.Source
}

// GetDestination returns the Destination State that's part of the process being executed.
func (td TransitionDef) GetDestination() plinko.State {
	return td.Destination
}

// GetTrigger returns the Trigger used to launch the transition
func (td TransitionDef) GetTrigger() plinko.Trigger {
	return td.Trigger
}

// Dispatch is responsible for executing a set of side effect definitions when called upon.  It is sensitive to the definition
//   in terms of what is called.
func Dispatch(stateAction plinko.StateAction, sideEffects []SideEffectDefinition, payload plinko.Payload, transitionInfo plinko.TransitionInfo) int {
	iCount := 0
	for _, sideEffectDefinition := range sideEffects {
		if sideEffectDefinition.Filter&getFilterDefinition(stateAction) > 0 {

			sideEffectDefinition.SideEffect(stateAction, payload, transitionInfo)
			iCount++
		}
	}
	return iCount
}