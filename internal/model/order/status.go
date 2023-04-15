package order

import (
	"go.uber.org/zap"
)

type status int8

func (s status) String() string {
	switch s {
	case StatusNew:
		return "NEW"
	case StatusProcessing:
		return "PROCESSING"
	case StatusInvalid:
		return "INVALID"
	case StatusProcessed:
		return "PROCESSED"
	}

	zap.L().Error("unknown status value", zap.Int8("status", int8(s)))
	return ""
}

const (
	StatusNew status = iota
	StatusProcessing
	StatusInvalid
	StatusProcessed
)
