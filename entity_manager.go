package akara

import (
	"sync/atomic"
)

// EID is an entity ID
type EID = uint64

// NewEntityManager creates a new entity manager
func NewEntityManager(w *World) *EntityManager {
	return &EntityManager{
		world:          w,
		ComponentFlags: make(map[EID]*BitSet),
		Subscriptions:  make([]*Subscription, 0),
	}
}

// EntityManager manages generating entity ID's and maintains a map of bit flags
// for each entity which describes the components which an entity has
type EntityManager struct {
	world          *World
	nextEntityID   EID
	ComponentFlags map[EID]*BitSet
	Subscriptions  []*Subscription
}

// AddSubscription will try to add the given subscription to the entity manager subscription slice.
// If there is already a subscription with an identical component filter, then the argument
// subscription will be mutated to point to this existing one. This allows many systems to share
// subscriptions
func (em *EntityManager) AddSubscription(s *Subscription) {
	for subIdx := range em.Subscriptions {
		if em.Subscriptions[subIdx].Filter.Equals(s.Filter) {
			s = em.Subscriptions[subIdx] // nolint:staticcheck // we're mutating the pointer...
			return
		}
	}

	em.Subscriptions = append(em.Subscriptions, s)
}

// NewEntity creates a new entity and Component BitSet
func (em *EntityManager) NewEntity() uint64 {
	atomic.AddUint64(&em.nextEntityID, 1)
	em.ComponentFlags[em.nextEntityID] = &BitSet{}

	return em.nextEntityID
}

// RemoveEntity removes an entity
func (em *EntityManager) RemoveEntity(id EID) {
	for subIdx := range em.Subscriptions {
		for entIdx := range em.Subscriptions[subIdx].entities {
			if em.Subscriptions[subIdx].entities[entIdx] == id {
				em.Subscriptions[subIdx].entities = append(
					em.Subscriptions[subIdx].entities[:entIdx],
					em.Subscriptions[subIdx].entities[entIdx+1:]...,
				)
			}
		}
	}

	delete(em.ComponentFlags, id)
}

// UpdateComponentFlags updates the component bitset for the entity.
// The bitset just says which components an entity currently has.
func (em *EntityManager) UpdateComponentFlags(id EID) {
	// for each component map ...
	for mapIdx := range em.world.ComponentManager.maps {
		// ... check if the entity has a component in this component map ...
		_, found := em.world.ComponentManager.maps[mapIdx].Get(id)

		// ... for the bit index that corresponds to the map id,
		// use the true/false value we just obtained as the value for that bit in the bitset.
		em.ComponentFlags[id].Set(int(em.world.ComponentManager.maps[mapIdx].ID()), found)
	}
}

// UpdateSubscriptions will iterate through all subscriptions and add the entity id
// to the subscription if the entity can pass through the subscription filter
func (em *EntityManager) UpdateSubscriptions(id EID) {
	em.UpdateComponentFlags(id)

	for idx := range em.Subscriptions {
		allowed := em.Subscriptions[idx].Filter.Allow(em.ComponentFlags[id])
		em.Subscriptions[idx].dirty = true
		em.Subscriptions[idx].entityBitVector.Set(int(id), allowed)
	}
}
