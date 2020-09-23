package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMurMur3Hash(t *testing.T) {
	assert.Equal(t, 0x5a0cb7c3, MurMur3Hash("hell"))
	assert.Equal(t, 0xd7c31989, MurMur3Hash("hello"))
	assert.Equal(t, 0x22ab2984, MurMur3Hash("hello w"))
	assert.Equal(t, 0xdf0ca123, MurMur3Hash("hello wo"))
	assert.Equal(t, 0xe7744d61, MurMur3Hash("hello wor"))

	assert.Equal(t, 0xe07db09c, MurMur3Hash("The quick brown fox jumps over the lazy dog"))
	assert.Equal(t, 0x4e63d2ad, MurMur3Hash("The quick brown fox jumps over the lazy cog"))
}
