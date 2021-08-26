package akara

import "time"

// System describes the bare minimum of what is considered a system
type System interface {
	Active() bool
	SetActive(bool)
	Destroy()
}

type baseSystem interface {
	System
	bind(*World)
}

// SystemInitializer is a system with an Init method
type SystemInitializer interface {
	System
	Init(*World)
	IsInitialized() bool
}

type SystemUpdater interface {
	System
	Update()
}

type SystemUpdaterTimed interface {
	System
	Update(time.Duration)
}
