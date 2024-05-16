package tango

type StepStatus string

const (
	NEXT  StepStatus = "NEXT"
	DONE  StepStatus = "DONE"
	ERROR StepStatus = "ERROR"
	SKIP  StepStatus = "SKIP"
	JUMP  StepStatus = "JUMP"
)

type StepResponse struct {
	Result     interface{}
	Status     StepStatus
	SkipCount  int
	JumpTarget string
}

func NewStepResponse[Result any](result Result, status StepStatus, skipCount int, jumpTarget string) *StepResponse {
	return &StepResponse{Result: result, Status: status, SkipCount: skipCount, JumpTarget: jumpTarget}
}

func Next[Result any](result Result) *StepResponse {
	return NewStepResponse(result, NEXT, 0, "")
}

func Done[Result any](result Result) *StepResponse {
	return NewStepResponse(result, DONE, 0, "")
}

func Error[Result any](result Result) *StepResponse {
	return NewStepResponse(result, ERROR, 0, "")
}

func Skip[Result any](result Result, count int) *StepResponse {
	return NewStepResponse(result, SKIP, count, "")
}

func Jump[Result any](result Result, target string) *StepResponse {
	return NewStepResponse(result, JUMP, 0, target)
}

type Step struct {
	Name             string
	Execute          func(ctx *MachineContext) (*StepResponse, error)
	BeforeExecute    func(ctx *MachineContext) error
	AfterExecute     func(ctx *MachineContext) error
	Compensate       func(ctx *MachineContext) (*StepResponse, error)
	BeforeCompensate func(ctx *MachineContext) error
	AfterCompensate  func(ctx *MachineContext) error
}

func NewStep(step *Step) *Step {
	return &Step{
		Name:             step.Name,
		Execute:          step.Execute,
		BeforeExecute:    step.BeforeExecute,
		AfterExecute:     step.AfterExecute,
		Compensate:       step.Compensate,
		BeforeCompensate: step.BeforeCompensate,
		AfterCompensate:  step.AfterCompensate,
	}
}
