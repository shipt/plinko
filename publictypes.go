package plinko

type State string
type Trigger string

type CallbackDefinitions struct {
	OnEntryFn func(pp PlinkoPayload, transitionInfo TransitionInfo) (PlinkoPayload, error)
	OnExitFn  func(pp PlinkoPayload, transitionInfo TransitionInfo) (PlinkoPayload, error)
}
type PlinkoStateMachine interface {
	Fire(payload PlinkoPayload, trigger Trigger) (PlinkoPayload, error)
	CanFire(payload PlinkoPayload, trigger Trigger) bool
	EnumerateActiveTriggers(payload PlinkoPayload) ([]Trigger, error)
}

type TransitionInfo interface {
	GetSource() State
	GetDestination() State
	GetTrigger() Trigger
}

type PlinkoDefinition interface {
	Configure(state State) StateDefinition
	Compile() PlinkoCompilerOutput
	RenderUml() (Uml, error)
}

type StateDefinition interface {
	//State() string
	OnEntry(entryFn func(pp PlinkoPayload, transitionInfo TransitionInfo) (PlinkoPayload, error)) StateDefinition
	OnExit(exitFn func(pp PlinkoPayload, transitionInfo TransitionInfo) (PlinkoPayload, error)) StateDefinition
	Permit(triggerName Trigger, destinationState State) StateDefinition
	//TBD: AllowReentrance by request, not default
}

type PlinkoPayload interface {
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
