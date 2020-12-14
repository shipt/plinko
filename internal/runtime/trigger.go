package runtime

import (
	"fmt"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/sideeffects"
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

	td := &sideeffects.TransitionDef{
		Source:      state,
		Destination: destinationState.State,
		Trigger:     trigger,
	}

	defer sideeffects.Dispatch(plinko.AfterTransition, psm.pd.SideEffects, payload, td)

	sideeffects.Dispatch(plinko.BeforeTransition, psm.pd.SideEffects, payload, td)

	payload, err := sd2.Callbacks.ExecuteExitChain(payload, td)

	if err != nil {
		payload, td, errSub := sd2.Callbacks.ExecuteErrorChain(payload, td, err)

		if errSub != nil {
			// this ensures that the error condition is trapped and not overriden to the caller of the trigger function
			err = errSub
		}
		sideeffects.Dispatch(plinko.BetweenStates, psm.pd.SideEffects, payload, td)
		return payload, err
	}

	sideeffects.Dispatch(plinko.BetweenStates, psm.pd.SideEffects, payload, td)

	payload, err = destinationState.Callbacks.ExecuteEntryChain(payload, td)
	if err != nil {
		var errSub error

		payload, mtd, errSub := destinationState.Callbacks.ExecuteErrorChain(payload, td, err)
		td = &sideeffects.TransitionDef{
			Source:      mtd.GetSource(),
			Destination: mtd.GetDestination(),
			Trigger:     mtd.GetTrigger(),
		}

		if errSub != nil {
			err = errSub
		}

		return payload, err
	}

	return payload, nil
}
