//go:build e2e

package api_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	resetSimulatorState()
	os.Exit(m.Run())
}
