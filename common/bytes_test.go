package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBytes(t *testing.T) {
	assert.Equal(t, "120 KiB", IBytes(123123))
	assert.Equal(t, "117 MiB", IBytes(123123123))
	assert.Equal(t, "115 GiB", IBytes(123123123123))
	assert.Equal(t, "112 TiB", IBytes(123123123123123))
}
