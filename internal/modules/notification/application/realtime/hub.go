package realtime

import (
	"sync"

	"github.com/Taishi-co/SteelLink-backend/internal/modules/notification/domain/entity"
)

type NotificationEnvelope struct {
	Type         string              `json:"type"`
	Notification entity.Notification `json:"notification"`
}

type Client struct {
	UserID uint64
	Send   chan NotificationEnvelope
}

type Hub struct {
	mu      sync.RWMutex
	clients map[uint64]map[*Client]struct{}
}

var DefaultHub = NewHub()

func NewHub() *Hub {
	return &Hub{clients: make(map[uint64]map[*Client]struct{})}
}

func (h *Hub) Register(client *Client) {
	if client == nil || client.UserID == 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]struct{})
	}
	h.clients[client.UserID][client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	if client == nil || client.UserID == 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	userClients := h.clients[client.UserID]
	if userClients == nil {
		return
	}
	if _, ok := userClients[client]; ok {
		delete(userClients, client)
		close(client.Send)
	}
	if len(userClients) == 0 {
		delete(h.clients, client.UserID)
	}
}

func (h *Hub) Publish(notification entity.Notification) {
	if notification.RecipientUserID == 0 {
		return
	}
	h.mu.RLock()
	userClients := h.clients[notification.RecipientUserID]
	clients := make([]*Client, 0, len(userClients))
	for client := range userClients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	envelope := NotificationEnvelope{Type: "notification.created", Notification: notification}
	for _, client := range clients {
		select {
		case client.Send <- envelope:
		default:
		}
	}
}
