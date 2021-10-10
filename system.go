package akara

// System describes the bare minimum of what is considered a system
type System interface {
	Init(*World)
	IsInitialized() bool
	Active() bool
	SetActive(bool)
	Destroy()
	Update()

	Tick()
	WaitForTick()
	TickRate() float64
	SetTickRate(float64)
	PreTickFunc()
	PostTickFunc()
	SetPreTickFunc(func())
	SetPostTickFunc(func())
}

// hasBaseSystem describes a System that is composed of another type of System.
// For example, BaseSubscriberSystem is composed of/based off of BaseSystem.
// The Base method returns the System's base system.
// That System's base system can also have its own base system, and so on, creating a composition tree or chain.
type hasBaseSystem interface {
	Base() System
}
