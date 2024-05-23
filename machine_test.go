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

func TestMachine_Run(t *testing.T) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	step1 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Next("Next"), nil
		},
	})
	step2 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Done"), nil
		},
	})
	m.AddStep(*step1)
	m.AddStep(*step2)

	// Run the machine
	response, err := m.Run()

	// Check the result and error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response == nil {
		t.Errorf("Expected result to be a string, got: %v", response)
	}
	if response.Result != "Done" {
		t.Errorf("Expected result to be 'Done', got: %v", response)
	}

	// Check the executed steps
	expectedExecutedSteps := []tango.Step[Services, State]{*step1, *step2}
	if len(m.ExecutedSteps) != len(expectedExecutedSteps) {
		t.Errorf("Expected executed steps to be %v, got: %v", len(expectedExecutedSteps), len(m.ExecutedSteps))
	}
}

func TestMachine_Compensate(t *testing.T) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	step1 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Next("Next"), nil
		},
		Compensate: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Compensated"), nil
		},
	})
	step2 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Error("I will be compensated"), nil
		},
		Compensate: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Done"), nil
		},
	})
	m.AddStep(*step1)
	m.AddStep(*step2)

	// Run the machine
	result, err := m.Run()

	// Check the result and error
	if err.Error() != "execution error at Step2" {
		fmt.Println(err)
		t.Errorf("Expected error to be 'Compensated', got: %v", err)
	}
	if result != nil {
		t.Errorf("Expected result to be nil, got: %v", result)
	}

	// Check the executed steps
	expectedExecutedSteps := []tango.Step[Services, State]{*step1, *step2}
	if len(m.ExecutedSteps) != len(expectedExecutedSteps) {
		t.Errorf("Expected executed steps to be %v, got: %v", len(expectedExecutedSteps), len(m.ExecutedSteps))
	}
}

func TestMachine_Reset(t *testing.T) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	step1 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Next("Next"), nil
		},
	})
	step2 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Done"), nil
		},
	})
	m.AddStep(*step1)
	m.AddStep(*step2)

	// Run the machine
	_, _ = m.Run()

	// Reset the machine
	m.Reset()

	// Check the steps
	if len(m.Steps) != 0 {
		t.Errorf("Expected steps to be empty, got: %v", m.Steps)
	}
	if len(m.ExecutedSteps) != 0 {
		t.Errorf("Expected executed steps to be empty, got: %v", m.ExecutedSteps)
	}
}

func TestMachine_Context_State(t *testing.T) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{State: State{Counter: 0}}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	step1 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			ctx.State.Counter++
			return m.Next("Next"), nil
		},
	})
	step2 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			ctx.State.Counter++
			return m.Done("Done"), nil
		},
	})
	m.AddStep(*step1)
	m.AddStep(*step2)

	// Run the machine
	_, _ = m.Run()

	// Check the state
	if m.Context.State.Counter != 2 {
		t.Errorf("Expected state counter to be 2, got: %v", m.Context.State.Counter)
	}
}

func TestMachine_Context_Services(t *testing.T) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{Services: Services{Database: "MySQL"}}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	step1 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			ctx.Services.Database = "PostgreSQL"
			return m.Next("Next"), nil
		},
	})
	step2 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			ctx.Services.Database = "SQLite"
			return m.Done("Done"), nil
		},
	})
	m.AddStep(*step1)
	m.AddStep(*step2)

	// Run the machine
	_, _ = m.Run()

	// Check the services
	if m.Context.Services.Database != "SQLite" {
		t.Errorf("Expected services database to be 'SQLite', got: %v", m.Context.Services.Database)
	}
}

func TestMachine_Step_Jump(t *testing.T) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	step1 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Jump("Jump", "Step3"), nil
		},
	})
	step2 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Error("I got skipped"), nil
		},
	})
	step3 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step3",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Done"), nil
		},
	})
	m.AddStep(*step1)
	m.AddStep(*step2)
	m.AddStep(*step3)

	// Run the machine
	response, err := m.Run()

	// Check the result and error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response == nil {
		t.Errorf("Expected result to be a string, got: %v", response)
	}
	if response.Result != "Done" {
		t.Errorf("Expected result to be 'Done', got: %v", response)
	}

	// Check the executed steps
	expectedExecutedSteps := []tango.Step[Services, State]{*step1, *step2}
	if len(m.ExecutedSteps) != len(expectedExecutedSteps) {
		t.Errorf("Expected executed steps to be %d, got: %d", len(expectedExecutedSteps), len(m.ExecutedSteps))
	}
}

func TestMachine_Step_Skip(t *testing.T) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	step1 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Skip("Skip", 1), nil
		},
	})
	step2 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Error("I will be skipped"), nil
		},
	})
	step3 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step3",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Done"), nil
		},
	})
	m.AddStep(*step1)
	m.AddStep(*step2)
	m.AddStep(*step3)

	// Run the machine
	response, err := m.Run()

	// Check the result and error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response == nil {
		t.Errorf("Expected result to be a string, got: %v", response)
	}
	if response.Result != "Done" {
		t.Errorf("Expected result to be 'Done', got: %v", response)
	}

	// Check the executed steps
	expectedExecutedSteps := []tango.Step[Services, State]{*step1, *step2}
	if len(m.ExecutedSteps) != len(expectedExecutedSteps) {
		t.Errorf("Expected executed steps to be %d, got: %d", len(expectedExecutedSteps), len(m.ExecutedSteps))
	}
}

func BenchmarkMachine_Run(b *testing.B) {
	// Create a new machine
	m := tango.NewMachine("TestMachine", []tango.Step[Services, State]{}, &tango.MachineContext[Services, State]{}, &tango.MachineConfig[Services, State]{
		Log: false,
	})

	// Add some steps to the machine
	step1 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Next("Next"), nil
		},
	})
	step2 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Done"), nil
		},
	})
	m.AddStep(*step1)
	m.AddStep(*step2)

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
	step1 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Next("Next"), nil
		},
		Compensate: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Compensated"), nil
		},
	})
	step2 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Error("I will be compensated"), nil
		},
		Compensate: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Done"), nil
		},
	})
	m.AddStep(*step1)
	m.AddStep(*step2)

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
	step1 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step1",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Next("Next"), nil
		},
	})
	step2 := m.NewStep(&tango.Step[Services, State]{
		Name: "Step2",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Done"), nil
		},
	})
	m.AddStep(*step1)
	m.AddStep(*step2)

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
		step := m.NewStep(&tango.Step[Services, State]{
			Name: fmt.Sprintf("Step%d", i),
			Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
				return m.Next("Next"), nil
			},
		})
		m.AddStep(*step)
	}

	m.AddStep(*m.NewStep(&tango.Step[Services, State]{
		Name: "LastStep",
		Execute: func(ctx *tango.MachineContext[Services, State]) (*tango.StepResponse[Services, State], error) {
			return m.Done("Done"), nil
		},
	}))

	// Run the machine
	for i := 0; i < b.N; i++ {
		_, _ = m.Run()
	}
}
