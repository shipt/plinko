/**
 * Copyright (c) Shipt.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package composition

import (
	"context"
	"runtime/debug"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/sideeffects"
	"github.com/shipt/plinko/plinkoerror"
)

type ChainedFunctionCall struct {
	Predicate plinko.Predicate
	Operation plinko.Operation
	Config    plinko.OperationConfig
}

type ChainedErrorCall struct {
	ErrorOperation plinko.ErrorOperation
	Config         plinko.OperationConfig
}

type CallbackDefinitions struct {
	OnEntryFn []ChainedFunctionCall
	OnExitFn  []ChainedFunctionCall
	OnErrorFn []ChainedErrorCall

	EntryFunctionChain []string
	ExitFunctionChain  []string
}

func (cd *CallbackDefinitions) AddError(errorOperation plinko.ErrorOperation, cfg plinko.OperationConfig) *CallbackDefinitions {
	cd.OnErrorFn = append(cd.OnErrorFn, ChainedErrorCall{
		ErrorOperation: errorOperation,
		Config:         cfg,
	})

	return cd
}

func (cd *CallbackDefinitions) AddEntry(predicate plinko.Predicate, operation plinko.Operation, cfg plinko.OperationConfig) *CallbackDefinitions {

	cd.OnEntryFn = append(cd.OnEntryFn, ChainedFunctionCall{
		Predicate: predicate,
		Operation: operation,
		Config:    cfg,
	})

	return cd

}

func (cd *CallbackDefinitions) AddExit(predicate plinko.Predicate, operation plinko.Operation, cfg plinko.OperationConfig) *CallbackDefinitions {

	cd.OnExitFn = append(cd.OnExitFn, ChainedFunctionCall{
		Predicate: predicate,
		Operation: operation,
		Config:    cfg,
	})

	return cd
}

func executeChain(ctx context.Context, funcs []ChainedFunctionCall, p plinko.Payload, t plinko.TransitionInfo) (retPayload plinko.Payload, err error) {
	var stepName string
	step := 0
	defer func() {
		if err1 := recover(); err1 != nil {
			stack := string(debug.Stack())
			retPayload = p
			err = plinkoerror.CreatePlinkoPanicError(err1, t, step, stepName, stack)
		}
	}()

	if len(funcs) > 0 {
		for _, fn := range funcs {
			stepName = fn.Config.Name
			if fn.Predicate != nil {
				if err = fn.Predicate(ctx, p, t); err != nil {
					// in this case, the predicate failed meaning the function should not be executed.
					continue
				}
			}
			var e error
			p, e = fn.Operation(ctx, p, t)
			step++
			if e != nil {
				return p, e
			}
		}
	}

	return p, nil

}

func executeErrorChain(ctx context.Context, funcs []ChainedErrorCall, p plinko.Payload, t *sideeffects.TransitionDef, err error) (retPayload plinko.Payload, retTd *sideeffects.TransitionDef, retErr error) {
	var stepName string
	step := 0
	defer func() {
		if err1 := recover(); err1 != nil {
			stack := string(debug.Stack())
			retPayload = p
			retTd = t
			retErr = plinkoerror.CreatePlinkoPanicError(err1, t, step, stepName, stack)
		}
	}()

	if len(funcs) > 0 {
		for _, fn := range funcs {
			stepName = fn.Config.Name
			var e error
			p, e = fn.ErrorOperation(ctx, p, t, err)

			if e != nil {
				return p, t, e
			}
		}
	}

	return p, t, err
}

func (cd *CallbackDefinitions) ExecuteExitChain(ctx context.Context, p plinko.Payload, t plinko.TransitionInfo) (plinko.Payload, error) {
	return executeChain(ctx, cd.OnExitFn, p, t)
}

func (cd *CallbackDefinitions) ExecuteEntryChain(ctx context.Context, p plinko.Payload, t plinko.TransitionInfo) (plinko.Payload, error) {
	return executeChain(ctx, cd.OnEntryFn, p, t)
}

func (cd *CallbackDefinitions) ExecuteErrorChain(ctx context.Context, p plinko.Payload, t *sideeffects.TransitionDef, err error, elapsedMilliseconds int64) (plinko.Payload, *sideeffects.TransitionDef, error) {
	p, mt, err := executeErrorChain(ctx, cd.OnErrorFn, p, t, err)

	return p, mt, err
}
