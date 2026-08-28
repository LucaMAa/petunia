package ws

import (
	"math"
	"sync"

	"github.com/google/uuid"
)

type ClientLocation struct {
	Lat float64
	Lng float64
}

type Hub struct {
	mu        sync.RWMutex
	clients   map[uuid.UUID]*Client
	locations map[uuid.UUID]ClientLocation
}

var GlobalHub = &Hub{
	clients:   make(map[uuid.UUID]*Client),
	locations: make(map[uuid.UUID]ClientLocation),
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.userID] = c
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if existing, ok := h.clients[c.userID]; ok && existing == c {
		delete(h.clients, c.userID)
		delete(h.locations, c.userID)
		close(c.send)
	}
}

func (h *Hub) SendToUser(userID uuid.UUID, msg Message) bool {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	select {
	case client.send <- msg:
		return true
	default:
		h.Unregister(client)
		return false
	}
}

func (h *Hub) UpdateLocation(userID uuid.UUID, lat, lng float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.locations[userID] = ClientLocation{Lat: lat, Lng: lng}
}

func (h *Hub) BroadcastNearby(lat, lng, radiusMeters float64, msg Message) {
	h.mu.RLock()
	var targets []uuid.UUID
	for uid, loc := range h.locations {
		if haversine(lat, lng, loc.Lat, loc.Lng) <= radiusMeters {
			targets = append(targets, uid)
		}
	}
	h.mu.RUnlock()
	for _, uid := range targets {
		h.SendToUser(uid, msg)
	}
}

func (h *Hub) IsOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000.0
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
