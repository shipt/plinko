package composition

import (
	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/sideeffects"
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

func executeChain(funcs []ChainedFunctionCall, p plinko.Payload, t plinko.TransitionInfo) (plinko.Payload, error) {
	if funcs != nil && len(funcs) > 0 {
		for _, fn := range funcs {
			if fn.Predicate != nil {
				if !fn.Predicate(p, t) {
					continue
				}
			}

			p, e := fn.Operation(p, t)

			if e != nil {
				return p, e
			}
		}
	}

	return p, nil

}

func executeErrorChain(funcs []ChainedErrorCall, p plinko.Payload, t *sideeffects.TransitionDef, err error) (plinko.Payload, *sideeffects.TransitionDef, error) {
	if funcs != nil && len(funcs) > 0 {
		for _, fn := range funcs {
			p, t1, e := fn.ErrorOperation(p, t, err)

			if e != nil {
				return p, t1.(*sideeffects.TransitionDef), e
			}
		}
	}

	return p, t, err
}

func (cd *CallbackDefinitions) ExecuteExitChain(p plinko.Payload, t plinko.TransitionInfo) (plinko.Payload, error) {
	return executeChain(cd.OnExitFn, p, t)
}

func (cd *CallbackDefinitions) ExecuteEntryChain(p plinko.Payload, t plinko.TransitionInfo) (plinko.Payload, error) {
	return executeChain(cd.OnEntryFn, p, t)
}

func (cd *CallbackDefinitions) ExecuteErrorChain(p plinko.Payload, t *sideeffects.TransitionDef, err error) (plinko.Payload, *sideeffects.TransitionDef, error) {
	p, mt, err := executeErrorChain(cd.OnErrorFn, p, t, err)

	return p, mt, err
}
