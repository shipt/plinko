package composition

import (
	"context"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/sideeffects"
	"github.com/shipt/plinko/plinkoerror"
)

type ChainedFunctionCall struct {
	Predicate plinko.Predicate
	Operation plinko.Operation
}

type ChainedErrorCall struct {
	ErrorOperation plinko.ErrorOperation
}

type CallbackDefinitions struct {
	OnEntryFn []ChainedFunctionCall
	OnExitFn  []ChainedFunctionCall
	OnErrorFn []ChainedErrorCall

	EntryFunctionChain []string
	ExitFunctionChain  []string
}

func (cd *CallbackDefinitions) AddError(errorOperation plinko.ErrorOperation) *CallbackDefinitions {
	cd.OnErrorFn = append(cd.OnErrorFn, ChainedErrorCall{
		ErrorOperation: errorOperation,
	})

	return cd
}

func (cd *CallbackDefinitions) AddEntry(predicate plinko.Predicate, operation plinko.Operation) *CallbackDefinitions {

	cd.OnEntryFn = append(cd.OnEntryFn, ChainedFunctionCall{
		Predicate: predicate,
		Operation: operation,
	})

	return cd

}

func (cd *CallbackDefinitions) AddExit(predicate plinko.Predicate, operation plinko.Operation) *CallbackDefinitions {

	cd.OnExitFn = append(cd.OnEntryFn, ChainedFunctionCall{
		Predicate: predicate,
		Operation: operation,
	})

	return cd
}

func executeChain(ctx context.Context, funcs []ChainedFunctionCall, p plinko.Payload, t plinko.TransitionInfo) (retPayload plinko.Payload, err error) {
	step := 0
	defer func() {
		if err1 := recover(); err1 != nil {
			retPayload = p
			err = plinkoerror.CreatePlinkoPanicError(err1, t, step)
		}
	}()

	if funcs != nil && len(funcs) > 0 {
		for _, fn := range funcs {

			if fn.Predicate != nil {
				if !fn.Predicate(ctx, p, t) {
					continue
				}
			}

			p, e := fn.Operation(ctx, p, t)
			step++
			if e != nil {
				return p, e
			}
		}
	}

	return p, nil

}

func executeErrorChain(ctx context.Context, funcs []ChainedErrorCall, p plinko.Payload, t *sideeffects.TransitionDef, err error) (retPayload plinko.Payload, retTd *sideeffects.TransitionDef, retErr error) {
	step := 0
	defer func() {
		if err1 := recover(); err1 != nil {
			retPayload = p
			retTd = t
			retErr = plinkoerror.CreatePlinkoPanicError(err1, t, step)
		}
	}()

	if funcs != nil && len(funcs) > 0 {
		for _, fn := range funcs {
			p, e := fn.ErrorOperation(ctx, p, t, err)

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
