package tests

import (
	"testing"

	"github.com/gravestench/akara"
)

func TestEntitiesShareCanonicalArchetypes(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	velocities := akara.Register[velocity](world)
	moving, err := world.Subscribe(akara.All(positions, velocities))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = moving.Close() }()

	const entityCount = 10_000
	entities := make([]akara.Entity, entityCount)
	for index := range entities {
		entities[index] = world.MustCreateEntity()
		_, _ = positions.Add(entities[index])
		_, _ = velocities.Add(entities[index])
	}
	// Empty, Position, and Position+Velocity are shared by all 10,000 entities.
	if got := world.ArchetypeCount(); got != 3 {
		t.Fatalf("archetype count = %d, want 3", got)
	}
	if moving.Len() != entityCount {
		t.Fatalf("moving entities = %d, want %d", moving.Len(), entityCount)
	}

	for _, entity := range entities {
		velocities.Remove(entity)
	}
	if got := world.ArchetypeCount(); got != 3 {
		t.Fatalf("reusing Position archetype changed count to %d", got)
	}
	if moving.Len() != 0 {
		t.Fatal("subscription did not follow archetype transitions")
	}
}

func TestNewArchetypeIsCachedByExistingSubscription(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	subscription, err := world.Subscribe(akara.All(positions))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = subscription.Close() }()
	entity := world.MustCreateEntity()
	_, _ = positions.Add(entity) // Creates the archetype after the subscription.
	if got := subscription.Entities(); len(got) != 1 || got[0] != entity {
		t.Fatalf("subscription entities = %v", got)
	}
}
