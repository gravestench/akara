package akara

type EventListener struct {
	fn   func(...interface{})
	once bool
}
