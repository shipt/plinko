# Plinko - an Finite State Machine for go

The project, as well as the example above was inspired by Simple State Machine and the Erlang Stateless State Machine Implementations.

## Features

* Simple support for states and triggers
* Entry/Exit events for states

Some useful extensions are also provided:

* Pushes state external to the structure - instantiate once, use many times.
* Reentrant states
* Export to dot graph and PlantUML

# Introspection
The state machine can provide a list of triggers for a given state to provide simple access to the list of triggers for any state.

