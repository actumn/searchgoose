package common

import "math/bits"

func MurMur3Hash(routing string) int {
	bytesToHash := make([]byte, len(routing)*2)

	for i := 0; i < len(routing); i++ {
		c := routing[i]
		bytesToHash[i*2] = c
		bytesToHash[i*2+1] = c >> 8
	}

	return int(murmur32(bytesToHash, 0, len(bytesToHash), 0))
}

func murmur32(data []byte, offset int, len int, seed uint32) uint32 {
	c1 := uint32(0xcc9e2d51)
	c2 := uint32(0x1b873593)

	h1 := seed
	roundedEnd := offset + (len & 0xfffffffc) // round down to 4 byte block

	for i := offset; i < roundedEnd; i += 4 {
		// little endian load order
		k1 := uint32((data[i] & 0xff) | ((data[i+1] & 0xff) << 8) | ((data[i+2] & 0xff) << 16) | (data[i+3] << 24))
		k1 *= c1
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2

		h1 ^= k1
		h1 = bits.RotateLeft32(h1, 13)
		h1 = h1*5 + 0xe6546b64
	}

	// tail
	k1 := uint32(0)
	switch len & 0x03 {
	case 3:
		k1 = uint32((data[roundedEnd+2] & 0xff) << 16)
		fallthrough
	case 2:
		k1 |= uint32((data[roundedEnd+1] & 0xff) << 8)
		fallthrough
	case 1:
		k1 |= uint32(data[roundedEnd] & 0xff)
		k1 *= c1
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2
		h1 ^= k1
	}

	// finalization
	h1 ^= uint32(len)

	// fmix(h1);
	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}
