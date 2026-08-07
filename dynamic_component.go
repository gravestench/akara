package akara

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"sync"
)

var (
	ErrInvalidSchema      = errors.New("akara: invalid component schema")
	ErrSchemaConflict     = errors.New("akara: component schema conflicts with existing registration")
	ErrFieldNotFound      = errors.New("akara: component field does not exist")
	ErrInvalidFieldValue  = errors.New("akara: invalid component field value")
	ErrComponentNotFound  = errors.New("akara: entity does not have component")
	ErrComponentReference = errors.New("akara: component reference is stale")
	ErrSchemaVersion      = errors.New("akara: schema version must increase")
)

// FieldKind is the storage and validation type of a runtime-defined field.
type FieldKind uint8

const (
	FieldBool FieldKind = iota + 1
	FieldInt64
	FieldUint64
	FieldFloat64
	FieldString
	FieldEntity
)

// Field defines one named field and its optional default value. A nil default
// selects the zero value for Kind.
type Field struct {
	Name    string
	Kind    FieldKind
	Default any
}

// Schema defines a runtime component type. Name is the stable application-level
// identity; Version is available to adapters implementing schema migration.
type Schema struct {
	Name    string
	Version uint32
	Fields  []Field
}

type compiledField struct {
	Field
	index int
}

type compiledSchema struct {
	Schema
	fieldsByName map[string]compiledField
	defaults     []any
}

type dynamicInstance struct {
	generation uint64
	values     []any
}

// DynamicStore owns instances of one runtime-defined schema.
type DynamicStore struct {
	world     *World
	key       ComponentKey
	schema    compiledSchema
	mu        sync.RWMutex
	instances map[Entity]*dynamicInstance
	nextGen   uint64
}

// Migration converts one component snapshot to values for a replacement schema.
type Migration func(entity Entity, previous map[string]any) (map[string]any, error)

func (s *DynamicStore) componentKey() ComponentKey { return s.key }
func (s *DynamicStore) Key() ComponentKey          { return s.key }

// Schema returns a defensive copy of the registered schema.
func (s *DynamicStore) Schema() Schema {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := s.schema.Schema
	result.Fields = append([]Field(nil), result.Fields...)
	return result
}

// Migrate transactionally replaces this store's schema and all of its component
// instances. The name must remain stable and the version must increase. If any
// conversion or validation fails, neither schema nor instances are changed.
func (s *DynamicStore) Migrate(next Schema, migrate Migration) error {
	if migrate == nil {
		return fmt.Errorf("%w: nil migration", ErrInvalidSchema)
	}
	compiled, err := compileSchema(next)
	if err != nil {
		return err
	}
	s.world.mutationMu.Lock()
	defer s.world.mutationMu.Unlock()
	if s.world.closed.Load() {
		return ErrWorldClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if compiled.Name != s.schema.Name {
		return fmt.Errorf("%w: migration changes name %q to %q", ErrSchemaConflict, s.schema.Name, compiled.Name)
	}
	if compiled.Version <= s.schema.Version {
		return fmt.Errorf("%w: current %d, next %d", ErrSchemaVersion, s.schema.Version, compiled.Version)
	}
	replacements := make(map[Entity]*dynamicInstance, len(s.instances))
	entities := make([]Entity, 0, len(s.instances))
	for entity := range s.instances {
		entities = append(entities, entity)
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i] < entities[j] })
	for _, entity := range entities {
		instance := s.instances[entity]
		previous := make(map[string]any, len(s.schema.Fields))
		for index, field := range s.schema.Fields {
			previous[field.Name] = instance.values[index]
		}
		values, err := migrate(entity, previous)
		if err != nil {
			return fmt.Errorf("akara: migrate %q entity %d: %w", s.schema.Name, entity, err)
		}
		normalized := append([]any(nil), compiled.defaults...)
		for name, value := range values {
			field, found := compiled.fieldsByName[name]
			if !found {
				return fmt.Errorf("akara: migrate %q entity %d: %w: %q", s.schema.Name, entity, ErrFieldNotFound, name)
			}
			value, err = normalizeFieldValue(field.Kind, value)
			if err != nil {
				return fmt.Errorf("akara: migrate %q entity %d: %w: %q: %v", s.schema.Name, entity, ErrInvalidFieldValue, name, err)
			}
			normalized[field.index] = value
		}
		s.nextGen++
		replacements[entity] = &dynamicInstance{generation: s.nextGen, values: normalized}
	}
	s.schema = compiled
	s.instances = replacements
	return nil
}

