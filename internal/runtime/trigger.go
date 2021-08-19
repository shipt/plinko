/**
 * Copyright (c) Shipt.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package runtime

import (
	"context"
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

func (psm plinkoStateMachine) CanFire(ctx context.Context, payload plinko.Payload, trigger plinko.Trigger) error {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return plinkoerror.CreatePlinkoStateError(state, fmt.Sprintf("State '%s' not defined", state))
	}

	triggerData := sd2.Triggers[trigger]
	if triggerData == nil {
		return plinkoerror.CreatePlinkoTriggerError(trigger, fmt.Sprintf("Triggers '%s' not defined for state '%s'", trigger, state))
	}

	if triggerData.Predicate != nil {
		return triggerData.Predicate(ctx, payload, sideeffects.TransitionDef{
			Destination: triggerData.DestinationState,
			Source:      state,
			Trigger:     triggerData.Name,
		})
	}

	return nil
}

func (psm plinkoStateMachine) Fire(ctx context.Context, payload plinko.Payload, trigger plinko.Trigger) (plinko.Payload, error) {
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

	if triggerData.Predicate != nil {
		if err := triggerData.Predicate(ctx, payload, td); err != nil {
			return payload, plinkoerror.CreatePlinkoTriggerError(trigger, fmt.Sprintf("Conditional Trigger '%s' conditions not met for state: %s", trigger, state))
		}
	}

	sideeffects.Dispatch(ctx, plinko.BeforeTransition, psm.pd.SideEffects, payload, td, time.Since(start).Milliseconds())

	payload, err := sd2.Callbacks.ExecuteExitChain(ctx, payload, td)

	if err != nil {
		payload, td, errSub := sd2.Callbacks.ExecuteErrorChain(ctx, payload, td, err, time.Since(start).Milliseconds())

		if errSub != nil {
			// this ensures that the error condition is trapped and not overriden to the caller of the trigger function
			err = errSub
		}
		sideeffects.Dispatch(ctx, plinko.BetweenStates, psm.pd.SideEffects, payload, td, time.Since(start).Milliseconds())
		return payload, err
	}

	sideeffects.Dispatch(ctx, plinko.BetweenStates, psm.pd.SideEffects, payload, td, time.Since(start).Milliseconds())

	payload, err = destinationState.Callbacks.ExecuteEntryChain(ctx, payload, td)
	if err != nil {
		var errSub error

		payload, mtd, errSub := destinationState.Callbacks.ExecuteErrorChain(ctx, payload, td, err, time.Since(start).Milliseconds())
		_ = &sideeffects.TransitionDef{
			Source:      mtd.GetSource(),
			Destination: mtd.GetDestination(),
			Trigger:     mtd.GetTrigger(),
		}

		if errSub != nil {
			err = errSub
		}

		return payload, err
	}

	sideeffects.Dispatch(ctx, plinko.AfterTransition, psm.pd.SideEffects, payload, td, time.Since(start).Milliseconds())

	return payload, nil
}
