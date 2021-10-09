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
			nextEntityID:       new(uint64),
			Subscriptions:      make([]*Subscription, 0),
			entityRemovalQueue: make([]EID, 0),
		},
		componentManagement: &componentManagement{
			nextFactoryID: new(uint64),
			registry:      make(componentRegistry),
			factories:     make(componentFactories),
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
	registry      componentRegistry
	factories     componentFactories
	nextFactoryID *uint64
}

type entityManagement struct {
	nextEntityID       *uint64
	ComponentFlags     sync.Map // map[EID]*bitset.BitSet // bitset for each entity, shows what components the entity has
	Subscriptions      []*Subscription
	entityRemovalQueue []EID
}

type systemManagement struct {
	Systems            []System
	systemStartQueue   []func()
	systemRemovalQueue []System
}

// World contains all of the Entities, Components, and Systems
type World struct {
	TimeDelta time.Duration
	*entityManagement
	*componentManagement
	*systemManagement
	mutex         sync.Mutex
	tickWaitgroup sync.WaitGroup
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

	nextId := atomic.AddUint64(w.nextEntityID, 1)
	factory := newComponentFactory(ComponentID(nextId))
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

// initializeSystem traverses a System's composition tree recursively, ensuring that every base struct's Init is called.
// for example, MovementSystem is composed of BaseSubscriberSystem, which is composed of BaseSystem. We need to call
// MovementSystem.BaseSystem.Init(), then MovementSystem.BaseSubscriberSystem.Init(), and finally MovementSystem.Init()
func (w *World) initializeSystem(s System) {
	if baseContainer, ok := s.(hasBaseSystem); ok {
		w.initializeSystem(baseContainer.Base())
	}

	// once we've traversed this system's whole composition tree, go ahead and call the top-level Init
	s.Init(w)
}

// AddSystem adds a system to the world
func (w *World) AddSystem(s System) *World {
	// make sure that we properly initialize all the Systems that this System is built on top of
	w.initializeSystem(s)

	// keep track of all our systems' subscriptions
	if subscriber, ok := s.(SubscriberSystem); ok {
		subs := subscriber.GetSubscriptions()
		for idx := range subs {
			subs[idx] = w.AddSubscription(subs[idx].Filter)
		}
	}

	w.mutex.Lock()
	w.Systems = append(w.Systems, s)
	w.mutex.Unlock()

	// add System start func to queue.
	// allows the system to update once per tick in its own thread once started
	w.systemStartQueue = append(w.systemStartQueue, func() {
		s.SetActive(true)

		for s.Active() {
			// blocks until the World ticks
			s.WaitForTick()

			s.PreTickFunc()
			s.Update()
			s.PostTickFunc()

			// inform the World that this System finished its update for the current tick
			w.tickWaitgroup.Done()
		}
	})

	return w
}

// RemoveSystem queues the given system for removal
func (w *World) RemoveSystem(s System) *World {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.systemRemovalQueue = append(w.systemRemovalQueue, s)

	return w
}

func (w *World) processSystemStartQueue() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, systemStartFunc := range w.systemStartQueue {
		go systemStartFunc()
	}

	w.systemStartQueue = nil
}

func (w *World) processRemoveQueues() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for remIdx := range w.systemRemovalQueue {
		for idx := range w.Systems {
			if w.Systems[idx] == w.systemRemovalQueue[remIdx] {
				w.Systems = append(w.Systems[:idx], w.Systems[idx+1:]...)
				w.systemRemovalQueue[remIdx].SetActive(false)
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

		w.ComponentFlags.Delete(id)
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

	w.processSystemStartQueue()

	w.processRemoveQueues()

	numSystems := len(w.Systems)
	if numSystems == 0 {
		return nil
	}

	// allow all the Systems to tick once, wait for all of them to finish
	w.tickWaitgroup.Add(numSystems)

	for _, system := range w.Systems {
		system.Tick()
	}

	w.tickWaitgroup.Wait()

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
	nextId := atomic.AddUint64(w.nextEntityID, 1)
	w.ComponentFlags.Store(nextId, &bitset.BitSet{})

	return nextId
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
		cf, cfFound := w.ComponentFlags.Load(id)
		if !cfFound {
			continue
		}

		cf.(*bitset.BitSet).Set(int(w.factories[idx].ID()), found)
	}
}

// updateSubscriptions will iterate through all subscriptions and add the entity id
// to the subscription if the entity can pass through the subscription filter
func (w *World) updateSubscriptions(id EID) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.updateComponentFlags(id)

	cfInterface, found := w.ComponentFlags.Load(id)
	if !found {
		// uhhhhh
		return
	}

	cf := cfInterface.(*bitset.BitSet)

	for idx := range w.Subscriptions {
		w.Subscriptions[idx].mutex.Lock()

		if w.Subscriptions[idx].Filter.Allow(cf) {
			w.Subscriptions[idx].AddEntity(id)
		} else {
			w.Subscriptions[idx].RemoveEntity(id)
		}

		w.Subscriptions[idx].mutex.Unlock()
	}
}