// RegisterSchema returns the canonical runtime-defined store for schema.Name.
// Re-registering an identical schema is idempotent; incompatible declarations
// return ErrSchemaConflict.
func RegisterSchema(world *World, schema Schema) (*DynamicStore, error) {
	if world == nil {
		return nil, fmt.Errorf("%w: nil world", ErrInvalidSchema)
	}
	compiled, err := compileSchema(schema)
	if err != nil {
		return nil, err
	}
	world.mutationMu.Lock()
	defer world.mutationMu.Unlock()
	if world.closed.Load() {
		return nil, ErrWorldClosed
	}
	world.registryMu.Lock()
	defer world.registryMu.Unlock()
	if existing := world.storesByName[compiled.Name]; existing != nil {
		if reflect.DeepEqual(existing.schema.Schema, compiled.Schema) {
			return existing, nil
		}
		return nil, fmt.Errorf("%w: %q", ErrSchemaConflict, compiled.Name)
	}
	if len(world.stores) > math.MaxUint32 {
		return nil, fmt.Errorf("%w: component type limit exhausted", ErrInvalidSchema)
	}
	store := &DynamicStore{
		world:     world,
		key:       ComponentKey{id: uint32(len(world.stores)), world: world},
		schema:    compiled,
		instances: make(map[Entity]*dynamicInstance),
	}
	world.storesByName[compiled.Name] = store
	world.stores = append(world.stores, store)
	return store, nil
}

// GetDynamicStore looks up an already registered runtime component by name.
func GetDynamicStore(world *World, name string) (*DynamicStore, bool) {
	if world == nil {
		return nil, false
	}
	world.registryMu.RLock()
	store, found := world.storesByName[name]
	world.registryMu.RUnlock()
	return store, found
}

func compileSchema(schema Schema) (compiledSchema, error) {
	schema.Name = strings.TrimSpace(schema.Name)
	if schema.Name == "" {
		return compiledSchema{}, fmt.Errorf("%w: empty name", ErrInvalidSchema)
	}
	result := compiledSchema{Schema: schema, fieldsByName: make(map[string]compiledField)}
	result.Fields = append([]Field(nil), schema.Fields...)
	result.defaults = make([]any, len(schema.Fields))
	for index, field := range result.Fields {
		field.Name = strings.TrimSpace(field.Name)
		if field.Name == "" {
			return compiledSchema{}, fmt.Errorf("%w: field %d has empty name", ErrInvalidSchema, index)
		}
		if _, exists := result.fieldsByName[field.Name]; exists {
			return compiledSchema{}, fmt.Errorf("%w: duplicate field %q", ErrInvalidSchema, field.Name)
		}
		value, err := normalizeFieldValue(field.Kind, field.Default)
		if err != nil {
			return compiledSchema{}, fmt.Errorf("%w: default for %q: %v", ErrInvalidSchema, field.Name, err)
		}
		field.Default = value
		result.Fields[index] = field
		result.defaults[index] = value
		result.fieldsByName[field.Name] = compiledField{Field: field, index: index}
	}
	return result, nil
}

func normalizeFieldValue(kind FieldKind, value any) (any, error) {
	if value == nil {
		switch kind {
		case FieldBool:
			return false, nil
		case FieldInt64:
			return int64(0), nil
		case FieldUint64:
			return uint64(0), nil
		case FieldFloat64:
			return float64(0), nil
		case FieldString:
			return "", nil
		case FieldEntity:
			return Entity(0), nil
		default:
			return nil, fmt.Errorf("unknown field kind %d", kind)
		}
	}
	switch kind {
	case FieldBool:
		if typed, ok := value.(bool); ok {
			return typed, nil
		}
	case FieldInt64:
		switch typed := value.(type) {
		case int:
			return int64(typed), nil
		case int64:
			return typed, nil
		}
	case FieldUint64:
		switch typed := value.(type) {
		case uint:
			return uint64(typed), nil
		case uint64:
			return typed, nil
		}
	case FieldFloat64:
		switch typed := value.(type) {
		case float32:
			return float64(typed), nil
		case float64:
			return typed, nil
		}
	case FieldString:
		if typed, ok := value.(string); ok {
			return typed, nil
		}
	case FieldEntity:
		if typed, ok := value.(Entity); ok {
			return typed, nil
		}
	default:
		return nil, fmt.Errorf("unknown field kind %d", kind)
	}
	return nil, fmt.Errorf("kind %d does not accept %T", kind, value)
}

// Add creates a component with schema defaults or returns the existing reference.
func (s *DynamicStore) Add(entity Entity) (*DynamicComponent, error) {
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
	instance, found := s.instances[entity]
	if !found {
		s.nextGen++
		instance = &dynamicInstance{
			generation: s.nextGen,
			values:     append([]any(nil), s.schema.defaults...),
		}
		s.instances[entity] = instance
	}
	s.mu.Unlock()
	if !found {
		s.world.transitionEntityLocked(entity, state, s.key.id, true)
	}
	return &DynamicComponent{store: s, entity: entity, generation: instance.generation}, nil
}

