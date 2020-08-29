package cluster

import (
	"github.com/actumn/searchgoose/common"
)

func GenerateNodeId() string {
	return common.RandomBase64()
}
