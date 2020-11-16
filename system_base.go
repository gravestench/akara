package akara

// BaseSystem is the base system type
type BaseSystem struct {
	*World
	active bool
}

// Active whether or not the system is active
func (s *BaseSystem) Active() bool {
	return s.active
}

// SetActive sets the active flag for the system
func (s *BaseSystem) SetActive(b bool) {
	s.active = b
}

// Destroy the system and remove it from the world
func (s *BaseSystem) Destroy() {
	s.World.RemoveSystem(s)
}
