package group

import (
	"backend/api/group"
)

type ControllerV1 struct{}

func NewV1() group.IGroupV1 {
	return &ControllerV1{}
}
