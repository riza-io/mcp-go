package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/riza-io/mcp-go"
)

type session struct {
	out chan *mcp.Message
}

type message struct {
	sessionID string
	msg       *mcp.Message
}

type Stream struct {
	mu       sync.RWMutex
	in       chan *mcp.Message
	sessions map[string]*session
}

func writeEvent(w http.ResponseWriter, id string, event string, data string) {
	fmt.Fprintf(w, "id: %d\n", id)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", data)
}

func NewStream(mux *http.ServeMux, sseRoute, messagesRoute string) *Stream {
	s := &Stream{
		in:       make(chan *mcp.Message),
		sessions: make(map[string]*session),
	}

	mux.HandleFunc("POST "+messagesRoute, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sessionID := r.FormValue("session_id")

		var msg mcp.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		msg.Metadata = map[string]string{
			"session": sessionID,
		}

		go func() {
			s.in <- &msg
		}()

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET "+sseRoute, func(w http.ResponseWriter, r *http.Request) {
		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Create a flusher to ensure data is sent immediately
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		id := uuid.New().String()
		out := make(chan *mcp.Message)

		s.mu.Lock()
		s.sessions[id] = &session{
			out: out,
		}
		s.mu.Unlock()

		session := messagesRoute + "?session_id=" + id
		fmt.Println(session)

		writeEvent(w, "1", "endpoint", session)
		flusher.Flush()

		for msg := range out {
			bs, err := json.Marshal(msg)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeEvent(w, msg.ID.String(), "message", string(bs))
			flusher.Flush()
		}
	})

	return s
}

func (s *Stream) Recv() (*mcp.Message, error) {
	return <-s.in, nil
}

func (s *Stream) Send(msg *mcp.Message) error {
	if msg.Metadata == nil {
		return fmt.Errorf("metadata is nil")
	}
	sessionID := msg.Metadata["session"]
	if sessionID == "" {
		return fmt.Errorf("session id is empty")
	}
	s.mu.RLock()
	session, ok := s.sessions[sessionID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session not found")
	}
	go func() {
		session.out <- msg
	}()
	return nil
}
