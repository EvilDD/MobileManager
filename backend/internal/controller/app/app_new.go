package app

import (
	"backend/api/app"
)

type ControllerV1 struct{}

func NewV1() app.IAppV1 {
	return &ControllerV1{}
}
