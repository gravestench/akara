package akara

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
)

var (
	ErrWorldClosed       = errors.New("akara: world is closed")
	ErrEntityNotFound    = errors.New("akara: entity does not exist")
	ErrForeignComponent  = errors.New("akara: component belongs to another world")
	ErrNilComponent      = errors.New("akara: component handle is nil")
	ErrNilSystem         = errors.New("akara: system is nil")
	ErrEntityIDExhausted = errors.New("akara: entity ID space exhausted")
	ErrSystemIDExhausted = errors.New("akara: system ID space exhausted")
)

// Entity is a world-local entity identifier. Zero is never a valid entity.
type Entity uint64

type entityState struct{ archetype *archetype }

// World owns entities, component stores, subscriptions, and deterministic systems.
type World struct {
	mutationMu     sync.Mutex
	entities       map[Entity]entityState
	nextEntityID   atomic.Uint64
	registryMu     sync.RWMutex
	storesByType   map[reflect.Type]componentStore
	stores         []componentStore
	subscriptions  map[*Subscription]struct{}
	archetypes     map[uint64][]*archetype
	archetypeCount int
	emptyArchetype *archetype
	systemsMu      sync.RWMutex
	systems        []systemEntry
	nextSystemID   SystemID
	updateMu       sync.Mutex
	closed         atomic.Bool
	closeOnce      sync.Once
	done           chan struct{}
}

// NewWorld creates an empty world.
func NewWorld() *World {
	world := &World{
		entities:      make(map[Entity]entityState),
		storesByType:  make(map[reflect.Type]componentStore),
		subscriptions: make(map[*Subscription]struct{}),
		archetypes:    make(map[uint64][]*archetype),
		done:          make(chan struct{}),
	}
	world.emptyArchetype = world.internArchetypeLocked(nil)
	return world
}

// NewWorldWithContext closes its world when ctx is canceled.
func NewWorldWithContext(ctx context.Context) *World {
	if ctx == nil {
		panic("akara: nil context")
	}
	world := NewWorld()
	go func() {
		select {
		case <-ctx.Done():
			_ = world.Close()
		case <-world.Done():
		}
	}()
	return world
}

// CreateEntity creates an entity with no components.
func (world *World) CreateEntity() (Entity, error) {
	world.mutationMu.Lock()
	defer world.mutationMu.Unlock()
	if world.closed.Load() {
		return 0, ErrWorldClosed
	}
	entity := Entity(world.nextEntityID.Add(1))
	if entity == 0 {
		return 0, ErrEntityIDExhausted
	}
	state := entityState{archetype: world.emptyArchetype}
	world.entities[entity] = state
	world.emptyArchetype.entities[entity] = struct{}{}
	return entity, nil
}

// MustCreateEntity creates an entity or panics. It is convenient during trusted
// world construction; runtime code should generally use CreateEntity.
func (world *World) MustCreateEntity() Entity {
	entity, err := world.CreateEntity()
	if err != nil {
		panic(err)
	}
	return entity
}

// DestroyEntity immediately removes an entity, all its components, and all its
// subscription memberships.
func (world *World) DestroyEntity(entity Entity) bool {
	world.mutationMu.Lock()
	defer world.mutationMu.Unlock()
	if world.closed.Load() {
		return false
	}
	state, found := world.entities[entity]
	if !found {
		return false
	}
	state.archetype.mask.eachSet(func(id uint32) { world.stores[id].removeLocked(entity) })
	delete(state.archetype.entities, entity)
	delete(world.entities, entity)
	return true
}

// EntityExists reports whether entity is live.
func (world *World) EntityExists(entity Entity) bool {
	world.mutationMu.Lock()
	_, found := world.entities[entity]
	world.mutationMu.Unlock()
	return found
}

// Entities returns all live entity IDs in ascending order.
func (world *World) Entities() []Entity {
	world.mutationMu.Lock()
	entities := make([]Entity, 0, len(world.entities))
	for entity := range world.entities {
		entities = append(entities, entity)
	}
	world.mutationMu.Unlock()
	sort.Slice(entities, func(i, j int) bool { return entities[i] < entities[j] })
	return entities
}

// Closed reports whether Close has been called.
func (world *World) Closed() bool { return world.closed.Load() }

// Done closes when the world closes.
func (world *World) Done() <-chan struct{} { return world.done }

// Close permanently closes the world. Existing component data remains readable,
// while mutation, subscription creation, and updates are rejected.
func (world *World) Close() error {
	world.closeOnce.Do(func() {
		world.mutationMu.Lock()
		world.systemsMu.Lock()
		world.closed.Store(true)
		close(world.done)
		world.systemsMu.Unlock()
		for subscription := range world.subscriptions {
			subscription.mu.Lock()
			subscription.closed = true
			subscription.archetypes = nil
			subscription.ignored = nil
			subscription.mu.Unlock()
		}
		world.subscriptions = make(map[*Subscription]struct{})
		world.mutationMu.Unlock()
	})
	return nil
}
