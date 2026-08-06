package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gravestench/akara"
)

func TestWorldCloseRejectsMutation(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	entity := world.MustCreateEntity()
	if err := world.Close(); err != nil {
		t.Fatal(err)
	}
	if !world.Closed() {
		t.Fatal("world does not report closed")
	}
	if _, err := world.CreateEntity(); !errors.Is(err, akara.ErrWorldClosed) {
		t.Fatalf("CreateEntity error = %v", err)
	}
	if _, err := positions.Add(entity); !errors.Is(err, akara.ErrWorldClosed) {
		t.Fatalf("Add error = %v", err)
	}
	if err := world.Update(time.Second); !errors.Is(err, akara.ErrWorldClosed) {
		t.Fatalf("Update error = %v", err)
	}
	if err := world.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestContextClosesWorld(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	world := akara.NewWorldWithContext(ctx)
	cancel()
	select {
	case <-world.Done():
	case <-time.After(time.Second):
		t.Fatal("context cancellation did not close world")
	}
}

func TestEntitySnapshotsAreIsolated(t *testing.T) {
	world := akara.NewWorld()
	first := world.MustCreateEntity()
	second := world.MustCreateEntity()
	entities := world.Entities()
	if len(entities) != 2 || entities[0] != first || entities[1] != second {
		t.Fatalf("entities = %v", entities)
	}
	entities[0] = 999
	if world.Entities()[0] != first {
		t.Fatal("caller mutated entity snapshot")
	}
}
