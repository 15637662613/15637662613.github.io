package test

import (
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestUuid(t *testing.T) {
	s := uuid.NewV4().String()
	println(s)
}
