package tango

import (
	"fmt"
	"sync"
)

type MachineContext[S, T any] struct {
	Services       S
	PreviousResult *StepResponse[S, T]
	State          T
	Machine        *Machine[S, T]
}

type MachineConfig[S, T any] struct {
	Log      bool
	LogLevel string
	Plugins  []Plugin[S, T]
}

type Machine[S, T any] struct {
	Name           string
	Context        *MachineContext[S, T]
	Steps          []Step[S, T]
	ExecutedSteps  []Step[S, T]
	InitialContext *MachineContext[S, T]
	Config         *MachineConfig[S, T]
	mu             sync.Mutex
}

func NewMachine[S, T any](
	name string,
	steps []Step[S, T],
	initialContext *MachineContext[S, T],
	config *MachineConfig[S, T],
) *Machine[S, T] {
	m := &Machine[S, T]{
		Name:           name,
		Steps:          steps,
		InitialContext: initialContext,
		Context:        initialContext,
		Config:         config,
	}
	m.Context.Machine = m
	return m
}

func (m *Machine[S, T]) AddStep(step Step[S, T]) {
	m.Steps = append(m.Steps, step)
}

func (m *Machine[S, T]) Reset() {
	m.Steps = nil
	m.Context = m.InitialContext
	m.ExecutedSteps = nil
}

func (m *Machine[S, T]) Run() (*StepResponse[S, T], error) {
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
			return nil, fmt.Errorf("execution error at %s", step.Name)
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

func (m *Machine[S, T]) executeStep(step Step[S, T]) (*StepResponse[S, T], error) {

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

	return response, nil
}

func (m *Machine[S, T]) Compensate() (*StepResponse[S, T], error) {
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

type Result interface{}

func (m *Machine[S, T]) NewStep(step *Step[S, T]) *Step[S, T] {
	return NewStep(step)
}

func (m *Machine[S, T]) Next(result Result) *StepResponse[S, T] {
	return Next[Result, S, T](result)
}

func (m *Machine[S, T]) Done(result Result) *StepResponse[S, T] {
	return Done[Result, S, T](result)
}

func (m *Machine[S, T]) Error(result Result) *StepResponse[S, T] {
	return Error[Result, S, T](result)
}

func (m *Machine[S, T]) Skip(result Result, count int) *StepResponse[S, T] {
	return Skip[Result, S, T](result, count)
}

func (m *Machine[S, T]) Jump(result any, target string) *StepResponse[S, T] {
	return Jump[Result, S, T](result, target)
}
