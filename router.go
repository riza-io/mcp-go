package mcp

import "sync"

type router struct {
	lock  sync.Mutex
	next  uint64
	boxes map[uint64]chan *Message
}

func newRouter() *router {
	return &router{
		boxes: make(map[uint64]chan *Message),
	}
}

func (r *router) Add() (uint64, chan *Message) {
	r.lock.Lock()
	id := r.next
	r.next++
	inbox := make(chan *Message, 1)
	r.boxes[id] = inbox
	r.lock.Unlock()
	return id, inbox
}

func (r *router) Remove(id uint64) (chan *Message, bool) {
	r.lock.Lock()
	inbox, ok := r.boxes[id]
	if ok {
		delete(r.boxes, id)
	}
	r.lock.Unlock()
	return inbox, ok
}
