package akara

import (
	"reflect"
	"strings"
	"sync"
	"time"
)

type componentRegistry = map[string]ComponentID
type componentFactories = map[ComponentID]*ComponentFactory

// NewWorld creates a new world instance
func NewWorld(cfg *WorldConfig) *World {
	world := &World{
		registry:    make(componentRegistry),
		factories:   make(componentFactories),
		Systems:     make([]System, 0),
		removeQueue: make([]System, 0),
		Events:      NewEventEmitter(),
	}

	world.EntityManager = NewEntityManager(world)

	for _, system := range cfg.systems {
		world.AddSystem(system)
	}

	for _, c := range cfg.components {
		world.RegisterComponent(c)
	}

	return world
}

// World contains all of the Systems, as well as an EntityManager and ComponentManager
type World struct {
	TimeDelta time.Duration
	*EntityManager
	Events      *EventEmitter
	registry    componentRegistry
	factories   componentFactories
	Systems     []System
	removeQueue []System
}

// RegisterComponent registers a component type, assigning and returning its component ID
func (w *World) RegisterComponent(c Component) ComponentID {
	name := strings.ToLower(reflect.TypeOf(c).Elem().Name())

	id, found := w.registry[name]
	if found {
		return id
	}

	nextID := ComponentID(len(w.factories))

	factory := newComponentFactory(nextID)
	factory.world = w

	factory.ComponentMap = &ComponentMap{
		mux:       &sync.Mutex{},
		base:      factory,
		instances: make(map[EID]Component),
	}

	factory.provider = func() Component {
		return c.New()
	}

	w.registry[name] = factory.id
	w.factories[factory.id] = factory

	return factory.id
}

func (w *World) GetComponentFactory(id ComponentID) *ComponentFactory {
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
	w.Systems = append(w.Systems, s)

	s.SetActive(true)

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
	w.removeQueue = append(w.removeQueue, s)

	return w
}

func (w *World) processRemoveQueue() {
	for remIdx := range w.removeQueue {
		for idx := range w.Systems {
			if w.Systems[idx] == w.removeQueue[remIdx] {
				w.Systems = append(w.Systems[:idx], w.Systems[idx+1:]...)
				break
			}
		}
	}
}

// UpdateEntity updates the entity in the world. This causes the entity manager to
// update all subscriptions for this entity ID.
func (w *World) UpdateEntity(id EID) {
	w.UpdateSubscriptions(id)
}

// Update iterates through all Systems and calls the update method if the system is active
func (w *World) Update(dt time.Duration) error {
	w.TimeDelta = dt

	for sysIdx := range w.Systems {
		if !w.Systems[sysIdx].Active() {
			continue
		}

		updater, ok := w.Systems[sysIdx].(SystemUpdater)
		if !ok {
			continue
		}

		updater.Update()
	}

	w.processRemoveQueue()

	return nil
}
