package checkstate

import (
	"fmt"
	"slices"
	"time"

	"gopkg.in/yaml.v3"
)

type StateMachinable interface {
	SetCheckedState(string)
	GetCurrentState() string
	AppendLogEntry(event Event, transition StateTransition) error
	PreviousTransitions() []StateTransition
}

type StateMachine interface {
	Move(event Event, state string) error
	HaveMoved(transition StateTransition) bool
}

type PermittedStateTransitions map[string][]string

type Event struct {
	Timestamp   time.Time
	Description string
}

type StateTransition struct {
	From string
	To   string
}

func PermittedStateTransitionsFromYAML(data []byte) (PermittedStateTransitions, error) {

	var result PermittedStateTransitions

	err := yaml.Unmarshal([]byte(data), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse yaml %w", err)
	}

	return result, nil

}

func CreateStateMachine(value StateMachinable, permittedtransitions PermittedStateTransitions) (StateMachine, error) {

	var result stateMachine

	initialStateIsValid := false
	for k := range permittedtransitions {
		if value.GetCurrentState() == k {
			initialStateIsValid = true
			break
		}
		if slices.Contains(permittedtransitions[k], value.GetCurrentState()) {
			initialStateIsValid = true
			break
		}
	}
	if !initialStateIsValid {
		return &result, fmt.Errorf("current state %s is not defined in permitted state transitions", value.GetCurrentState())

	}

	value.SetCheckedState(value.GetCurrentState())
	result.PermittedStateTransitions = permittedtransitions
	result.StateMachinable = value

	return &result, nil

}

type stateMachine struct {
	StateMachinable
	PermittedStateTransitions
}

func (s *stateMachine) Move(event Event, state string) error {

	transition := StateTransition{
		From: s.StateMachinable.GetCurrentState(),
		To:   state,
	}

	isMovePermitted := false
	if transitions, exists := s.PermittedStateTransitions[s.StateMachinable.GetCurrentState()]; exists {
		if slices.Contains(transitions, state) {
			isMovePermitted = true
		}
	}
	if !isMovePermitted {
		return fmt.Errorf("invalid state transition from %s to %s triggered by event: %s", transition.From, transition.To, event.Description)
	}

	s.StateMachinable.AppendLogEntry(event, transition)
	s.StateMachinable.SetCheckedState(state)

	return nil

}

func (s stateMachine) HaveMoved(transition StateTransition) bool {

	for _, entry := range s.StateMachinable.PreviousTransitions() {
		if entry.From == transition.From &&
			entry.To == transition.To {
			return true
		}
	}

	return false

}
