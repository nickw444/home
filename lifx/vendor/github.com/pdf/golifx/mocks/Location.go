package mocks

import "github.com/stretchr/testify/mock"

type Location struct {
	Group
	mock.Mock
}
