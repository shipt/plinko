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
	"testing"

	"github.com/shipt/plinko"
	"github.com/stretchr/testify/assert"
)

const (
	NewOrder plinko.State = "NewOrder"
)

func TestGetCallerHelper(t *testing.T) {
	n := getCallerHelper(func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		return nil, nil
	})

	// this follow test is based on the location of the above line
	assert.Equal(t, "internal/runtime/internal_test.go:23", n)

}

func TestStateDefinition(t *testing.T) {
	state := InternalStateDefinition{
		State:    "NewOrder",
		Triggers: make(map[plinko.Trigger]*TriggerDefinition),
	}

	assert.Panics(t, func() {
		state.Permit("Submit", "PublishedOrder").
			Permit("Review", "ReviewOrder").
			Permit("Submit", "foo")
	})

}

func TestStateRedeclarationPanic(t *testing.T) {
	p := createPlinkoDefinition()

	p.SideEffect(nil)
	p.FilteredSideEffect(plinko.AllowAfterTransition, nil)

	p.Configure("Open")
	assert.Panics(t, func() { p.Configure("Open") })

	assert.Panics(t, func() {
		p.Configure("Close").
			Permit("Go", "Open").
			Permit("Go", "Open")
	})
}

func OnNewOrderEntry(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	fmt.Printf("onentry: %+v", transitionInfo)
	return pp, nil
}

func OnRetrieveWithName(name string) func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	return func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		fmt.Printf("onentry with %s: %+v", name, transitionInfo)
		return pp, nil
	}
}

func OnRetrieveWithName2(name string) func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	return func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		fmt.Printf("onentry with %s: %+v", name, transitionInfo)
		return pp, nil
	}
}
func TestInferredNaming(t *testing.T) {
	n := nameOf(OnNewOrderEntry)

	assert.Equal(t, "OnNewOrderEntry", n)
}

func TestInferredNamingOnLambdaLiftedClosure(t *testing.T) {
	n := nameOf(OnRetrieveWithName2("foo"))

	assert.Equal(t, "OnRetrieveWithName2", n)
}

func TestInferredNamingOnLambdaLiftedClosurePart2(t *testing.T) {
	n := nameOf(OnRetrieveWithName("foo"))

	assert.Equal(t, "OnRetrieveWithName", n)
}

func TestInferredNamingWithAnonymousFunction(t *testing.T) {
	n := nameOf(func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		return nil, nil
	})

	assert.Equal(t, "TestInferredNamingWithAnonymousFunction", n)
}

func TestCleanFileName(t *testing.T) {
	assert.Equal(t, "", cleanFileName(""))
	assert.Equal(t, "foo", cleanFileName("foo"))
	assert.Equal(t, "foo/bar", cleanFileName("foo/bar"))
	assert.Equal(t, "foo/bar/baz", cleanFileName("foo/bar/baz"))
	assert.Equal(t, "bar/baz/fizz", cleanFileName("foo/bar/baz/fizz"))
	assert.Equal(t, "//", cleanFileName("///"))
}
