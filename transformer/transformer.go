package transformer

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/dop251/goja"
	"github.com/mgjules/harvit/plan"
)

// Transform allows the user to transform any harvested data into a desired format
// using a custom transformer.
func Transform(ctx context.Context, transformer string, fields []plan.Field, data map[string]any) (any, error) {
	vm := goja.New()

	go func() {
		select {
		case <-time.After(2 * time.Second):
			vm.Interrupt("halt")
		case <-ctx.Done():
			vm.Interrupt("halt")
		}
	}()

	src, err := ioutil.ReadFile(filepath.Clean(transformer))
	if err != nil {
		return nil, fmt.Errorf("failed to read transformer: %w", err)
	}

	if err = vm.Set("fields", fields); err != nil {
		return nil, fmt.Errorf("failed to set fields: %w", err)
	}

	if err = vm.Set("data", data); err != nil {
		return nil, fmt.Errorf("failed to set data: %w", err)
	}

	if _, err = vm.RunString(string(src)); err != nil {
		return nil, fmt.Errorf("failed to run transformer: %w", err)
	}

	return vm.Get("data").Export(), nil
}
