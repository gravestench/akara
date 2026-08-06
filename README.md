# Akara

Akara is a small, typed entity-component-system runtime for Go. Components are
ordinary Go data, component stores are generic, subscriptions update immediately,
and systems run deterministically in registration order.

Akara deliberately does not require component interfaces, constructor methods,
factory wrappers, runtime casts, or public bit masks.

## Components

Define components as plain data:

```go
type Position struct {
	X, Y float64
}

type Velocity struct {
	X, Y float64
}
```

Register each type once during application composition:

```go
world := akara.NewWorld()
defer world.Close()

positions := akara.Register[Position](world)
velocities := akara.Register[Velocity](world)
```

`Register[T]` is idempotent. Every call for the same type and world returns the
same `*Store[T]`. Code that does not participate in registration can look up an
existing handle without registering implicitly:

```go
positions, found := akara.GetStore[Position](world)
if !found {
	return errors.New("Position was not registered")
}
```

## Entities and stores

```go
entity, err := world.CreateEntity()
if err != nil {
	return err
}

position, err := positions.Add(entity)
if err != nil {
	return err
}
position.X = 10

velocity, err := velocities.Set(entity, Velocity{X: 4, Y: -2})
if err != nil {
	return err
}

position, found := positions.Get(entity)
removed := velocities.Remove(entity)
destroyed := world.DestroyEntity(entity)
```

`Add` creates the component's zero value or returns the existing pointer. `Set`
replaces it with a copied value. Adding a component to an unknown entity returns
`ErrEntityNotFound`; orphan components cannot be created.

`DestroyEntity` immediately removes every component and subscription membership.
Entity IDs are world-local, begin at one, and are never reused.

Stores and ECS indexes are safe for concurrent use. A component pointer returned
by `Add`, `Set`, or `Get` is ordinary application data: callers must coordinate
concurrent field mutation themselves.

## Sharing store handles

Prefer explicit dependency injection. A small application registry makes system
dependencies visible and avoids service-location during updates:

```go
type Components struct {
	Positions  *akara.Store[Position]
	Velocities *akara.Store[Velocity]
}

func RegisterComponents(world *akara.World) Components {
	return Components{
		Positions:  akara.Register[Position](world),
		Velocities: akara.Register[Velocity](world),
	}
}
```

Pass `Components` or the individual stores to system constructors. Use
`GetStore[T]` for occasional lookup across package boundaries.

## Subscriptions

Subscriptions accept typed store handles directly:

```go
moving, err := world.Subscribe(
	akara.All(positions, velocities),
	akara.Any(playerControlled, aiControlled),
	akara.None(frozen),
)
if err != nil {
	return err
}
defer moving.Close()

for _, entity := range moving.Entities() {
	position, _ := positions.Get(entity)
	velocity, _ := velocities.Get(entity)
	// Update component data.
}
```

- `All` requires every supplied component.
- `Any` requires at least one supplied component.
- `None` forbids every supplied component.
- An empty filter matches every entity.

Membership changes immediately when components are added or removed. `Entities`
returns a sorted snapshot, making iteration stable even if the world changes later.
A filter cannot accidentally use a store from another world; `Subscribe` returns
`ErrForeignComponent`.

Subscriptions are independent. `Ignore` and `Include` affect only that subscription.
Call `Close` when a subscription is no longer needed.

## Deterministic systems

Akara does not start background tick goroutines. The application supplies delta
time, and systems run sequentially in registration order:

```go
type MovementSystem struct {
	positions *akara.Store[Position]
	velocities *akara.Store[Velocity]
	moving *akara.Subscription
}

func (system *MovementSystem) Update(_ *akara.World, delta time.Duration) error {
	seconds := delta.Seconds()
	for _, entity := range system.moving.Entities() {
		position, _ := system.positions.Get(entity)
		velocity, _ := system.velocities.Get(entity)
		position.X += velocity.X * seconds
		position.Y += velocity.Y * seconds
	}
	return nil
}

movement := &MovementSystem{
	positions: positions,
	velocities: velocities,
	moving: moving,
}

systemID, err := world.AddSystem(movement)
if err != nil {
	return err
}

if err := world.Update(frameDelta); err != nil {
	return err
}
```

`SystemFunc` adapts functions to `System`. `RemoveSystem` uses the `SystemID`
returned by registration. An update stops and wraps the first system error.
Concurrent calls to `Update` are serialized.

## Lifecycle

`NewWorldWithContext` binds world shutdown to context cancellation:

```go
world := akara.NewWorldWithContext(gameContext)
defer world.Close()
```

`Close` is idempotent. It closes `Done` and rejects subsequent mutations,
subscriptions, system registration, and updates with `ErrWorldClosed`. Existing
component values remain readable for diagnostics or shutdown persistence.

## Application integration

Create one world and one typed component registry in the application's composition
root. Construct systems with explicit store dependencies, register them in the desired
simulation order, and call `world.Update(delta)` from the main loop. This produces
deterministic component ownership without scheduler goroutines or hidden registration.

Entity IDs and component handles are runtime-local. Save games and network protocols
should use stable application IDs and schema names rather than Akara internals.

## Internal representation

Akara interns each distinct component combination as an archetype. Every entity
holds one archetype pointer; entities with the same component types share one dense
component mask. Adding or removing a component moves the entity along a cached
archetype transition edge.

Subscriptions cache matching archetypes rather than duplicating an entity-membership
map. A filter is evaluated once when a new archetype appears, after which all entities
in that archetype match implicitly.

This design supports large component registries without allocating a mask per entity.
For example, a 10,000-type registry needs roughly 1.25 KB for a fully expanded mask,
but that mask is paid once per used component combination rather than once per entity.
Unused registered types consume no entity-mask memory. `ArchetypeCount` is available
for diagnostics.

The mask and archetype graph are private implementation details. Akara has no
dependency on the standalone BitSet module and can change representations without
affecting application code.

## Development

```bash
go test ./...
go test -race ./...
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run ./...
```
