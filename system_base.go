package akara

func NewBaseSystem() *BaseSystem {
	return &BaseSystem{
		Events: NewEventEmitter(),
		active: true,
	}
}

type hasBaseSystem interface {
	Base() *BaseSystem
}

// BaseSystem is the base system type
type BaseSystem struct {
	*World
	Events *EventEmitter
	active bool
}

func (s *BaseSystem) bind(world *World) {
	s.World = world
}

func (s *BaseSystem) Base() *BaseSystem {
	return s
}

// Active whether or not the system is active
func (s *BaseSystem) Active() bool {
	return s.active
}

// SetActive sets the active flag for the system
func (s *BaseSystem) SetActive(b bool) {
	s.active = b
}

// Destroy removes the system from the world
func (s *BaseSystem) Destroy() {
	s.World.RemoveSystem(s)
}

// InjectComponent is shorthand for registering a component and placing the factory in the given destination
func (s *BaseSystem) InjectComponent(c Component, dst **ComponentFactory) {
	*dst = s.GetComponentFactory(s.RegisterComponent(c))
}
