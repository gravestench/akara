package tests

import (
	"sync"
	"testing"

	"github.com/gravestench/akara"
)

func TestConcurrentEntityLifecycle(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	velocities := akara.Register[velocity](world)
	moving, err := world.Subscribe(akara.All(positions, velocities))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = moving.Close() }()

	const workers = 8
	const entitiesPerWorker = 250
	var wait sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for index := 0; index < entitiesPerWorker; index++ {
				entity, createErr := world.CreateEntity()
				if createErr != nil {
					t.Errorf("CreateEntity: %v", createErr)
					return
				}
				if _, addErr := positions.Add(entity); addErr != nil {
					t.Errorf("positions.Add: %v", addErr)
					return
				}
				if _, addErr := velocities.Add(entity); addErr != nil {
					t.Errorf("velocities.Add: %v", addErr)
					return
				}
				world.DestroyEntity(entity)
			}
		}()
	}
	wait.Wait()
	if len(world.Entities()) != 0 || positions.Len() != 0 || velocities.Len() != 0 || moving.Len() != 0 {
		t.Fatalf("lifecycle leaked state: entities=%d positions=%d velocities=%d moving=%d",
			len(world.Entities()), positions.Len(), velocities.Len(), moving.Len())
	}
}

func BenchmarkCreateAddSubscribeDestroy(b *testing.B) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	velocities := akara.Register[velocity](world)
	moving, err := world.Subscribe(akara.All(positions, velocities))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = moving.Close() })
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		entity, _ := world.CreateEntity()
		_, _ = positions.Add(entity)
		_, _ = velocities.Add(entity)
		world.DestroyEntity(entity)
	}
}
