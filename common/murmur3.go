package common

import (
	"math/bits"
	"unsafe"
)

const (
	c1 = 0xcc9e2d51
	c2 = 0x1b873593
)

func MurMur3Hash(routing string) int {
	bytesToHash := make([]byte, len(routing)*2)

	for i := 0; i < len(routing); i++ {
		c := routing[i]
		bytesToHash[i*2] = c
		bytesToHash[i*2+1] = c >> 8
	}
	return int(murmur32(bytesToHash, 0))
}

func murmur32(data []byte, seed uint32) uint32 {
	h1 := seed

	nblocks := len(data) / 4
	p := uintptr(unsafe.Pointer(&data[0]))
	p1 := p + uintptr(4*nblocks)
	for ; p < p1; p += 4 {
		k1 := *(*uint32)(unsafe.Pointer(p))

		k1 *= c1
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2

		h1 ^= k1
		h1 = bits.RotateLeft32(h1, 13)
		h1 = h1*4 + h1 + 0xe6546b64
	}

	tail := data[nblocks*4:]
	var k1 uint32
	switch len(tail) & 3 {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2
		h1 ^= k1
	}

	h1 ^= uint32(len(data))

	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}
