package akara

import (
	"sort"
	"sync"
)

// Subscription caches the archetypes matching an immutable component filter.
// Entity membership follows archetype transitions without per-entity filter work.
type Subscription struct {
	world      *World
	filter     componentFilter
	mu         sync.RWMutex
	archetypes map[*archetype]struct{}
	ignored    map[Entity]struct{}
	closed     bool
}

// Subscribe creates an independent subscription. An empty filter matches every
// live entity. Component handles from another world return ErrForeignComponent.
func (world *World) Subscribe(options ...FilterOption) (*Subscription, error) {
	spec := filterSpec{}
	for _, option := range options {
		if option != nil {
			option.apply(&spec)
		}
	}
	filter, err := world.compileFilter(spec)
	if err != nil {
		return nil, err
	}

	world.mutationMu.Lock()
	defer world.mutationMu.Unlock()
	if world.closed.Load() {
		return nil, ErrWorldClosed
	}
	subscription := &Subscription{
		world:      world,
		filter:     filter,
		archetypes: make(map[*archetype]struct{}),
		ignored:    make(map[Entity]struct{}),
	}
	for _, bucket := range world.archetypes {
		for _, archetype := range bucket {
			if filter.allow(archetype.mask) {
				subscription.archetypes[archetype] = struct{}{}
			}
		}
	}
	world.subscriptions[subscription] = struct{}{}
	return subscription, nil
}

func (world *World) compileFilter(spec filterSpec) (componentFilter, error) {
	if spec.invalid {
		return componentFilter{}, ErrNilComponent
	}
	filter := componentFilter{}
	sets := []struct {
		keys []ComponentKey
		mask *componentMask
	}{
		{spec.all, &filter.all},
		{spec.any, &filter.any},
		{spec.none, &filter.none},
	}
	for _, set := range sets {
		for _, key := range set.keys {
			if key.world != world {
				return componentFilter{}, ErrForeignComponent
			}
			set.mask.set(key.id, true)
		}
	}
	return filter, nil
}

// Entities returns matching entity IDs in ascending order.
func (subscription *Subscription) Entities() []Entity {
	subscription.world.mutationMu.Lock()
	subscription.mu.RLock()
	entities := make([]Entity, 0)
	if !subscription.closed {
		for archetype := range subscription.archetypes {
			for entity := range archetype.entities {
				if _, ignored := subscription.ignored[entity]; !ignored {
					entities = append(entities, entity)
				}
			}
		}
	}
	subscription.mu.RUnlock()
	subscription.world.mutationMu.Unlock()
	sort.Slice(entities, func(i, j int) bool { return entities[i] < entities[j] })
	return entities
}

// Len returns the current number of matching entities.
func (subscription *Subscription) Len() int {
	subscription.world.mutationMu.Lock()
	subscription.mu.RLock()
	length := 0
	if !subscription.closed {
		for archetype := range subscription.archetypes {
			for entity := range archetype.entities {
				if _, ignored := subscription.ignored[entity]; !ignored {
					length++
				}
			}
		}
	}
	subscription.mu.RUnlock()
	subscription.world.mutationMu.Unlock()
	return length
}

// Ignore excludes entity from this subscription until Include is called.
func (subscription *Subscription) Ignore(entity Entity) {
	subscription.mu.Lock()
	if !subscription.closed {
		subscription.ignored[entity] = struct{}{}
	}
	subscription.mu.Unlock()
}

// Include removes an explicit exclusion.
func (subscription *Subscription) Include(entity Entity) {
	subscription.mu.Lock()
	if !subscription.closed {
		delete(subscription.ignored, entity)
	}
	subscription.mu.Unlock()
}

// Close unregisters the subscription. It is safe to call repeatedly.
func (subscription *Subscription) Close() error {
	subscription.world.mutationMu.Lock()
	defer subscription.world.mutationMu.Unlock()
	subscription.mu.Lock()
	if !subscription.closed {
		subscription.closed = true
		delete(subscription.world.subscriptions, subscription)
		subscription.archetypes = nil
		subscription.ignored = nil
	}
	subscription.mu.Unlock()
	return nil
}
