# Golang Entity Component System
A Golang Entity Component System implementation

## Foreword
This module only provides a concrete implementation of a `World`, a `ComponentManager`, an
 `EntityManager`, and a couple of utility systems. You will need to create concrete
  implementations of all components and "actual" systems
 and I'll describe how you should go about doing that below.
 
 **If you really want to just see how this package is used, skip ahead to the Examples section.**

### World
The `World` contains the `EntityManager`, `ComponentManager`, and a slice of all `Systems`.
The world is created using a `WorldConfig`, which is built from systems:
```golang
cfg := akara.NewWorldConfig().
        With(NewMovementSystem()).
        With(NewRenderSystem()).
        With(NewPhysicsSystem())

world := akara.NewWorld(cfg)
```

After the world is created, calling the `Update` method will call the `Process` methods of all of
 the systems
 ```golang
elapsedTimeSinceLastUpdate = time.Millisecond * 16
world.Update(durationSinceLastUpdate)
 ```

### Entities and the EntityManager
Entities are just unique `uint64`'s. They are used to create associations with components
 (among other things...).
 
 For now, all you need to know is that this is essentially what an entity is:
 ```golang
// EID is an entity ID
type EID = uint64
```

The `EntityManager` is responsible for creating new entity IDs, as well as a `BitSet` for each
 entity. Bitsets describe which components an entity currently has. We will talk about BitSets later...

### Components, ComponentMaps, and the ComponentManager
 - **Component** -- Something that has a unique identifier (for the type of component), and a
  means of creating a `ComponentMap`. 

 - **ComponentMap** -- Responsible for creating, retrieving, and deleting instances of
  components. The component map is what maintains the mapping of entity IDs to component instances.

 - **ComponentManager** -- A container for all component maps. Internally, it used the component
  ID as the key into a map of ComponentMaps.

Both `Components` and `ComponentMap`s need to implement these two interfaces.

 ```golang
type ComponentIdentifier interface {
	ID() ComponentID
}

type ComponentMapProvider interface {
	NewMap() ComponentMap
}
```

**It's really important that the component type IDs are unique!**

I think it's worth stating again for clarity the _component IDs are NOT unique to an instance of
 a component, it is an ID for the TYPE of component._

 ### Systems
Systems are fairly simple in that they need only implement this interface:
 ```golang
 type System interface {
 	Active() bool
 	SetActive(bool)
 	Process()
 }
```

However, there are a couple of concrete system types provided which you can use to create your own
 systems.
 
 The first is `BaseSystem`, and it has its own `Active` and `SetActive` methods.
 ```golang
type BaseSystem struct {
	*World
	active bool
}

func (s *BaseSystem) Active() bool {
	return s.active
}

func (s *BaseSystem) SetActive(b bool) {
	s.active = b
} 
```

You can embed the `BaseSystem` in your own system like this:
```golang
type ExampleSystem struct {
    *BaseSystem
}

func (s *ExampleSystem) Process() {
    // do stuff
}
```

However, this is a pretty boring system, and you're probably wondering where the entities are
 going to come from. This leads into...

### Subscriptions, ComponentFilters, and BitSets! Oh, my!
Before we can talk about `Subscription`s or `ComponentFilter`s, we need to know what a `BitSet` is.

#### BitSets
BitSets are just a bunch of booleans, packed up in a slice of `uint64`'s.
```golang
type BitSet struct {
	groups []uint64
}
```

These are used by the EntityManager to signify which components an entity currently has
. **Whenever a ComponentMap adds or removes a component for an entity, the EntityManager will
 update the entity's bitset**.
 
 Remember how the Component types all have a unique ID? That ID corresponds to
  the bit index that is toggled in the BitSet when a component is being added or removed!

#### ComponentFilters
ComponentFilters also use BitSets, but they use them for comparisons against an entity bitset
.
```golang
type ComponentFilter struct {
	Required    *BitSet
	OneRequired *BitSet
	Forbidden   *BitSet
}
```

When an entity bitset is evaluated by a ComponentFilter, each of the Filter's Bitsets is used
to determine if the entity should be allowed to pass through the filter.

When determining if an entity's component bitset will pass through the ComponentFilter:
 - **Required** -- The entity bitset must contain `true` bits for all `true` bits present in the
  Required bitset.
 - **OneRequired** -- The entity bitset must have at least one `true` bit in common with the
  `OneRequired` BitSet
 - **Forbidden** -- The entity bitset must **not** contain any `true` bits for any `true` bits
  present in the `Forbidden` bitset

#### Subscriptions
Subscriptions are simply a combination of a ComponentFilter and a slice of entity ID's:
```golang
type Subscription struct {
	Filter          *ComponentFilter
	entities        []EID
}
```

As Components are added and removed from entities, the entity manager will pass the updated entity
 bitset to the subscription. If the entity bitset passes through the subscription's filter, the
  entity ID is added to the slice of entities for that subscription.
  
  This leads us to the second utility system that is provided... The `SubscriberSystem`!
```golang
type SubscriberSystem struct {
	*BaseSystem
	Subscriptions []*Subscription
}
```

## Examples

#### Enumerating Component ID's
You may want to put all of the component ID's into a single spot for easy enumeration, like this:
```golang
const (
	RenderableCID ComponentID = iota
	PositionCID
	VelocityCID
)
```

