package tests

import (
	"fmt"
	"github.com/gravestench/akara"
	"math"
	"math/rand"
	"testing"
	"time"
)

func between(min, max float64) float64 {
	range_ := max - min

	n := rand.Float64()

	return math.Round(min + (n * range_))
}

func Test_ExampleMovementSystem(t *testing.T) {
	sys := &MovementSystem{}
	systemTicks := 0
	sys.SetPostTickCallback(func() {
		systemTicks += 1
	})

	cfg := akara.NewWorldConfig().With(sys)
	world := akara.NewWorld(cfg)

	const numEntities = 4

	for idx := 0; idx < numEntities; idx++ {
		e := world.NewEntity()
		p := sys.AddPosition(e)
		v := sys.AddVelocity(e)

		p.X, p.Y = between(-10, 10), between(-10, 10)
		v.X, v.Y = between(-10, 10), between(-10, 10)
	}

	world.Update()

	numUpdates := 4
	loopsWaited := 0
	for systemTicks < numUpdates {
		if loopsWaited > 5 {
			t.Fail()
		}

		time.Sleep(10 * time.Millisecond)
		loopsWaited += 1
	}
}

func BenchmarkExampleMovementSystem(b *testing.B) {
	tests := []int{
		10,
		100,
		1000,
		10000,
	}

	for _, n := range tests {
		name := fmt.Sprintf("%d entities", n)
		b.Run(name, func(b *testing.B) {
			Bench_ExampleMovementSystemN(n, b)
		})
	}
}

func Bench_ExampleMovementSystemN(numEntities int, b *testing.B) {
	sys := &MovementSystem{}

	sys.disableLog = true

	cfg := akara.NewWorldConfig().With(sys)
	world := akara.NewWorld(cfg)

	for idx := 0; idx < numEntities; idx++ {
		e := world.NewEntity()
		p := sys.AddPosition(e)
		v := sys.AddVelocity(e)

		p.X, p.Y = between(-10, 10), between(-10, 10)
		v.X, v.Y = between(-10, 10), between(-10, 10)
	}

	numUpdates := b.N
	for numUpdates > 0 {
		err := world.Update()
		if err != nil {
			b.Errorf("failed to update world: %s", err)
			b.Fail()
		}
		numUpdates--
	}
}

// static check that MovementSystem implements the System interface
var _ akara.System = &MovementSystem{}

// MovementSystem handles entity movement based on velocity and position components
type MovementSystem struct {
	akara.BaseSystem
	PositionFactory
	VelocityFactory
	movableEntities *akara.Subscription
	disableLog      bool
}

// Init initializes the system with the given world
func (m *MovementSystem) Init(_ *akara.World) {
	filter := m.NewComponentFilter().
		Require(
			&Position{},
			&Velocity{},
		).Build()

	m.movableEntities = m.AddSubscription(filter)

	positionID := m.RegisterComponent(&Position{})
	velocityID := m.RegisterComponent(&Velocity{})

	m.Position = m.GetComponentFactory(positionID)
	m.Velocity = m.GetComponentFactory(velocityID)
}

// Update positions of all entities with their velocities
func (m *MovementSystem) Update() {
	entities := m.movableEntities.GetEntities()

	for entIdx := range entities {
		m.move(entities[entIdx])
	}
}

// move updates an individual entity in the movement system
func (m *MovementSystem) move(id akara.EID) {
	p, found := m.GetPosition(id)
	if !found {
		return
	}

	v, found := m.GetVelocity(id)
	if !found {
		return
	}

	s := float64(m.TimeDelta) / float64(time.Second)
	newX := p.X + (v.X * s)
	newY := p.Y + (v.Y * s)

	const strFmt = "p(%+.0f, %+.0f) + v(%+.0f, %+.0f)@%.0fs => p(%+.0f, %+.0f)\n"
	if !m.disableLog {
		fmt.Printf(strFmt, p.X, p.Y, v.X, v.Y, s, newX, newY)
	}

	p.X = newX
	p.Y = newY
}

// static check that Velocity implements Component
var _ akara.Component = &Velocity{}

// Velocity contains an embedded velocity as a vector
type Velocity struct {
	X, Y float64
}

// New creates a new Velocity. By default, the velocity is (0,0).
func (*Velocity) New() akara.Component {
	return &Velocity{}
}

// VelocityFactory is a wrapper for the generic component factory that returns Velocity component instances.
// This can be embedded inside of a system to give them the methods for adding, retrieving, and removing a Velocity.
type VelocityFactory struct {
	Velocity *akara.ComponentFactory
}

// AddVelocity adds a Velocity component to the given entity and returns it
func (m *VelocityFactory) AddVelocity(id akara.EID) *Velocity {
	return m.Velocity.Add(id).(*Velocity)
}

// GetVelocity returns the Velocity component for the given entity, and a bool for whether or not it exists
func (m *VelocityFactory) GetVelocity(id akara.EID) (*Velocity, bool) {
	component, found := m.Velocity.Get(id)
	return component.(*Velocity), found
}

// static check that Position implements Component
var _ akara.Component = &Position{}

// Position contains an embedded d2vector.Position, which is a vector with
// helper methods for translating between screen, isometric, tile, and sub-tile space.
type Position struct {
	X, Y float64
}

// New creates a new Position. By default, the position is (0,0)
func (*Position) New() akara.Component {
	return &Position{}
}

// PositionFactory is a wrapper for the generic component factory that returns Position component instances.
// This can be embedded inside of a system to give them the methods for adding, retrieving, and removing a Position.
type PositionFactory struct {
	Position *akara.ComponentFactory
}

// AddPosition adds a Position component to the given entity and returns it
func (m *PositionFactory) AddPosition(id akara.EID) *Position {
	return m.Position.Add(id).(*Position)
}

// GetPosition returns the Position component for the given entity, and a bool for whether or not it exists
func (m *PositionFactory) GetPosition(id akara.EID) (*Position, bool) {
	component, found := m.Position.Get(id)
	return component.(*Position), found
}
