# Plinko - a Stateless State Machine for Go

The project, as well as the example below, are inspired by the [Erlang stateless State Machine](https://erlang.org/doc/design_principles/statem.html) and [Stateless project](https://github.com/dotnet-state-machine/stateless) implementations along with the [Tinder State Machine](https://github.com/Tinder/StateMachine).  The goal is to create the fastest state machine that can be reused across many entities with the least amount of overhead in the process.

## Why Stateless
Most state machine implementations keep track of an in-memory state during the running of an application. This makes sense for desktop applications or games where the journey of that state is critical to the user-facing process, but that doesn't map well to a service that is shepherding things like Orders and Products that number in the thousands-to-millions on any given day.

Stateless State Machines are simply the extraction of the state from the mechanics of state transition.  This allows the state machine to be reduced to a simple data structure, and enables the cost of wiring up the machine to happen only once.  In turn, the state machine can shared across multiple threads and executed concurrently without interference between discrete runs.

There are a number of good articles on this front, there are a couple that focus on state design from the [esoteric around soundness of the design](https://en.wikibooks.org/wiki/Haskell/Understanding_monads/State) to the more [functional programming based definition of a state machine](https://hexdocs.pm/as_fsm/readme.html).

## Common implementation pattern in web services
Many times, a web service may have controllers that span the lifecycle of the entity they are coordinating.  This pattern allows the controller to play the role of traffic cop and defers execution decisions to the state machine.  The state machine introduces two key notions: State and Trigger.  Triggers are mapped to states and execution paths can be different based on sttaes. Applying this to an MVC pattern, the entity contains the state and the state modifying [POST|PUT|PATCH] is the trigger.  For example:

An order can be in different states during it's lifecycle:  Open, Claimed, Delivered, etc.   If someone wishes to cancel that order, there are different protocols and processes involved in each of those states.  In this example a `/cancel/{id}` endpoint is called.  The controller loads the order into a payload and fires the `Cancel` trigger at it using the state machine.  The state machine selects the proper flow and returns the status when complete.

## Features

* Simple support for states and triggers
* Entry/Exit events for states

Some useful extensions are also provided:

* Pushes state external to the structure - instantiate once, use many times.
* Reentrant states
* Export to PlantUML

# Introspection
The state machine can provide a list of triggers for a given state to provide simple access to the list of triggers for any state.

## Creating a state machine
A state machine is created by articulating the states,  the triggers that can be used at each state and the destination state where they land.  Below, a state machine is created describing a set of states an order can progress through along with the triggers that can be used.

```golang
p := plinko.CreateDefinition()

p.CreateState(Created).
	OnEntry(OnNewOrderEntry).
	Permit(Open, Opened).
	Permit(Cancel, Canceled)

p.CreateState(Opened).
	Permit(AddItemToOrder, Opened).
	Permit(Claim, Claimed).
	Permit(Cancel, Canceled)

p.CreateState(Claimed).
	Permit(AddItemToOrder, Claimed).
	Permit(Submit, ArriveAtStore).
	Permit(Cancel, Canceled)

p.CreateState(ArriveAtStore).
	Permit(Submit, MarkedAsPickedUp).
	Permit(Cancel, Canceled)

p.CreateState(MarkedAsPickedUp).
	Permit(Deliver, Delivered).
	Permit(Cancel, Canceled)

p.CreateState(Delivered).
	Permit(Return, Returned)

p.CreateState(Canceled).
	Permit(Reinstate, Created)
	
p.CreateState(Returned)
```

Once created, the next step is compiling the state machine.  This means the state machine is validated for complete-ness.  At this stage, Errors and Warnings are raised.  This incidentally allows the state machine definition to be fully testable in the build pipeline before deployment.

```golang
co := p.Compile()

if co.error {
    // exit
}

fsm := co.StateMachine
```

Once we have the state machine, we can pass that around explicitly or through things like controller context to make it available where needed.

We can trigger the state processes by creating a PlinkoPayload and handing it to the statemachine like so:

```golang
payload := appPayload{ /* ... */ }
fsm.Fire(appPayload, Submit)
```

## State Machine documentation
The fsm can document itself upon a successful compile - emitting PlantUML which can, in turn, be rendered into a state diagram:

```golang
uml, err := p.RenderUml()

if err != nil {
    // exit...
}

fmt.Println(string(uml))
```

![PlantUML Rendered State Diagram](./docs/sample_state_diagram.png)

