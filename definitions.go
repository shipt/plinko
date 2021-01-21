package plinko

type State string
type Trigger string

type Predicate func(Payload, TransitionInfo) bool
type Operation func(Payload, TransitionInfo) (Payload, error)
type ErrorOperation func(Payload, ModifiableTransitionInfo, error) (Payload, error)

type StateDefinition interface {
	//State() string
	OnEntry(Operation) StateDefinition
	OnTriggerEntry(Trigger, Operation) StateDefinition
	OnExit(Operation) StateDefinition
	OnTriggerExit(Trigger, Operation) StateDefinition
	Permit(Trigger, State) StateDefinition
	PermitIf(Predicate, Trigger, State) StateDefinition
	OnError(ErrorOperation) StateDefinition
	//TBD: AllowReentrance by request, not default
}

type StateMachine interface {
	Fire(payload Payload, trigger Trigger) (Payload, error)
	CanFire(payload Payload, trigger Trigger) bool
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

type SideEffect func(StateAction, Payload, TransitionInfo, int64)

type PlinkoDefinition interface {
	Configure(State) StateDefinition
	SideEffect(SideEffect) PlinkoDefinition
	FilteredSideEffect(SideEffectFilter, SideEffect) PlinkoDefinition
	Compile() CompilerOutput
	RenderUml() (Uml, error)
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
	BeforeTransition StateAction = "BeforeStateExit"
	BetweenStates    StateAction = "AfterStateExit"
	AfterTransition  StateAction = "BeforeStateEntry"
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
