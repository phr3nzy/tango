package tango_test

import (
	"fmt"
	"testing"

	"github.com/phr3nzy/tango"
)

type Services struct {
	Database string
}

type State struct {
	Counter int
}

type testCase struct {
	name              string
	steps             []tango.Step[Services, State]
	expectedResult    string
	expectedError     error
	expectedStepNames []string
}

func TestMachine_Run(t *testing.T) {
	tests := []testCase{
		{
			name: "SimpleRun",
			steps: []tango.Step[Services, State]{
				{
					Name: "Step1",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Next("Next"), nil
					},
				},
				{
					Name: "Step2",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Done("Done"), nil
					},
				},
			},
			expectedResult:    "Done",
			expectedError:     nil,
			expectedStepNames: []string{"Step1", "Step2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{}, &tango.MachineConfig[Services, State]{
				Log: false,
			})

			for _, step := range tt.steps {
				m.AddStep(step)
			}

			response, err := m.Run()

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if response == nil {
				t.Errorf("expected response to be non-nil")
			} else if response.Result != tt.expectedResult {
				t.Errorf("expected result to be %v, got %v", tt.expectedResult, response.Result)
			}

			if len(m.ExecutedSteps) != len(tt.expectedStepNames) {
				t.Errorf("expected %v executed steps, got %v", len(tt.expectedStepNames), len(m.ExecutedSteps))
			}

			for i, step := range m.ExecutedSteps {
				if step.Name != tt.expectedStepNames[i] {
					t.Errorf("expected step %v, got %v", tt.expectedStepNames[i], step.Name)
				}
			}
		})
	}
}

type compensateTestCase struct {
	name              string
	steps             []tango.Step[Services, State]
	expectedError     string
	expectedResult    *tango.Response[Services, State]
	expectedStepNames []string
}

func TestMachine_Compensate(t *testing.T) {
	tests := []compensateTestCase{
		{
			name: "CompensateOnError",
			steps: []tango.Step[Services, State]{
				{
					Name: "Step1",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Next("Next"), nil
					},
					Compensate: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Done("Compensated"), nil
					},
				},
				{
					Name: "Step2",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Error("I will be compensated"), nil
					},
					Compensate: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Done("Done"), nil
					},
				},
			},
			expectedError:     "step Step2 failed: I will be compensated",
			expectedResult:    nil,
			expectedStepNames: []string{"Step1", "Step2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &tango.MachineContext[Services, State]{}
			m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, context, &tango.MachineConfig[Services, State]{
				Log: false,
			})
			context.Machine = m

			for _, step := range tt.steps {
				m.AddStep(step)
			}

			result, err := m.Run()

			if err == nil || err.Error() != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if result != tt.expectedResult {
				t.Errorf("expected result %v, got %v", tt.expectedResult, result)
			}

			if len(m.ExecutedSteps) != len(tt.expectedStepNames) {
				t.Errorf("expected %v executed steps, got %v", len(tt.expectedStepNames), len(m.ExecutedSteps))
			}

			for i, step := range m.ExecutedSteps {
				if step.Name != tt.expectedStepNames[i] {
					t.Errorf("expected step %v, got %v", tt.expectedStepNames[i], step.Name)
				}
			}
		})
	}
}

type compensateStateTestCase struct {
	name              string
	steps             []tango.Step[Services, State]
	expectedError     string
	expectedResult    *tango.Response[Services, State]
	expectedStepNames []string
}

func TestMachine_Compensate_State(t *testing.T) {
	tests := []compensateStateTestCase{
		{
			name: "CompensateStateOnError",
			steps: []tango.Step[Services, State]{
				{
					Name: "Step1",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						ctx.State.Counter++
						return ctx.Machine.Next("Next"), nil
					},
					Compensate: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						ctx.State.Counter--
						return ctx.Machine.Done("Compensated"), nil
					},
				},
				{
					Name: "Step2",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						ctx.State.Counter++
						return ctx.Machine.Error("I will be compensated"), nil
					},
					Compensate: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						ctx.State.Counter--
						return ctx.Machine.Done("Done"), nil
					},
				},
			},
			expectedError:     "step Step2 failed: I will be compensated",
			expectedResult:    nil,
			expectedStepNames: []string{"Step1", "Step2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &tango.MachineContext[Services, State]{State: State{Counter: 0}}
			m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, context, &tango.MachineConfig[Services, State]{
				Log: false,
			})
			context.Machine = m

			for _, step := range tt.steps {
				m.AddStep(step)
			}

			result, err := m.Run()

			if err == nil || err.Error() != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if result != tt.expectedResult {
				t.Errorf("expected result %v, got %v", tt.expectedResult, result)
			}

			if len(m.ExecutedSteps) != len(tt.expectedStepNames) {
				t.Errorf("expected %v executed steps, got %v", len(tt.expectedStepNames), len(m.ExecutedSteps))
			}

			for i, step := range m.ExecutedSteps {
				if step.Name != tt.expectedStepNames[i] {
					t.Errorf("expected step %v, got %v", tt.expectedStepNames[i], step.Name)
				}
			}

			if m.Context.State.Counter != 0 {
				t.Errorf("expected state counter to be 0, got %v", m.Context.State.Counter)
			}

		})
	}
}

