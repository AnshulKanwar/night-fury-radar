package hiccup

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/anshulkanwar/night-fury-radar/internal/storage"
	"github.com/anshulkanwar/night-fury-radar/internal/types"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type Server struct {
	ctx     context.Context
	storage storage.Storage
	clients map[*Client]bool
	mu      sync.Mutex
	metrics chan types.Metric
	cancel  context.CancelFunc
}

type Client struct {
	conn       *websocket.Conn
	metricType string
}

func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		ctx:     ctx,
		storage: *storage.NewStorage(),
		clients: make(map[*Client]bool),
		metrics: make(chan types.Metric, 100),
		cancel:  cancel,
	}
}

func (s *Server) Start() {
	s.startMetricsReceiver()
	s.startMetricsBroadcaster()
}

func (s *Server) Stop() {
	s.cancel()
	close(s.metrics)
	for client := range s.clients {
		s.removeClient(client)
	}
	s.storage.Close()
}

func (s *Server) startMetricsReceiver() {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case notification := <-s.storage.Listener.Notify:
				var metric types.Metric
				err := json.Unmarshal([]byte(notification.Extra), &metric)
				if err != nil {
					log.Fatal(err)
				}
				s.metrics <- metric
			}
		}
	}()
}

func (s *Server) startMetricsBroadcaster() {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case metric := <-s.metrics:
				s.mu.Lock()
				for client := range s.clients {
					if client.metricType == metric.Type {
						err := client.conn.WriteJSON(metric)
						if err != nil {
							log.Println(err)
						}
					}
				}
				s.mu.Unlock()
			}
		}
	}()
}

func (s *Server) setupClient(client *Client) {
	s.addClient(client)
	s.sendPastData(client)
}

func (s *Server) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client] = true
}

func (s *Server) sendPastData(client *Client) {
	metrics := s.storage.ReadLast100Points(client.metricType)
	for _, metric := range metrics {
		client.conn.WriteJSON(metric)
	}
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[client]; ok {
		client.conn.Close()
		delete(s.clients, client)
	}
}

func (s *Server) HandleMetricRequest(metricType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}

		client := &Client{
			conn:       c,
			metricType: metricType,
		}
		s.setupClient(client)

		go func() {
			defer s.removeClient(client)

			for {
				_, _, err := c.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("websocket error: %v\n", err)
					}
					return
				}
			}
		}()
	}
}
