package tests

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/gravestench/akara"
)

func TestSystemsRunDeterministically(t *testing.T) {
	world := akara.NewWorld()
	order := make([]int, 0, 2)
	first, err := world.AddSystem(akara.SystemFunc(func(_ *akara.World, delta time.Duration) error {
		if delta != 16*time.Millisecond {
			t.Fatalf("delta = %v", delta)
		}
		order = append(order, 1)
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	second, err := world.AddSystem(akara.SystemFunc(func(_ *akara.World, _ time.Duration) error {
		order = append(order, 2)
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := world.Update(16 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []int{1, 2}) {
		t.Fatalf("system order = %v", order)
	}
	if !reflect.DeepEqual(world.Systems(), []akara.SystemID{first, second}) {
		t.Fatalf("system IDs = %v", world.Systems())
	}
	if !world.RemoveSystem(first) || world.RemoveSystem(first) {
		t.Fatal("RemoveSystem result is incorrect")
	}
}

func TestSystemErrorStopsUpdate(t *testing.T) {
	world := akara.NewWorld()
	want := errors.New("boom")
	_, _ = world.AddSystem(akara.SystemFunc(func(*akara.World, time.Duration) error { return want }))
	called := false
	_, _ = world.AddSystem(akara.SystemFunc(func(*akara.World, time.Duration) error {
		called = true
		return nil
	}))
	err := world.Update(time.Second)
	if !errors.Is(err, want) || called {
		t.Fatalf("Update error = %v, later called = %v", err, called)
	}
}

func TestNilSystemRejected(t *testing.T) {
	world := akara.NewWorld()
	var system akara.SystemFunc
	if _, err := world.AddSystem(system); !errors.Is(err, akara.ErrNilSystem) {
		t.Fatalf("AddSystem error = %v, want ErrNilSystem", err)
	}
}

func TestMovementSystem(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	velocities := akara.Register[velocity](world)
	moving, _ := world.Subscribe(akara.All(positions, velocities))
	defer func() { _ = moving.Close() }()
	_, err := world.AddSystem(akara.SystemFunc(func(_ *akara.World, delta time.Duration) error {
		seconds := delta.Seconds()
		for _, entity := range moving.Entities() {
			position, hasPosition := positions.Get(entity)
			velocity, hasVelocity := velocities.Get(entity)
			if hasPosition && hasVelocity {
				position.X += velocity.X * seconds
				position.Y += velocity.Y * seconds
			}
		}
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	entity := world.MustCreateEntity()
	_, _ = positions.Set(entity, position{X: 10, Y: 20})
	_, _ = velocities.Set(entity, velocity{X: 2, Y: -4})
	if err := world.Update(500 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	got, _ := positions.Get(entity)
	if got.X != 11 || got.Y != 18 {
		t.Fatalf("position = %+v", got)
	}
}
