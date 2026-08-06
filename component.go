package akara

import (
	"math"
	"reflect"
	"sort"
	"sync"
)

// ComponentKey is an opaque, world-local component type identifier. Keys are
// stable for the lifetime of a World but must not be persisted.
type ComponentKey struct {
	id    uint32
	world *World
}

// ComponentType is implemented by Store[T] and accepted by filter builders.
// Its private method prevents forged component keys.
type ComponentType interface {
	componentKey() ComponentKey
}

type componentStore interface {
	componentKey() ComponentKey
	has(Entity) bool
	removeLocked(Entity)
}

// Store owns the components of one Go type in one World. Register returns the
// canonical Store[T] for that type and world.
type Store[T any] struct {
	world     *World
	key       ComponentKey
	typeName  string
	mu        sync.RWMutex
	instances map[Entity]*T
}

// Register returns the world's canonical store for T. T does not implement an
// interface and its zero value is used when Add creates a component.
// Register panics if world is nil or closed.
func Register[T any](world *World) *Store[T] {
	if world == nil {
		panic("akara: cannot register a component in a nil world")
	}
	world.mutationMu.Lock()
	defer world.mutationMu.Unlock()
	if world.Closed() {
		panic("akara: cannot register a component in a closed world")
	}
	typ := reflect.TypeFor[T]()

	world.registryMu.Lock()
	defer world.registryMu.Unlock()
	if existing, found := world.storesByType[typ]; found {
		store, ok := existing.(*Store[T])
		if !ok {
			panic("akara: component registry invariant violated")
		}
		return store
	}

	if len(world.stores) > math.MaxUint32 {
		panic("akara: component type limit exhausted")
	}
	key := ComponentKey{id: uint32(len(world.stores)), world: world}
	store := &Store[T]{
		world:     world,
		key:       key,
		typeName:  typ.String(),
		instances: make(map[Entity]*T),
	}
	world.storesByType[typ] = store
	world.stores = append(world.stores, store)
	return store
}

// GetStore returns the already registered store for T.
func GetStore[T any](world *World) (*Store[T], bool) {
	if world == nil {
		return nil, false
	}
	world.registryMu.RLock()
	store, found := world.storesByType[reflect.TypeFor[T]()]
	world.registryMu.RUnlock()
	if !found {
		return nil, false
	}
	typed, ok := store.(*Store[T])
	return typed, ok
}

func (s *Store[T]) componentKey() ComponentKey { return s.key }

// Key returns the opaque component key used by filters and diagnostics.
func (s *Store[T]) Key() ComponentKey { return s.key }

// TypeName returns the registered Go type's name.
func (s *Store[T]) TypeName() string { return s.typeName }

// Add returns the entity's existing component or creates its zero value.
func (s *Store[T]) Add(entity Entity) (*T, error) {
	s.world.mutationMu.Lock()
	defer s.world.mutationMu.Unlock()
	if s.world.closed.Load() {
		return nil, ErrWorldClosed
	}
	state, exists := s.world.entities[entity]
	if !exists {
		return nil, ErrEntityNotFound
	}

	s.mu.Lock()
	component, found := s.instances[entity]
	if !found {
		component = new(T)
		s.instances[entity] = component
	}
	s.mu.Unlock()
	if !found {
		s.world.transitionEntityLocked(entity, state, s.key.id, true)
	}
	return component, nil
}

// Set replaces entity's component with a copy of value and returns its address.
func (s *Store[T]) Set(entity Entity, value T) (*T, error) {
	s.world.mutationMu.Lock()
	defer s.world.mutationMu.Unlock()
	if s.world.closed.Load() {
		return nil, ErrWorldClosed
	}
	state, exists := s.world.entities[entity]
	if !exists {
		return nil, ErrEntityNotFound
	}
	component := new(T)
	*component = value
	s.mu.Lock()
	_, found := s.instances[entity]
	s.instances[entity] = component
	s.mu.Unlock()
	if !found {
		s.world.transitionEntityLocked(entity, state, s.key.id, true)
	}
	return component, nil
}

// Get returns entity's component. The returned pointer remains valid until the
// component is removed or replaced by Set; callers synchronize concurrent field access.
func (s *Store[T]) Get(entity Entity) (*T, bool) {
	s.mu.RLock()
	component, found := s.instances[entity]
	s.mu.RUnlock()
	return component, found
}

// Has reports whether entity has this component.
func (s *Store[T]) Has(entity Entity) bool { return s.has(entity) }

func (s *Store[T]) has(entity Entity) bool {
	s.mu.RLock()
	_, found := s.instances[entity]
	s.mu.RUnlock()
	return found
}

// Remove deletes entity's component and reports whether it existed.
func (s *Store[T]) Remove(entity Entity) bool {
	s.world.mutationMu.Lock()
	defer s.world.mutationMu.Unlock()
	if s.world.closed.Load() {
		return false
	}
	state, exists := s.world.entities[entity]
	if !exists {
		return false
	}
	s.mu.Lock()
	_, found := s.instances[entity]
	if found {
		delete(s.instances, entity)
	}
	s.mu.Unlock()
	if found {
		s.world.transitionEntityLocked(entity, state, s.key.id, false)
	}
	return found
}

func (s *Store[T]) removeLocked(entity Entity) {
	s.mu.Lock()
	delete(s.instances, entity)
	s.mu.Unlock()
}

// Len returns the number of stored components.
func (s *Store[T]) Len() int {
	s.mu.RLock()
	length := len(s.instances)
	s.mu.RUnlock()
	return length
}

// Entities returns sorted IDs for entities with this component.
func (s *Store[T]) Entities() []Entity {
	s.mu.RLock()
	entities := make([]Entity, 0, len(s.instances))
	for entity := range s.instances {
		entities = append(entities, entity)
	}
	s.mu.RUnlock()
	sort.Slice(entities, func(i, j int) bool { return entities[i] < entities[j] })
	return entities
}
