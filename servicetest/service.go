// Package servicetest provides testing helpers for the service package.
package servicetest

import (
	"log/slog"
	"testing"

	"app/service"
)

// NewFat for testing, with a logger of its own.
//
// No operation is wired: a test wires the ones it exercises, with the capabilities it is already
// asserting against.
func NewFat(t *testing.T) *service.Fat {
	t.Helper()

	return service.NewFat(service.NewFatOptions{
		Log: slog.New(slog.NewTextHandler(&testWriter{t: t}, nil)),
	})
}

// testWriter logs through the test it belongs to, so an operation says what it did under the test
// that ran it and stays quiet unless that test fails.
type testWriter struct {
	t *testing.T
}

func (t *testWriter) Write(p []byte) (n int, err error) {
	t.t.Log(string(p))
	return len(p), nil
}
