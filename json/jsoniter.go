//go:build jsoniter
// +build jsoniter

package json

import (
	stdjson "encoding/json"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	// Marshal refers to 'github.com/json-iterator/go.Marshal'.
	Marshal = json.Marshal
	// Unmarshal refers to 'github.com/json-iterator/go.Unmarshal'.
	Unmarshal = json.Unmarshal
	// MarshalIndent refers to 'github.com/json-iterator/go.MarshalIndent'.
	MarshalIndent = json.MarshalIndent
	// NewDecoder refers to 'github.com/json-iterator/go.NewDecoder'.
	NewDecoder = json.NewDecoder
	// NewEncoder refers to 'github.com/json-iterator/go.NewEncoder'.
	NewEncoder = json.NewEncoder
)

// RawMessage refers to 'encoding/json.RawMessage'.
type RawMessage = stdjson.RawMessage
