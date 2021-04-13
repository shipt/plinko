package composition

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/sideeffects"
	"github.com/shipt/plinko/plinkoerror"
	"github.com/stretchr/testify/assert"
)

type testPayload struct {
	value string
}

func (t testPayload) GetState() plinko.State {
	return "stub"
}

func TestAddEntry(t *testing.T) {
	cd := CallbackDefinitions{}

	cd.AddEntry(nil, func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		return pp, nil
	}, plinko.OperationConfig{})

	assert.Equal(t, 1, len(cd.OnEntryFn))

	cd.AddEntry(nil, func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		return pp, nil
	}, plinko.OperationConfig{})

	assert.Equal(t, 2, len(cd.OnEntryFn))
}

func TestAddError(t *testing.T) {
	cd := CallbackDefinitions{}

	cd.AddError(func(_ context.Context, pp plinko.Payload, ti plinko.ModifiableTransitionInfo, err error) (plinko.Payload, error) {
		return pp, err
	}, plinko.OperationConfig{})

	assert.Equal(t, 1, len(cd.OnErrorFn))

	cd.AddError(func(_ context.Context, pp plinko.Payload, ti plinko.ModifiableTransitionInfo, err error) (plinko.Payload, error) {
		return pp, err
	}, plinko.OperationConfig{})

	assert.Equal(t, 2, len(cd.OnErrorFn))
}
func TestAddExit(t *testing.T) {
	cd := CallbackDefinitions{}

	cd.AddExit(nil, func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		return pp, nil
	}, plinko.OperationConfig{})

	assert.Equal(t, 1, len(cd.OnExitFn))

	cd.AddExit(nil, func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		return pp, nil
	}, plinko.OperationConfig{})

	assert.Equal(t, 2, len(cd.OnExitFn))
}

func TestExecuteErrorChainSingleFunctionWithModifiedDestination(t *testing.T) {
	const Woo plinko.State = "woo"
	const ErrorState plinko.State = "bar2"
	const GoodState plinko.State = "bar"
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: GoodState,
		Trigger:     "baz",
	}
	list := []ChainedErrorCall{
		ChainedErrorCall{
			ErrorOperation: func(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
				m.SetDestination(ErrorState)
				return p, nil
			},
		},
	}

	p, t1, e := executeErrorChain(context.TODO(), list, nil, &transitionDef, errors.New("wizard"))

	assert.Equal(t, ErrorState, t1.GetDestination())
	assert.Equal(t, errors.New("wizard"), e)
	assert.Equal(t, p, nil)

}

func TestExecuteErrorChainMultiFunctionWithError(t *testing.T) {
	const Woo plinko.State = "woo"
	const ErrorState plinko.State = "bar2"
	const GoodState plinko.State = "bar"
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: GoodState,
		Trigger:     "baz",
	}
	counter := 0
	list := []ChainedErrorCall{
		ChainedErrorCall{
			ErrorOperation: func(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
				counter++
				return p, errors.New("notwizard")
			},
		},
		ChainedErrorCall{
			ErrorOperation: func(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
				m.SetDestination(ErrorState)
				counter++
				return p, nil
			},
		},
	}

	p, t1, e := executeErrorChain(context.TODO(), list, nil, &transitionDef, errors.New("wizard"))

	assert.Equal(t, GoodState, t1.GetDestination())
	assert.Equal(t, 1, counter)
	assert.Equal(t, errors.New("notwizard"), e)
	assert.Equal(t, p, nil)
}

func TestChainedFunctionWithError(t *testing.T) {
	payload := testPayload{
		value: "foo",
	}
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: "GoodState",
		Trigger:     "baz",
	}

	list := []ChainedFunctionCall{
		{
			Operation: func(_ context.Context, p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				te := p.(testPayload)
				assert.Equal(t, "foo", te.value)

				// here we'll return a new payload instance
				te2 := testPayload{
					value: "foo2",
				}

				return te2, errors.New("foo")
			},
		},
	}

	p, err := executeChain(context.TODO(), list, payload, transitionDef)

	assert.NotNil(t, p)
	assert.NotNil(t, err)
	assert.Equal(t, "foo", err.Error())
}

func TestChainedFunctionWithFailedPredicate(t *testing.T) {
	payload := testPayload{
		value: "foo",
	}
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: "GoodState",
		Trigger:     "baz",
	}

	list := []ChainedFunctionCall{
		{
			Predicate: func(_ context.Context, _ plinko.Payload, _ plinko.TransitionInfo) error {
				return errors.New("predicate_error")
			},
			Operation: func(_ context.Context, p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				te := p.(testPayload)
				assert.Equal(t, "foo", te.value)

				// here we'll return a new payload instance
				te2 := testPayload{
					value: "foo2",
				}

				return te2, nil
			},
		},
	}

	p, err := executeChain(context.TODO(), list, payload, transitionDef)
	p1 := p.(testPayload)

	assert.NotNil(t, p1)
	assert.Nil(t, err)
	assert.Equal(t, "foo", p1.value)
}

