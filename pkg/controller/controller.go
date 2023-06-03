package controller

// Controller manages distributed MariaDB cluster in each database server.
// The controller forms as a state machine.
type Controller interface {
	/// GetState returns the current state of the controller.
	GetState() State
	// PreMakeDecisionHandler is triggered before calling MakeDecision()
	PreMakeDecisionHandler() error
	// MakeDecision determines next state that the controller should transition.
	MakeDecision() State
	// OnStateHandler is an implementation of the root state handler.
	// All controller must trigget the state handler on the given state.
	OnStateHandler(nextState State) error
	// OnExit is triggered when the Start() received context.Context.Done().
	OnExit() error
}

// State specifies the controller state.
type State string

const (
	StateInitial State = "Initial"
	StatePrimary State = "Primary"
	StateReplica State = "Replica"
	StateFault   State = "Fault"
)

// UnimplementedController implements Controller interface.
// each method of the Controller does nothing.
type UnimplementedController struct{}

var _ Controller = &UnimplementedController{}

// GetState implements Controller
func (*UnimplementedController) GetState() State {
	return StateFault
}

// MakeDecision implements Controller
func (*UnimplementedController) MakeDecision() State {
	return StateFault
}

// OnExit implements Controller
func (*UnimplementedController) OnExit() error {
	return nil
}

// OnStateHandler implements Controller
func (*UnimplementedController) OnStateHandler(nextState State) error {
	return nil
}

// PreMakeDecisionHandler implements Controller
func (*UnimplementedController) PreMakeDecisionHandler() error {
	return nil
}