type resetTestCase struct {
	name              string
	steps             []tango.Step[Services, State]
	expectedSteps     int
	expectedExecSteps int
}

func TestMachine_Reset(t *testing.T) {
	tests := []resetTestCase{
		{
			name: "ResetAfterRun",
			steps: []tango.Step[Services, State]{
				{
					Name: "Step1",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Next("Next"), nil
					},
				},
				{
					Name: "Step2",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Done("Done"), nil
					},
				},
			},
			expectedSteps:     0,
			expectedExecSteps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &tango.MachineContext[Services, State]{}
			m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, context, &tango.MachineConfig[Services, State]{
				Log: false,
			})
			context.Machine = m

			for _, step := range tt.steps {
				m.AddStep(step)
			}

			_, _ = m.Run()

			m.Reset()

			if len(m.Steps) != tt.expectedSteps {
				t.Errorf("expected steps to be %v, got %v", tt.expectedSteps, len(m.Steps))
			}
			if len(m.ExecutedSteps) != tt.expectedExecSteps {
				t.Errorf("expected executed steps to be %v, got %v", tt.expectedExecSteps, len(m.ExecutedSteps))
			}
		})
	}
}

type stateTestCase struct {
	name            string
	initialState    State
	steps           []tango.Step[Services, State]
	expectedCounter int
}

func TestMachine_Context_State(t *testing.T) {
	tests := []stateTestCase{
		{
			name:         "IncrementCounter",
			initialState: State{Counter: 0},
			steps: []tango.Step[Services, State]{
				{
					Name: "Step1",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						ctx.State.Counter++
						return ctx.Machine.Next("Next"), nil
					},
				},
				{
					Name: "Step2",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						ctx.State.Counter++
						return ctx.Machine.Done("Done"), nil
					},
				},
			},
			expectedCounter: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &tango.MachineContext[Services, State]{State: tt.initialState}
			m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, context, &tango.MachineConfig[Services, State]{
				Log: false,
			})
			context.Machine = m

			for _, step := range tt.steps {
				m.AddStep(step)
			}

			_, _ = m.Run()

			if m.Context.State.Counter != tt.expectedCounter {
				t.Errorf("expected state counter to be %v, got %v", tt.expectedCounter, m.Context.State.Counter)
			}
		})
	}
}

type servicesTestCase struct {
	name             string
	initialServices  Services
	steps            []tango.Step[Services, State]
	expectedDatabase string
}

func TestMachine_Context_Services(t *testing.T) {
	tests := []servicesTestCase{
		{
			name:            "ChangeDatabaseService",
			initialServices: Services{Database: "MySQL"},
			steps: []tango.Step[Services, State]{
				{
					Name: "Step1",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						ctx.Services.Database = "PostgreSQL"
						return ctx.Machine.Next("Next"), nil
					},
				},
				{
					Name: "Step2",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						ctx.Services.Database = "SQLite"
						return ctx.Machine.Done("Done"), nil
					},
				},
			},
			expectedDatabase: "SQLite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &tango.MachineContext[Services, State]{Services: tt.initialServices}
			m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, context, &tango.MachineConfig[Services, State]{
				Log: false,
			})
			context.Machine = m

			for _, step := range tt.steps {
				m.AddStep(step)
			}

			_, _ = m.Run()

			if m.Context.Services.Database != tt.expectedDatabase {
				t.Errorf("expected services database to be %v, got %v", tt.expectedDatabase, m.Context.Services.Database)
			}
		})
	}
}

type jumpTestCase struct {
	name              string
	steps             []tango.Step[Services, State]
	expectedResult    string
	expectedError     error
	expectedStepNames []string
}

