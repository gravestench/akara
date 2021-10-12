package akara

import "time"

// BaseSystem is the base system type
type BaseSystem struct {
	*World
	TimeDelta        time.Duration
	active           bool
	tickFrequency    float64
	tickPeriod		 time.Duration
	lastTick         time.Time
	preTickCallback  func()
	tickCallback	 func()
	postTickCallback func()
}

var DefaultTickRate float64 = 100

func (s *BaseSystem) base() baseSystem {
	return s
}

func (s *BaseSystem) Name() string {
	return "BaseSystem"
}

func (s *BaseSystem) IsInitialized() bool {
	return s.World != nil
}

func (s *BaseSystem) Init(world *World, tickFunc func()) {
	s.World = world
	s.tickCallback = tickFunc

	if s.tickFrequency == 0 {
		s.SetTickFrequency(DefaultTickRate)
	}
}

// Active returns true if the system is active, otherwise false
func (s *BaseSystem) Active() bool {
	return s.active
}

// Deactivate marks the system inactive and stops it from ticking automatically in the background.
// The system can be re-activated by calling the Activate method.
func (s *BaseSystem) Deactivate() {
	s.active = false
}

// Activate calls Tick repeatedly at the target TickRate. This method blocks the thread.
func (s *BaseSystem) Activate() {
	s.active = true

	go func() {
		ticker := time.NewTicker(s.TickPeriod())

		for range ticker.C {
			if !s.Active() {
				break
			}

			s.Tick()
		}

		ticker.Stop()
	}()
}

// InjectComponent is shorthand for registering a component and placing the factory in the given destination
func (s *BaseSystem) InjectComponent(c Component, dst **ComponentFactory) {
	*dst = s.GetComponentFactory(s.RegisterComponent(c))
}

// Tick performs a single tick. This is called automatically when the System is Active, but can be called manually
// to single-step the System, regardless of the System's TickRate.
func (s *BaseSystem) Tick() {
	s.preTickFunc()
	s.tickFunc()
	s.postTickFunc()

	s.TimeDelta = time.Since(s.lastTick)
	s.lastTick = time.Now()
}

// TickPeriod returns the length of one tick as a time.Duration
func (s *BaseSystem) TickPeriod() time.Duration {
	return s.tickPeriod
}

// TickFrequency returns the maximum number of ticks per second this system will perform.
func (s *BaseSystem) TickFrequency() float64 {
	return s.tickFrequency
}

func (s *BaseSystem) SetTickFrequency(rate float64) {
	s.tickFrequency = rate
	s.tickPeriod = calculateTickPeriod(rate)
}

func (s *BaseSystem) SetPreTickCallback(fn func()) {
	s.preTickCallback = fn
}

func (s *BaseSystem) SetPostTickCallback(fn func()) {
	s.postTickCallback = fn
}

func (s *BaseSystem) preTickFunc() {
	if s.preTickCallback != nil {
		s.preTickCallback()
	}
}

func (s *BaseSystem) tickFunc() {
	if s.tickCallback != nil {
		s.tickCallback()
	}
}

func (s *BaseSystem) postTickFunc() {
	if s.postTickCallback != nil {
		s.postTickCallback()
	}
}

func calculateTickPeriod(freq float64) time.Duration {
	return time.Duration(float64(time.Second) * (1 / freq))
}
