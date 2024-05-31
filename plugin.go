package tango

type Plugin[Services, State any] struct {
	Init                    func(ctx *MachineContext[Services, State]) error
	Execute                 func(ctx *MachineContext[Services, State]) error
	Cleanup                 func(ctx *MachineContext[Services, State]) error
	ModifyExecutionStrategy func(m *Machine[Services, State]) ExecutionStrategy[Services, State]
}
