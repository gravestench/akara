package akara

import (
	"errors"
	"reflect"
	"testing"
)

type dynamicTestPosition struct{ X, Y float64 }

func TestDynamicComponentsShareTypedArchetypes(t *testing.T) {
	world := NewWorld()
	positions := Register[dynamicTestPosition](world)
	velocities, err := RegisterSchema(world, Schema{
		Name:    "movement.velocity",
		Version: 1,
		Fields: []Field{
			{Name: "x", Kind: FieldFloat64},
			{Name: "y", Kind: FieldFloat64, Default: float64(2)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	moving, err := world.Subscribe(All(positions, velocities))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := moving.Close(); err != nil {
			t.Errorf("close subscription: %v", err)
		}
	})

	entity := world.MustCreateEntity()
	if _, err := positions.Set(entity, dynamicTestPosition{X: 10}); err != nil {
		t.Fatal(err)
	}
	if moving.Len() != 0 {
		t.Fatal("entity matched before dynamic component was present")
	}
	velocity, err := velocities.Set(entity, map[string]any{"x": 4.5})
	if err != nil {
		t.Fatal(err)
	}
	if got := moving.Entities(); !reflect.DeepEqual(got, []Entity{entity}) {
		t.Fatalf("matching entities = %v", got)
	}
	if y, err := velocity.Get("y"); err != nil || y != float64(2) {
		t.Fatalf("default y = %v, err = %v", y, err)
	}
	if err := velocity.Set("x", float64(9)); err != nil {
		t.Fatal(err)
	}
	if snapshot, err := velocity.Snapshot(); err != nil || snapshot["x"] != float64(9) {
		t.Fatalf("snapshot = %v, err = %v", snapshot, err)
	}
}

func TestDynamicSchemaRegistrationAndValidation(t *testing.T) {
	world := NewWorld()
	schema := Schema{Name: "status.poisoned", Version: 1, Fields: []Field{{Name: "ticks", Kind: FieldInt64}}}
	first, err := RegisterSchema(world, schema)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RegisterSchema(world, schema)
	if err != nil || first != second {
		t.Fatalf("idempotent registration = %p, %p, %v", first, second, err)
	}
	if _, err := RegisterSchema(world, Schema{Name: schema.Name, Version: 2}); !errors.Is(err, ErrSchemaConflict) {
		t.Fatalf("schema conflict error = %v", err)
	}
	if _, err := RegisterSchema(world, Schema{Name: "bad", Fields: []Field{{Name: "x", Kind: FieldBool, Default: 3}}}); !errors.Is(err, ErrInvalidSchema) {
		t.Fatalf("invalid default error = %v", err)
	}
}

func TestDynamicStoresAreSortedRegistrySnapshot(t *testing.T) {
	world := NewWorld()
	defer world.Close()
	second, err := RegisterSchema(world, Schema{Name: "world.velocity"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := RegisterSchema(world, Schema{Name: "world.position"})
	if err != nil {
		t.Fatal(err)
	}
	stores := DynamicStores(world)
	if len(stores) != 2 || stores[0] != first || stores[1] != second {
		t.Fatalf("stores = %v", stores)
	}
	stores[0] = nil
	if got := DynamicStores(world); len(got) != 2 || got[0] != first {
		t.Fatalf("registry changed through returned slice: %v", got)
	}
	if DynamicStores(nil) != nil {
		t.Fatal("nil world returned stores")
	}
}

func TestDynamicReferenceInvalidatedByReplacementAndRemoval(t *testing.T) {
	world := NewWorld()
	store, err := RegisterSchema(world, Schema{Name: "test.value", Fields: []Field{{Name: "value", Kind: FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	entity := world.MustCreateEntity()
	first, err := store.Set(entity, map[string]any{"value": "first"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Set(entity, map[string]any{"value": "second"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Get("value"); !errors.Is(err, ErrComponentReference) {
		t.Fatalf("replaced reference error = %v", err)
	}
	if !store.Remove(entity) {
		t.Fatal("component was not removed")
	}
	if _, err := second.Get("value"); !errors.Is(err, ErrComponentReference) {
		t.Fatalf("removed reference error = %v", err)
	}
}

func TestDynamicAddReturnsExistingComponent(t *testing.T) {
	world := NewWorld()
	store, err := RegisterSchema(world, Schema{Name: "counter", Fields: []Field{{Name: "value", Kind: FieldInt64}}})
	if err != nil {
		t.Fatal(err)
	}
	entity := world.MustCreateEntity()
	first, err := store.Set(entity, map[string]any{"value": int64(7)})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Add(entity)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := second.Get("value"); err != nil || got != int64(7) {
		t.Fatalf("existing value = %v, err = %v", got, err)
	}
	if err := first.Set("value", int64(8)); err != nil {
		t.Fatalf("Add invalidated existing reference: %v", err)
	}
}

func TestDynamicEntityReferenceField(t *testing.T) {
	world := NewWorld()
	store, err := RegisterSchema(world, Schema{Name: "target", Fields: []Field{{Name: "entity", Kind: FieldEntity}}})
	if err != nil {
		t.Fatal(err)
	}
	entity, target := world.MustCreateEntity(), world.MustCreateEntity()
	component, err := store.Set(entity, map[string]any{"entity": target})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := component.Get("entity"); err != nil || got != target {
		t.Fatalf("target = %v, err = %v", got, err)
	}
	if err := component.Set("entity", uint64(target)); !errors.Is(err, ErrInvalidFieldValue) {
		t.Fatalf("untyped entity error = %v", err)
	}
}

func TestDynamicSchemaMigrationIsTransactional(t *testing.T) {
	world := NewWorld()
	store, err := RegisterSchema(world, Schema{
		Name: "position", Version: 1,
		Fields: []Field{{Name: "x", Kind: FieldFloat64}},
	})
	if err != nil {
		t.Fatal(err)
	}
	firstEntity, secondEntity := world.MustCreateEntity(), world.MustCreateEntity()
	first, _ := store.Set(firstEntity, map[string]any{"x": float64(3)})
	_, _ = store.Set(secondEntity, map[string]any{"x": float64(4)})
	next := Schema{
		Name: "position", Version: 2,
		Fields: []Field{{Name: "x", Kind: FieldFloat64}, {Name: "y", Kind: FieldFloat64}},
	}
	failure := errors.New("migration failed")
	err = store.Migrate(next, func(entity Entity, previous map[string]any) (map[string]any, error) {
		if entity == secondEntity {
			return nil, failure
		}
		return map[string]any{"x": previous["x"], "y": float64(9)}, nil
	})
	if !errors.Is(err, failure) || store.Schema().Version != 1 {
		t.Fatalf("failed migration changed schema: version=%d err=%v", store.Schema().Version, err)
	}
	if value, err := first.Get("x"); err != nil || value != float64(3) {
		t.Fatalf("failed migration changed instance: value=%v err=%v", value, err)
	}

	err = store.Migrate(next, func(_ Entity, previous map[string]any) (map[string]any, error) {
		return map[string]any{"x": previous["x"], "y": float64(9)}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Get("x"); !errors.Is(err, ErrComponentReference) {
		t.Fatalf("migration did not invalidate old reference: %v", err)
	}
	migrated, found := store.Get(firstEntity)
	if !found {
		t.Fatal("migrated component missing")
	}
	if value, err := migrated.Get("y"); err != nil || value != float64(9) {
		t.Fatalf("migrated y=%v err=%v", value, err)
	}
}
