package akara

import (
	"github.com/gravestench/bitset"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type componentRegistry = map[string]ComponentID
type componentFactories = map[ComponentID]*ComponentFactory

// NewWorld creates a new world instance
func NewWorld(cfg *WorldConfig) *World {
	world := &World{
		entityManagement: &entityManagement{
			ComponentFlags: make(map[EID]*bitset.BitSet),
			Subscriptions:  make([]*Subscription, 0),
			entityRemovalQueue: make([]EID, 0),
		},
		componentManagement: &componentManagement{
			registry:  make(componentRegistry),
			factories: make(componentFactories),
		},
		systemManagement: &systemManagement{
			Systems:            make([]System, 0),
			systemRemovalQueue: make([]System, 0),
		},
	}

	for _, system := range cfg.systems {
		world.AddSystem(system)
	}

	for _, c := range cfg.components {
		world.RegisterComponent(c)
	}

	return world
}

type componentManagement struct {
	registry  componentRegistry
	factories componentFactories
}

type entityManagement struct {
	nextEntityID       EID
	ComponentFlags     map[EID]*bitset.BitSet // bitset for each entity, shows what components the entity has
	Subscriptions      []*Subscription
	entityRemovalQueue []EID
}

type systemManagement struct {
	Systems            []System
	systemRemovalQueue []System
}

// World contains all of the Entities, Components, and Systems
type World struct {
	TimeDelta time.Duration
	*entityManagement
	*componentManagement
	*systemManagement
	mutex sync.Mutex
}

// RegisterComponent registers a component type, assigning and returning its component ID
func (w *World) RegisterComponent(c Component) ComponentID {
	name := strings.ToLower(reflect.TypeOf(c).Elem().Name())

	w.mutex.Lock()
	defer w.mutex.Unlock()

	id, found := w.registry[name]
	if found {
		return id
	}

	nextID := ComponentID(len(w.factories))

	factory := newComponentFactory(nextID)
	factory.world = w

	factory.provider = func() Component {
		return c.New()
	}

	w.registry[name] = factory.id
	w.factories[factory.id] = factory

	return factory.id
}

// GetComponentFactory returns the ComponentFactory for the given ComponentID
func (w *World) GetComponentFactory(id ComponentID) *ComponentFactory {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.factories[id]
}

// NewComponentFilter creates a builder for creating
func (w *World) NewComponentFilter() *ComponentFilterBuilder {
	return &ComponentFilterBuilder{
		world:      w,
		require:    make([]Component, 0),
		requireOne: make([]Component, 0),
		forbid:     make([]Component, 0),
	}
}

// AddSystem adds a system to the world
func (w *World) AddSystem(s System) *World {
	w.mutex.Lock()

	w.Systems = append(w.Systems, s)

	s.SetActive(true)

	w.mutex.Unlock()

	if baseContainer, ok := s.(hasBaseSystem); ok {
		baseContainer.Base().World = w
	}

	if initializer, ok := s.(SystemInitializer); ok {
		initializer.Init(w)
	}

	if subscriber, ok := s.(SubscriberSystem); ok {
		subs := subscriber.GetSubscriptions()
		for idx := range subs {
			subs[idx] = w.AddSubscription(subs[idx].Filter)
		}
	}

	return w
}

// RemoveSystem queues the given system for removal
func (w *World) RemoveSystem(s System) *World {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.systemRemovalQueue = append(w.systemRemovalQueue, s)

	return w
}

func (w *World) processRemoveQueue() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for remIdx := range w.systemRemovalQueue {
		for idx := range w.Systems {
			if w.Systems[idx] == w.systemRemovalQueue[remIdx] {
				w.Systems = append(w.Systems[:idx], w.Systems[idx+1:]...)
				break
			}
		}
	}

	for _, id := range w.entityRemovalQueue {
		for subIdx := range w.Subscriptions {
			for entIdx := range w.Subscriptions[subIdx].entities {
				if w.Subscriptions[subIdx].entities[entIdx] == id {
					w.Subscriptions[subIdx].entities = append(
						w.Subscriptions[subIdx].entities[:entIdx],
						w.Subscriptions[subIdx].entities[entIdx+1:]...,
					)
					break
				}
			}
		}

		delete(w.ComponentFlags, id)
	}
}

