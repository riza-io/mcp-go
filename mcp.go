package mcp

type Request[T any] struct {
	Params *T

	method   string
	id       string
	metadata map[string]string
}

func (r *Request[T]) ID() string {
	return r.id
}

func (r *Request[T]) Metadata() map[string]string {
	return r.metadata
}

func NewRequest[T any](params *T) *Request[T] {
	return &Request[T]{
		Params: params,
	}
}

// Any returns the concrete request params as an empty interface, so that
// *Request implements the [AnyRequest] interface.
func (r *Request[_]) Any() any {
	return r.Params
}

func (r *Request[_]) Method() string {
	return r.method
}

// internalOnly implements AnyRequest.
func (r *Request[_]) internalOnly() {}

// AnyRequest is the common method set of every [Request], regardless of type
// parameter. It's used in unary interceptors.
//
// To preserve our ability to add methods to this interface without breaking
// backward compatibility, only types defined in this package can implement
// AnyRequest.
type AnyRequest interface {
	Any() any
	ID() string
	Method() string
	internalOnly()
}

type Response[T any] struct {
	Result *T

	id string
}

func NewResponse[T any](result *T) *Response[T] {
	return &Response[T]{
		Result: result,
	}
}

// Any returns the concrete result as an empty interface, so that
// *Response implements the [AnyResponse] interface.
func (r *Response[_]) Any() any {
	return r.Result
}

func (r *Response[_]) ID() string {
	return r.id
}

// internalOnly implements AnyResponse.
func (r *Response[_]) internalOnly() {}

// AnyResponse is the common method set of every [Response], regardless of type
// parameter. It's used in unary interceptors.
//
// Headers and trailers beginning with "Connect-" and "Grpc-" are reserved for
// use by the gRPC and Connect protocols: applications may read them but
// shouldn't write them.
//
// To preserve our ability to add methods to this interface without breaking
// backward compatibility, only types defined in this package can implement
// AnyResponse.
type AnyResponse interface {
	Any() any
	ID() string

	internalOnly()
}

type Error struct {
	code int
	err  error
}

func (e *Error) Error() string {
	return e.err.Error()
}

func NewError(code int, underlying error) *Error {
	return &Error{code: code, err: underlying}
}
