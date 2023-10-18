package util

import (
	"github.com/google/uuid"
)

var FissionTempDir = "/tmp/fission-" + uuid.New().String()
var KnTempDir = "/tmp/kn-" + uuid.New().String()
var FaasTempDir = "template"

type KnComponent int32

const (
	TYPE_SERVICE  KnComponent = 0
	TYPE_BROKER   KnComponent = 1
	TYPE_SEQUENCE KnComponent = 2
)
