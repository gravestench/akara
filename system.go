package akara

// System describes the bare minimum of what is considered a system
type System interface {
	Active() bool
	SetActive(bool)
	Process()
}

// SystemInitializer is a system with an Init method
type SystemInitializer interface {
	System
	Init(*World)
}
