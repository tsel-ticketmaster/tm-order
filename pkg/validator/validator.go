package validator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	vld         *validator.Validate
	vldSyncOnce sync.Once
)

func new() *validator.Validate {
	vld := validator.New()

	return vld
}

func Get() *validator.Validate {
	vldSyncOnce.Do(func() {
		vld = new()
	})

	return vld
}
