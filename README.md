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
        With(&MovementSystem{}).
        With(&RenderSystem{}).
        With(&PhysicsSystem{})

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

### Components & Component Factories
`Components` only need to implement this interface:
 ```golang
type Component interface {
    New() Component
}
```

Here's an example `Velocity` component:
```golang
type Velocity struct {
	x, y float64
}

func (*Velocity) New() akara.Component {
	return &Velocity{}
}
```

Components must be registered in the ECS world. 
The reason for this is because each distinct type of component is given a `ComponentID` internally.
This ID ends up being used to keep track of what components each entity has.
It's also used to retrieve a `ComponentFactory` instance:
```golang
velocityID := world.RegisterComponent(&Velocity{})
factory := world.GetComponentFactory(velocityID)
```

Component factories are used to create, retrieve, and remove components from entities:
```golang
id := world.RegisterComponent(&Velocity{})
factory := world.GetComponentFactory(id)

e := world.NewEntity()

// returns a akara.Component interface!
c := factory.Add(e) 

// we must cast the interface to our concrete component type
v := c.(*Velocity) 
```

Notice how we always have to cast the returned interface back to our concrete component implementation?
We can get around this by making a concrete component factory:
```golang
type VelocityFactory struct {
	*akara.ComponentFactory
}

func (m *VelocityFactory) Add(id akara.EID) *Velocity {
	return m.ComponentFactory.Add(id).(*Velocity)
}

func (m *VelocityFactory) Get(id akara.EID) (*Velocity, bool) {
	component, found := m.ComponentFactory.Get(id)
	if !found {
		return nil, found
	}

	return component.(*Velocity), found
}
```

this allows us to just use `Add` and `Get` without having to cast the returned value.

It's worth mentioning that each distinct component type that is registered will only have one
component factory and one component ID.
```golang
id1 := world.RegisterComponent(&Velocity{})
id2 := world.RegisterComponent(&Velocity{})

isSame := id1 == id2 // true 
```

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

The second type of system is a `SubscriberSystem`, but before we talk about that we need to talk about subscriptions...

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

#### Creating a world
```golang
// make a world config
cfg := akara.NewWorldConfig()

// add systems to the config
cfg.With(&MovementSystem{})

// create a world instance using the world config
world := akara.NewWorld(cfg) 
```

#### Declaring a Component
Here is the bare minimum required to create a new component:
```golang
type Velocity struct {
	*Vector3
}

func (*Velocity) New() akara.Component {
	return &Velocity{
		Vector3: NewVector3(0, 0, 0),
	}
}
```
Initialization logic specific to this component (like creating instances of another *struct) belongs 
inside of the `New` method. 

#### Concrete Component Factories
A concrete component factory is just a wrapper for the generic component factory, 
but it casts the returned values from `Add` and `Get` to the concrete component implementation.
This is just to prevent you from having to cast the component interface to struct pointers.
```golang
type VelocityFactory struct {
	*akara.ComponentFactory
}

func (m *VelocityFactory) Add(id akara.EID) *Velocity {
	return m.ComponentFactory.Add(id).(*Velocity)
}

func (m *VelocityFactory) Get(id akara.EID) (*Velocity, bool) {
	component, found := m.ComponentFactory.Get(id)
	if !found {
		return nil, found
	}

	return component.(*Velocity), found
}
```

#### Creating a ComponentFilter
```golang
cfg := akara.NewFilter()

cfg.Require(
    components.Position, 
    components.Velocity,
    )

filter := cfg.Build()
```

#### Example System (with Subscriptions!)
For systems that use subscriptions, it is recommended that you embed an `akara.BaseSubscriberSystem` as it 
provides the generic methods for dealing with subscriptions. It also contains an `akara.BaseSystem`, which has other
generic system methods.

It is also recommended that all component factories be placed inside of a common struct and given explicit names.
This helps to keep the code clear when writing complicated systems.

As of writing, there is no general guide for how subscriptions are managed, but just embedding them in the system struct
and giving them descriptive names is sufficient.
```golang
type MovementSystem struct {
    akara.SubscriberSystem
    components struct {
        positions   PositionFactory
        velocities  VelocityFactory
    }
    movingEntities []akara.EID
}
```

As of writing, all systems should set their `World` and call `SetActive(false)` if world is nil. 
After that, actual system initialization logic is added: 
```golang
func (m *MovementSystem) Init(world *akara.World) {
	m.World = world

	if world == nil {
		m.SetActive(false)
		return
	}

    m.setupComponents()
    m.setupSubscriptions()
}
```

Here, we use `BaseSystem.InjectComponent`, which registers a component and assigns a component factory to the given destination.
```golang
func (m *MovementSystem) setupComponents() {
    m.InjectComponent(&Position{}, &m.components.Position.ComponentFactory)
    m.InjectComponent(&Velocity{}, &m.components.Velocity.ComponentFactory)
}
```

Here, we set up our only subscription. For this example, our `MovementSystem` is interested in
`Position` and `Velocity` components.
```golang
func (m *MovementSystem) setupSubscriptions() {
    filterBuilder := m.NewComponentFilter()
    
    filterBuilder.Require(&Position{})
    filterBuilder.Require(&Velocity{})

    filter := filterBuilder.Build()

	m.movingEntities = m.World.AddSubscription(filter)
}
```

Our `Update` method is simple; we iterate over entities that are in our subscription. 
Remember, as components are added and removed from entities, _all subscriptions are updated_.
```golang
func (m *MovementSystem) Update() {
	for _, eid := range m.movingEntities.GetEntities() {
        m.moveEntity(eid)
	}
}
```

This is where our system actually does work. For the given incoming entity id, we retreive 
the position and velocity components and apply the velocity to the position. If either of those 
components does not exist for the entity, we return. 
```golang
func (m *MovementSystem) moveEntity(id akara.EID) {
	position, found := m.components.positions.Get(id)
	if !found {
		return
	}

	velocity, found := m.components.velocities.Get(id)
	if !found {
		return
	}

	s := float64(m.World.TimeDelta) / float64(time.Second)
	position.Vector.Add(velocity.Vector.Clone().Scale(s))
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