func TestChainedFunctionPassingProperly(t *testing.T) {
	payload := testPayload{
		value: "foo",
	}
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: "GoodState",
		Trigger:     "baz",
	}

	list := []ChainedFunctionCall{

		ChainedFunctionCall{
			Operation: func(_ context.Context, p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				te := p.(testPayload)
				assert.Equal(t, "foo", te.value)

				// here we'll return a new payload instance
				te2 := testPayload{
					value: "foo2",
				}

				return te2, nil
			},
		},
		ChainedFunctionCall{
			Operation: func(_ context.Context, p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				te := p.(testPayload)
				assert.Equal(t, "foo2", te.value)

				// here we'll test a mutating value
				te.value = "foo3"

				return te, nil
			},
		},
	}

	p, err := executeChain(context.TODO(), list, payload, transitionDef)

	assert.NotNil(t, p)
	assert.Nil(t, err)
	assert.Equal(t, "foo3", p.(testPayload).value)
}

func TestChainedFunctionChainWithPanic(t *testing.T) {
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: "GoodState",
		Trigger:     "baz",
	}

	list := []ChainedFunctionCall{

		ChainedFunctionCall{
			Operation: func(_ context.Context, p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				return p, nil
			},
		},
		ChainedFunctionCall{
			Operation: func(_ context.Context, p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				panic(errors.New("panic-error"))
				//return p, errors.New("notwizard")
			},
		},
	}

	p, err := executeChain(context.TODO(), list, nil, transitionDef)

	assert.Nil(t, p)
	assert.NotNil(t, err)

	e := err.(*plinkoerror.PlinkoPanicError)
	assert.NotNil(t, e)

	assert.Equal(t, "panic-error", e.InnerError.Error())
	assert.Nil(t, e.UnknownInnerError)
	assert.Equal(t, 1, e.StepNumber)
	assert.True(t, strings.Contains(e.Stack, "internal/composition/composition_test.go"))

}

func TestErrorFunctionChainWithPanic(t *testing.T) {
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: "GoodState",
		Trigger:     "baz",
	}

	list := []ChainedErrorCall{
		ChainedErrorCall{
			ErrorOperation: func(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {

				panic(errors.New("panic-error"))
			},
		},
		ChainedErrorCall{
			ErrorOperation: func(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {

				return p, nil
			},
		},
	}

	p, td2, err := executeErrorChain(context.TODO(), list, nil, &transitionDef, errors.New("encompassing-error"))

	assert.Nil(t, p)
	assert.NotNil(t, err)
	assert.Equal(t, plinko.State("GoodState"), td2.Destination)

	e := err.(*plinkoerror.PlinkoPanicError)
	assert.NotNil(t, e)

	assert.Equal(t, "panic-error", e.InnerError.Error())
	assert.Nil(t, e.UnknownInnerError)
	assert.Equal(t, 0, e.StepNumber)
}

func TestExecuteErrorChain(t *testing.T) {
	cd := CallbackDefinitions{}
	tp := testPayload{
		value: "foo",
	}

	p, td, e := cd.ExecuteErrorChain(context.TODO(), &tp, &sideeffects.TransitionDef{}, errors.New("foo"), 100)

	p1 := p.(*testPayload)
	assert.Equal(t, "foo", p1.value)
	assert.NotNil(t, td)
	assert.NotNil(t, e)
	assert.Equal(t, "foo", e.Error())
}

func TestExecuteEntryChain(t *testing.T) {
	cd := CallbackDefinitions{}
	tp := &testPayload{
		value: "foo",
	}

	p, e := cd.ExecuteEntryChain(context.TODO(), tp, nil)
	p1 := p.(*testPayload)

	assert.NotNil(t, p1)
	assert.Nil(t, e)

	assert.Equal(t, "foo", p1.value)
}

func TestExecuteExitChain(t *testing.T) {
	cd := CallbackDefinitions{}
	tp := &testPayload{
		value: "foo",
	}

	p, e := cd.ExecuteExitChain(context.TODO(), tp, nil)
	p1 := p.(*testPayload)

	assert.NotNil(t, p1)
	assert.Nil(t, e)

	assert.Equal(t, "foo", p1.value)
}
