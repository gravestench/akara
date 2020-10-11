package akara

func NewComponentFilter(all, oneOf, none *BitSet) *ComponentFilter {
	return &ComponentFilter{
		all,
		oneOf,
		none,
	}
}

// ComponentFilter describes the component requirements of entities
type ComponentFilter struct {
	Required    *BitSet
	OneRequired *BitSet
	Forbidden   *BitSet
}

// Equals checks if this component filter is equal to the argument component filter
func (cf *ComponentFilter) Equals(other *ComponentFilter) bool {
	return cf.Required.Equals(other.Required) &&
		cf.OneRequired.Equals(other.OneRequired) &&
		cf.Forbidden.Equals(other.Forbidden)
}

// Allow returns true if the given bitset is not rejected by the component filter
func (cf *ComponentFilter) Allow(bs *BitSet) bool {
	if cf == nil {
		return false
	}

	if cf.Required != nil {
		if !cf.Required.ContainsAll(bs) {
			return false
		}
	}

	if cf.OneRequired != nil {
		if !cf.OneRequired.Intersects(bs) && !cf.OneRequired.Empty() {
			return false
		}
	}

	if cf.Forbidden != nil {
		if cf.Forbidden.Intersects(bs) {
			return false
		}
	}

	return true
}
