//go:build !jsoniter
// +build !jsoniter

package json

import "encoding/json"

var (
	// Marshal refers to 'encoding/json.Marshal'.
	Marshal = json.Marshal
	// Unmarshal refers to 'encoding/json.Unmarshal'.
	Unmarshal = json.Unmarshal
	// MarshalIndent refers to 'encoding/json.MarshalIndent'.
	MarshalIndent = json.MarshalIndent
	// NewDecoder refers to 'encoding/json.NewDecoder'.
	NewDecoder = json.NewDecoder
	// NewEncoder refers to 'encoding/json.NewEncoder'.
	NewEncoder = json.NewEncoder
)

// RawMessage refers to 'encoding/json.RawMessage'.
type RawMessage = json.RawMessage
