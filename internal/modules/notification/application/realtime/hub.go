package realtime

import (
	"sync"

	notificationEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/notification/domain/entity"
)

type NotificationEnvelope struct {
	Type         string                          `json:"type"`
	Notification notificationEntity.Notification `json:"notification"`
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
	if userClients, ok := h.clients[client.UserID]; ok {
		delete(userClients, client)
		if len(userClients) == 0 {
			delete(h.clients, client.UserID)
		}
	}
	close(client.Send)
}

func (h *Hub) Publish(notification notificationEntity.Notification) {
	if notification.RecipientUserID == 0 {
		return
	}
	envelope := NotificationEnvelope{Type: "notification.created", Notification: notification}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients[notification.RecipientUserID] {
		select {
		case client.Send <- envelope:
		default:
		}
	}
}
