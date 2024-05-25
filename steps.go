package tango

type StepStatus string

const (
	NEXT  StepStatus = "NEXT"
	DONE  StepStatus = "DONE"
	ERROR StepStatus = "ERROR"
	SKIP  StepStatus = "SKIP"
	JUMP  StepStatus = "JUMP"
)

type StepResponse[S, T any] struct {
	Result     interface{}
	Status     StepStatus
	SkipCount  int
	JumpTarget string
	NewMachine *Machine[S, T] // New field to allow nested machine execution
}

func NewStepResponse[Result, S, T any](result Result, status StepStatus, skipCount int, jumpTarget string, newMachine *Machine[S, T]) *StepResponse[S, T] {
	return &StepResponse[S, T]{Result: result, Status: status, SkipCount: skipCount, JumpTarget: jumpTarget, NewMachine: newMachine}
}

func Next[Result, S, T any](result Result) *StepResponse[S, T] {
	return NewStepResponse[Result, S, T](result, NEXT, 0, "", nil)
}

func Done[Result, S, T any](result Result) *StepResponse[S, T] {
	return NewStepResponse[Result, S, T](result, DONE, 0, "", nil)
}

func Error[Result, S, T any](result Result) *StepResponse[S, T] {
	return NewStepResponse[Result, S, T](result, ERROR, 0, "", nil)
}

func Skip[Result, S, T any](result Result, count int) *StepResponse[S, T] {
	return NewStepResponse[Result, S, T](result, SKIP, count, "", nil)
}

func Jump[Result, S, T any](result Result, target string) *StepResponse[S, T] {
	return NewStepResponse[Result, S, T](result, JUMP, 0, target, nil)
}

func RunNewMachine[Result, S, T any](result Result, newMachine *Machine[S, T]) *StepResponse[S, T] {
	return NewStepResponse(result, NEXT, 0, "", newMachine)
}

type Step[S, T any] struct {
	Name             string
	Execute          func(ctx *MachineContext[S, T]) (*StepResponse[S, T], error)
	BeforeExecute    func(ctx *MachineContext[S, T]) error
	AfterExecute     func(ctx *MachineContext[S, T]) error
	Compensate       func(ctx *MachineContext[S, T]) (*StepResponse[S, T], error)
	BeforeCompensate func(ctx *MachineContext[S, T]) error
	AfterCompensate  func(ctx *MachineContext[S, T]) error
}

func NewStep[S, T any](step *Step[S, T]) *Step[S, T] {
	return &Step[S, T]{
		Name:             step.Name,
		Execute:          step.Execute,
		BeforeExecute:    step.BeforeExecute,
		AfterExecute:     step.AfterExecute,
		Compensate:       step.Compensate,
		BeforeCompensate: step.BeforeCompensate,
		AfterCompensate:  step.AfterCompensate,
	}
}
