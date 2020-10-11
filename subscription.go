package akara

// NewSubscription creates a new subscription with the given component filter
func NewSubscription(cf *ComponentFilter) *Subscription {
	return &Subscription{
		Filter:          cf,
		entityBitVector: NewBitSet(),
		entities:        make([]EID, 0),
	}
}

// Subscription is a component filter and a slice of entity ID's for which the filter applies
type Subscription struct {
	Filter          *ComponentFilter
	entityBitVector *BitSet
	entities        []EID
	dirty           bool
}

// GetEntities returns the entities for the system
func (s *Subscription) GetEntities() []EID {
	if s.entityBitVector.Empty() {
		return []EID{}
	}

	if s.dirty {
		s.rebuildCache()
		s.dirty = false
	}

	return s.entities
}

func (s *Subscription) rebuildCache() {
	s.entities = s.entityBitVector.ToIntArray()
}
