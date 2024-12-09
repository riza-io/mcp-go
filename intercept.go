package mcp

import (
	"context"
	"slices"
)

type UnaryFunc func(context.Context, AnyRequest) (AnyResponse, error)

type UnaryInterceptorFunc func(UnaryFunc) UnaryFunc

func (f UnaryInterceptorFunc) WrapUnary(next UnaryFunc) UnaryFunc { return f(next) }

type Interceptor interface {
	WrapUnary(UnaryFunc) UnaryFunc
}

type stack struct {
	interceptors []Interceptor
}

func newStack(interceptors []Interceptor) *stack {
	slices.Reverse(interceptors)
	return &stack{interceptors: interceptors}
}

func (s *stack) WrapUnary(next UnaryFunc) UnaryFunc {
	for _, interceptor := range s.interceptors {
		next = interceptor.WrapUnary(next)
	}
	return next
}
