package tests

import (
	"errors"
	"testing"

	"github.com/gravestench/akara"
)

func TestSubscriptionMembershipTracksComponents(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	velocities := akara.Register[velocity](world)
	frozenStore := akara.Register[frozen](world)
	subscription, err := world.Subscribe(akara.All(positions, velocities), akara.None(frozenStore))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = subscription.Close() }()

	entity := world.MustCreateEntity()
	if _, err := positions.Add(entity); err != nil {
		t.Fatal(err)
	}
	if subscription.Len() != 0 {
		t.Fatal("partially matching entity subscribed")
	}
	if _, err := velocities.Add(entity); err != nil {
		t.Fatal(err)
	}
	if subscription.Len() != 1 {
		t.Fatal("matching entity did not subscribe")
	}
	if _, err := frozenStore.Add(entity); err != nil {
		t.Fatal(err)
	}
	if subscription.Len() != 0 {
		t.Fatal("forbidden component did not unsubscribe entity")
	}
	if !frozenStore.Remove(entity) || subscription.Len() != 1 {
		t.Fatal("removing forbidden component did not restore entity")
	}
	if !velocities.Remove(entity) || subscription.Len() != 0 {
		t.Fatal("removing required component did not unsubscribe entity")
	}
}

func TestAnyAndEmptyFilters(t *testing.T) {
	world := akara.NewWorld()
	positions := akara.Register[position](world)
	velocities := akara.Register[velocity](world)
	first := world.MustCreateEntity()
	second := world.MustCreateEntity()
	_, _ = positions.Add(first)
	_, _ = velocities.Add(second)

	any, err := world.Subscribe(akara.Any(positions, velocities))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = any.Close() }()
	if any.Len() != 2 {
		t.Fatalf("Any filter contains %d entities, want 2", any.Len())
	}

	allEntities, err := world.Subscribe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = allEntities.Close() }()
	third := world.MustCreateEntity()
	if got := allEntities.Entities(); len(got) != 3 || got[2] != third {
		t.Fatalf("empty filter entities = %v", got)
	}
}

func TestForeignComponentFilterRejected(t *testing.T) {
	first := akara.NewWorld()
	second := akara.NewWorld()
	positions := akara.Register[position](first)
	if _, err := second.Subscribe(akara.All(positions)); !errors.Is(err, akara.ErrForeignComponent) {
		t.Fatalf("Subscribe error = %v, want ErrForeignComponent", err)
	}
}

func TestNilComponentFilterRejected(t *testing.T) {
	world := akara.NewWorld()
	var positions *akara.Store[position]
	if _, err := world.Subscribe(akara.All(positions)); !errors.Is(err, akara.ErrNilComponent) {
		t.Fatalf("Subscribe error = %v, want ErrNilComponent", err)
	}
}

func TestIgnoreIncludeAndClose(t *testing.T) {
	world := akara.NewWorld()
	subscription, err := world.Subscribe()
	if err != nil {
		t.Fatal(err)
	}
	entity := world.MustCreateEntity()
	subscription.Ignore(entity)
	if subscription.Len() != 0 {
		t.Fatal("ignored entity remains subscribed")
	}
	subscription.Include(entity)
	if subscription.Len() != 1 {
		t.Fatal("included entity did not return")
	}
	if err := subscription.Close(); err != nil {
		t.Fatal(err)
	}
	if err := subscription.Close(); err != nil {
		t.Fatal(err)
	}
	if subscription.Len() != 0 {
		t.Fatal("closed subscription retained entities")
	}
}

func TestDestroyRemovesSubscriptionMembership(t *testing.T) {
	world := akara.NewWorld()
	subscription, _ := world.Subscribe()
	defer func() { _ = subscription.Close() }()
	entity := world.MustCreateEntity()
	world.DestroyEntity(entity)
	if subscription.Len() != 0 {
		t.Fatal("destroyed entity remains subscribed")
	}
}
