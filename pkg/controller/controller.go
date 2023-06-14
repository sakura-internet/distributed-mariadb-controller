package controller

// Controller manages distributed MariaDB cluster in each database server.
// The controller forms as a state machine.
type Controller interface {
	/// GetState returns the current state of the controller.
	GetState() State
	// PreDecideNextStateHandler is triggered before calling MakeDecision()
	PreDecideNextStateHandler() error
	// DecideNextState determines next state that the controller should transition.
	DecideNextState() State
	// OnStateHandler is an implementation of the root state handler.
	// All controller must trigget the state handler on the given state.
	OnStateHandler(nextState State) error
	// OnExit is triggered when the Start() received context.Context.Done().
	OnExit() error
}

// State specifies the controller state.
type State string

const (
	StateInitial State = "initial"
	StatePrimary State = "primary"
	StateReplica State = "replica"
	StateFault   State = "fault"
)

// UnimplementedController implements Controller interface.
// each method of the Controller does nothing.
type UnimplementedController struct{}

// DecideNextState implements Controller
func (*UnimplementedController) DecideNextState() State {
	return StateFault
}

// PreDecideNextStateHandler implements Controller
func (*UnimplementedController) PreDecideNextStateHandler() error {
	return nil
}

var _ Controller = &UnimplementedController{}

// GetState implements Controller
func (*UnimplementedController) GetState() State {
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
