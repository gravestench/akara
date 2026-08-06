package tests

import (
	"errors"
	"sync"
	"testing"

	"github.com/gravestench/akara"
)

type position struct{ X, Y float64 }
type velocity struct{ X, Y float64 }
type frozen struct{}

func TestPlainComponentsAndCanonicalStores(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	if again := akara.Register[position](world); again != positions {
		t.Fatal("Register did not return the canonical typed store")
	}
	if lookedUp, found := akara.GetStore[position](world); !found || lookedUp != positions {
		t.Fatal("GetStore did not find the canonical store")
	}
	if _, found := akara.GetStore[velocity](world); found {
		t.Fatal("GetStore found an unregistered type")
	}

	entity := world.MustCreateEntity()
	component, err := positions.Add(entity)
	if err != nil {
		t.Fatal(err)
	}
	component.X = 12
	stored, found := positions.Get(entity)
	if !found || stored != component || stored.X != 12 {
		t.Fatal("typed component did not round trip")
	}
	if duplicate, err := positions.Add(entity); err != nil || duplicate != component {
		t.Fatal("Add did not preserve the existing component")
	}
}

func TestStoreRejectsUnknownEntity(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	if _, err := positions.Add(999); !errors.Is(err, akara.ErrEntityNotFound) {
		t.Fatalf("Add error = %v, want ErrEntityNotFound", err)
	}
	if positions.Len() != 0 {
		t.Fatal("unknown entity left an orphan component")
	}
}

func TestDestroyEntityRemovesEveryComponent(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	velocities := akara.Register[velocity](world)
	entity := world.MustCreateEntity()
	if _, err := positions.Add(entity); err != nil {
		t.Fatal(err)
	}
	if _, err := velocities.Add(entity); err != nil {
		t.Fatal(err)
	}

	if !world.DestroyEntity(entity) || world.EntityExists(entity) {
		t.Fatal("entity was not destroyed")
	}
	if positions.Has(entity) || velocities.Has(entity) {
		t.Fatal("destroyed entity retained components")
	}
	if world.DestroyEntity(entity) {
		t.Fatal("destroyed entity was removed twice")
	}
}

func TestConcurrentAddReturnsOneComponent(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	entity := world.MustCreateEntity()
	const count = 64
	results := make(chan *position, count)
	var wait sync.WaitGroup
	for index := 0; index < count; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			component, err := positions.Add(entity)
			if err != nil {
				t.Errorf("Add: %v", err)
				return
			}
			results <- component
		}()
	}
	wait.Wait()
	close(results)
	var first *position
	for component := range results {
		if first == nil {
			first = component
		} else if component != first {
			t.Fatal("concurrent Add returned multiple instances")
		}
	}
}

func TestStoreEntitiesAreSortedSnapshots(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	first := world.MustCreateEntity()
	second := world.MustCreateEntity()
	_, _ = positions.Add(second)
	_, _ = positions.Add(first)
	entities := positions.Entities()
	if len(entities) != 2 || entities[0] != first || entities[1] != second {
		t.Fatalf("unexpected store entities: %v", entities)
	}
	entities[0] = 999
	if positions.Entities()[0] != first {
		t.Fatal("caller mutated store entity snapshot")
	}
}
