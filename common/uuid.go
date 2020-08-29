package common

import (
	"encoding/base64"
	"math/rand"
	"time"
)

func RandomBase64() string {
	uuid := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	rand.Read(uuid)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(uuid)
}
