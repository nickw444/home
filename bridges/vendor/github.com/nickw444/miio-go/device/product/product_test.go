package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetModel(t *testing.T) {
	p, err := GetModel("chuangmi.plug.m1")
	assert.NoError(t, err)
	assert.Equal(t, PowerPlug, p)

	p, err = GetModel("fake")
	assert.Error(t, err)
	assert.Equal(t, Unknown, p)
}
