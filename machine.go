package tango

import (
	"fmt"
	"sync"
)

// ResponseStatus is a type that represents the status of a response.
type MachineContext[Services, State any] struct {
	Services       Services
	PreviousResult *Response[Services, State]
	State          State
	Machine        *Machine[Services, State]
}

// Plugin is an interface that represents a machine plugin.
type MachineConfig[Services, State any] struct {
	Log      bool
	LogLevel string
	Plugins  []Plugin[Services, State]
}

// Machine is a struct that represents a machine.
type Machine[Services, State any] struct {
	Name           string
	Context        *MachineContext[Services, State]
	Steps          []Step[Services, State]
	ExecutedSteps  []Step[Services, State]
	InitialContext *MachineContext[Services, State]
	Config         *MachineConfig[Services, State]
	mu             sync.Mutex
}

// NewMachine creates a new machine.
func NewMachine[Services, State any](
	name string,
	steps []Step[Services, State],
	initialContext *MachineContext[Services, State],
	config *MachineConfig[Services, State],
) *Machine[Services, State] {
	m := &Machine[Services, State]{
		Name:           name,
		Steps:          steps,
		InitialContext: initialContext,
		Context:        initialContext,
		Config:         config,
	}
	m.Context.Machine = m
	return m
}

// AddStep adds a step to the machine.
func (m *Machine[Services, State]) AddStep(step Step[Services, State]) {
	m.Steps = append(m.Steps, step)
}

// Reset resets the machine to its initial state. It clears the context and executed steps.
func (m *Machine[Services, State]) Reset() {
	m.Steps = nil
	m.Context = m.InitialContext
	m.ExecutedSteps = nil
}

// Run executes the machine steps.
func (m *Machine[Services, State]) Run() (*Response[Services, State], error) {
	if len(m.Steps) == 0 {
		return nil, fmt.Errorf("no steps to execute")
	}

	for i := 0; i < len(m.Steps); i++ {
		step := m.Steps[i]

		response, err := m.executeStep(step)
		if err != nil {
			return nil, err
		}

		m.mu.Lock()
		m.ExecutedSteps = append(m.ExecutedSteps, step)
		m.Context.PreviousResult = response
		m.mu.Unlock()

		switch response.Status {
		case NEXT:
			continue
		case DONE:
			return response, nil
		case ERROR:
			cResponse, err := m.Compensate()
			if err != nil {
				return nil, fmt.Errorf("compensate error: %v", err)
			}
			return cResponse, fmt.Errorf("step %s failed: %v", step.Name, response.Result)
		case SKIP:
			i += response.SkipCount
		case JUMP:
			targetIndex := -1
			for index, s := range m.Steps {
				if s.Name == response.JumpTarget {
					targetIndex = index
					break
				}
			}
			if targetIndex >= 0 {
				i = targetIndex - 1
			} else {
				return nil, fmt.Errorf("jump target '%s' not found at %s", response.JumpTarget, step.Name)
			}
		}
	}

	for _, plugin := range m.Config.Plugins {
		if err := plugin.Cleanup(m.Context); err != nil {
			return nil, fmt.Errorf("plugin cleanup error: %v", err)
		}
	}

	return nil, nil
}

// executeStep runs the step and its before and after functions.
func (m *Machine[Services, State]) executeStep(step Step[Services, State]) (*Response[Services, State], error) {
	if m.Config.Log {
		fmt.Printf("Executing step: %s\n", step.Name)
	}

	for _, plugin := range m.Config.Plugins {
		if err := plugin.Execute(m.Context); err != nil {
			return nil, fmt.Errorf("plugin before step error: %v", err)
		}
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

	return response, nil
}

// Compensate runs the compensate functions of the executed steps.
func (m *Machine[Services, State]) Compensate() (*Response[Services, State], error) {
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

// Result is an alias for any.
type Result interface{}

// NewStep creates a new step.
func (m *Machine[Services, State]) NewStep(step *Step[Services, State]) {
	m.AddStep(*NewStep(step))
}

// Next creates a response with status NEXT.
func (m *Machine[Services, State]) Next(result Result) *Response[Services, State] {
	return Next[Result, Services, State](result)
}

// Done creates a response with status DONE.
func (m *Machine[Services, State]) Done(result Result) *Response[Services, State] {
	return Done[Result, Services, State](result)
}

// Error creates a response with status ERROR.
func (m *Machine[Services, State]) Error(result Result) *Response[Services, State] {
	return Error[Result, Services, State](result)
}

// Skip creates a response with status SKIP.
func (m *Machine[Services, State]) Skip(result Result, count int) *Response[Services, State] {
	return Skip[Result, Services, State](result, count)
}

// Jump creates a response with status JUMP.
func (m *Machine[Services, State]) Jump(result any, target string) *Response[Services, State] {
	return Jump[Result, Services, State](result, target)
}
