package runtime

import (
	"fmt"
	"time"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/sideeffects"
	"github.com/shipt/plinko/plinkoerror"
)

func (psm plinkoStateMachine) EnumerateActiveTriggers(payload plinko.Payload) ([]plinko.Trigger, error) {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return nil, plinkoerror.CreatePlinkoStateError(state, fmt.Sprintf("State %s not found in state machine definition", state))
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
	start := time.Now()
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return payload, plinkoerror.CreatePlinkoStateError(state, fmt.Sprintf("State not found in definition of states: %s", state))
	}

	triggerData := sd2.Triggers[trigger]
	if triggerData == nil {
		return payload, plinkoerror.CreatePlinkoTriggerError(trigger, fmt.Sprintf("Trigger '%s' not found in definition for state: %s", trigger, state))
	}

	destinationState := (*psm.pd.States)[triggerData.DestinationState]

	td := &sideeffects.TransitionDef{
		Source:      state,
		Destination: destinationState.State,
		Trigger:     trigger,
	}

	sideeffects.Dispatch(plinko.BeforeTransition, psm.pd.SideEffects, payload, td, time.Since(start).Milliseconds())

	payload, err := sd2.Callbacks.ExecuteExitChain(payload, td)

	if err != nil {
		payload, td, errSub := sd2.Callbacks.ExecuteErrorChain(payload, td, err, time.Since(start).Milliseconds())

		if errSub != nil {
			// this ensures that the error condition is trapped and not overriden to the caller of the trigger function
			err = errSub
		}
		sideeffects.Dispatch(plinko.BetweenStates, psm.pd.SideEffects, payload, td, time.Since(start).Milliseconds())
		return payload, err
	}

	sideeffects.Dispatch(plinko.BetweenStates, psm.pd.SideEffects, payload, td, time.Since(start).Milliseconds())

	payload, err = destinationState.Callbacks.ExecuteEntryChain(payload, td)
	if err != nil {
		var errSub error

		payload, mtd, errSub := destinationState.Callbacks.ExecuteErrorChain(payload, td, err, time.Since(start).Milliseconds())
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

	sideeffects.Dispatch(plinko.AfterTransition, psm.pd.SideEffects, payload, td, time.Since(start).Milliseconds())

	return payload, nil
}
