package composition

import "github.com/shipt/plinko/pkg/plinko"

type ChainedFunctionCall struct {
	Predicate plinko.Predicate
	Operation plinko.Operation
}

type CallbackDefinitions struct {
	OnEntryFn []ChainedFunctionCall
	OnExitFn  []ChainedFunctionCall

	EntryFunctionChain []string
	ExitFunctionChain  []string
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

func (cd *CallbackDefinitions) ExecuteExitChain(p plinko.Payload, t plinko.TransitionInfo) (plinko.Payload, error) {
	return executeChain(cd.OnExitFn, p, t)
}

func (cd *CallbackDefinitions) ExecuteEntryChain(p plinko.Payload, t plinko.TransitionInfo) (plinko.Payload, error) {
	return executeChain(cd.OnEntryFn, p, t)
}
