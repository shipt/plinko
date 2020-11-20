package composition

import "github.com/shipt/plinko/pkg/plinko"

type ChainedFunctionCall struct {
	Predicate plinko.Predicate
	Operation plinko.Operation
}

type CallbackDefinitions struct {
	OnEntryFn []ChainedFunctionCall
	OnExitFn  func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error)

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
