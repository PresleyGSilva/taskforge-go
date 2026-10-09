package ws

import (
	"encoding/json"
	"sync"
)

type Message struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

func (m Message) JSON() []byte {
	b, _ := json.Marshal(m)
	return b
}

type Client struct {
	Workspace string
	Send      chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub() *Hub { return &Hub{clients: map[*Client]struct{}{}} }

func (h *Hub) Add(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) Remove(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
}

// Broadcast sends msg to all clients in workspace, except optional skip.
func (h *Hub) Broadcast(workspace string, msg []byte, skip *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.Workspace != workspace || c == skip {
			continue
		}
		select {
		case c.Send <- msg:
		default:
		}
	}
}

func (h *Hub) Emit(workspace, typ string, payload any, skip *Client) {
	h.Broadcast(workspace, Message{Type: typ, Payload: payload}.JSON(), skip)
}
