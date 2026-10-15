package akara

import (
	"errors"
	"reflect"
	"testing"
)

func TestDynamicCollectionFieldsNormalizeEveryScalarKind(t *testing.T) {
	world := NewWorld()
	t.Cleanup(func() { _ = world.Close() })

	entity := world.MustCreateEntity()
	target := world.MustCreateEntity()
	cases := []struct {
		name  string
		kind  FieldKind
		value any
	}{
		{
			name:  "bool",
			kind:  FieldBool,
			value: true,
		},
		{
			name:  "int64",
			kind:  FieldInt64,
			value: int64(-7),
		},
		{
			name:  "uint64",
			kind:  FieldUint64,
			value: uint64(8),
		},
		{
			name:  "float64",
			kind:  FieldFloat64,
			value: float64(1.5),
		},
		{
			name:  "string",
			kind:  FieldString,
			value: "crawl",
		},
		{
			name:  "entity",
			kind:  FieldEntity,
			value: target,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, err := RegisterSchema(world, Schema{
				Name: "collections." + tc.name,
				Fields: []Field{
					{
						Name:        "list",
						Kind:        FieldList,
						ElementKind: tc.kind,
					},
					{
						Name:        "map",
						Kind:        FieldMap,
						ElementKind: tc.kind,
					},
				},
			})

			if err != nil {
				t.Fatal(err)
			}

			component, err := store.Add(entity)

			if err != nil {
				t.Fatal(err)
			}

			listDefault, err := component.Get("list")

			if err != nil || !reflect.DeepEqual(listDefault, []any{}) {
				t.Fatalf("list default = %#v, %v", listDefault, err)
			}

			mapDefault, err := component.Get("map")

			if err != nil || !reflect.DeepEqual(mapDefault, map[string]any{}) {
				t.Fatalf("map default = %#v, %v", mapDefault, err)
			}

			list := []any{tc.value}
			mapping := map[string]any{"one": tc.value}
			component, err = store.Set(entity, map[string]any{
				"list": list,
				"map":  mapping,
			})

			if err != nil {
				t.Fatal(err)
			}

			list[0] = nil
			mapping["one"] = nil

			storedList, _ := component.Get("list")
			storedMap, _ := component.Get("map")

			if !reflect.DeepEqual(storedList, []any{tc.value}) ||
				!reflect.DeepEqual(storedMap, map[string]any{"one": tc.value}) {
				t.Fatal("Set retained caller-owned collection memory")
			}

			storedList.([]any)[0] = nil
			storedMap.(map[string]any)["one"] = nil

			snapshot, err := component.Snapshot()

			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(snapshot["list"], []any{tc.value}) ||
				!reflect.DeepEqual(snapshot["map"], map[string]any{"one": tc.value}) {
				t.Fatal("Get returned an alias to stored collection memory")
			}

			snapshot["list"].([]any)[0] = nil
			snapshot["map"].(map[string]any)["one"] = nil

			again, _ := component.Get("list")

			if !reflect.DeepEqual(again, []any{tc.value}) {
				t.Fatal("Snapshot returned an alias to stored memory")
			}
		})
	}
}

func TestDynamicCollectionsRejectWrongShapesAndElementKinds(t *testing.T) {
	world := NewWorld()
	t.Cleanup(func() { _ = world.Close() })

	store, err := RegisterSchema(world, Schema{
		Name: "collection.validation",
		Fields: []Field{
			{
				Name:        "list",
				Kind:        FieldList,
				ElementKind: FieldInt64,
			},
			{
				Name:        "map",
				Kind:        FieldMap,
				ElementKind: FieldString,
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	entity := world.MustCreateEntity()

	for _, values := range []map[string]any{
		{"list": []any{int64(1), "two"}},
		{"list": []any{nil}},
		{"list": map[string]any{"1": int64(1)}},
		{"map": map[string]any{"one": int64(1)}},
		{"map": map[int]any{1: "one"}},
	} {
		if _, err := store.Set(entity, values); !errors.Is(err, ErrInvalidFieldValue) {
			t.Fatalf("invalid collection %#v: %v", values, err)
		}
	}

	if _, err := RegisterSchema(world, Schema{
		Name: "invalid.nested",
		Fields: []Field{
			{
				Name:        "items",
				Kind:        FieldList,
				ElementKind: FieldMap,
			},
		},
	}); !errors.Is(err, ErrInvalidSchema) {
		t.Fatalf("nested collection schema: %v", err)
	}

	if _, err := RegisterSchema(world, Schema{
		Name: "collection.validation",
		Fields: []Field{
			{
				Name:        "list",
				Kind:        FieldList,
				ElementKind: FieldString,
			},
			{
				Name:        "map",
				Kind:        FieldMap,
				ElementKind: FieldString,
			},
		},
	}); !errors.Is(err, ErrSchemaConflict) {
		t.Fatalf("changed element kind re-registration: %v", err)
	}
}

func TestDynamicCollectionDefaultsAndMigrationStayDetached(t *testing.T) {
	world := NewWorld()
	t.Cleanup(func() { _ = world.Close() })

	listDefault := []any{"original"}
	mapDefault := map[string]any{"one": int64(1)}
	store, err := RegisterSchema(world, Schema{
		Name:    "collection.defaults",
		Version: 1,
		Fields: []Field{
			{
				Name:        "list",
				Kind:        FieldList,
				ElementKind: FieldString,
				Default:     listDefault,
			},
			{
				Name:        "map",
				Kind:        FieldMap,
				ElementKind: FieldInt64,
				Default:     mapDefault,
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	listDefault[0] = "changed"
	mapDefault["one"] = int64(2)

	first := world.MustCreateEntity()
	second := world.MustCreateEntity()
	firstComponent, err := store.Add(first)

	if err != nil {
		t.Fatal(err)
	}

	secondComponent, err := store.Add(second)

	if err != nil {
		t.Fatal(err)
	}

	firstList, _ := firstComponent.Get("list")
	firstList.([]any)[0] = "edited"

	secondList, _ := secondComponent.Get("list")

	if !reflect.DeepEqual(secondList, []any{"original"}) {
		t.Fatalf("component defaults leaked across entities: %#v", secondList)
	}

	declared := store.Schema()
	declared.Fields[0].Default.([]any)[0] = "schema edit"

	if got := store.Schema().Fields[0].Default; !reflect.DeepEqual(got, []any{"original"}) {
		t.Fatalf("schema default alias: %#v", got)
	}

	failure := errors.New("stop migration")
	err = store.Migrate(Schema{
		Name:    "collection.defaults",
		Version: 2,
		Fields: []Field{
			{
				Name:        "list",
				Kind:        FieldList,
				ElementKind: FieldString,
			},
			{
				Name:        "map",
				Kind:        FieldMap,
				ElementKind: FieldInt64,
			},
		},
	}, func(entity Entity, previous map[string]any) (map[string]any, error) {
		previous["list"].([]any)[0] = "callback edit"
		previous["map"].(map[string]any)["one"] = int64(9)

		if entity == second {
			return nil, failure
		}

		return previous, nil
	})

	if !errors.Is(err, failure) {
		t.Fatalf("migration error = %v", err)
	}

	after, _ := firstComponent.Get("list")

	if !reflect.DeepEqual(after, []any{"original"}) {
		t.Fatalf("failed migration changed live state: %#v", after)
	}
}
