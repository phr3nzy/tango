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
	Strategy       ExecutionStrategy[Services, State]
}

// NewMachine creates a new machine.
func NewMachine[Services, State any](
	name string,
	steps []Step[Services, State],
	initialContext *MachineContext[Services, State],
	config *MachineConfig[Services, State],
	strategy ExecutionStrategy[Services, State],
) *Machine[Services, State] {
	m := &Machine[Services, State]{
		Name:           name,
		Steps:          steps,
		InitialContext: initialContext,
		Context:        initialContext,
		Config:         config,
		Strategy:       strategy,
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

	for _, plugin := range m.Config.Plugins {
		if err := plugin.Init(m.Context); err != nil {
			return nil, fmt.Errorf("plugin setup error: %v", err)
		}
		newStrategy := plugin.ModifyExecutionStrategy(m)
		if newStrategy != nil {
			m.Strategy = newStrategy
		}
	}

	response, err := m.Strategy.Execute(m)
	if err != nil {
		return nil, err
	}

	for _, plugin := range m.Config.Plugins {
		if err := plugin.Cleanup(m.Context); err != nil {
			return nil, fmt.Errorf("plugin cleanup error: %v", err)
		}
	}

	return response, nil
}

// executeStep runs the step and its before and after functions.
func (m *Machine[Services, State]) executeStep(step Step[Services, State]) (*Response[Services, State], error) {
	if m.Config.Log {
		fmt.Printf("executing step: %s\n", step.Name)
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
		return nil, fmt.Errorf("step %s has no execute function", step.Name)
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
	return m.Strategy.Compensate(m)
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
