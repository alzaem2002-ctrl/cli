package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSession(t *testing.T) {
	assert.True(t, IsSessionID("00000000-0000-0000-0000-000000000000"))
	assert.True(t, IsSessionID("e2fa49d2-f164-4a56-ab99-498090b8fcdf"))
	assert.True(t, IsSessionID("E2FA49D2-F164-4A56-AB99-498090B8FCDF"))

	assert.False(t, IsSessionID(""))
	assert.False(t, IsSessionID(" "))
	assert.False(t, IsSessionID("\n"))
	assert.False(t, IsSessionID("not-a-uuid"))
	assert.False(t, IsSessionID("000000000000000000000000000000000000"))
	assert.False(t, IsSessionID("00000000-0000-0000-0000-000000000000-extra"))
}
