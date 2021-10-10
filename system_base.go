package akara

import "time"

// BaseSystem is the base system type
type BaseSystem struct {
	baseSystem
}

type baseSystem struct {
	*World
	active       bool
	tickRate     float64
	lastTick     time.Time
	tickChan     chan interface{}
	preTickFunc  func()
	postTickFunc func()
}

var DefaultTickRate float64 = 100

func (s *BaseSystem) Base() System {
	return &s.baseSystem
}

func (s *baseSystem) IsInitialized() bool {
	return s.World != nil
}

func (s *baseSystem) Init(world *World) {
	s.World = world
	if s.tickRate == 0 {
		s.tickRate = DefaultTickRate
	}
	s.tickChan = make(chan interface{})
}

// Active whether or not the system is active
func (s *baseSystem) Active() bool {
	return s.active
}

// SetActive sets the active flag for the system
func (s *baseSystem) SetActive(b bool) {
	s.active = b
}

// Destroy removes the system from the world
func (s *baseSystem) Destroy() {
	s.World.RemoveSystem(s)
}

// InjectComponent is shorthand for registering a component and placing the factory in the given destination
func (s *baseSystem) InjectComponent(c Component, dst **ComponentFactory) {
	*dst = s.GetComponentFactory(s.RegisterComponent(c))
}

// Update is called once per tick
func (s *baseSystem) Update() {}

// Tick tells the System to tick
func (s *baseSystem) Tick() {
	s.tickChan <- nil
}

// WaitForTick blocks until the System is told to tick
func (s *baseSystem) WaitForTick() {
	for {
		select {
		case <-s.tickChan:
			return
		}
	}
}

func (s *baseSystem) TickRate() float64 {
	return s.tickRate
}

func (s *baseSystem) SetTickRate(rate float64) {
	s.tickRate = rate
}

func (s *baseSystem) SetPreTickFunc(fn func()) {
	s.preTickFunc = fn
}

func (s *baseSystem) SetPostTickFunc(fn func()) {
	s.postTickFunc = fn
}

func (s *baseSystem) PreTickFunc() {
	if s.preTickFunc != nil {
		s.preTickFunc()
	}
}

func (s *baseSystem) PostTickFunc() {
	if s.postTickFunc != nil {
		s.postTickFunc()
	}
}
