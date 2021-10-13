package akara

import "github.com/gravestench/bitset"

func NewComponentFilter(all, oneOf, none *bitset.BitSet) *ComponentFilter {
	return &ComponentFilter{
		all,
		oneOf,
		none,
	}
}

// ComponentFilter describes the component requirements of entities
type ComponentFilter struct {
	Required    *bitset.BitSet
	OneRequired *bitset.BitSet
	Forbidden   *bitset.BitSet
}

// Equals checks if this component filter is equal to the argument component filter
func (cf *ComponentFilter) Equals(other *ComponentFilter) bool {
	return cf.Required.Equals(other.Required) &&
		cf.OneRequired.Equals(other.OneRequired) &&
		cf.Forbidden.Equals(other.Forbidden)
}

// Allow returns true if the given bitset is not rejected by the component filter
func (cf *ComponentFilter) Allow(other *bitset.BitSet) bool {
	if cf == nil {
		return false
	}

	if cf.Required != nil && !cf.Required.Empty() {
		if !other.ContainsAll(cf.Required) {
			return false
		}
	}

	if cf.OneRequired != nil {
		if !cf.OneRequired.Intersects(other) && !cf.OneRequired.Empty() {
			return false
		}
	}

	if cf.Forbidden != nil && !cf.Forbidden.Empty() {
		if cf.Forbidden.Intersects(other) {
			return false
		}
	}

	return true
}