// UpdateEntity updates the entity in the world. This causes the entity manager to
// update all subscriptions for this entity ID.
func (w *World) UpdateEntity(id EID) {
	w.updateSubscriptions(id)
}

// Update iterates through all Systems and calls the update method if the system is active
func (w *World) Update(dt time.Duration) error {
	w.TimeDelta = dt

	for sysIdx := range w.Systems {
		if !w.Systems[sysIdx].Active() {
			continue
		}

		if sys, ok := w.Systems[sysIdx].(SystemInitializer); ok {
			if !sys.IsInitialized() {
				if base, ok := sys.(baseSystem); ok {
					base.bind(w)
				}

				sys.Init(w)
				continue
			}
		}

		if sys, ok := w.Systems[sysIdx].(SystemUpdater); ok {
			sys.Update()
			continue
		}

		if sys, ok := w.Systems[sysIdx].(SystemUpdaterTimed); ok {
			sys.Update(dt)
			continue
		}
	}

	w.processRemoveQueue()

	return nil
}

// AddSubscription will look for an identical component filter and return an existing
// subscription if it can. Otherwise, it creates a new subscription and returns it.
func (w *World) AddSubscription(input interface{}) *Subscription {
	var s *Subscription
	var cf *ComponentFilter

	switch t := input.(type) {
	case *Subscription:
		s = t
		cf = s.Filter
	case *ComponentFilter:
		cf = t
		s = NewSubscription(cf)
	default:
		return nil
	}

	for subIdx := range w.Subscriptions {
		if w.Subscriptions[subIdx].Filter.Equals(cf) {
			return w.Subscriptions[subIdx]
		}
	}

	w.Subscriptions = append(w.Subscriptions, s)

	// need to inform new subscriptions about existing entities
	for _, cid := range cf.Required.ToIntArray() {
		for eid := range w.factories[ComponentID(cid)].instances {
			w.UpdateEntity(eid)
		}
	}

	for _, cid := range cf.OneRequired.ToIntArray() {
		for eid := range w.factories[ComponentID(cid)].instances {
			w.UpdateEntity(eid)
		}
	}

	for _, cid := range cf.Forbidden.ToIntArray() {
		for eid := range w.factories[ComponentID(cid)].instances {
			w.UpdateEntity(eid)
		}
	}

	return s
}

// NewEntity creates a new entity and Component BitSet
func (w *World) NewEntity() EID {
	atomic.AddUint64(&w.nextEntityID, 1)
	w.ComponentFlags[w.nextEntityID] = &bitset.BitSet{}

	return w.nextEntityID
}

// RemoveEntity removes an entity
func (w *World) RemoveEntity(id EID) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.entityRemovalQueue = append(w.entityRemovalQueue, id)

}

// UpdateComponentFlags updates the component bitset for the entity.
// The bitset just says which components an entity currently has.
func (w *World) updateComponentFlags(id EID) {
	// for each component map ...
	for idx := range w.factories {
		// ... check if the entity has a component in this component map ...
		_, found := w.factories[idx].Get(id)

		// ... for the bit index that corresponds to the map id,
		// use the true/false value we just obtained as the value for that bit in the bitset.
		w.ComponentFlags[id].Set(int(w.factories[idx].ID()), found)
	}
}

// updateSubscriptions will iterate through all subscriptions and add the entity id
// to the subscription if the entity can pass through the subscription filter
func (w *World) updateSubscriptions(id EID) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.updateComponentFlags(id)

	for idx := range w.Subscriptions {
		w.Subscriptions[idx].mutex.Lock()

		if w.Subscriptions[idx].Filter.Allow(w.ComponentFlags[id]) {
			w.Subscriptions[idx].AddEntity(id)
		} else {
			w.Subscriptions[idx].RemoveEntity(id)
		}

		w.Subscriptions[idx].mutex.Unlock()
	}
}
