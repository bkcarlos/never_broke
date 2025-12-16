package testutil

import (
	"github.com/stretchr/testify/mock"
)

// MockStore is a mock implementation of repository.Store
// Note: This is a basic mock. For tests that need repository mocks,
// define them in the test files to avoid import cycles.
type MockStore struct {
	mock.Mock
}
