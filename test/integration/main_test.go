//go:build integration

package integration_test

import (
	"os"
	"testing"

	"github.com/zlatovlaska707/avito-kitchen/test/suite"
)

func TestMain(m *testing.M) {
	code := m.Run()
	suite.Close()
	os.Exit(code)
}
