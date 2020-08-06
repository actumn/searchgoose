package cluster

import (
	"encoding/base64"
	"math/rand"
)

func GenerateNodeId() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(uuid)
}
