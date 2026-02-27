package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type IpcServer struct {
	mu        sync.RWMutex
	clients   map[*websocket.Conn]bool
	upgrader  websocket.Upgrader
	onCommand func(cmd IpcCommand)
}

type IpcCommand struct {
	Type   string          `json:"type"`
	Params json.RawMessage `json:"params"`
}

func NewIpcServer(onCommand func(cmd IpcCommand)) *IpcServer {
	return &IpcServer{
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
		onCommand: onCommand,
	}
}

func (s *IpcServer) Start(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/", s.handleWebSocket)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = httpServer.Shutdown(context.Background())
	}()

	err := httpServer.ListenAndServe()
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *IpcServer) Broadcast(snapshot *IpcSnapshot) {
	if snapshot == nil {
		return
	}

	payload, err := json.Marshal(snapshot)
	if err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for conn := range s.clients {
		if writeErr := conn.WriteMessage(websocket.TextMessage, payload); writeErr != nil {
			_ = conn.Close()
			delete(s.clients, conn)
		}
	}
}

func (s *IpcServer) ClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

func (s *IpcServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	go s.readClientCommands(conn)
}

func (s *IpcServer) readClientCommands(conn *websocket.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		_ = conn.Close()
	}()

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var cmd IpcCommand
		if unmarshalErr := json.Unmarshal(payload, &cmd); unmarshalErr != nil {
			continue
		}

		if s.onCommand != nil {
			s.onCommand(cmd)
		}
	}
}
