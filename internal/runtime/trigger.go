package runtime

import (
	"fmt"

	"github.com/shipt/plinko/internal/sideeffects"
	"github.com/shipt/plinko/pkg/plinko"
)

func (psm plinkoStateMachine) EnumerateActiveTriggers(payload plinko.Payload) ([]plinko.Trigger, error) {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return nil, fmt.Errorf("State %s not found in state machine definition", state)
	}

	keys := make([]plinko.Trigger, 0, len(sd2.Triggers))
	for k := range sd2.Triggers {
		keys = append(keys, k)
	}

	return keys, nil

}

func (psm plinkoStateMachine) CanFire(payload plinko.Payload, trigger plinko.Trigger) bool {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return false
	}

	triggerData := sd2.Triggers[trigger]
	if triggerData == nil {
		return false
	}

	if triggerData.Predicate != nil {
		return triggerData.Predicate(payload, sideeffects.TransitionDef{
			Destination: triggerData.DestinationState,
			Source:      state,
			Trigger:     triggerData.Name,
		})
	}

	return true
}

func (psm plinkoStateMachine) Fire(payload plinko.Payload, trigger plinko.Trigger) (plinko.Payload, error) {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return payload, fmt.Errorf("State not found in definition of states: %s", state)
	}

	triggerData := sd2.Triggers[trigger]
	if triggerData == nil {
		return payload, fmt.Errorf("Trigger '%s' not found in definition for state: %s", trigger, state)
	}

	destinationState := (*psm.pd.States)[triggerData.DestinationState]

	td := sideeffects.TransitionDef{
		Source:      state,
		Destination: destinationState.State,
		Trigger:     trigger,
	}

	sideeffects.Dispatch(plinko.BeforeTransition, psm.pd.SideEffects, payload, td)

	sd2.Callbacks.ExecuteExitChain(payload, td)

	sideeffects.Dispatch(plinko.BetweenStates, psm.pd.SideEffects, payload, td)

	destinationState.Callbacks.ExecuteEntryChain(payload, td)

	sideeffects.Dispatch(plinko.AfterTransition, psm.pd.SideEffects, payload, td)

	return payload, nil
}
