package akara

import (
	"fmt"
	"sort"
)

func scalarElementKind(kind FieldKind) bool {
	return kind >= FieldBool && kind <= FieldEntity
}

func normalizeFieldValue(field Field, value any) (any, error) {
	if field.Kind != FieldList && field.Kind != FieldMap {
		if field.ElementKind != 0 {
			return nil, fmt.Errorf("scalar field has element kind %d", field.ElementKind)
		}

		return normalizeScalarValue(field.Kind, value)
	}

	if !scalarElementKind(field.ElementKind) {
		return nil, fmt.Errorf("invalid collection element kind %d", field.ElementKind)
	}

	if value == nil {
		if field.Kind == FieldList {
			return []any{}, nil
		}

		return map[string]any{}, nil
	}

	if field.Kind == FieldList {
		entries, ok := value.([]any)

		if !ok {
			return nil, fmt.Errorf("expected list, got %T", value)
		}

		result := make([]any, len(entries))

		for index, entry := range entries {
			if entry == nil {
				return nil, fmt.Errorf("entry %d is nil", index+1)
			}

			normalized, err := normalizeScalarValue(field.ElementKind, entry)

			if err != nil {
				return nil, fmt.Errorf("entry %d: %w", index+1, err)
			}

			result[index] = normalized
		}

		return result, nil
	}

	entries, ok := value.(map[string]any)

	if !ok {
		return nil, fmt.Errorf("expected string-keyed map, got %T", value)
	}

	result := make(map[string]any, len(entries))
	keys := make([]string, 0, len(entries))

	for key := range entries {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		entry := entries[key]

		if entry == nil {
			return nil, fmt.Errorf("key %q is nil", key)
		}

		normalized, err := normalizeScalarValue(field.ElementKind, entry)

		if err != nil {
			return nil, fmt.Errorf("key %q: %w", key, err)
		}

		result[key] = normalized
	}

	return result, nil
}

func cloneDynamicValue(value any) any {
	switch typed := value.(type) {
	case []any:
		result := make([]any, len(typed))
		copy(result, typed)

		return result

	case map[string]any:
		copy := make(map[string]any, len(typed))

		for key, entry := range typed {
			copy[key] = entry
		}

		return copy

	default:
		return value
	}
}

func cloneDynamicDefaults(defaults []any) []any {
	result := make([]any, len(defaults))

	for index, value := range defaults {
		result[index] = cloneDynamicValue(value)
	}

	return result
}
