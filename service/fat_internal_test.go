package service

import (
	"log/slog"
	"reflect"
	"testing"

	"maragu.dev/is"
)

func TestNewFat(t *testing.T) {
	t.Run("discards log output when given no logger", func(t *testing.T) {
		f := NewFat(NewFatOptions{})

		is.True(t, f.log.Handler() == slog.DiscardHandler)
	})

	t.Run("stocks the tracer the package traces with", func(t *testing.T) {
		f := NewFat(NewFatOptions{})

		is.True(t, f.tracer != nil)
	})
}

func TestSetup(t *testing.T) {
	t.Run("wires every operation", func(t *testing.T) {
		// From inside the package and over the fields themselves, because a delegate only knows whether
		// its own field was set, and a hand-written list only covers the operations someone remembered:
		// one that gets a wiring function but never a line in Setup would go missing in production and
		// nowhere else. The capabilities can be nil, since nothing calls them.
		f := NewFat(NewFatOptions{})
		Setup(f, nil, nil, nil)

		fields := reflect.ValueOf(f).Elem()
		var operations int
		for i := range fields.NumField() {
			if fields.Field(i).Kind() != reflect.Func {
				continue
			}
			operations++
			is.True(t, !fields.Field(i).IsNil(), "operation "+fields.Type().Field(i).Name+" is not wired")
		}
		is.True(t, operations > 0, "expected at least one operation")
	})
}
