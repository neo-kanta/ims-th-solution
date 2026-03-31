package workflow

import (
	"context"

	"github.com/go-chi/chi/v5"
)

type State int

const (
	Not_Started_State = iota
	Investment_Day_Start
	Manager_Approval
	Transaction_Closing
	Accounting_Closing
)

type Event int

const (
	Start_Day_Event = iota
	Manager_Approval_Event
	Close_Transactions_Event
	Close_Accounting_Event
)

type StateMachine struct {
	currentState State
	trnasitions  map[State]map[Event]State
	action       func(ctx context.Context, event Event) error
}

// Module Workflow Management — day-start, manager-approval, transaction-closing, accounting-closing.
type Module struct {
	// Dependencies will be injected here during wire-up.
}

// NewModule creates a new workflow module with its dependencies.
func NewModule() *Module {
	return &Module{}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(r chi.Router) {
	// TODO: register routes for workflow module
}
