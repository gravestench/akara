package akara

import (
	"reflect"
	"testing"
)

func TestComponentMaskAcrossWords(t *testing.T) {
	var mask componentMask
	for _, id := range []uint32{0, 63, 64, 65, 130} {
		mask.set(id, true)
	}
	var got []uint32
	mask.eachSet(func(id uint32) { got = append(got, id) })
	want := []uint32{0, 63, 64, 65, 130}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("set IDs = %v, want %v", got, want)
	}
	required := componentMask{}
	required.set(64, true)
	required.set(130, true)
	if !mask.containsAll(required) || !mask.intersects(required) {
		t.Fatal("multi-word mask comparison failed")
	}
	mask.set(64, false)
	if mask.containsAll(required) {
		t.Fatal("cleared bit remains contained")
	}
}

func TestHighComponentMaskArchetypeIsInterned(t *testing.T) {
	world := NewWorld()
	mask := (componentMask{}).with(9_999, true)
	if got, want := len(mask), 157; got != want {
		t.Fatalf("mask words = %d, want %d", got, want)
	}
	first := world.internArchetypeLocked(mask)
	second := world.internArchetypeLocked(append(componentMask(nil), mask...))
	if first != second {
		t.Fatal("equivalent high-component masks produced different archetypes")
	}
	if world.ArchetypeCount() != 2 {
		t.Fatalf("archetype count = %d, want 2", world.ArchetypeCount())
	}
	if got := (componentMask{}).with(9_999, false); len(got) != 0 {
		t.Fatalf("clearing an absent high component retained %d words", len(got))
	}
}