// Set replaces a component using schema defaults overridden by values.
func (s *DynamicStore) Set(entity Entity, values map[string]any) (*DynamicComponent, error) {
	normalized, err := s.values(values)
	if err != nil {
		return nil, err
	}
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
	_, found := s.instances[entity]
	s.nextGen++
	instance := &dynamicInstance{generation: s.nextGen, values: normalized}
	s.instances[entity] = instance
	s.mu.Unlock()
	if !found {
		s.world.transitionEntityLocked(entity, state, s.key.id, true)
	}
	return &DynamicComponent{store: s, entity: entity, generation: instance.generation}, nil
}

func (s *DynamicStore) values(values map[string]any) ([]any, error) {
	result := append([]any(nil), s.schema.defaults...)
	for name, value := range values {
		field, found := s.schema.fieldsByName[name]
		if !found {
			return nil, fmt.Errorf("%w: %q", ErrFieldNotFound, name)
		}
		normalized, err := normalizeFieldValue(field.Kind, value)
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %v", ErrInvalidFieldValue, name, err)
		}
		result[field.index] = normalized
	}
	return result, nil
}

// Get returns a generation-checked reference to entity's component.
func (s *DynamicStore) Get(entity Entity) (*DynamicComponent, bool) {
	s.mu.RLock()
	instance, found := s.instances[entity]
	s.mu.RUnlock()
	if !found {
		return nil, false
	}
	return &DynamicComponent{store: s, entity: entity, generation: instance.generation}, true
}

func (s *DynamicStore) Has(entity Entity) bool { return s.has(entity) }
func (s *DynamicStore) has(entity Entity) bool {
	s.mu.RLock()
	_, found := s.instances[entity]
	s.mu.RUnlock()
	return found
}

func (s *DynamicStore) Remove(entity Entity) bool {
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

func (s *DynamicStore) remove(entity Entity) bool { return s.Remove(entity) }

func (s *DynamicStore) removeLocked(entity Entity) {
	s.mu.Lock()
	delete(s.instances, entity)
	s.mu.Unlock()
}
func (s *DynamicStore) Len() int {
	s.mu.RLock()
	length := len(s.instances)
	s.mu.RUnlock()
	return length
}
func (s *DynamicStore) Entities() []Entity {
	s.mu.RLock()
	entities := make([]Entity, 0, len(s.instances))
	for entity := range s.instances {
		entities = append(entities, entity)
	}
	s.mu.RUnlock()
	sort.Slice(entities, func(i, j int) bool { return entities[i] < entities[j] })
	return entities
}

// DynamicComponent is a checked reference to one runtime component instance.
// Replacing or removing the component invalidates prior references.
type DynamicComponent struct {
	store      *DynamicStore
	entity     Entity
	generation uint64
}

func (c *DynamicComponent) Get(name string) (any, error) {
	field, found := c.store.schema.fieldsByName[name]
	if !found {
		return nil, fmt.Errorf("%w: %q", ErrFieldNotFound, name)
	}
	c.store.mu.RLock()
	defer c.store.mu.RUnlock()
	instance, found := c.store.instances[c.entity]
	if !found || instance.generation != c.generation {
		return nil, ErrComponentReference
	}
	return instance.values[field.index], nil
}

func (c *DynamicComponent) Set(name string, value any) error {
	field, found := c.store.schema.fieldsByName[name]
	if !found {
		return fmt.Errorf("%w: %q", ErrFieldNotFound, name)
	}
	normalized, err := normalizeFieldValue(field.Kind, value)
	if err != nil {
		return fmt.Errorf("%w: %q: %v", ErrInvalidFieldValue, name, err)
	}
	c.store.mu.Lock()
	defer c.store.mu.Unlock()
	instance, found := c.store.instances[c.entity]
	if !found || instance.generation != c.generation {
		return ErrComponentReference
	}
	instance.values[field.index] = normalized
	return nil
}

// Snapshot returns a defensive field-value map.
func (c *DynamicComponent) Snapshot() (map[string]any, error) {
	c.store.mu.RLock()
	defer c.store.mu.RUnlock()
	instance, found := c.store.instances[c.entity]
	if !found || instance.generation != c.generation {
		return nil, ErrComponentReference
	}
	result := make(map[string]any, len(c.store.schema.Fields))
	for index, field := range c.store.schema.Fields {
		result[field.Name] = instance.values[index]
	}
	return result, nil
}
