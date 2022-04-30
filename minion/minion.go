package minion

import (
	"fmt"

	"github.com/google/uuid"
)

// Minion does nothing but stores a key (sad).
type Minion struct {
	name string
	key  string
}

// New creates a new Minion.
func New(name, key string) (*Minion, error) {
	if key == "" {
		uuid, err := uuid.NewUUID()
		if err != nil {
			return nil, fmt.Errorf("minion-key: %w", err)
		}

		key = uuid.String()
	}

	return &Minion{name, key}, nil
}

// Key returns the minion's key.
func (m *Minion) String() string {
	return fmt.Sprintf("My name is '%s' and I have a secret key '%s'.", m.name, m.key)
}
