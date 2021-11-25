package akara

import "github.com/gravestench/bitset"

// NewComponentFilter creates a component filter using the given BitSets.
//
// The first BitSet declares those bits which are required in
// order to pass through the filter.
//
// The second BitSet declares those bits of which at least one is required in
// order to pass through the filter.
//
// The third BitSet declares those bits which none are allowed to
// pass through the filter.
func NewComponentFilter(all, oneOf, none *bitset.BitSet) *ComponentFilter {
	return &ComponentFilter{
		all,
		oneOf,
		none,
	}
}

// ComponentFilter is a group of 3 BitSets which are used to filter any other given BitSet.
// A target (fourth, external) BitSet is allowed to "pass" through the filter under
// the following conditions:
//
// The target bitset contains all "required" bits
// The target bitset contains at least one of the "one required" bits
// The target bitset contains none of the "forbidden" bits
//
// If the target bitset invalidates any of these three rules, the target
// bitset is said to have been "rejected" by the filter.
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

	if cf.Required != nil && !cf.Required.Empty() && !other.ContainsAll(cf.Required) {
		return false
	}

	if cf.OneRequired != nil && !cf.OneRequired.Intersects(other) && !cf.OneRequired.Empty() {
		return false
	}

	if cf.Forbidden != nil && !cf.Forbidden.Empty() && cf.Forbidden.Intersects(other) {
		return false
	}

	return true
}
