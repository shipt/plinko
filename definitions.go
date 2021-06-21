package plinko

import (
	"context"
)

type State string
type Trigger string

type Predicate func(context.Context, Payload, TransitionInfo) error
type TriggerPredicate func(context.Context, Payload, TransitionInfo) bool
type Operation func(context.Context, Payload, TransitionInfo) (Payload, error)
type ErrorOperation func(context.Context, Payload, ModifiableTransitionInfo, error) (Payload, error)

type StateDefinition interface {
	//State() string
	OnEntry(Operation, ...OperationOption) StateDefinition
	OnError(ErrorOperation, ...OperationOption) StateDefinition
	OnExit(Operation, ...OperationOption) StateDefinition
	OnTriggerEntry(Trigger, Operation, ...OperationOption) StateDefinition
	OnTriggerExit(Trigger, Operation, ...OperationOption) StateDefinition
	Permit(Trigger, State) StateDefinition
	PermitIf(Predicate, Trigger, State) StateDefinition
	PermitReentry(Trigger) StateDefinition
	PermitReentryIf(Predicate, Trigger) StateDefinition
}

type StateMachine interface {
	Fire(context.Context, Payload, Trigger) (Payload, error)
	CanFire(context.Context, Payload, Trigger) error
	EnumerateActiveTriggers(payload Payload) ([]Trigger, error)
}

type TransitionInfo interface {
	GetSource() State
	GetDestination() State
	GetTrigger() Trigger
}

type ModifiableTransitionInfo interface {
	GetSource() State
	GetDestination() State
	GetTrigger() Trigger
	SetDestination(State)
}

type SideEffect func(context.Context, StateAction, Payload, TransitionInfo, int64)

type PlinkoDefinition interface {
	Configure(State, ...StateOption) StateDefinition
	SideEffect(SideEffect) PlinkoDefinition
	FilteredSideEffect(SideEffectFilter, SideEffect) PlinkoDefinition
	Compile() CompilerOutput
	RenderUml() (Uml, error)
	Render(Renderer) error
}

type Renderer interface {
	Render(Graph) error
}

type Graph interface {
	Edges(func(State, State, Trigger))
	Nodes(func(State, StateConfig))
}

type Payload interface {
	GetState() State
}

type CompilerMessage struct {
	CompileMessage CompilerReportType
	Message        string
}

type CompilerReportType string

const (
	CompileError   CompilerReportType = "Compile Error"
	CompileWarning CompilerReportType = "Compile Warning"
	// CompileInfo CompilerReportType "Compile Info"
)

type StateAction string

const (
	BeforeTransition StateAction = "BeforeTransition"
	BetweenStates    StateAction = "MiddleTransition"
	AfterTransition  StateAction = "AfterTransition"
)

type SideEffectFilter int

const (
	AllowBeforeTransition SideEffectFilter = 1
	AllowBetweenStates    SideEffectFilter = 2
	AllowAfterTransition  SideEffectFilter = 4
)

type Uml string

type CompilerOutput struct {
	StateMachine StateMachine
	Messages     []CompilerMessage
}

type OperationConfig struct {
	Name string
}

type OperationOption func(c *OperationConfig)

type StateConfig struct {
	Name        string
	Description string
}

type StateOption func(c *StateConfig)
