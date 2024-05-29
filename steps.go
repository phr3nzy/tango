package tango

// ResponseStatus is a type that represents the status of a response.
type ResponseStatus string

// ResponseStatus is a type that represents the status of a response.
const (
	NEXT  ResponseStatus = "NEXT"
	DONE  ResponseStatus = "DONE"
	ERROR ResponseStatus = "ERROR"
	SKIP  ResponseStatus = "SKIP"
	JUMP  ResponseStatus = "JUMP"
)

// Response is a struct that represents the response of a step execution.
type Response[State, Services any] struct {
	Result     interface{}
	Status     ResponseStatus
	SkipCount  int
	JumpTarget string
	NewMachine *Machine[State, Services] // New field to allow nested machine execution
}

// NewResponse creates a new response.
func NewResponse[Result, State, Services any](result Result, status ResponseStatus, skipCount int, jumpTarget string, newMachine *Machine[State, Services]) *Response[State, Services] {
	return &Response[State, Services]{Result: result, Status: status, SkipCount: skipCount, JumpTarget: jumpTarget, NewMachine: newMachine}
}

// Next creates a response with status NEXT.
func Next[Result, State, Services any](result Result) *Response[State, Services] {
	return NewResponse[Result, State, Services](result, NEXT, 0, "", nil)
}

// Done creates a response with status DONE.
func Done[Result, State, Services any](result Result) *Response[State, Services] {
	return NewResponse[Result, State, Services](result, DONE, 0, "", nil)
}

// Error creates a response with status ERROR.
func Error[Result, State, Services any](result Result) *Response[State, Services] {
	return NewResponse[Result, State, Services](result, ERROR, 0, "", nil)
}

// Skip creates a response with status SKIP.
func Skip[Result, State, Services any](result Result, count int) *Response[State, Services] {
	return NewResponse[Result, State, Services](result, SKIP, count, "", nil)
}

// Jump creates a response with status JUMP.
func Jump[Result, State, Services any](result Result, target string) *Response[State, Services] {
	return NewResponse[Result, State, Services](result, JUMP, 0, target, nil)
}

// RunNewMachine creates a response with status NEXT and a new machine.
func RunNewMachine[Result, State, Services any](result Result, newMachine *Machine[State, Services]) *Response[State, Services] {
	return NewResponse(result, NEXT, 0, "", newMachine)
}

// Step is a struct that represents a step in a machine.
type Step[State, Services any] struct {
	Name             string
	Execute          func(ctx *MachineContext[State, Services]) (*Response[State, Services], error)
	BeforeExecute    func(ctx *MachineContext[State, Services]) error
	AfterExecute     func(ctx *MachineContext[State, Services]) error
	Compensate       func(ctx *MachineContext[State, Services]) (*Response[State, Services], error)
	BeforeCompensate func(ctx *MachineContext[State, Services]) error
	AfterCompensate  func(ctx *MachineContext[State, Services]) error
}

// NewStep creates a new step.
func NewStep[State, Services any](step *Step[State, Services]) *Step[State, Services] {
	return &Step[State, Services]{
		Name:             step.Name,
		Execute:          step.Execute,
		BeforeExecute:    step.BeforeExecute,
		AfterExecute:     step.AfterExecute,
		Compensate:       step.Compensate,
		BeforeCompensate: step.BeforeCompensate,
		AfterCompensate:  step.AfterCompensate,
	}
}