func TestMachine_Step_Jump(t *testing.T) {
	tests := []jumpTestCase{
		{
			name: "JumpToStep3",
			steps: []tango.Step[Services, State]{
				{
					Name: "Step1",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Jump("Jump", "Step3"), nil
					},
				},
				{
					Name: "Step2",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Error("I got skipped"), nil
					},
				},
				{
					Name: "Step3",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Done("Done"), nil
					},
				},
			},
			expectedResult:    "Done",
			expectedError:     nil,
			expectedStepNames: []string{"Step1", "Step3"},
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &tango.MachineContext[Services, State]{}
			m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, context, &tango.MachineConfig[Services, State]{
				Log: false,
			})
			context.Machine = m

			for _, step := range tt.steps {
				m.AddStep(step)
			}

			response, err := m.Run()

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if response == nil {
				t.Errorf("expected response to be non-nil")
			} else if response.Result != tt.expectedResult {
				t.Errorf("expected result to be %v, got %v", tt.expectedResult, response.Result)
			}

			if len(m.ExecutedSteps) != len(tt.expectedStepNames) {
				t.Errorf("expected %v executed steps, got %v", len(tt.expectedStepNames), len(m.ExecutedSteps))
			}

			for i, step := range m.ExecutedSteps {
				if step.Name != tt.expectedStepNames[i] {
					t.Errorf("expected step %v, got %v", tt.expectedStepNames[i], step.Name)
				}
			}
		})
	}
}

type stepSkipTestCase struct {
	name                  string
	steps                 []tango.Step[Services, State]
	expectedExecutedSteps []string
	expectedResult        string
	expectedError         error
}

func TestMachine_Step_Skip(t *testing.T) {
	tests := []stepSkipTestCase{
		{
			name: "SkipStep",
			steps: []tango.Step[Services, State]{
				{
					Name: "Step1",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Skip("Skip", 1), nil
					},
				},
				{
					Name: "Step2",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Error("I will be skipped"), nil
					},
				},
				{
					Name: "Step3",
					Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
						return ctx.Machine.Done("Done"), nil
					},
				},
			},
			expectedExecutedSteps: []string{"Step1", "Step3"},
			expectedResult:        "Done",
			expectedError:         nil,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &tango.MachineContext[Services, State]{}
			m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, context, &tango.MachineConfig[Services, State]{
				Log: false,
			})
			context.Machine = m

			for _, step := range tt.steps {
				m.AddStep(step)
			}

			response, err := m.Run()

			if err != tt.expectedError {
				t.Errorf("unexpected error: %v", err)
			}
			if response == nil {
				t.Errorf("expected response to be non-nil, got nil")
			} else if response.Result != tt.expectedResult {
				t.Errorf("expected result to be %v, got %v", tt.expectedResult, response.Result)
			}

			executedStepNames := []string{}
			for _, step := range m.ExecutedSteps {
				executedStepNames = append(executedStepNames, step.Name)
			}
			if len(executedStepNames) != len(tt.expectedExecutedSteps) {
				t.Errorf("expected executed steps to be %v, got %v", tt.expectedExecutedSteps, executedStepNames)
			}
		})
	}
}

func BenchmarkMachine_Run(b *testing.B) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
			return m.Next("Next"), nil
		},
	})
	m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
			return m.Done("Done"), nil
		},
	})

	// Run the machine
	for i := 0; i < b.N; i++ {
		_, _ = m.Run()
	}
}

func BenchmarkMachine_Compensate(b *testing.B) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
			return m.Next("Next"), nil
		},
		Compensate: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
			return m.Done("Compensated"), nil
		},
	})
	m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
			return m.Error("I will be compensated"), nil
		},
		Compensate: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
			return m.Done("Done"), nil
		},
	})

	// Run the machine
	for i := 0; i < b.N; i++ {
		_, _ = m.Run()
	}
}

func BenchmarkMachine_Reset(b *testing.B) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
			return m.Next("Next"), nil
		},
	})
	m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
			return m.Done("Done"), nil
		},
	})

	// Run the machine
	_, _ = m.Run()

	// Reset the machine
	for i := 0; i < b.N; i++ {
		m.Reset()
	}
}

func BenchmarkMachine_100Steps_Run(b *testing.B) {
	// Create a new machine
	m := tango.NewMachine(
		"TestMachine",
		[]tango.Step[Services, State]{},
		&tango.MachineContext[Services, State]{},
		&tango.MachineConfig[Services, State]{
			Log: false,
		})

	// Add 100 steps to the machine
	for i := 0; i < 100; i++ {
		m.NewStep(&tango.Step[Services, State]{
			Name: fmt.Sprintf("Step%d", i),
			Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
				return m.Next("Next"), nil
			},
		})
	}

	m.NewStep(&tango.Step[Services, State]{
		Name: "LastStep",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.Response[Services, State], error) {
			return m.Done("Done"), nil
		},
	})

	// Run the machine
	for i := 0; i < b.N; i++ {
		_, _ = m.Run()
	}
}
