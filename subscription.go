package akara

import (
	"sort"
	"sync"
)

type entityMap map[EID]*empty

type empty struct{} // intentionally empty!

// NewSubscription creates a new subscription with the given component filter
func NewSubscription(cf *ComponentFilter) *Subscription {
	return &Subscription{
		Filter:    cf,
		entityMap: make(entityMap),
		entities:  make([]EID, 0),
		ignoredEntities: make(entityMap),
	}
}

// Subscription is a component filter and a slice of entity ID's for which the filter applies
type Subscription struct {
	Filter    *ComponentFilter
	entityMap       // we use (abuse) the lookup ability of maps for adding/removing EIDs
	entities  []EID // we sort the map keys when GetEntities is called, only if dirty==true
	dirty     bool
	ignoredEntities entityMap
	mutex     sync.Mutex
}

// AddEntity adds an entity to the subscription entity map
func (s *Subscription) AddEntity(id EID) {
	if _, found := s.entityMap[id]; !found {
		s.dirty = true
		s.entityMap[id] = &empty{}
	}
}

// RemoveEntity removes an entity from the subscription entity map
func (s *Subscription) RemoveEntity(id EID) {
	if _, found := s.entityMap[id]; found {
		s.dirty = true
		delete(s.entityMap, id)
	}
}

// GetEntities returns the entities for the system
func (s *Subscription) GetEntities() []EID {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.dirty || s.entities == nil {
		s.rebuildCache()
		s.dirty = false
	}

	return s.entities
}

func (s *Subscription) rebuildCache() {
	s.entities = make([]EID, len(s.entityMap))

	idx := 0
	for eid := range s.entityMap {
		s.entities[idx] = eid
		idx++
	}

	sort.Slice(s.entities, func(i, j int) bool {
		return s.entities[i] < s.entities[j]
	})
}

func (s *Subscription) IgnoreEntity(id EID) {
	s.mutex.Lock()
	s.ignoredEntities[id] = &empty{}
	s.RemoveEntity(id)
	s.mutex.Unlock()
}

func (s *Subscription) EntityIsIgnored(id EID) bool {
	_, ok := s.ignoredEntities[id]
	return ok
}
