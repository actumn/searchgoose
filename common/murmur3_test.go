package common

import "testing"

func TestMurMur3Hash(t *testing.T) {
	println(MurMur3Hash("hello"))
	println(0xd7c31989)
}
