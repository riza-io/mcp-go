package stdio

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/riza-io/mcp-go"
)

type Stream struct {
	rlock sync.Mutex
	scan  *bufio.Scanner
	w     io.Writer
	wlock sync.Mutex
}

func NewStream(r io.Reader, w io.Writer) *Stream {
	return &Stream{
		scan: bufio.NewScanner(r),
		w:    w,
	}
}

func (s *Stream) Recv() (*mcp.Message, error) {
	if !s.scan.Scan() {
		return nil, s.scan.Err()
	}
	line := s.scan.Bytes()
	var msg mcp.Message
	if err := json.Unmarshal(line, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (s *Stream) Send(msg *mcp.Message) error {
	bs, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	s.wlock.Lock()
	_, err = fmt.Fprintln(s.w, string(bs))
	s.wlock.Unlock()
	return err
}
