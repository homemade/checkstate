package checkstate

import (
	"time"
)

type ModelWithState struct {
	State              string                    `json:"string" bson:"string"`
	StateTransitionLog []StateTransitionLogEntry `json:"stateTransitionLog" bson:"stateTransitionLog"`
}

type StateTransitionLogEntry struct {
	Timestamp   time.Time `json:"timestamp" bson:"timestamp"`
	From        string    `json:"from" bson:"from"`
	To          string    `json:"to" bson:"to"`
	Description string    `json:"description" bson:"description"`
}

func (m *ModelWithState) SetCheckedState(state string) {

	m.State = state

}

func (m ModelWithState) GetCurrentState() string {

	return m.State

}

func (m ModelWithState) PreviousTransitions() []StateTransition {

	var result []StateTransition

	for _, l := range m.StateTransitionLog {
		result = append(result, StateTransition{
			From: l.From,
			To:   l.To,
		})
	}

	return result

}

func (m *ModelWithState) AppendLogEntry(event Event, transition StateTransition) error {

	m.StateTransitionLog = append(m.StateTransitionLog, StateTransitionLogEntry{
		Timestamp:   event.Timestamp,
		From:        transition.From,
		To:          transition.To,
		Description: event.Description,
	})

	return nil

}
