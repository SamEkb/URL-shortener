// Code generated by mockery v2. Do not edit.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
)

// URLDeleter is an autogenerated mock type for the URLDeleter type
type URLDeleter struct {
	mock.Mock
}

// DeleteURL provides a mock function with given fields: alias
func (_m *URLDeleter) DeleteURL(alias string) error {
	ret := _m.Called(alias)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(alias)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
