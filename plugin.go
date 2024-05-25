package tango

type Plugin[S, T any] interface {
	Init(ctx *MachineContext[S, T]) error
	Execute(ctx *MachineContext[S, T]) error
	Cleanup(ctx *MachineContext[S, T]) error
}
