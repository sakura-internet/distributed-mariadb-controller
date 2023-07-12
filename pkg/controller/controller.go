// Copyright 2023 The distributed-mariadb-controller Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
