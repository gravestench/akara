// Package akara provides a typed, deterministic entity-component-system runtime.
//
// Components are arbitrary Go data types registered through generic Store values.
// Subscriptions use opaque store handles and cache shared entity archetypes. Systems
// execute sequentially in a caller-driven update loop. Component masks and archetype
// transitions are private; the package does not expose or depend on a bit-set API.
package akara
