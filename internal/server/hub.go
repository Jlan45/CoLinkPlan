package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"CoLinkPlan/internal/protocol"
	"CoLinkPlan/pkg/logger"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients map[*ClientConn]bool

	register   chan *ClientConn
	unregister chan *ClientConn

	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*ClientConn]bool),
		register:   make(chan *ClientConn),
		unregister: make(chan *ClientConn),
	}
}

func (h *Hub) Run() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			logger.Log.Info("New client connected", "client_id", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				logger.Log.Info("Client disconnected", "client_id", client.ID)
			}
			h.mu.Unlock()

		case <-ticker.C:
			h.mu.RLock()
			for client := range h.clients {
				client.ConnMutex.Lock()
				err := client.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))
				client.ConnMutex.Unlock()

				if err != nil {
					client.Conn.Close()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) SelectClient(model string) (*ClientConn, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var bestClient *ClientConn
	lowestRatio := 1.0 // Ratio: ActiveTasks / MaxParallel

	for c := range h.clients {
		if !c.SupportedModels[model] {
			continue
		}

		// Check penalty
		if time.Now().Before(c.PenaltyUntil) {
			continue
		}

		if c.MaxParallel == 0 {
			continue // Not registered yet
		}

		if c.ActiveTasks >= c.MaxParallel {
			continue // Fully booked
		}

		ratio := float64(c.ActiveTasks) / float64(c.MaxParallel)
		if bestClient == nil || ratio < lowestRatio {
			bestClient = c
			lowestRatio = ratio
		}
	}

	if bestClient == nil {
		return nil, fmt.Errorf("no available clients for model: %s", model)
	}

	return bestClient, nil
}

// RouteCall finds a client, sends the payload and returns the stream channel.
// Performs Failover: silent retries up to 3 times on disonnects or BUSY.
func (h *Hub) RouteCall(ctx context.Context, requestID, model string, payload interface{}) (chan protocol.WSPayload, error) {
	var lastErr error
	var bestClient *ClientConn

	for i := 0; i < 3; i++ {
		c, err := h.SelectClient(model)
		if err != nil {
			return nil, fmt.Errorf("scheduling failed: %w (last err: %v)", err, lastErr)
		}

		bestClient = c
		streamCh := make(chan protocol.WSPayload, 10)

		bestClient.PendingMutex.Lock()
		bestClient.PendingStreams[requestID] = streamCh
		bestClient.PendingMutex.Unlock()

		bestClient.Hub.mu.Lock()
		bestClient.ActiveTasks++
		bestClient.Hub.mu.Unlock()

		sndErr := bestClient.SendMessage(protocol.WSPayload{
			Type: protocol.MsgTypeCall,
			Data: protocol.CallData{
				RequestID: requestID,
				Model:     model,
				Payload:   payload,
			},
		})

		if sndErr != nil {
			lastErr = sndErr
			bestClient.Hub.mu.Lock()
			bestClient.ActiveTasks--
			bestClient.PenaltyUntil = time.Now().Add(60 * time.Second) // Penalty 60s
			bestClient.Hub.mu.Unlock()

			bestClient.PendingMutex.Lock()
			delete(bestClient.PendingStreams, requestID)
			bestClient.PendingMutex.Unlock()
			continue // Retry
		}

		// Successfully dispatched to client
		return streamCh, nil
	}

	return nil, fmt.Errorf("failed to route call after 3 retries, last error: %v", lastErr)
}

// ListModels returns the set of model names currently advertised by at least one
// connected, non-penalized client node.
func (h *Hub) ListModels() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	seen := make(map[string]bool)
	for c := range h.clients {
		if c.MaxParallel == 0 {
			continue // not yet registered
		}
		if time.Now().Before(c.PenaltyUntil) {
			continue // penalized
		}
		for m := range c.SupportedModels {
			seen[m] = true
		}
	}

	models := make([]string, 0, len(seen))
	for m := range seen {
		models = append(models, m)
	}
	return models
}

func (h *Hub) CompleteTask(client *ClientConn, requestID string) {
	client.Hub.mu.Lock()
	client.ActiveTasks--
	if client.ActiveTasks < 0 {
		client.ActiveTasks = 0
	}
	client.Hub.mu.Unlock()

	client.PendingMutex.Lock()
	delete(client.PendingStreams, requestID)
	client.PendingMutex.Unlock()
}
