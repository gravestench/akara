package akara

import "time"

// System describes the bare minimum of what is considered a system
type System interface {
	systemInfo
	systemControls
	Update()
}

type systemControls interface {
	systemStatusControls
	systemFrequencyControl
}

type systemStatusControls interface {
	Active() bool
	Activate()
	Deactivate()
}

type systemFrequencyControl interface {
	Tick()
	TickFrequency() float64
	TickPeriod() time.Duration
	SetTickFrequency(float64)
	SetPreTickCallback(func())
	SetPostTickCallback(func())
	TickCount() uint
}

type systemInfo interface {
	Name() string
	Uptime() time.Duration
}

type Initializer interface {
	Init(*World)
	IsInitialized() bool
}

type baseSystem interface {
	Init(*World, func())
}

// hasBaseSystem describes a System that is composed of another type of System.
// The Base method returns the System's base system.
// That System's base system can also have its own base system, and so on, creating a composition tree or chain.
type hasBaseSystem interface {
	base() baseSystem
}
