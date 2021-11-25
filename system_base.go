package akara

import "time"

type timeManagement struct {
	TimeDelta        time.Duration
	tickFrequency    float64
	tickPeriod		 time.Duration
	lastTick         time.Time
	preTickCallback  func()
	tickCallback	 func()
	postTickCallback func()
}

type systemDebugging struct {
	tickCount uint
	uptime time.Duration
}

// BaseSystem is the base system type
type BaseSystem struct {
	*World
	timeManagement
	systemDebugging
	active           bool
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

	// prevent the system from thinking that the last tick was 1970-01-01...
	s.lastTick = time.Now()

	go s.startTicking()
}

func (s *BaseSystem) startTicking() {
	// if the TickFrequency is set, try to tick that frequently. Otherwise, we tick as fast as we can
	if s.TickFrequency() != 0 {
		ticker := time.NewTicker(s.TickPeriod())

		for range ticker.C {
			if !s.Active() {
				break
			}

			s.Tick()
		}

		ticker.Stop()
	} else {
		for {
			if !s.Active() {
				break
			}

			s.Tick()
		}
	}
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
	s.tickCount += 1
	s.uptime += s.TimeDelta
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
	if rate != 0 {
		s.tickPeriod = calculateTickPeriod(rate)
	} else {
		s.tickPeriod = 0
	}
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

func (s *BaseSystem) TickCount() uint {
	return s.tickCount
}

func (s *BaseSystem) Uptime() time.Duration {
	return s.uptime
}

func calculateTickPeriod(freq float64) time.Duration {
	return time.Duration(float64(time.Second) * (1 / freq))
}
