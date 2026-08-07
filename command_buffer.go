package akara

import (
	"errors"
	"fmt"
	"sync"
)

var ErrCommandBufferApplied = errors.New("akara: command buffer already applied")

type structuralCommand func() error

// CommandBuffer defers structural world mutations until a safe barrier. It is
// safe to populate concurrently. Apply executes commands once in submission order.
type CommandBuffer struct {
	mu       sync.Mutex
	commands []structuralCommand
	applied  bool
}

// NewCommandBuffer creates an empty structural command buffer.
func NewCommandBuffer() *CommandBuffer { return &CommandBuffer{} }

// AddDynamic queues addition or replacement of a runtime-defined component.
// The values map is copied at submission time.
func (buffer *CommandBuffer) AddDynamic(store *DynamicStore, entity Entity, values map[string]any) {
	copyValues := make(map[string]any, len(values))
	for name, value := range values {
		copyValues[name] = value
	}
	buffer.append(func() error {
		if store == nil {
			return ErrNilComponent
		}
		_, err := store.Set(entity, copyValues)
		return err
	})
}

// Remove queues removal of a component type from an entity.
func (buffer *CommandBuffer) Remove(component ComponentType, entity Entity) {
	buffer.append(func() error {
		if component == nil {
			return ErrNilComponent
		}
		key := component.componentKey()
		if key.world == nil {
			return ErrForeignComponent
		}
		key.world.registryMu.RLock()
		if int(key.id) >= len(key.world.stores) {
			key.world.registryMu.RUnlock()
			return ErrForeignComponent
		}
		store := key.world.stores[key.id]
		key.world.registryMu.RUnlock()
		store.remove(entity)
		return nil
	})
}

// Destroy queues entity destruction.
func (buffer *CommandBuffer) Destroy(world *World, entity Entity) {
	buffer.append(func() error {
		if world == nil {
			return ErrEntityNotFound
		}
		world.DestroyEntity(entity)
		return nil
	})
}

func (buffer *CommandBuffer) append(command structuralCommand) {
	buffer.mu.Lock()
	if !buffer.applied {
		buffer.commands = append(buffer.commands, command)
	}
	buffer.mu.Unlock()
}

// Apply executes every queued mutation once. It stops at the first error.
func (buffer *CommandBuffer) Apply() error {
	buffer.mu.Lock()
	if buffer.applied {
		buffer.mu.Unlock()
		return ErrCommandBufferApplied
	}
	buffer.applied = true
	commands := append([]structuralCommand(nil), buffer.commands...)
	buffer.commands = nil
	buffer.mu.Unlock()
	for index, command := range commands {
		if err := command(); err != nil {
			return fmt.Errorf("akara: structural command %d: %w", index, err)
		}
	}
	return nil
}
