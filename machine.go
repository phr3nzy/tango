package tango

import (
	"fmt"
)

type MachineContext struct {
	Services       any
	PreviousResult *StepResponse
	State          any
}

type MachineConfig struct {
	Log      bool
	LogLevel string
}

type Machine struct {
	Name           string
	Context        *MachineContext
	Steps          []Step
	ExecutedSteps  []Step
	InitialContext *MachineContext
	Config         *MachineConfig
}

func NewMachine(name string, steps []Step, initialContext *MachineContext, config *MachineConfig) *Machine {
	return &Machine{Name: name, Steps: steps, InitialContext: initialContext, Context: initialContext, Config: config}
}

func (m *Machine) AddStep(step Step) {
	m.Steps = append(m.Steps, step)
}

func (m *Machine) Reset() {
	m.Steps = nil
	m.Context = m.InitialContext
	m.ExecutedSteps = nil
}

func (m *Machine) Run() (*StepResponse, error) {
	if m.Steps == nil || len(m.Steps) == 0 {
		return nil, fmt.Errorf("no steps to execute")
	}

	i := 0
	for i < len(m.Steps) {
		step := m.Steps[i]
		if m.Config.Log {
			fmt.Printf("Executing step: %s\n", step.Name)
		}

		if step.BeforeExecute != nil {
			if err := step.BeforeExecute(m.Context); err != nil {
				return nil, err
			}
		}

		if step.Execute == nil {
			return nil, fmt.Errorf("step %s has no Execute function", step.Name)
		}
		response, err := step.Execute(m.Context)
		if err != nil {
			return nil, err
		}

		if step.AfterExecute != nil {
			if err := step.AfterExecute(m.Context); err != nil {
				return nil, err
			}
		}

		m.ExecutedSteps = append(m.ExecutedSteps, step)
		if m.Config.Log {
			fmt.Printf("Step result for %s: %s\n", step.Name, response.Status)
		}

		switch response.Status {
		case NEXT:
			m.Context.PreviousResult = response
			i++
		case DONE:
			return response, nil
		case ERROR:
			return nil, fmt.Errorf("execution error at %s", step.Name)
		case SKIP:
			i += response.SkipCount + 1
		case JUMP:
			targetIndex := -1
			for index, s := range m.Steps {
				if s.Name == response.JumpTarget {
					targetIndex = index
					break
				}
			}
			if targetIndex >= 0 {
				i = targetIndex
			} else {
				return nil, fmt.Errorf("jump target '%s' not found at %s", response.JumpTarget, step.Name)
			}
		}
	}

	return nil, nil
}

func (m *Machine) Compensate() (*StepResponse, error) {
	m.Context = m.InitialContext
	for i := len(m.ExecutedSteps) - 1; i >= 0; i-- {
		step := m.ExecutedSteps[i]
		if step.BeforeCompensate != nil {
			if err := step.BeforeCompensate(m.Context); err != nil {
				return nil, err
			}
		}
		if step.Compensate == nil {
			return nil, fmt.Errorf("step %s has no Compensate function", step.Name)
		}
		if _, err := step.Compensate(m.Context); err != nil {
			return nil, err
		}
		if step.AfterCompensate != nil {
			if err := step.AfterCompensate(m.Context); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func (m *Machine) NewStep(step *Step) *Step {
	return NewStep(step)
}

func (m *Machine) Next(result any) *StepResponse {
	return Next(result)
}

func (m *Machine) Done(result any) *StepResponse {
	return Done(result)
}

func (m *Machine) Error(result any) *StepResponse {
	return Error(result)
}

func (m *Machine) Skip(result any, count int) *StepResponse {
	return Skip(result, count)
}

func (m *Machine) Jump(result any, target string) *StepResponse {
	return Jump(result, target)
}
