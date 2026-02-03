package events

// Payload defines an abstraction for data that can be serialized to JSON for event delivery.
// Implementations must provide ToJSON which returns a valid JSON string representation of the payload.
// The method should perform necessary marshaling/formatting and ensure the result is ready for inclusion
// in event messages and logs (e.g., appropriate escaping and omission of sensitive fields).
type Payload interface {
	ToJSON() string
}
