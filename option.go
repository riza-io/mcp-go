package mcp

type Option interface {
	ClientOption
	ServerOption
}

type ClientOption interface {
	applyToClient(c *Client)
}

type ServerOption interface {
	applyToServer(s *serverConfig)
}

func WithInterceptors(interceptors ...Interceptor) Option {
	return &interceptorsOption{interceptors}
}

type interceptorsOption struct {
	Interceptors []Interceptor
}

func (o *interceptorsOption) applyToClient(c *Client) {
	c.interceptors = o.Interceptors
}

func (o *interceptorsOption) applyToServer(s *serverConfig) {
	s.interceptors = o.Interceptors
}