#### Example: Component
```golang
type PositionComponent struct {
	*Vector
}

// ID returns a unique identifier for the component type
func (*PositionComponent) ID() akara.ComponentID {
	return PositionCID
}

// NewMap returns a new component map the component type
func (*PositionComponent) NewMap() akara.ComponentMap {
	return NewPositionMap()
}

```

#### Example: ComponentMap
```golang
// NewPositionMap creates a new map of entity ID's to position components
func NewPositionMap() *PositionMap {
	cm := &PositionMap{
		components: make(map[akara.EID]*PositionComponent),
	}

	return cm
}
```

```golang
// PositionMap is a map of entity ID's to position components
type PositionMap struct {
	world      *akara.World
	components map[akara.EID]*PositionComponent
}

// Init initializes the component map with the given world
func (cm *PositionMap) Init(world *akara.World) {
	cm.world = world
}

// ID returns a unique identifier for the component type
func (*PositionMap) ID() akara.ComponentID {
	return PositionCID
}

// NewMap returns a new component map for this component type
func (*PositionMap) NewMap() akara.ComponentMap {
	return NewPositionMap()
}

// Add a new PositionComponent for the given entity id, return that component.
// If the entity already has a component, just return that one.
func (cm *PositionMap) Add(id akara.EID) akara.Component {
	if com, has := cm.components[id]; has {
		return com
	}

	position := NewVector(0, 0)
	cm.components[id] = &PositionComponent{Position: &position}

	cm.world.UpdateEntity(id)

	return cm.components[id]
}

// AddPosition adds a new PositionComponent for the given entity id and returns it.
// If the entity already has a position component, just return that one.
// this is a convenience method for the generic Add method, as it returns a
// *PositionComponent instead of an akara.Component interface
func (cm *PositionMap) AddPosition(id akara.EID) *PositionComponent {
	return cm.Add(id).(*PositionComponent)
}

// Get returns the component associated with the given entity id
func (cm *PositionMap) Get(id akara.EID) (akara.Component, bool) {
	entry, found := cm.components[id]
	return entry, found
}

// GetPosition returns the position component associated with the given entity id
func (cm *PositionMap) GetPosition(id akara.EID) (*PositionComponent, bool) {
	entry, found := cm.components[id]
	return entry, found
}

// Remove a component for the given entity id, return the component.
func (cm *PositionMap) Remove(id akara.EID) {
	delete(cm.components, id)
	cm.world.UpdateEntity(id)
}
```

#### Example: Component `nil` Instance
This may seem silly, and perhaps it is, but look for how it is used in the following examples.
```golang
// Position is a convenient reference to be used as a component identifier
var Position = (*PositionComponent)(nil)
```

#### Example: Creating a ComponentFilter with the filter builder
```golang
cfg := akara.NewFilter().Require(components.Position, components.Velocity)

filter := cfg.Build()
```

#### Example: System
**Subscriber System Implementation**
```golang
// MovementSystem handles entity movement based on velocity and position components
type MovementSystem struct {
	*akara.SubscriberSystem
	positions  *components.PositionMap
	velocities *components.VelocityMap
}

// Init initializes the system with the given world
func (m *MovementSystem) Init(world *akara.World) {
	m.World = world

	if world == nil {
		m.SetActive(false)
		return
	}

	for subIdx := range m.Subscriptions {
		m.AddSubscription(m.Subscriptions[subIdx])
	}

	// try to inject the components we require, then cast the returned
	// abstract ComponentMap back to the concrete implementation
	m.positions = m.InjectMap(components.Position).(*components.PositionMap)
	m.velocities = m.InjectMap(components.Velocity).(*components.VelocityMap)
}

// Process processes all of the Entities
func (m *MovementSystem) Process() {
	for subIdx := range m.Subscriptions {
		entities := m.Subscriptions[subIdx].GetEntities()
		for entIdx := range entities {
			m.ProcessEntity(entities[entIdx])
		}
	}
}

// ProcessEntity updates an individual entity in the movement system
func (m *MovementSystem) ProcessEntity(id akara.EID) {
	position, found := m.positions.GetPosition(id)
	if !found {
		return
	}

	velocity, found := m.velocities.GetVelocity(id)
	if !found {
		return
	}

	s := float64(m.World.TimeDelta) / float64(time.Second)
	position.Vector = *position.Vector.Add(velocity.Vector.Clone().Scale(s))
}
```

**System Factory Function**
```golang
// NewMovementSystem creates a movement system
func NewMovementSystem() *MovementSystem {
	cfg := akara.NewFilter().Require(components.Position, components.Velocity)

	filter := cfg.Build()

	return &MovementSystem{
		SubscriberSystem: akara.NewSubscriberSystem(filter),
	}
}
```

#### Example: Static Checks for Interface Implementation
Wherever you define components and systems, it's good practice to add static checks. These prevent
you from compiling if things don't implement the interfaces that they should.
```golang
// static check that MovementSystem implements the System interface
var _ akara.System = &MovementSystem{}
```
```golang
// static check that PositionComponent implements Component
var _ akara.Component = &PositionComponent{}
```
```golang
// static check that PositionMap implements ComponentMap
var _ akara.ComponentMap = &PositionMap{}
```

#### Creating a world
```golang
cfg := akara.NewWorldConfig() // make a world config

cfg.With(NewMovementSystem()) // add the system

world := akara.NewWorld(cfg) // initialize the world with the config
```