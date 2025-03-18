package mcp

import (
	"context"
	"strconv"
)

type base struct {
	router       *router
	stream       Stream
	interceptors []Interceptor
}

func (b *base) listen(ctx context.Context, handler func(ctx context.Context, msg *Message) error) error {
	for {
		msg, err := b.stream.Recv()
		if err != nil {
			return err
		}
		if msg == nil {
			continue
		}
		if msg.Method != nil {
			go func() {
				handler(ctx, msg)
			}()
		} else {
			id, err := strconv.ParseUint(msg.ID.String(), 10, 64)
			if err != nil {
				continue
			}
			if inbox, ok := b.router.Remove(id); ok {
				inbox <- msg
			}
		}
	}
}
