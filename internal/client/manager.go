package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"CoLinkPlan/internal/adapter"
	"CoLinkPlan/internal/config"
	"CoLinkPlan/internal/protocol"
	"CoLinkPlan/pkg/logger"

	"github.com/gorilla/websocket"
)

type Manager struct {
	Cfg           *config.ClientConfig
	Conn          *websocket.Conn
	ConnMutex     sync.Mutex
	ActiveWorkers int
	WorkerMutex   sync.Mutex
	Adapters      map[string]adapter.ProviderAdapter
	ModelMapping  map[string]ModelRoute // server_mapping -> ModelRoute
}

type ModelRoute struct {
	Provider adapter.ProviderAdapter
	Local    string
}

func NewManager(cfg *config.ClientConfig) *Manager {
	m := &Manager{
		Cfg:          cfg,
		Adapters:     make(map[string]adapter.ProviderAdapter),
		ModelMapping: make(map[string]ModelRoute),
	}

	for _, p := range cfg.Providers {
		var ad adapter.ProviderAdapter
		if p.Type == "openai" {
			ad = adapter.NewOpenAIAdapter(p.APIKey, p.BaseURL)
		} else if p.Type == "claude" {
			ad = adapter.NewClaudeAdapter(p.APIKey, p.BaseURL)
		} else {
			logger.Log.Warn("Unknown provider type", "type", p.Type)
			continue
		}
		m.Adapters[p.Type] = ad

		for _, model := range p.Models {
			m.ModelMapping[model.ServerMapping] = ModelRoute{
				Provider: ad,
				Local:    model.Local,
			}
		}
	}
	return m
}

func (m *Manager) Start(ctx context.Context) {
	backoff := 2 * time.Second

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Manager shutting down")
			return
		default:
		}

		err := m.connect(ctx)
		if err != nil {
			logger.Log.Error("Connection error", "err", err)
			logger.Log.Info("Reconnecting...", "backoff", backoff)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > 32*time.Second {
				backoff = 32 * time.Second
			}
			continue
		}

		// Reset backoff on successful connect and loop
		backoff = 2 * time.Second
		m.readLoop(ctx)
	}
}

func (m *Manager) connect(ctx context.Context) error {
	header := http.Header{}
	header.Set("Client-Token", m.Cfg.ClientToken)

	logger.Log.Info("Dialing server", "url", m.Cfg.ServerURL)
	c, _, err := websocket.DefaultDialer.DialContext(ctx, m.Cfg.ServerURL, header)
	if err != nil {
		return err
	}
	m.ConnMutex.Lock()
	m.Conn = c
	m.ConnMutex.Unlock()

	logger.Log.Info("Connected to server successfully")

	// Register
	var serverModels []string
	for m := range m.ModelMapping {
		serverModels = append(serverModels, m)
	}

	regData := protocol.RegisterData{
		MaxParallel: m.Cfg.MaxParallel,
		Models:      serverModels,
	}

	err = m.sendMessage(protocol.WSPayload{
		Type: protocol.MsgTypeRegister,
		Data: regData,
	})

	if err != nil {
		c.Close()
		return fmt.Errorf("failed to send register: %w", err)
	}
	return nil
}

func (m *Manager) readLoop(ctx context.Context) {
	defer func() {
		m.ConnMutex.Lock()
		if m.Conn != nil {
			m.Conn.Close()
			m.Conn = nil
		}
		m.ConnMutex.Unlock()
	}()

	m.Conn.SetPingHandler(func(appData string) error {
		m.ConnMutex.Lock()
		defer m.ConnMutex.Unlock()
		return m.Conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(5*time.Second))
	})

	for {
		_, message, err := m.Conn.ReadMessage()
		if err != nil {
			logger.Log.Error("Read error", "err", err)
			return
		}

		var payload protocol.WSPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			logger.Log.Error("Failed to unmarshal strict payload", "err", err)
			continue
		}

		if payload.Type == protocol.MsgTypeCall {
			dataBytes, _ := json.Marshal(payload.Data)
			var callData protocol.CallData
			json.Unmarshal(dataBytes, &callData)

			m.handleCall(ctx, callData)
		}
	}
}

func (m *Manager) handleCall(ctx context.Context, callData protocol.CallData) {
	m.WorkerMutex.Lock()
	if m.ActiveWorkers >= m.Cfg.MaxParallel {
		m.WorkerMutex.Unlock()
		// Reject
		m.sendMessage(protocol.WSPayload{
			Type: protocol.MsgTypeError,
			Data: protocol.ErrorData{
				RequestID: callData.RequestID,
				Code:      http.StatusServiceUnavailable,
				Message:   "BUSY: Local concurrency limit reached",
			},
		})
		return
	}
	m.ActiveWorkers++
	m.WorkerMutex.Unlock()

	go func() {
		defer func() {
			m.WorkerMutex.Lock()
			m.ActiveWorkers--
			m.WorkerMutex.Unlock()
		}()
		m.executeTask(ctx, callData)
	}()
}

func (m *Manager) executeTask(ctx context.Context, callData protocol.CallData) {
	logger.Log.Info("Executing task", "request_id", callData.RequestID, "model", callData.Model)

	route, ok := m.ModelMapping[callData.Model]
	if !ok {
		m.sendMessage(protocol.WSPayload{
			Type: protocol.MsgTypeError,
			Data: protocol.ErrorData{
				RequestID: callData.RequestID,
				Code:      http.StatusBadRequest,
				Message:   "Model not supported natively by this client",
			},
		})
		return
	}

	payloadBytes, _ := json.Marshal(callData.Payload)
	// context for adapter run
	streamCh := make(chan interface{})
	errCh := make(chan error, 1)

	// Context for adapter run
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go route.Provider.Call(subCtx, callData.RequestID, route.Local, payloadBytes, streamCh, errCh)

	for {
		select {
		case err, ok := <-errCh:
			if ok && err != nil {
				logger.Log.Error("Adapter error", "err", err)
				m.sendMessage(protocol.WSPayload{
					Type: protocol.MsgTypeError,
					Data: protocol.ErrorData{
						RequestID: callData.RequestID,
						Code:      http.StatusInternalServerError,
						Message:   err.Error(),
					},
				})
			}
		case chunk, ok := <-streamCh:
			if !ok {
				// Stream finished normally
				m.sendMessage(protocol.WSPayload{
					Type: protocol.MsgTypeFinish,
					Data: protocol.FinishData{
						RequestID: callData.RequestID,
					},
				})
				return
			}

			m.sendMessage(protocol.WSPayload{
				Type: protocol.MsgTypeStream,
				Data: protocol.StreamData{
					RequestID: callData.RequestID,
					Chunk:     chunk,
				},
			})
		}
	}
}

func (m *Manager) sendMessage(payload protocol.WSPayload) error {
	m.ConnMutex.Lock()
	defer m.ConnMutex.Unlock()
	if m.Conn == nil {
		return fmt.Errorf("no active connection")
	}
	return m.Conn.WriteJSON(payload)
}
