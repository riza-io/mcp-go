package mcp

type Option interface {
	ClientOption
	ServerOption
}

type ClientOption interface {
	applyToClient(c *StdioClient)
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

func (o *interceptorsOption) applyToClient(c *StdioClient) {
	c.interceptors = o.Interceptors
}

func (o *interceptorsOption) applyToServer(s *serverConfig) {
	s.interceptors = o.Interceptors
}
