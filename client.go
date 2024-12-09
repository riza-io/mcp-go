package mcp

import (
	"context"
	"io"
	"sync"
	"github.com/riza-io/mcp-go/internal/jsonrpc"
)

type StdioClient struct {
	in  io.Reader
	out io.Writer
	scanner *bufio.Scanner
	next int
	lock sync.Mutex
}

func NewStdioClient(stdin io.Reader, stdout io.Writer) *StdioClient {
	return &StdioClient{
		in: stdin,
		out: stdout,
		scanner: bufio.NewScanner(stdin),
	}
}

func (c *StdioClient) sendMessage(ctx context.Context, method string, params, result any) error {
	// Ensure that we are not sending multiple requests at the same time
	c.lock.Lock()
	defer c.lock.Unlock()

	msg := jsonrpc.Message{
		ID:      json.Number(strconv.Itoa(c.next)),
		JsonRPC: "2.0",
		Method:  method,
		Params:  params,
	}

		bs, err := json.Marshal(msg)
		if err != nil {
			return err
		}

	fmt.Fprintln(c.out, string(bs))

	for c.scanner.Scan() {
		line := c.scanner.Text()
	}

	if err := c.scanner.Err(); err != nil {
		return err
	}
	
	// Increment the ID counter
	c.next++

	return nil
}


func (c *Client) Initialize(ctx context.Context, request *Request[InitializeRequest]) (*Response[InitializeResponse], error) {
	var result InitializeResponse
	if err := c.sendMessage(ctx, "initialize", request.Params, &result); err != nil {
		return nil, NewError(1, err)
	}
	return NewResponse(&result), nil
}	

func (c *Client) ListResources(ctx context.Context, request *Request[ListResourcesRequest]) (*Response[ListResourcesResponse], error) {
	return nil, nil
}
