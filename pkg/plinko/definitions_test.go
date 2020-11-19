package plinko

/*func TestStateActionToFilterConversion(t *testing.T) {
	assert.Equal(t, AllowBeforeStateExit, getFilterDefinition(BeforeStateExit))
	assert.Equal(t, AllowAfterStateExit, getFilterDefinition(AfterStateExit))
	assert.Equal(t, AllowBeforeStateEntry, getFilterDefinition(BeforeStateEntry))
	assert.Equal(t, AllowAfterStateEntry, getFilterDefinition(AfterStateEntry))
}


*/

const Created State = "Created"
const Opened State = "Opened"
const Claimed State = "Claimed"
const ArriveAtStore State = "ArrivedAtStore"
const MarkedAsPickedUp State = "MarkedAsPickedup"
const Delivered State = "Delivered"
const Canceled State = "Canceled"
const Returned State = "Returned"
const NewOrder State = "NewOrder"

const Submit Trigger = "Submit"
const Cancel Trigger = "Cancel"
const Open Trigger = "Open"
const Claim Trigger = "Claim"
const Deliver Trigger = "Deliver"
const Return Trigger = "Return"
const Reinstate Trigger = "Reinstate"

/*
func TestCallSideEffectsWithNilSet(t *testing.T) {

	result := runtime.callSideEffects(BeforeStateExit, nil, nil, nil)

	assert.True(t, result == 0)
}

func TestCallEffects(t *testing.T) {
	var effects []runtime.sideEffectDefinition
	callCount := 0

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa StateAction, p Payload, ti TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	payload := testPayload{}
	trInfo := transitionDef{}

	result := callSideEffects(BeforeStateExit, effects, payload, trInfo)

	assert.Equal(t, result, 1)
}

func TestCallEffects_Multiple(t *testing.T) {
	var effects []sideEffectDefinition
	callCount := 0

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa StateAction, p Payload, ti TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa StateAction, p Payload, ti TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa StateAction, p Payload, ti TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, sideEffectDefinition{Filter: AllowAfterStateExit, SideEffect: func(sa StateAction, p Payload, ti TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	payload := testPayload{}
	trInfo := transitionDef{}

	count := callSideEffects(BeforeStateExit, effects, payload, trInfo)

	assert.Equal(t, 3, callCount)
	assert.Equal(t, 3, count)

	callCount = 0
	count = callSideEffects(AfterStateExit, effects, payload, trInfo)

	assert.Equal(t, 4, callCount)
	assert.Equal(t, 4, count)
}

type testPayload struct {
	state     State
	condition bool
}

func (p testPayload) GetState() State {
	return p.state
}

func (p testPayload) PutState(state State) {
	p.state = state
}

func TestStateMachine(t *testing.T) {
	p := CreateDefinition()

	p.Configure(NewOrder).
		OnEntry(OnNewOrderEntry).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder").
		OnEntry(OnNewOrderEntry).
		Permit("Submit", NewOrder)

	p.Configure("UnderReview").
		Permit("CompleteReview", "PublishedOrder").
		Permit("RejectOrder", "RejectedOrder")

	p.Configure("RejectedOrder")

	visitCount := 0
	p.SideEffect(func(sa StateAction, payload Payload, ti TransitionInfo) {
		visitCount += 1
	})

	compilerOutput := p.Compile()
	psm := compilerOutput.StateMachine

	payload := testPayload{state: NewOrder}

	psm.Fire(payload, "Submit")

	assert.Equal(t, 4, visitCount)

}

func TestOnEntryTriggerOperation(t *testing.T) {

	p := CreateDefinition()
	counter := 0

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder").
		OnTriggerEntry("Resupply", func(pp Payload, transitionInfo TransitionInfo) (Payload, error) {
			counter++
			return pp, nil
		}).
		Permit("Resupply", "PublishedOrder").
		Permit("Resubmit", "PublishedOrder")

	compilerOutput := p.Compile()
	psm := compilerOutput.StateMachine

	payload := testPayload{state: "PublishedOrder"}

	psm.Fire(payload, "Resupply")
	assert.Equal(t, 1, counter)

	psm.Fire(payload, "Resubmit")
	assert.Equal(t, 1, counter)

	psm.Fire(payload, "Resupply")
	assert.Equal(t, 2, counter)

}

func TestCanFire(t *testing.T) {
	p := CreateDefinition()

	p.Configure(Created).
		Permit(Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := testPayload{state: Created}

	assert.True(t, psm.CanFire(payload, Open))
	assert.False(t, psm.CanFire(payload, Deliver))
}

func PermitIfPredicate(p Payload, t TransitionInfo) bool {
	tp := p.(testPayload)

	return tp.condition
}

func TestCanFireWithPermitIf(t *testing.T) {
	p := CreateDefinition()

	p.Configure(Created).
		PermitIf(PermitIfPredicate, Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := testPayload{
		state:     Created,
		condition: true,
	}
	assert.True(t, psm.CanFire(payload, Open))

	payload.condition = false
	assert.False(t, psm.CanFire(payload, Open))

}

func findTrigger(triggers []Trigger, trigger Trigger) bool {
	for _, v := range triggers {
		if v == trigger {
			return true
		}
	}

	return false
}
func TestEnumerateTriggers(t *testing.T) {
	p := CreateDefinition()

	p.Configure(Created).
		Permit(Open, Opened).
		Permit(Cancel, Canceled)

	p.Configure(Opened)
	p.Configure(Canceled)

	co := p.Compile()

	psm := co.StateMachine
	payload := testPayload{state: Created}
	triggers, err := psm.EnumerateActiveTriggers(payload)

	assert.Nil(t, err)
	assert.True(t, findTrigger(triggers, Open))
	assert.True(t, findTrigger(triggers, Cancel))
	assert.False(t, findTrigger(triggers, Claim))

	payload = testPayload{state: Opened}
	triggers, err = psm.EnumerateActiveTriggers(payload)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(triggers))

	// request a state that doesn't exist in the state machine definiton and get an error thrown
	payload = testPayload{state: Claimed}
	triggers, err = psm.EnumerateActiveTriggers(payload)

	assert.NotNil(t, err)
}

func TestDiagramming(t *testing.T) {
	p := CreateDefinition()

	p.Configure(Created).
		OnEntry(OnNewOrderEntry).
		Permit(Open, Opened).
		Permit(Cancel, Canceled)

	p.Configure(Opened).
		Permit("AddItemToOrder", Opened).
		Permit(Claim, Claimed).
		Permit(Cancel, Canceled)

	p.Configure(Claimed).
		Permit("AddItemToOrder", Claimed).
		Permit(Submit, ArriveAtStore).
		Permit(Cancel, Canceled)

	p.Configure(ArriveAtStore).
		Permit(Submit, MarkedAsPickedUp).
		Permit(Cancel, Canceled)

	p.Configure(MarkedAsPickedUp).
		Permit(Deliver, Delivered).
		Permit(Cancel, Canceled)

	p.Configure(Delivered).
		Permit(Return, Returned)

	p.Configure(Canceled).
		Permit(Reinstate, Created)

	p.Configure(Returned)

	co := p.Compile()
	fmt.Printf("%+v\n", co.Messages)
	uml, err := p.RenderUml()

	fmt.Println(err)

	fmt.Println(uml)

}

const (
	NewOrder State = "NewOrder"
	Reviewed State = "Reviewed"
)

func IsPlatform(pp Payload) bool {
	return true
}

func OnNewOrderEntry(pp Payload, transitionInfo TransitionInfo) (Payload, error) {
	fmt.Printf("onentry: %+v", transitionInfo)
	return pp, nil
}
*/
