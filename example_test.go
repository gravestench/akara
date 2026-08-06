package akara_test

import (
	"fmt"
	"time"

	"github.com/gravestench/akara"
)

type examplePosition struct{ X, Y float64 }
type exampleVelocity struct{ X, Y float64 }

func ExampleRegister() {
	world := akara.NewWorld()
	defer func() { _ = world.Close() }()
	positions := akara.Register[examplePosition](world)
	entity := world.MustCreateEntity()
	position, _ := positions.Set(entity, examplePosition{X: 12, Y: 4})
	fmt.Println(position.X, position.Y)
	// Output: 12 4
}

func ExampleWorld_Subscribe() {
	world := akara.NewWorld()
	defer func() { _ = world.Close() }()
	positions := akara.Register[examplePosition](world)
	velocities := akara.Register[exampleVelocity](world)
	moving, _ := world.Subscribe(akara.All(positions, velocities))
	defer func() { _ = moving.Close() }()

	entity := world.MustCreateEntity()
	_, _ = positions.Add(entity)
	_, _ = velocities.Set(entity, exampleVelocity{X: 2})
	fmt.Println(moving.Entities())
	// Output: [1]
}

func ExampleWorld_Update() {
	world := akara.NewWorld()
	defer func() { _ = world.Close() }()
	_, _ = world.AddSystem(akara.SystemFunc(func(_ *akara.World, delta time.Duration) error {
		fmt.Println(delta)
		return nil
	}))
	_ = world.Update(16 * time.Millisecond)
	// Output: 16ms
}
