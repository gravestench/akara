package akara

// archetype is one interned component combination. Its mask is immutable after
// creation and shared by every entity in the archetype.
type archetype struct {
	mask        componentMask
	entities    map[Entity]struct{}
	transitions map[archetypeTransition]*archetype
}

type archetypeTransition struct {
	componentID uint32
	present     bool
}

func (world *World) internArchetypeLocked(mask componentMask) *archetype {
	hash := mask.hash()
	for _, existing := range world.archetypes[hash] {
		if existing.mask.equals(mask) {
			return existing
		}
	}
	archetype := &archetype{
		mask:        mask,
		entities:    make(map[Entity]struct{}),
		transitions: make(map[archetypeTransition]*archetype),
	}
	world.archetypes[hash] = append(world.archetypes[hash], archetype)
	world.archetypeCount++
	for subscription := range world.subscriptions {
		subscription.mu.Lock()
		if !subscription.closed && subscription.filter.allow(mask) {
			subscription.archetypes[archetype] = struct{}{}
		}
		subscription.mu.Unlock()
	}
	return archetype
}

func (world *World) transitionEntityLocked(entity Entity, state entityState, componentID uint32, present bool) {
	transition := archetypeTransition{componentID: componentID, present: present}
	target, found := state.archetype.transitions[transition]
	if !found {
		mask := state.archetype.mask.with(componentID, present)
		target = world.internArchetypeLocked(mask)
		state.archetype.transitions[transition] = target
		target.transitions[archetypeTransition{componentID: componentID, present: !present}] = state.archetype
	}
	if target == state.archetype {
		return
	}
	delete(state.archetype.entities, entity)
	target.entities[entity] = struct{}{}
	state.archetype = target
	world.entities[entity] = state
}

// ArchetypeCount returns the number of component combinations interned by the
// world, including the empty archetype.
func (world *World) ArchetypeCount() int {
	world.mutationMu.Lock()
	count := world.archetypeCount
	world.mutationMu.Unlock()
	return count
}
