package health

import (
	"context"
	"sync"
	"time"

	"github.com/alexliesenfeld/health"
)

// Check represents a single health check.
type Check struct {
	// The Name must be unique among all checks. Name is a required attribute.
	Name string // Required

	// Check is the check function that will be executed to check availability.
	// This function must return an error if the checked service is considered
	// not available. Check is a required attribute.
	Check func(ctx context.Context) error // Required

	// Timeout will override the global timeout value, if it is smaller than
	// the global timeout (see WithTimeout).
	Timeout time.Duration // Optional

	// MaxTimeInError will set a duration for how long a service must be
	// in an error state until it is considered down/unavailable.
	MaxTimeInError time.Duration // Optional

	// MaxContiguousFails will set a maximum number of contiguous
	// check fails until the service is considered down/unavailable.
	MaxContiguousFails uint // Optional

	// StatusListener allows to set a listener that will be called
	// whenever the AvailabilityStatus (e.g. from "up" to "down").
	StatusListener func(ctx context.Context, name string, state health.CheckState) // Optional

	RefreshPeriod time.Duration
	InitialDelay  time.Duration
}

// Checks holds a list of Check from a list of health.Check.
type Checks struct {
	mu    sync.Mutex
	items map[string]*Check
}

// New returns a new list of Check.
func New() *Checks {
	return &Checks{
		items: make(map[string]*Check),
	}
}

// RegisterChecks registers a list of health.Check.
func (c *Checks) RegisterChecks(checks ...Check) {
	for i := range checks {
		check := &checks[i]

		if check.Name == "" {
			continue
		}

		c.mu.Lock()
		c.items[check.Name] = check
		c.mu.Unlock()
	}
}

// CompileHealthCheckerOption takes a list of Check and returns health.CheckerOption.
func (c *Checks) CompileHealthCheckerOption() []health.CheckerOption {
	var opts []health.CheckerOption
	for _, c := range c.items {
		opts = append(opts, health.WithPeriodicCheck(c.RefreshPeriod, c.InitialDelay, health.Check{
			Name:               c.Name,
			Timeout:            c.Timeout,
			MaxTimeInError:     c.MaxTimeInError,
			MaxContiguousFails: c.MaxContiguousFails,
			StatusListener:     c.StatusListener,
			Check:              c.Check,
		}))
	}

	return opts
}
