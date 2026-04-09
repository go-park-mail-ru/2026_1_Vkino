package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	user := &User{}
	assert.Equal(t, "users", user.Name())
}

func TestSession(t *testing.T) {
	session := &TokenPair{}
	assert.Equal(t, "sessions", session.Name())
}
