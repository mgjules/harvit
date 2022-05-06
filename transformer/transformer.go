package transformer

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/dop251/goja"
	"github.com/mgjules/harvit/logger"
	"github.com/mgjules/harvit/plan"
)

// Transform allows the user to transform any harvested data into a desired format
// using a custom transformer.
func Transform(ctx context.Context, transformer string, fields []plan.Field, data map[string]any) any {
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
		logger.Log.ErrorwContext(ctx, "Failed to read transformer", "transformer", transformer, "error", err)

		return data
	}

	if err = vm.Set("fields", fields); err != nil {
		logger.Log.ErrorwContext(ctx, "Failed to set fields", "error", err)

		return data
	}

	if err = vm.Set("data", data); err != nil {
		logger.Log.ErrorwContext(ctx, "Failed to set data", "error", err)

		return data
	}

	if _, err = vm.RunString(string(src)); err != nil {
		logger.Log.ErrorwContext(ctx, "Failed to run transformer", "error", err)

		return data
	}

	return vm.Get("data").Export()
}
