# Plinko - a Fluent State Machine for Go


[![Build Status](https://drone.shipt.com/api/badges/shipt/plinko/status.svg)](https://drone.shipt.com/shipt/plinko) [![codecov](https://codecov.io/gh/shipt/plinko/branch/main/graph/badge.svg?token=8UX649KGGV)](https://codecov.io/gh/shipt/plinko) Build Status

## Create state machines and lightweight state machine-based workflows directly in golang code

The project, as well as the example below, are inspired by the [Erlang State Machine](https://erlang.org/doc/design_principles/statem.html) and [Stateless project](https://github.com/dotnet-state-machine/stateless) implementations. The goal is to create the fastest state machine that can be reused across many entities with the least amount of overhead in the process.

## Why State Machines?
Some state machine implementations keep track of an in-memory state during the running of an application. This makes sense for desktop applications or games where the journey of that state is critical to the user-facing process, but that doesn't map well to web services shepherding things like Orders and Products that number in the thousands-to-millions on any given day.

This allows the state machine to be reduced to a simple data structure, and enables the cost of wiring up the machine to happen only once but reused multiple times.  In turn, the state machine can be shared across multiple threads and executed concurrently without interference between discrete runs.

There are a number of good articles on this front, there are a couple that focus on state design from the [esoteric around soundness of the design](https://en.wikibooks.org/wiki/Haskell/Understanding_monads/State) to the more [functional programming based definition of a state machine](https://hexdocs.pm/as_fsm/readme.html).

## Common implementation pattern in web services
Many times, a web service may have controllers that span the lifecycle of the entity they are coordinating.  This pattern allows the controller to play the key, narrow role of traffic coordinator and defers execution decisions to the state machine.  The state machine introduces two key notions: State and Trigger.  Triggers are mapped to states and execution paths can be different based on states. Applying this to an MVC pattern, the entity contains the state and the state modifying `[POST|PUT|PATCH]` endpoint is the trigger.  For example:

An order can be in different states during it's lifecycle:  Open, Claimed, Delivered, etc.   If someone wishes to cancel that order, there are different protocols and processes involved in each of those states.  In this approach a `/cancel/{id}` endpoint is called.  The controller loads the order into a payload and fires the `Cancel` trigger at it using the state machine.  The state machine selects the proper flow and returns the status when complete.

## Features

* Simple support for states and triggers
* Entry/Exit events for states
* Side Effect support for supporting uniform functionality when modifying state
* Error events used to properly respond to errors raised during a state transition

Some useful extensions are also provided:

* Pushes state external to the structure - instantiate once, use many times.
* Reentrant states
* Export to PlantUML

# Installing 

Using Plinko is easy.   First, use `go get` to istall the latest version of the library.  This command will install everything you need - in fact, one design goal of Plinko is to minimize dependencies.  There are no runtime dependencies required for Plinko, and the only dependencies used by the project are used for unit testing.

```go
go get -u github.com/shipt/plinko`
```

Next, include Plinko in your application:

```go
import "github.com/shipt/plinko"
```

You will define state machine using the examples below, and compiling the state machine once to reuse again and again.  Efficiency is front of mind,  meaning the compilation process is fast and runs in far less than 1/10,000th of a second on a reasonable VM. Or, given a single thread on an x86 processor, a statemachine can be fully compiled and ready to run more than 10,000,000 times a second.  

# Introspection
The state machine can provide a list of triggers for a given state to provide simple access to the list of triggers for any state.

## Creating a state machine
A state machine is created by articulating the states,  the triggers that can be used at each state and the destination state where they land. Here is a sample declaration of the states and triggers we will use:

```go
const Created          State = "Created"
const Opened           State = "Opened"
const Claimed          State = "Claimed"
const ArriveAtStore    State = "ArrivedAtStore"
const MarkedAsPickedUp State = "MarkedAsPickedup"
const Delivered        State = "Delivered"
const Canceled         State = "Canceled"
const Returned         State = "Returned"

const Submit    Trigger = "Submit"
const Cancel    Trigger = "Cancel"
const Open      Trigger = "Open"
const Claim     Trigger = "Claim"
const Deliver   Trigger = "Deliver"
const Return    Trigger = "Return"
const Reinstate Trigger = "Reinstate"
```

 Below, a state machine is created describing a set of states an order can progress through along with the triggers that can be used.

```go
p := plinko.CreateDefinition()

p.Configure(Created).
   OnEntry(OnNewOrderEntry).
   Permit(Open, Opened).
   Permit(Cancel, Canceled)

p.Configure(Opened).
   Permit(AddItemToOrder, Opened).
   Permit(Claim, Claimed).
   Permit(Cancel, Canceled)

p.Configure(Claimed).
   Permit(AddItemToOrder, Claimed).
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
```

Once created, the next step is compiling the state machine.  This means the state machine is validated for complete-ness.  At this stage, Errors and Warnings are raised.  This incidentally allows the state machine definition to be fully testable in the build pipeline before deployment.

```go
co := p.Compile()

if co.error {
   // exit
}

fsm := co.StateMachine
```

Once we have the state machine, we can pass that around explicitly or through things like controller context to make it available where needed.

We can trigger the state processes by creating a PlinkoPayload and handing it to the statemachine like so:

```go
payload := appPayload{ /* ... */ }
fsm.Fire(ctx, appPayload, Submit)
```

## Permitted Transitions

The state machine allows the definitions of transitions using the `Permit` function.  This means I can declare that a triggered action can happen on one state, but not another using:

```go
p.Configure(Opened).
   // ...
   Permit(Cancel, Canceled)

p.Configure(Claimed).
   // The key here is that Canceling from a Claimed state is not permitted.
```

This is useful, because I now have guard rails around when a `Cancel` trigger can be used and when it cannot.  Furthermore, I can use the `CanFire()` method of the state machine to ask if I have a valid action:

```go
if !fsm.CanFire(ctx, payload, Cancel) {
   return "Cannot perform this action"
}
```

Furthermore, let's say `Cancel` is allowed within a timeframe described in the payload.  In this case, let's say it's valid when our order is more than 1 hour from being scheduled to shop.  In this, we first define a predicate function:

```go
func IsOrderCancellable(p Payload, t TransitionInfo) bool {
   return p.ScheduledToShop().Sub(time.Now()).Hours() >= 1
}
```

In this case, I define the trigger differently with:

```go
p.Configure(Claimed).
   PermitIf(IsOrderCancellable, Cancel, Cancelled)
```

Using `PermitIf` now allows the `fsm.CanFire` code block above to be executed without modification,  but now the state machine validates if the trigger can be used based on the order's scheduled to shop time.

### Reentrancy
Reentrancy is a state transition where the destination is the same State.   This means `OnExit` functions get called for the current state, followed by the `OnEntry` calls for the current state.  All the SideEffects are also accordingly raised as expected with the source and destination states being the same.

A simple reentrant state is defined as:

```go
p.Configure(Claimed).
   PermitReentry(AddItemToOrder)
```

Likewise, a conditional `PermitReentryIf` can be defined that relies on a predicate to decide if the trigger may be fired.  For this example, a rule might determine that an item can only be added based on timing or other conditions described in a function called `ItemAddRule`.

```go
p.Configure(Claimed).
   PermitReentryIf(ItemAddRule, AddItemToOrder)
```   

## Functional Composition

When entering or exiting a state, a series of functions need to act to make that transition complete.  Some transitions are simple, and some are complex.  The key here is creating a series of steps that are testable and operate based on a standard pattern. 

Let's take a look at a piece of code we setup earlier:


```go 
p.Configure(Created).
   OnEntry(OnNewOrderEntry).
   Permit(Open, Opened).
   Permit(AddItem, Created)
```

OnNewOrderEntry is function defined as such:

``` go
func OnNewOrderEntry(p plinko.Payload, t plinko.TransitionInfo) (plinko.Payload, error) {
   // perform a series of steps based on the 
   // payload and transition info
   // ...

   return p, nil
}
```

This is useful for a couple of reasons: First, this becomes one distinct action that can succeed or fail.  When it succeeds, the chain continues and works toward the successful transition to the new state. And second, this is an operation that can be tested in isolation.  

Both of these reasons are significant when building a complex set of transitions.

Next, we have a variation on the chaining where we can say "only run this function if a particular trigger triggered the transition".   This is the `OnTriggerEntry(trigger, func)` function.

```go 
p.Configure(Created).
   OnTriggerEntry(AddItem, RecalculateTotals).
   Permit(Open, Opened).
   Permit(AddItem, Created)
```

 In the example above, the `RecalculateTotals` function is only executed when the `AddItem` trigger is raised.   This allows us to explicitly describe the transition steps without placing that complexity inside the `RecalculateTotals` function.


## Side-Effect Support

Side-Effect supports wiring up common functions that respond to state changes happening.   This is a great place for logging and recording movement in a uniform way.

Side Effects are raised at different phases of a state transition.  Given an order that's sitting in a `Created` state that has been actioned with an `Open` trigger, we'll see the following calls to the SideEffect functions.

| State | Action |  Trigger | Destination State |
| --- | --- | --- | --- |
| Created | BeforeTransition | Open | Opened |
| Created | BetweenStates | Open | Opened |
| Created | AfterTransition | Open | Opened |

In the above list, you can see the registered function is called 4 times throughout the lifecycle of the transition.   This gives us consistency and observability throughout the process.

We can better understand how this works by looking at a standard configuration.  

```go
// given a standard definition ...
p := plinko.CreateDefinition()

p.Configure(Created).
   OnEntry(OnNewOrderEntry).
   Permit(Open, Opened).
   Permit(Cancel, Canceled)

p.Configure(Opened).
   Permit(AddItemToOrder, Opened).
   Permit(Claim, Claimed).
   Permit(Cancel, Canceled)


// we register for side effects like this.
p.SideEffect(StateLogging)
p.SideEffect(MetricsRecording)
p.FilteredSideEffect(AfterStateEntry, StateEntryRecording)
```

In the case above, we registered two functions that get executed whenever a change happens.  These functions will always be called in the order they are registered for a given state transition.

In addition, we registered a FilteredSideEffect that only gets called on the requested action.

These are functions that have signature including the starting state, the destination state, the trigger used to kick off the transition and the payload.

```go
func StateLogging(action StateAction, payload Payload, transitionInfo TransitionInfo) {
   // this can typically be broken out into a function on the logger, but keeping
   // it here for clarity in demonstration

   logEntry := StateLog {
      Action:           action,
      SourceState:      transitionInfo.GetSource(),
      DestinationState: transitionInfo.GetDestination(),
      Trigger:          transitionInfo.GetTrigger(),
      OrderID:          payload.GetOrderID(),
   }

   // call to our logger that will decorate the entry with timing information and the like.
   logger.LogStateInfo(logEntry)
}

func MetricsRecording(action StateAction, payload Payload, transitionInfo TransitionInfo) {
   // this can be a simple function that pulls apart the details and sends them to
   // things like graphite, influx or any timeseries metrics database for graphing and alerting.
   metrics.RecordStateMovement(action, payload, transitionInfo)
}

```


## Error Handling

State Machine error handling follows the same pattern that we see in golang in general, when an error occurs that cannot be rectified and causes the state change to fail, an error is raised from the function.   Plinko redirects the flow to the `OnError` definition for remediation. An error in this situation can mean that a Payloads state is moved to something other than the original destination.  Depending on the system, this might be mean it goes back to an old state, continues on to the new state or it lands in a _triage_ state.  Equally important is that this information can be recorded reliably with the Side-Effect support documented above.  Plinko ensures the ability to adjust the destination state and make that consistent with SideEffects.

While the `OnEntry` and `OnExit` function definitions take a `TransitionInfo` parameter that is immutable, and error operation is defined with a `ModifiableTransitionInfo` interface that allows the function to change the `DestinationState`.  In addition, the function also accepts the error raised during the `On[Entry|Exit]` operation so it can be interrogated when necessary.  The definition of an error operation handler looks like this:

```go
 ErrorOperation func(Payload, ModifiableTransitionInfo, error) (Payload, error)
```

An ErrorOperation function implements this signature and tests the error case.  Here is an example where we redirect based on a match.

```go
func RedirectOnDeactivatedCustomer(p Payload, m ModifiableTransitionInfo, e error) (Payload, error) {
   if e == DeactivatedCustomerError {
      m.SetDestination(DeactivatedTriage)
      return RecordOrder(p, m)
   }

   // we didn't identify the error, so we'll pass this through for further error handling
   return p, nil
}
```

There are a couple things to note.   If you return a non-nil `error` during an `OnError` routine, this is regarded as a fatal error that is floated to the caller who initiated the `.Fire(..)` command.  This condition is floated to the registered SideEffect handlers as well.

Some key pieces to remember when building up a set of error handlers.    First, you don't have to handle _every_ error case.  This is done by returning `(payload, nil)`) to the caller.  Plinko will call any subsequent error handlers in this case to give each handler an opportunity to perform it's role in the set of operations.  This is powerful, as handlers can take on different aspects of error handling, including custom messaging and metrics. This allows these functions to be simple, focused operations that compose a larger set of responsiblities (through additional functions) when an error occurs.

Lastly, here is a sample plinko configuration that uses error handling to perform the proper state destination redirect shown above when an order transitions to `Opened` and the user has been deactivated.  Note the separation of concerns - one to perform the redirect and save state, and the other to perform a system notification.

```golang
p.Configure(Created).
   Permit(Open, Opened).
   Permit(Cancel, Canceled)

p.Configure(Opened).
   OnEntry(OnOrderOpen).
   OnError(RedirectOnDeactivatedCustomer).
   OnError(GenerateSlackMessageNotification).
   Permit(AddItemToOrder, Opened).
   Permit(Claim, Claimed).
   Permit(Cancel, Canceled)
```

## Panic Support
On calls to Entry or Exit Functions, Plinko will capture any panics.  These panics are recorded as a structured error, containing when and where the error occured.  The `OnError` handlers can then respond as appropriate.

## State Machine self-documentation
The fsm can document itself upon a successful compile - emitting PlantUML which can, in turn, be rendered into a state diagram:

```go
uml, err := p.RenderUml()

if err != nil {
   // exit...
}

fmt.Println(string(uml))
```

![PlantUML Rendered State Diagram](./docs/sample_state_diagram.png)

