package akara

import "reflect"

// FilterOption contributes one component constraint to a subscription filter.
type FilterOption interface{ apply(*filterSpec) }

type filterOption func(*filterSpec)

func (option filterOption) apply(spec *filterSpec) { option(spec) }

type filterSpec struct {
	all     []ComponentKey
	any     []ComponentKey
	none    []ComponentKey
	invalid bool
}

// All requires every supplied component type.
func All(components ...ComponentType) FilterOption {
	return filterOption(func(spec *filterSpec) { appendKeys(spec, &spec.all, components) })
}

// Any requires at least one supplied component type. Any() with no arguments
// adds no constraint.
func Any(components ...ComponentType) FilterOption {
	return filterOption(func(spec *filterSpec) { appendKeys(spec, &spec.any, components) })
}

// None forbids every supplied component type.
func None(components ...ComponentType) FilterOption {
	return filterOption(func(spec *filterSpec) { appendKeys(spec, &spec.none, components) })
}

func appendKeys(spec *filterSpec, destination *[]ComponentKey, components []ComponentType) {
	for _, component := range components {
		if component == nil || reflect.ValueOf(component).IsNil() {
			spec.invalid = true
			continue
		}
		*destination = append(*destination, component.componentKey())
	}
}

type componentFilter struct {
	all  componentMask
	any  componentMask
	none componentMask
}

func (filter componentFilter) allow(mask componentMask) bool {
	return mask.containsAll(filter.all) &&
		(filter.any.empty() || mask.intersects(filter.any)) &&
		!mask.intersects(filter.none)
}
