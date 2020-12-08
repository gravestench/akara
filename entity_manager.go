package akara

import (
	"sync/atomic"
)

// EntityID is an entity ID
type EntityID = uint64

// EID is shorthand for EntityID
type EID = EntityID

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

// AddSubscription will look for an identical component filter and return an existing
// subscription if it can. Otherwise, it creates a new subscription and returns it.
func (em *EntityManager) AddSubscription(cf *ComponentFilter) *Subscription {
	for subIdx := range em.Subscriptions {
		if em.Subscriptions[subIdx].Filter.Equals(cf) {
			return em.Subscriptions[subIdx]
		}
	}

	s := NewSubscription(cf)
	em.Subscriptions = append(em.Subscriptions, s)

	// need to inform new subscriptions about existing entities
	for _, cid := range cf.Required.ToIntArray() {
		for eid := range em.world.factories[ComponentID(cid)].instances {
			em.world.UpdateEntity(eid)
		}
	}

	for _, cid := range cf.OneRequired.ToIntArray() {
		for eid := range em.world.factories[ComponentID(cid)].instances {
			em.world.UpdateEntity(eid)
		}
	}

	for _, cid := range cf.Forbidden.ToIntArray() {
		for eid := range em.world.factories[ComponentID(cid)].instances {
			em.world.UpdateEntity(eid)
		}
	}

	return s
}

// NewEntity creates a new entity and Component BitSet
func (em *EntityManager) NewEntity() EID {
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
				break
			}
		}
	}

	delete(em.ComponentFlags, id)
}

// UpdateComponentFlags updates the component bitset for the entity.
// The bitset just says which components an entity currently has.
func (em *EntityManager) UpdateComponentFlags(id EID) {
	// for each component map ...
	for idx := range em.world.factories {
		// ... check if the entity has a component in this component map ...
		_, found := em.world.factories[idx].Get(id)

		// ... for the bit index that corresponds to the map id,
		// use the true/false value we just obtained as the value for that bit in the bitset.
		em.ComponentFlags[id].Set(int(em.world.factories[idx].ID()), found)
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
