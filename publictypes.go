package plinko

type State string
type Trigger string

type CallbackDefinitions struct {
	OnEntryFn func(pp Payload, transitionInfo TransitionInfo) (Payload, error)
	OnExitFn  func(pp Payload, transitionInfo TransitionInfo) (Payload, error)
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

type PlinkoDefinition interface {
	Configure(state State) StateDefinition
	Compile() CompilerOutput
	RenderUml() (Uml, error)
}

type StateDefinition interface {
	//State() string
	OnEntry(entryFn func(pp Payload, transitionInfo TransitionInfo) (Payload, error)) StateDefinition
	OnExit(exitFn func(pp Payload, transitionInfo TransitionInfo) (Payload, error)) StateDefinition
	Permit(triggerName Trigger, destinationState State) StateDefinition
	//TBD: AllowReentrance by request, not default
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
