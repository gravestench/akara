package akara

import (
	"errors"
	"testing"
)

func TestCommandBufferAppliesStructuralChangesInOrder(t *testing.T) {
	world := NewWorld()
	tag, err := RegisterSchema(world, Schema{Name: "tag.selected"})
	if err != nil {
		t.Fatal(err)
	}
	entity := world.MustCreateEntity()
	buffer := NewCommandBuffer()
	buffer.AddDynamic(tag, entity, nil)
	buffer.Remove(tag, entity)
	buffer.AddDynamic(tag, entity, nil)
	if tag.Has(entity) {
		t.Fatal("commands mutated the world before Apply")
	}
	if err := buffer.Apply(); err != nil {
		t.Fatal(err)
	}
	if !tag.Has(entity) {
		t.Fatal("submission order was not preserved")
	}
	if err := buffer.Apply(); !errors.Is(err, ErrCommandBufferApplied) {
		t.Fatalf("second Apply error = %v", err)
	}
}

func TestCommandBufferCopiesDynamicValues(t *testing.T) {
	world := NewWorld()
	store, err := RegisterSchema(world, Schema{Name: "name", Fields: []Field{{Name: "value", Kind: FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	entity := world.MustCreateEntity()
	values := map[string]any{"value": "before"}
	buffer := NewCommandBuffer()
	buffer.AddDynamic(store, entity, values)
	values["value"] = "after"
	if err := buffer.Apply(); err != nil {
		t.Fatal(err)
	}
	component, found := store.Get(entity)
	if !found {
		t.Fatal("component missing")
	}
	if got, err := component.Get("value"); err != nil || got != "before" {
		t.Fatalf("value = %v, err = %v", got, err)
	}
}
